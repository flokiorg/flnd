package lnwire

import (
	"fmt"
	"io"

	"github.com/flokiorg/flnd/tlv"
	"github.com/flokiorg/go-flokicoin/chainutil"
)

const (
	// mLokScale is a value that's used to scale satoshis to milli-satoshis, and
	// the other way around.
	mLokScale uint64 = 1000

	// MaxMilliLoki is the maximum number of mloks that can be expressed
	// in this data type.
	MaxMilliLoki = ^MilliLoki(0)
)

// MilliLoki are the native unit of the Lightning Network. A milli-satoshi
// is simply 1/1000th of a satoshi. There are 1000 milli-satoshis in a single
// satoshi. Within the network, all HTLC payments are denominated in
// milli-satoshis. As milli-satoshis aren't deliverable on the native
// blockchain, before settling to broadcasting, the values are rounded down to
// the nearest satoshi.
type MilliLoki uint64

// NewMLokFromLokis creates a new MilliLoki instance from a target amount
// of satoshis.
func NewMLokFromLokis(sat chainutil.Amount) MilliLoki {
	return MilliLoki(uint64(sat) * mLokScale)
}

// ToBTC converts the target MilliLoki amount to its corresponding value
// when expressed in BTC.
func (m MilliLoki) ToFLC() float64 {
	lok := m.ToLokis()
	return lok.ToFLC()
}

// ToSatoshis converts the target MilliLoki amount to satoshis. Simply, this
// sheds a factor of 1000 from the mSAT amount in order to convert it to SAT.
func (m MilliLoki) ToLokis() chainutil.Amount {
	return chainutil.Amount(uint64(m) / mLokScale)
}

// String returns the string representation of the mSAT amount.
func (m MilliLoki) String() string {
	return fmt.Sprintf("%v mLOK", uint64(m))
}

// TODO(roasbeef): extend with arithmetic operations?

// Record returns a TLV record that can be used to encode/decode a MilliLoki
// to/from a TLV stream.
func (m *MilliLoki) Record() tlv.Record {
	mlok := uint64(*m)

	return tlv.MakeDynamicRecord(
		0, m, tlv.SizeBigSize(&mlok), encodeMilliLokis,
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
