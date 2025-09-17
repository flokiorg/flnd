package flnwallet

import (
	"context"
	"sync"
	"time"

	"github.com/flokiorg/flnd"
	"github.com/flokiorg/flnd/lncfg"
	"github.com/flokiorg/flnd/lnrpc"
	"github.com/flokiorg/go-flokicoin/chaincfg"
	"github.com/flokiorg/go-flokicoin/chainutil"
)

type Status string

const (
	StatusInit        Status = "init"
	StatusNone        Status = "none"
	StatusLocked      Status = "locked"
	StatusUnlocked    Status = "unlocked"
	StatusSyncing     Status = "syncing"
	StatusReady       Status = "ready"
	StatusNoWallet    Status = "noWallet"
	StatusDown        Status = "down"
	StatusTransaction Status = "tx"
	StatusBlock       Status = "block"
	StatusScanning    Status = "scanning"
	StatusQuit        Status = "quit"
)

type Update struct {
	State                     Status
	Err                       error
	Transaction               *lnrpc.Transaction
	BlockHeight, SyncedHeight uint32
	BlockHash                 string
}

type ServiceConfig struct {
	Walletdir         string        `short:"w" long:"walletdir"  description:"Directory for Flokicoin Lightning Network"`
	RegressionTest    bool          `long:"regtest" description:"Use the regression test network"`
	Testnet           bool          `long:"testnet" description:"Use the test network"`
	ConnectionTimeout time.Duration `short:"t" long:"connectiontimeout" default:"50s" description:"The timeout value for network connections. Valid time units are {ms, s, m, h}."`
	DebugLevel        string        `short:"d" long:"debuglevel" default:"info" description:"Logging level for all subsystems {trace, debug, info, warn, error, critical}"`
	ConnectPeers      []string      `long:"connect" description:"Connect only to the specified peers at startup"`
	Feeurl            string        `long:"feeurl" description:"Custom fee estimation API endpoint (Required on mainnet)"`

	TLSExtraIPs     []string `long:"tlsextraip" description:"Adds an extra ip to the generated certificate"`
	TLSExtraDomains []string `long:"tlsextradomain" description:"Adds an extra domain to the generated certificate"`
	TLSAutoRefresh  bool     `long:"tlsautorefresh" description:"Re-generate TLS certificate and key if the IPs or domains are changed"`

	RawRPCListeners  []string `long:"rpclisten" description:"Add an interface/port/socket to listen for RPC connections"`
	RawRESTListeners []string `long:"restlisten" description:"Add an interface/port/socket to listen for REST connections"`
	RawListeners     []string `long:"listen" description:"Add an interface/port to listen for peer connections"`

	RestCORS []string `long:"restcors" description:"Add an ip:port/hostname to allow cross origin access from. To allow all origins, set as \"*\"."`

	Network *chaincfg.Params
}

type Service struct {
	subMu sync.Mutex
	subs  []chan *Update

	ctx    context.Context
	cancel context.CancelFunc

	flndConfig *flnd.Config
	client     *Client
	daemon     *daemon
	cmux       sync.Mutex
	wg         sync.WaitGroup
	running    bool
	lastEvent  *Update
}

func New(pctx context.Context, cfg *ServiceConfig) *Service {

	ctx, cancel := context.WithCancel(pctx)

	conf := flnd.DefaultConfig()
	conf.LndDir = cfg.Walletdir
	conf.Bitcoin.Node = flnd.NeutrinoBackendName
	conf.NeutrinoMode.ConnectPeers = cfg.ConnectPeers
	conf.DebugLevel = cfg.DebugLevel
	conf.Fee.URL = cfg.Feeurl
	conf.ProtocolOptions = &lncfg.ProtocolOptions{}
	conf.Pprof = &lncfg.Pprof{}
	conf.LogConfig.Console.Disable = true
	conf.ConnectionTimeout = cfg.ConnectionTimeout
	conf.TLSExtraDomains = append(conf.TLSExtraDomains, cfg.TLSExtraDomains...)
	conf.TLSExtraIPs = append(conf.TLSExtraIPs, cfg.TLSExtraIPs...)
	conf.RawRPCListeners = append(conf.RawRPCListeners, cfg.RawRPCListeners...)
	conf.RawRESTListeners = append(conf.RawRESTListeners, cfg.RawRESTListeners...)
	conf.RawListeners = append(conf.RawListeners, cfg.RawListeners...)
	conf.RestCORS = append(conf.RestCORS, cfg.RestCORS...)
	conf.TLSAutoRefresh = cfg.TLSAutoRefresh

	switch cfg.Network {
	case &chaincfg.MainNetParams:
		conf.Bitcoin.MainNet = true
	case &chaincfg.TestNet3Params:
		conf.Bitcoin.TestNet3 = true
	case &chaincfg.TestNet4Params:
		conf.Bitcoin.TestNet4 = true
	case &chaincfg.SimNetParams:
		conf.Bitcoin.SigNet = true
	case &chaincfg.RegressionNetParams:
		conf.Bitcoin.RegTest = true
	case &chaincfg.SigNetParams:
		conf.Bitcoin.SigNet = true
	}

	s := &Service{
		lastEvent:  &Update{State: StatusInit},
		flndConfig: &conf,
		ctx:        ctx,
		cancel:     cancel,
	}

	go s.run()

	return s
}

