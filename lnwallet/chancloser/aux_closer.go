package chancloser

import (
	"github.com/flokiorg/flnd/fn"
	"github.com/flokiorg/flnd/lnwallet"
	"github.com/flokiorg/flnd/lnwallet/types"
	"github.com/flokiorg/flnd/lnwire"
	"github.com/flokiorg/go-flokicoin/wire"
)

// AuxCloseOutputs is used to specify extra outputs that should be used when
// constructing the co-op close transaction.
type AuxCloseOutputs struct {
	// ExtraCloseOutputs is a set of extra outputs that should be included
	// in the close transaction.
	ExtraCloseOutputs []lnwallet.CloseOutput

	// CustomSort is a custom function that can be used to sort the
	// transaction outputs. If this isn't set, then the default BIP-69
	// sorting is used.
	CustomSort lnwallet.CloseSortFunc
}

// AuxChanCloser is used to allow an external caller to modify the co-op close
// transaction.
type AuxChanCloser interface {
	// ShutdownBlob returns the set of custom records that should be
	// included in the shutdown message.
	ShutdownBlob(req types.AuxShutdownReq) (fn.Option[lnwire.CustomRecords],
		error)

	// AuxCloseOutputs returns the set of custom outputs that should be used
	// to construct the co-op close transaction.
	AuxCloseOutputs(desc types.AuxCloseDesc) (fn.Option[AuxCloseOutputs],
		error)

	// FinalizeClose is called after the close transaction has been agreed
	// upon.
	FinalizeClose(desc types.AuxCloseDesc, closeTx *wire.MsgTx) error
}
