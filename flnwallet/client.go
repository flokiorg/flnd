package flnwallet

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/flokiorg/flnd"
	"github.com/flokiorg/flnd/aezeed"
	"github.com/flokiorg/flnd/lnrpc"
	"github.com/flokiorg/flnd/lnrpc/chainrpc"
	"github.com/flokiorg/flnd/lnrpc/walletrpc"
	"github.com/flokiorg/flnd/rpcperms"
	"github.com/flokiorg/go-flokicoin/chainutil"
	"github.com/flokiorg/go-flokicoin/chainutil/psbt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	ErrWalletNotFound      = errors.New("wallet not found")
	ErrWalletAlreadyExists = errors.New("wallet already exists")
	ErrWalletMustBeLocked  = errors.New("wallet must be locked to change password")
)

type txCache struct {
	Txs         []*lnrpc.Transaction
	NextOffset  uint32 // Offset to fetch the next page
	LastUpdated time.Time
}

type Client struct {
	unlockerClient lnrpc.WalletUnlockerClient
	lnClient       lnrpc.LightningClient
	walletKit      walletrpc.WalletKitClient
	stateClient    lnrpc.StateClient
	ntfClient      chainrpc.ChainNotifierClient
	chainKit       chainrpc.ChainKitClient

	health      chan *Update
	config      *flnd.Config
	ctx         context.Context
	adminMacHex string

	subTxsOnce sync.Once
	cache      *txCache
	closing    bool

	syncPollingActive bool
	isSynced          bool
	syncedHeight      uint32
	mu                sync.Mutex

	txFetchLimit uint32
}

const (
	defaultRPCTimeout       = 5 * time.Second
	transactionFetchTimeout = 30 * time.Second
	transactionPageSize     = 1000
)

func NewClient(ctx context.Context, conn *grpc.ClientConn, config *flnd.Config) *Client {
	c := &Client{
		unlockerClient: lnrpc.NewWalletUnlockerClient(conn),
		lnClient:       lnrpc.NewLightningClient(conn),
		walletKit:      walletrpc.NewWalletKitClient(conn),
		stateClient:    lnrpc.NewStateClient(conn),
		ntfClient:      chainrpc.NewChainNotifierClient(conn),
		chainKit:       chainrpc.NewChainKitClient(conn),
		// Buffer health updates to avoid dropping important state transitions
		health: make(chan *Update, 16),
		ctx:    ctx,
		config: config,
		cache: &txCache{
			Txs:         []*lnrpc.Transaction{},
			NextOffset:  0,
			LastUpdated: time.Time{},
		},
	}

	go c.subscribeState()

	return c
}

func (c *Client) subscribeTransactions() {

	stream, err := c.lnClient.SubscribeTransactions(c.withMacaroon(), &lnrpc.GetTransactionsRequest{})
	if err != nil {
		c.kill(err)
		return
	}

	for {
		r, err := stream.Recv()
		if err != nil {
			c.kill(err)
			return
		}

		c.submitHealth(Update{State: StatusTransaction, Transaction: r})
	}
}

func (c *Client) subscribeBlocks() {

	stream, err := c.ntfClient.RegisterBlockEpochNtfn(c.withMacaroon(), &chainrpc.BlockEpoch{})
	if err != nil {
		c.kill(err)
		return
	}

	for {
		r, err := stream.Recv()
		if err != nil {
			c.kill(err)
			return
		}

		state := StatusBlock
		var syncedHeight uint32

		c.mu.Lock()
		if !c.isSynced {
			state = StatusScanning
			syncedHeight = c.syncedHeight
		}
		c.mu.Unlock()

		c.submitHealth(Update{State: state, SyncedHeight: syncedHeight, BlockHeight: r.Height, BlockHash: hex.EncodeToString(r.Hash)})
	}
}

