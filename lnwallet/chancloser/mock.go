package chancloser

import (
	"sync/atomic"

	"github.com/flokiorg/go-flokicoin/crypto"

	"github.com/flokiorg/flnd/chainntnfs"
	"github.com/flokiorg/flnd/channeldb"
	"github.com/flokiorg/flnd/fn"
	"github.com/flokiorg/flnd/input"
	"github.com/flokiorg/flnd/lnwallet"
	"github.com/flokiorg/flnd/lnwallet/chainfee"
	"github.com/flokiorg/flnd/lnwire"
	"github.com/flokiorg/go-flokicoin/chaincfg/chainhash"
	"github.com/flokiorg/go-flokicoin/chainutil"
	"github.com/flokiorg/go-flokicoin/wire"
	"github.com/stretchr/testify/mock"
)

type dummyAdapters struct {
	mock.Mock

	msgSent atomic.Bool

	confChan  chan *chainntnfs.TxConfirmation
	spendChan chan *chainntnfs.SpendDetail
}

func newDaemonAdapters() *dummyAdapters {
	return &dummyAdapters{
		confChan:  make(chan *chainntnfs.TxConfirmation, 1),
		spendChan: make(chan *chainntnfs.SpendDetail, 1),
	}
}

func (d *dummyAdapters) SendMessages(pub crypto.PublicKey,
	msgs []lnwire.Message) error {

	defer d.msgSent.Store(true)

	args := d.Called(pub, msgs)

	return args.Error(0)
}

func (d *dummyAdapters) BroadcastTransaction(tx *wire.MsgTx,
	label string) error {

	args := d.Called(tx, label)

	return args.Error(0)
}

func (d *dummyAdapters) DisableChannel(op wire.OutPoint) error {
	args := d.Called(op)

	return args.Error(0)
}

func (d *dummyAdapters) RegisterConfirmationsNtfn(txid *chainhash.Hash,
	pkScript []byte, numConfs, heightHint uint32,
	opts ...chainntnfs.NotifierOption,
) (*chainntnfs.ConfirmationEvent, error) {

	args := d.Called(txid, pkScript, numConfs)

	err := args.Error(0)

	return &chainntnfs.ConfirmationEvent{
		Confirmed: d.confChan,
	}, err
}

func (d *dummyAdapters) RegisterSpendNtfn(outpoint *wire.OutPoint,
	pkScript []byte, heightHint uint32) (*chainntnfs.SpendEvent, error) {

	args := d.Called(outpoint, pkScript, heightHint)

	err := args.Error(0)

	return &chainntnfs.SpendEvent{
		Spend: d.spendChan,
	}, err
}

type mockFeeEstimator struct {
	mock.Mock
}

func (m *mockFeeEstimator) EstimateFee(chanType channeldb.ChannelType,
	localTxOut, remoteTxOut *wire.TxOut,
	idealFeeRate chainfee.SatPerKWeight) chainutil.Amount {

	args := m.Called(chanType, localTxOut, remoteTxOut, idealFeeRate)
	return args.Get(0).(chainutil.Amount)
}

type mockChanObserver struct {
	mock.Mock
}

func (m *mockChanObserver) NoDanglingUpdates() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *mockChanObserver) DisableIncomingAdds() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockChanObserver) DisableOutgoingAdds() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockChanObserver) MarkCoopBroadcasted(txn *wire.MsgTx,
	local bool) error {

	args := m.Called(txn, local)
	return args.Error(0)
}

func (m *mockChanObserver) MarkShutdownSent(deliveryAddr []byte,
	isInitiator bool) error {

	args := m.Called(deliveryAddr, isInitiator)
	return args.Error(0)
}

func (m *mockChanObserver) FinalBalances() fn.Option[ShutdownBalances] {
	args := m.Called()
	return args.Get(0).(fn.Option[ShutdownBalances])
}

func (m *mockChanObserver) DisableChannel() error {
	args := m.Called()
	return args.Error(0)
}

type mockErrorReporter struct {
	mock.Mock
}

func (m *mockErrorReporter) ReportError(err error) {
	m.Called(err)
}

type mockCloseSigner struct {
	mock.Mock
}

func (m *mockCloseSigner) CreateCloseProposal(fee chainutil.Amount,
	localScript []byte, remoteScript []byte,
	closeOpt ...lnwallet.ChanCloseOpt) (
	input.Signature, *wire.MsgTx, chainutil.Amount, error) {

	args := m.Called(fee, localScript, remoteScript, closeOpt)

	return args.Get(0).(input.Signature), args.Get(1).(*wire.MsgTx),
		args.Get(2).(chainutil.Amount), args.Error(3)
}

func (m *mockCloseSigner) CompleteCooperativeClose(localSig,
	remoteSig input.Signature,
	localScript, remoteScript []byte,
	fee chainutil.Amount, closeOpt ...lnwallet.ChanCloseOpt,
) (*wire.MsgTx, chainutil.Amount, error) {

	args := m.Called(
		localSig, remoteSig, localScript, remoteScript, fee, closeOpt,
	)

	return args.Get(0).(*wire.MsgTx), args.Get(1).(chainutil.Amount),
		args.Error(2)
}
