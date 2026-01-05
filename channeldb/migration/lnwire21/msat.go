package lnwire

import (
	"fmt"
	"io"

	"github.com/flokiorg/flnd/tlv"
	"github.com/flokiorg/go-flokicoin/chainutil"
)

const (
	// mSatScale is a value that's used to scale loki to milli-loki, and
	// the other way around.
	mSatScale uint64 = 1000

	// MaxMilliLoki is the maximum number of msats that can be expressed
	// in this data type.
	MaxMilliLoki = ^MilliLoki(0)
)

// MilliLoki are the native unit of the Lightning Network. A milli-satoshi
// is simply 1/1000th of a satoshi. There are 1000 milli-loki in a single
// satoshi. Within the network, all HTLC payments are denominated in
// milli-loki. As milli-loki aren't deliverable on the native
// blockchain, before settling to broadcasting, the values are rounded down to
// the nearest satoshi.
type MilliLoki uint64

// NewMSatFromLokis creates a new MilliLoki instance from a target amount
// of loki.
func NewMSatFromLokis(sat chainutil.Amount) MilliLoki {
	return MilliLoki(uint64(sat) * mSatScale)
}

// ToBTC converts the target MilliLoki amount to its corresponding value
// when expressed in FLC.
func (m MilliLoki) ToFLC() float64 {
	sat := m.ToLokis()
	return sat.ToFLC()
}

// ToLokis converts the target MilliLoki amount to loki. Simply, this
// sheds a factor of 1000 from the mSAT amount in order to convert it to SAT.
func (m MilliLoki) ToLokis() chainutil.Amount {
	return chainutil.Amount(uint64(m) / mSatScale)
}

// String returns the string representation of the mSAT amount.
func (m MilliLoki) String() string {
	return fmt.Sprintf("%v mSAT", uint64(m))
}

// TODO(roasbeef): extend with arithmetic operations?

// Record returns a TLV record that can be used to encode/decode a MilliLoki
// to/from a TLV stream.
func (m *MilliLoki) Record() tlv.Record {
	msat := uint64(*m)

	return tlv.MakeDynamicRecord(
		0, m, tlv.SizeBigSize(&msat), encodeMilliLokis,
		decodeMilliLokis,
	)
}
func encodeMilliLokis(w io.Writer, val interface{}, buf *[8]byte) error {
	if v, ok := val.(*MilliLoki); ok {
		bigSize := uint64(*v)

		return tlv.EBigSize(w, &bigSize, buf)
	}

	return tlv.NewTypeForEncodingErr(val, "lnwire.MilliLoki")
}

func decodeMilliLokis(r io.Reader, val interface{}, buf *[8]byte,
	l uint64) error {

	if v, ok := val.(*MilliLoki); ok {
		var bigSize uint64
		err := tlv.DBigSize(r, &bigSize, buf, l)
		if err != nil {
			return err
		}

		*v = MilliLoki(bigSize)

		return nil
	}

	return tlv.NewTypeForDecodingErr(val, "lnwire.MilliLoki", l, l)
}