func (c *Client) subscribeState() {
	stream, err := c.stateClient.SubscribeState(c.ctx, &lnrpc.SubscribeStateRequest{})
	if err != nil {
		c.kill(err)
		return
	}

	for {
		r, err := stream.Recv()
		if err != nil {
			c.kill(err)
			return
		}

		switch r.State {
		case lnrpc.WalletState_NON_EXISTING:
			c.submitHealth(Update{State: StatusNoWallet})

		case lnrpc.WalletState_LOCKED:
			c.submitHealth(Update{State: StatusLocked})

		case lnrpc.WalletState_UNLOCKED:
			adminMacHex, err := readMacaroon(c.config.AdminMacPath)
			if err != nil {
				c.kill(err)
				return
			}
			c.adminMacHex = adminMacHex
			c.submitHealth(Update{State: StatusUnlocked})

		case lnrpc.WalletState_WAITING_TO_START:
			c.submitHealth(Update{State: StatusNone})

		case lnrpc.WalletState_RPC_ACTIVE:
			synced, blockHeight, err := c.IsSynced()
			if err != nil {
				continue
			} else if synced {
				c.submitHealth(Update{State: StatusNone, BlockHeight: blockHeight})
			} else {
				c.submitHealth(Update{State: StatusSyncing, BlockHeight: blockHeight})
				go c.pollSyncStatus()
			}

		case lnrpc.WalletState_SERVER_ACTIVE:
			synced, blockHeight, err := c.IsSynced()
			if err != nil {
				c.kill(err)
				return
			}

			if !synced {
				c.submitHealth(Update{State: StatusSyncing, BlockHeight: blockHeight})
				go c.pollSyncStatus()
			}

			c.subTxsOnce.Do(func() {
				go c.subscribeTransactions()
				go c.subscribeBlocks()
			})

		}

	}
}

func (c *Client) LoadMacaroon(path string) error {
	adminMacHex, err := readMacaroon(path)
	if err != nil {
		return fmt.Errorf("unable to read macaroon file. %v", err)
	}
	c.adminMacHex = adminMacHex
	return nil
}

func (c *Client) pollSyncStatus() {
	c.mu.Lock()
	if c.syncPollingActive {
		c.mu.Unlock()
		return
	}
	c.syncPollingActive = true
	c.mu.Unlock()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	defer func() {
		c.mu.Lock()
		c.syncPollingActive = false
		c.mu.Unlock()
	}()

	for {
		select {
		case <-c.ctx.Done():
			return

		case <-ticker.C:
			synced, blockHeight, err := c.IsSynced()
			if err != nil {
				continue
			}
			if synced {
				c.submitHealth(Update{State: StatusReady, BlockHeight: blockHeight})
				return
			}

			c.submitHealth(Update{
				State:       StatusSyncing,
				BlockHeight: blockHeight,
			})
		}
	}
}

func (c *Client) close() {
	c.closing = true
}

func (c *Client) kill(err error) {
	if matchRPCErrorMessage(err, context.Canceled) || c.closing {
		c.submitHealth(Update{State: StatusDown})
	} else {
		c.submitHealth(Update{State: StatusDown, Err: err})
	}
}

func (c *Client) submitHealth(change Update) {
	select {
	case c.health <- &change:
	default:
	}
}

func (c *Client) Health() <-chan *Update {
	return c.health
}

func (c *Client) WalletExists() (bool, error) {
	if c.closing {
		return false, ErrDaemonNotRunning
	}

	ctx, cancel := c.rpcContext(0)
	defer cancel()
	_, err := c.lnClient.GetInfo(ctx, &lnrpc.GetInfoRequest{})
	if err == nil {
		return true, nil // wallet exists and is accessible
	}

	// Wallet does not exist
	if matchRPCErrorMessage(err, rpcperms.ErrNoWallet) {
		return false, nil
	}

	// For other errors, assume wallet exists but something else failed
	// OR log and decide how strict you want this to be
	return true, nil
}

func (c *Client) IsSynced() (bool, uint32, error) {
	if c.closing {
		return false, 0, ErrDaemonNotRunning
	}

	ctx, cancel := c.rpcContext(0)
	defer cancel()
	resp, err := c.lnClient.GetInfo(ctx, &lnrpc.GetInfoRequest{})

	// If the RPC server is still starting up, treat it as not synced yet,
	// but don't surface an error so callers can keep polling smoothly.
	if err != nil && matchRPCErrorMessage(err, rpcperms.ErrRPCStarting) {
		err = nil
		resp = nil
	}

	var blockHeight uint32
	var synced bool
	if resp != nil {
		blockHeight = resp.BlockHeight
		synced = err == nil && resp.SyncedToChain
	}

	c.mu.Lock()
	c.isSynced = synced
	c.syncedHeight = blockHeight
	c.mu.Unlock()

	return synced, blockHeight, err
}

func (c *Client) Unlock(passphrase string) error {
	if c.closing {
		return ErrDaemonNotRunning
	}

	_, err := c.unlockerClient.UnlockWallet(c.ctx, &lnrpc.UnlockWalletRequest{
		WalletPassword: []byte(passphrase),
		RecoveryWindow: 255,
	})

	if err != nil && matchRPCErrorMessage(err, rpcperms.ErrWalletUnlocked) {
		return nil
	}
	return err
}

