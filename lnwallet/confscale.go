package lnwallet

import (
	"github.com/flokiorg/flnd/lnwire"
	"github.com/flokiorg/go-flokicoin/chainutil"
)

const (
	// MaxFundingAmount is the protocol-level maximum channel size for
	// non-wumbo channels.
	MaxFundingAmount = chainutil.Amount(16777215)

	// minRequiredConfs is the minimum number of confirmations we'll require
	// for a channel to be considered open.
	minRequiredConfs = 3

	// maxRequiredConfs is the maximum number of confirmations we'll require
	// for a channel to be considered open.
	maxRequiredConfs = 6
)

// ScaleNumConfs returns the number of confirmations required for a channel
// to be considered open, given its capacity and push amount.
func ScaleNumConfs(chanAmt chainutil.Amount,
	pushAmt lnwire.MilliLoki) uint16 {

	// For large channels we increase the number of confirmations we require
	// for the channel to be considered open. As it is always the responder
	// that gets to choose value, the pushAmt is value being pushed to us.
	// This means we have more to lose in the case this gets re-orged out,
	// and we will require more confirmations before we consider it open.

	// If this is a wumbo channel, then we'll require the max amount of
	// confirmations.
	if chanAmt > MaxFundingAmount {
		return uint16(maxRequiredConfs)
	}

	// If not we return a value scaled linearly between 3 and 6, depending on
	// channel size.
	maxChannelSize := uint64(
		lnwire.NewMSatFromLokis(MaxFundingAmount))
	stake := lnwire.NewMSatFromLokis(chanAmt) + pushAmt
	conf := uint64(maxRequiredConfs) * uint64(stake) / maxChannelSize
	if conf < minRequiredConfs {
		conf = minRequiredConfs
	}
	if conf > maxRequiredConfs {
		conf = maxRequiredConfs
	}
	return uint16(conf)
}

// FundingConfsForAmounts returns the number of confirmations required for a
// channel to be considered open, given its capacity and push amount.
func FundingConfsForAmounts(chanAmt chainutil.Amount,
	pushAmt lnwire.MilliLoki) uint16 {

	return ScaleNumConfs(chanAmt, pushAmt)
}