func (s *Service) run() {
	s.wg.Add(1)
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			return

		default:

			s.notifySubscribers(&Update{State: StatusNone})
			d := newDaemon(s.ctx, s.flndConfig)
			c, err := d.start()
			if err != nil {
				s.notifySubscribers(&Update{State: StatusDown, Err: err})
				continue
			}
			s.running = true
			ctx, cancel := context.WithCancel(s.ctx)
			go func() {
				for {

					select {
					case <-ctx.Done():
						d.stop()
						return

					case health := <-c.Health():
						s.notifySubscribers(health)
						switch health.State {
						case StatusDown:
							d.stop()
						default:
						}
					}
				}
			}()
			s.registerConnection(d, c)
			d.waitForShutdown()
			cancel()
			s.running = false
		}
	}
}

func (s *Service) Stop() {
	if !s.running {
		return
	}
	s.cancel()
	s.unsubscribeAll()
	s.wg.Wait()
	s.running = false
}

func (s *Service) Restart(pctx context.Context) {
	if s.daemon != nil {
		s.daemon.stop()
	}
}

func (s *Service) registerConnection(d *daemon, c *Client) {
	s.cmux.Lock()
	defer s.cmux.Unlock()
	s.client = c
	s.daemon = d
}

func (s *Service) Subscribe() <-chan *Update {
	ch := make(chan *Update, 5)
	s.subMu.Lock()
	s.subs = append(s.subs, ch)
	s.subMu.Unlock()
	ch <- s.lastEvent
	return ch
}

func (s *Service) Unsubscribe(ch <-chan *Update) {
	s.subMu.Lock()
	defer s.subMu.Unlock()

	for i := 0; i < len(s.subs); i++ {
		if s.subs[i] == ch {
			s.subs = append(s.subs[:i], s.subs[i+1:]...)
			break
		}
	}
}

func (s *Service) notifySubscribers(u *Update) {
	s.subMu.Lock()
	defer s.subMu.Unlock()
	s.lastEvent = u

	for _, ch := range s.subs {
		select {
		case ch <- u:
		default:
		}
	}
}

func (s *Service) unsubscribeAll() {
	s.subMu.Lock()
	defer s.subMu.Unlock()

	if len(s.subs) == 0 {
		return
	}

	finalUpdate := &Update{
		State: StatusDown,
	}

	for _, ch := range s.subs {
		select {
		case ch <- finalUpdate:
		case <-time.After(5 * time.Second):
		}
		close(ch)
	}

	s.subs = s.subs[:0]
}

func (s *Service) CreateWallet(passphrase string) (string, []string, error) {
	s.cmux.Lock()
	defer s.cmux.Unlock()
	return s.client.Create(passphrase)
}

func (s *Service) Balance() (*lnrpc.WalletBalanceResponse, error) {
	s.cmux.Lock()
	defer s.cmux.Unlock()
	return s.client.Balance()
}

func (s *Service) IsLocked() (bool, error) {
	s.cmux.Lock()
	defer s.cmux.Unlock()
	return s.client.IsLocked()
}

func (s *Service) Unlock(passphrase string) error {
	s.cmux.Lock()
	defer s.cmux.Unlock()
	return s.client.Unlock(passphrase)
}

func (s *Service) WalletExists() (bool, error) {
	s.cmux.Lock()
	defer s.cmux.Unlock()
	return s.client.WalletExists()
}

func (s *Service) FetchTransactions() ([]*lnrpc.Transaction, error) {
	s.cmux.Lock()
	defer s.cmux.Unlock()
	return s.client.FetchTransactions()
}

func (s *Service) GetNextAddress(t lnrpc.AddressType) (chainutil.Address, error) {
	s.cmux.Lock()
	defer s.cmux.Unlock()
	return s.client.GetNextAddress(t)
}

func (s *Service) RestoreByMnemonic(mnemonic []string, passphrase string) (string, error) {
	s.cmux.Lock()
	defer s.cmux.Unlock()
	return s.client.RestoreByMnemonic(mnemonic, passphrase)
}

func (s *Service) RestoreByEncipheredSeed(strEncipheredSeed, passphrase string) ([]string, error) {
	s.cmux.Lock()
	defer s.cmux.Unlock()
	return s.client.RestoreByEncipheredSeed(strEncipheredSeed, passphrase)
}

func (s *Service) ChangePassphrase(old, new string) error {
	s.cmux.Lock()
	defer s.cmux.Unlock()
	return s.client.ChangePassphrase(old, new)
}

func (s *Service) Transfer(address chainutil.Address, amount chainutil.Amount, lokiPerVbyte uint64) (string, error) {
	s.cmux.Lock()
	defer s.cmux.Unlock()
	return s.client.SimpleTransfer(address, amount, lokiPerVbyte)
}

func (s *Service) Fee(address chainutil.Address, amount chainutil.Amount) (*lnrpc.EstimateFeeResponse, error) {
	s.cmux.Lock()
	defer s.cmux.Unlock()
	return s.client.SimpleTransferFee(address, amount)
}

func (s *Service) GetLastEvent() *Update {
	return s.lastEvent
}