func (c *Client) IsLocked() (bool, error) {

	if c.closing {
		return false, ErrDaemonNotRunning
	}

	_, err := c.lnClient.GetInfo(c.withMacaroon(), &lnrpc.GetInfoRequest{})
	if err == nil {
		// Wallet is unlocked
		return false, nil
	}

	_, err = c.unlockerClient.GenSeed(c.ctx, &lnrpc.GenSeedRequest{})
	if err == nil {
		// Wallet is locked (GenSeed is only available when locked)
		return true, nil
	}

	if matchRPCErrorMessage(err, rpcperms.ErrWalletUnlocked, fmt.Errorf("wallet already exists")) {
		return true, nil
	}

	return false, err
}

func (c *Client) Create(passphrase string) (string, []string, error) {

	if c.closing {
		return "", nil, ErrDaemonNotRunning
	}

	seedResp, err := c.unlockerClient.GenSeed(context.Background(), &lnrpc.GenSeedRequest{})
	if err != nil {
		return "", nil, err
	}

	_, err = c.unlockerClient.InitWallet(context.Background(), &lnrpc.InitWalletRequest{
		WalletPassword:     []byte(passphrase),
		CipherSeedMnemonic: seedResp.CipherSeedMnemonic,
		RecoveryWindow:     0,
	})
	if err != nil {
		return "", nil, err
	}

	return hex.EncodeToString(seedResp.EncipheredSeed), seedResp.CipherSeedMnemonic, nil
}

func (c *Client) RestoreByEncipheredSeed(strEncipheredSeed, passphrase string) ([]string, error) {
	if c.closing {
		return nil, ErrDaemonNotRunning
	}

	encipheredSeed, err := hex.DecodeString(strEncipheredSeed)
	if err != nil {
		return nil, err
	}

	if len(encipheredSeed) == 32 { // legacy version
		return nil, fmt.Errorf("wallets from tWallet 0.1.x must be restored using the same version")
	}

	if len(encipheredSeed) != aezeed.EncipheredCipherSeedSize {
		return nil, fmt.Errorf("invalid seed length: possibly corrupted or unsupported format")
	}

	mnemonic, err := aezeed.CipherTextToMnemonic([aezeed.EncipheredCipherSeedSize]byte(encipheredSeed))
	if err != nil {
		return nil, err
	}

	_, err = c.unlockerClient.InitWallet(c.ctx, &lnrpc.InitWalletRequest{
		WalletPassword:     []byte(passphrase),
		CipherSeedMnemonic: mnemonic[:],
		RecoveryWindow:     255,
	})
	if err != nil {
		return nil, err
	}

	return mnemonic[:], nil
}

func (c *Client) RestoreByMnemonic(mnemonic []string, passphrase string) (string, error) {
	if c.closing {
		return "", ErrDaemonNotRunning
	}
	var seedMnemonic aezeed.Mnemonic
	copy(seedMnemonic[:], mnemonic)
	cipherSeed, err := seedMnemonic.ToCipherSeed([]byte{})
	if err != nil {
		return "", fmt.Errorf("%v. Wallets from tWallet 0.1.x must be restored using the same version", err) // include legacy notice
	}

	encipheredSeed, err := cipherSeed.Encipher([]byte{})
	if err != nil {
		return "", err
	}

	_, err = c.unlockerClient.InitWallet(c.ctx, &lnrpc.InitWalletRequest{
		WalletPassword:     []byte(passphrase),
		CipherSeedMnemonic: mnemonic,
		RecoveryWindow:     255,
	})
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(encipheredSeed[:]), nil
}

func (c *Client) Balance() (*lnrpc.WalletBalanceResponse, error) {
	if c.closing {
		return nil, ErrDaemonNotRunning
	}
	ctx, cancel := c.rpcContext(0)
	defer cancel()
	resp, err := c.lnClient.WalletBalance(ctx, &lnrpc.WalletBalanceRequest{
		MinConfs: 0,
	})
	if err != nil {
		// if matchRPCErrorMessage(err, rpcperms.ErrRPCStarting) {
		// 	// Treat as not-ready, return zero balance without error.
		// 	return &lnrpc.WalletBalanceResponse{}, nil
		// }
		return nil, err
	}

	return resp, nil
}

func (c *Client) ChangePassphrase(old, new string) error {
	if c.closing {
		return ErrDaemonNotRunning
	}
	locked, err := c.IsLocked()
	if err != nil {
		return err
	}
	if !locked {
		return ErrWalletMustBeLocked
	}

	_, err = c.unlockerClient.ChangePassword(c.withMacaroon(), &lnrpc.ChangePasswordRequest{
		CurrentPassword: []byte(old),
		NewPassword:     []byte(new),
	})

	if err != nil {
		return err
	}

	return nil
}

