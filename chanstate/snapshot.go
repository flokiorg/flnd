package chanstate

import (
	"github.com/flokiorg/flnd/lnwire"
	"github.com/flokiorg/go-flokicoin/chaincfg/chainhash"
	"github.com/flokiorg/go-flokicoin/chainutil"
	"github.com/flokiorg/go-flokicoin/crypto"
	"github.com/flokiorg/go-flokicoin/wire"
)

// ChannelSnapshot is a frozen snapshot of the current channel state. A
// snapshot is detached from the original channel that generated it, providing
// read-only access to the current or prior state of an active channel.
//
// TODO(roasbeef): remove all together? pretty much just commitment.
type ChannelSnapshot struct {
	// RemoteIdentity is the identity public key of the remote node that we
	// are maintaining the open channel with.
	RemoteIdentity crypto.PublicKey

	// ChanPoint is the outpoint that created the channel. This output is
	// found within the funding transaction and uniquely identified the
	// channel on the resident chain.
	ChannelPoint wire.OutPoint

	// ChainHash is the genesis hash of the chain that the channel resides
	// within.
	ChainHash chainhash.Hash

	// Capacity is the total capacity of the channel.
	Capacity chainutil.Amount

	// TotalMSatSent is the total number of milli-loki we've sent
	// within this channel.
	TotalMSatSent lnwire.MilliLoki

	// TotalMSatReceived is the total number of milli-loki we've
	// received within this channel.
	TotalMSatReceived lnwire.MilliLoki

	// ChannelCommitment is the current up-to-date commitment for the
	// target channel.
	ChannelCommitment
}
