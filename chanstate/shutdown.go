package chanstate

import (
	"github.com/flokiorg/flnd/lntypes"
	"github.com/flokiorg/flnd/lnwire"
	"github.com/flokiorg/flnd/tlv"
)

// ShutdownInfo contains various info about the shutdown initiation of a
// channel.
type ShutdownInfo struct {
	// DeliveryScript is the address that we have included in any previous
	// Shutdown message for a particular channel and so should include in
	// any future re-sends of the Shutdown message.
	DeliveryScript tlv.RecordT[tlv.TlvType0, lnwire.DeliveryAddress]

	// LocalInitiator is true if we sent a Shutdown message before ever
	// receiving a Shutdown message from the remote peer.
	LocalInitiator tlv.RecordT[tlv.TlvType1, bool]
}

// NewShutdownInfo constructs a new ShutdownInfo object.
func NewShutdownInfo(deliveryScript lnwire.DeliveryAddress,
	locallyInitiated bool) *ShutdownInfo {

	return &ShutdownInfo{
		DeliveryScript: tlv.NewRecordT[tlv.TlvType0](deliveryScript),
		LocalInitiator: tlv.NewPrimitiveRecord[tlv.TlvType1](
			locallyInitiated,
		),
	}
}

// Closer identifies the ChannelParty that initiated the coop-closure process.
func (s ShutdownInfo) Closer() lntypes.ChannelParty {
	if s.LocalInitiator.Val {
		return lntypes.Local
	}

	return lntypes.Remote
}