func (c *Client) SimpleTransfer(address chainutil.Address, amount chainutil.Amount, lokiPerVbyte uint64) (string, error) {
	if c.closing {
		return "", ErrDaemonNotRunning
	}
	resp, err := c.lnClient.SendCoins(c.withMacaroon(), &lnrpc.SendCoinsRequest{
		Addr:             address.String(),
		Amount:           int64(amount),
		SatPerVbyte:      lokiPerVbyte,
		SpendUnconfirmed: true,
	})
	if err != nil {
		return "", err
	}
	return resp.Txid, nil
}

func (c *Client) SimpleTransferFee(address chainutil.Address, amount chainutil.Amount) (*lnrpc.EstimateFeeResponse, error) {
	if c.closing {
		return nil, ErrDaemonNotRunning
	}

	entry := map[string]int64{}
	entry[address.String()] = int64(amount.ToUnit(chainutil.AmountLoki))

	resp, err := c.lnClient.EstimateFee(c.withMacaroon(), &lnrpc.EstimateFeeRequest{
		AddrToAmount:          entry,
		TargetConf:            1,
		CoinSelectionStrategy: lnrpc.CoinSelectionStrategy_STRATEGY_RANDOM,
		SpendUnconfirmed:      true,
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) FundPsbt(addrToAmount map[string]int64, lokiPerVbyte uint64, lockExpirationSeconds uint64) (*FundedPsbt, error) {
	if c.closing {
		return nil, ErrDaemonNotRunning
	}

	outputs := make(map[string]uint64, len(addrToAmount))
	for a, v := range addrToAmount {
		outputs[a] = uint64(v)
	}

	template := &walletrpc.TxTemplate{
		Outputs: outputs,
	}

	req := &walletrpc.FundPsbtRequest{
		Template: &walletrpc.FundPsbtRequest_Raw{
			Raw: template,
		},
		Fees: &walletrpc.FundPsbtRequest_SatPerVbyte{
			SatPerVbyte: lokiPerVbyte,
		},
		LockExpirationSeconds: lockExpirationSeconds,
	}

	resp, err := c.walletKit.FundPsbt(c.withMacaroon(), req)
	if err != nil {
		return nil, err
	}

	packet, err := psbt.NewFromRawBytes(bytes.NewReader(resp.FundedPsbt), false)
	if err != nil {
		return nil, err
	}

	locks := make([]*OutputLock, 0, len(resp.LockedUtxos))
	for _, utxo := range resp.LockedUtxos {
		if utxo == nil || utxo.Outpoint == nil {
			continue
		}
		locks = append(locks, &OutputLock{
			ID:       utxo.Id,
			Outpoint: utxo.Outpoint,
		})
	}

	return &FundedPsbt{
		Packet: packet,
		Locks:  locks,
	}, nil
}

func (c *Client) FinalizePsbt(packet *psbt.Packet) (*chainutil.Tx, error) {
	if c.closing {
		return nil, ErrDaemonNotRunning
	}

	var buf bytes.Buffer
	if err := packet.Serialize(&buf); err != nil {
		return nil, err
	}

	resp, err := c.walletKit.FinalizePsbt(c.withMacaroon(), &walletrpc.FinalizePsbtRequest{
		FundedPsbt: buf.Bytes(),
	})
	if err != nil {
		return nil, err
	}

	tx, err := chainutil.NewTxFromBytes(resp.RawFinalTx)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (c *Client) PublishTransaction(tx *chainutil.Tx) error {
	if c.closing {
		return ErrDaemonNotRunning
	}

	bytes, err := tx.MsgTx().Bytes()
	if err != nil {
		return err
	}

	resp, err := c.walletKit.PublishTransaction(c.withMacaroon(), &walletrpc.Transaction{
		TxHex: bytes,
	})
	if err != nil {
		return err
	}

	if resp.PublishError != "" {
		return fmt.Errorf(resp.PublishError)
	}

	return nil
}

func (c *Client) ReleaseOutputs(locks []*OutputLock) error {
	if len(locks) == 0 {
		return nil
	}
	if c.closing {
		return ErrDaemonNotRunning
	}

	for _, lock := range locks {
		if lock == nil || len(lock.ID) == 0 || lock.Outpoint == nil {
			continue
		}

		_, err := c.walletKit.ReleaseOutput(c.withMacaroon(), &walletrpc.ReleaseOutputRequest{
			Id:       lock.ID,
			Outpoint: lock.Outpoint,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) SimpleManyTransfer(addrToAmount map[string]int64, lokiPerVbyte uint64) (string, error) {
	if c.closing {
		return "", ErrDaemonNotRunning
	}
	resp, err := c.lnClient.SendMany(c.withMacaroon(), &lnrpc.SendManyRequest{
		AddrToAmount: addrToAmount,
		SatPerVbyte:  lokiPerVbyte,
	})
	if err != nil {
		return "", err
	}
	return resp.Txid, nil
}

func (c *Client) SimpleManyTransferFee(addrToAmount map[string]int64) (*lnrpc.EstimateFeeResponse, error) {
	if c.closing {
		return nil, ErrDaemonNotRunning
	}

	resp, err := c.lnClient.EstimateFee(c.withMacaroon(), &lnrpc.EstimateFeeRequest{
		AddrToAmount:          addrToAmount,
		TargetConf:            1,
		CoinSelectionStrategy: lnrpc.CoinSelectionStrategy_STRATEGY_RANDOM,
		SpendUnconfirmed:      true,
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) GetNextAddress(addrType lnrpc.AddressType) (chainutil.Address, error) {
	if c.closing {
		return nil, ErrDaemonNotRunning
	}
	resp, err := c.lnClient.NewAddress(c.withMacaroon(), &lnrpc.NewAddressRequest{Type: addrType})
	if err != nil {
		return nil, err
	}
	return chainutil.DecodeAddress(resp.Address, c.config.ActiveNetParams.Params)
}

func (c *Client) ListAddresses() ([]*walletrpc.AccountWithAddresses, error) {
	if c.closing {
		return nil, ErrDaemonNotRunning
	}
	resp, err := c.walletKit.ListAddresses(c.withMacaroon(), &walletrpc.ListAddressesRequest{})
	if err != nil {
		return nil, err
	}
	return resp.AccountWithAddresses, nil
}

func (c *Client) FetchTransactions() ([]*lnrpc.Transaction, error) {
	if c.closing {
		return nil, ErrDaemonNotRunning
	}

	offset := uint32(0)
	allTxs := []*lnrpc.Transaction{}

	maxResults := c.txFetchLimit
	for {
		ctx, cancel := c.rpcContext(transactionFetchTimeout)
		resp, err := c.lnClient.GetTransactions(ctx, &lnrpc.GetTransactionsRequest{
			MaxTransactions: transactionPageSize,
			IndexOffset:     offset,
		})
		cancel()
		if err != nil {
			if matchRPCErrorMessage(err, rpcperms.ErrRPCStarting) {
				// Not ready yet; return whatever we have (cache or empty)
				if len(c.cache.Txs) > 0 {
					return c.cache.Txs, nil
				}
				return []*lnrpc.Transaction{}, nil
			}
			if matchRPCErrorMessage(err, context.DeadlineExceeded) {
				return nil, fmt.Errorf("rpc connection timeout")
			}
			return nil, err
		}

		allTxs = append(allTxs, resp.Transactions...)
		offset += uint32(len(resp.Transactions))
		if maxResults > 0 && len(allTxs) >= int(maxResults) {
			if len(allTxs) > int(maxResults) {
				allTxs = allTxs[:maxResults]
			}
			break
		}

		if uint32(len(resp.Transactions)) < transactionPageSize {
			break
		}
	}

	c.cache.Txs = allTxs
	c.cache.NextOffset = offset
	c.cache.LastUpdated = time.Now()

	sort.Slice(c.cache.Txs, func(i, j int) bool {
		return c.cache.Txs[i].NumConfirmations < c.cache.Txs[j].NumConfirmations
	})

	return c.cache.Txs, nil
}

func (c *Client) withMacaroon() context.Context {
	md := metadata.Pairs("macaroon", c.adminMacHex)
	return metadata.NewOutgoingContext(c.ctx, md)
}

func (c *Client) rpcContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		timeout = defaultRPCTimeout
	}
	if c.config.ConnectionTimeout > 0 && timeout > c.config.ConnectionTimeout {
		timeout = c.config.ConnectionTimeout
	}
	ctx, cancel := context.WithTimeout(c.ctx, timeout)
	md := metadata.Pairs("macaroon", c.adminMacHex)
	return metadata.NewOutgoingContext(ctx, md), cancel
}

func (c *Client) SetMaxTransactionsLimit(limit uint32) {
	if limit == 0 {
		c.txFetchLimit = 0
	} else {
		c.txFetchLimit = limit
	}
}

func readMacaroon(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(data), nil
}

func matchRPCErrorMessage(err error, targets ...error) bool {
	st, ok := status.FromError(err)
	if !ok {
		return false
	}
	for _, t := range targets {
		if st.Message() == t.Error() {
			return true
		}
	}
	return false
}
