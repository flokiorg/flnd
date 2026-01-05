package chainfee

import (
	"fmt"

	"github.com/flokiorg/flnd/lntypes"
	"github.com/flokiorg/go-flokicoin/blockchain"
	"github.com/flokiorg/go-flokicoin/chainutil"
)

const (
	// FeePerKwFloor is the lowest fee rate in sat/kw that we should use for
	// estimating transaction fees before signing.
	FeePerKwFloor SatPerKWeight = 253

	// AbsoluteFeePerKwFloor is the lowest fee rate in sat/kw of a
	// transaction that we should ever _create_. This is the equivalent
	// of 1 sat/byte in sat/kw.
	AbsoluteFeePerKwFloor SatPerKWeight = 250
)

// SatPerVByte represents a fee rate in sat/vbyte.
type SatPerVByte chainutil.Amount

// FeePerKWeight converts the current fee rate from sat/vb to sat/kw.
func (s SatPerVByte) FeePerKWeight() SatPerKWeight {
	return SatPerKWeight(s * 1000 / blockchain.WitnessScaleFactor)
}

// FeePerKVByte converts the current fee rate from sat/vb to sat/kvb.
func (s SatPerVByte) FeePerKVByte() SatPerKVByte {
	return SatPerKVByte(s * 1000)
}

// String returns a human-readable string of the fee rate.
func (s SatPerVByte) String() string {
	return fmt.Sprintf("%v sat/vb", int64(s))
}

// SatPerKVByte represents a fee rate in sat/kb.
type SatPerKVByte chainutil.Amount

// FeeForVSize calculates the fee resulting from this fee rate and the given
// vsize in vbytes.
func (s SatPerKVByte) FeeForVSize(vbytes lntypes.VByte) chainutil.Amount {
	return chainutil.Amount(s) * chainutil.Amount(vbytes) / 1000
}

// FeePerKWeight converts the current fee rate from sat/kb to sat/kw.
func (s SatPerKVByte) FeePerKWeight() SatPerKWeight {
	return SatPerKWeight(s / blockchain.WitnessScaleFactor)
}

// String returns a human-readable string of the fee rate.
func (s SatPerKVByte) String() string {
	return fmt.Sprintf("%v sat/kvb", int64(s))
}

// SatPerKWeight represents a fee rate in sat/kw.
type SatPerKWeight chainutil.Amount

// NewSatPerKWeight creates a new fee rate in sat/kw.
func NewSatPerKWeight(fee chainutil.Amount, wu lntypes.WeightUnit) SatPerKWeight {
	return SatPerKWeight(fee.MulF64(1000 / float64(wu)))
}

// FeeForWeight calculates the fee resulting from this fee rate and the given
// weight in weight units (wu).
func (s SatPerKWeight) FeeForWeight(wu lntypes.WeightUnit) chainutil.Amount {
	// The resulting fee is rounded down, as specified in BOLT#03.
	return chainutil.Amount(s) * chainutil.Amount(wu) / 1000
}

// FeeForWeightRoundUp calculates the fee resulting from this fee rate and the
// given weight in weight units (wu), rounding up to the nearest satoshi.
func (s SatPerKWeight) FeeForWeightRoundUp(
	wu lntypes.WeightUnit) chainutil.Amount {

	return (chainutil.Amount(s)*chainutil.Amount(wu) + 999) / 1000
}

// FeeForVByte calculates the fee resulting from this fee rate and the given
// size in vbytes (vb).
func (s SatPerKWeight) FeeForVByte(vb lntypes.VByte) chainutil.Amount {
	return s.FeePerKVByte().FeeForVSize(vb)
}

// FeePerKVByte converts the current fee rate from sat/kw to sat/kb.
func (s SatPerKWeight) FeePerKVByte() SatPerKVByte {
	return SatPerKVByte(s * blockchain.WitnessScaleFactor)
}

// FeePerVByte converts the current fee rate from sat/kw to sat/vb.
func (s SatPerKWeight) FeePerVByte() SatPerVByte {
	return SatPerVByte(s * blockchain.WitnessScaleFactor / 1000)
}

// String returns a human-readable string of the fee rate.
func (s SatPerKWeight) String() string {
	return fmt.Sprintf("%v sat/kw", int64(s))
}
