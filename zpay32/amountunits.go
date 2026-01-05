package zpay32

import (
	"fmt"
	"strconv"

	"github.com/flokiorg/flnd/lnwire"
)

var (
	// toMSat is a map from a unit to a function that converts an amount
	// of that unit to milliloki.
	toMSat = map[byte]func(uint64) (lnwire.MilliLoki, error){
		'm': mBtcToMSat,
		'u': uBtcToMSat,
		'n': nBtcToMSat,
		'p': pBtcToMSat,
	}

	// fromMSat is a map from a unit to a function that converts an amount
	// in milliloki to an amount of that unit.
	fromMSat = map[byte]func(lnwire.MilliLoki) (uint64, error){
		'm': mSatToMBtc,
		'u': mSatToUBtc,
		'n': mSatToNBtc,
		'p': mSatToPBtc,
	}
)

// mBtcToMSat converts the given amount in milliBTC to milliloki.
func mBtcToMSat(m uint64) (lnwire.MilliLoki, error) {
	return lnwire.MilliLoki(m) * 100000000, nil
}

// uBtcToMSat converts the given amount in microBTC to milliloki.
func uBtcToMSat(u uint64) (lnwire.MilliLoki, error) {
	return lnwire.MilliLoki(u * 100000), nil
}

// nBtcToMSat converts the given amount in nanoBTC to milliloki.
func nBtcToMSat(n uint64) (lnwire.MilliLoki, error) {
	return lnwire.MilliLoki(n * 100), nil
}

// pBtcToMSat converts the given amount in picoBTC to milliloki.
func pBtcToMSat(p uint64) (lnwire.MilliLoki, error) {
	if p < 10 {
		return 0, fmt.Errorf("minimum amount is 10p")
	}
	if p%10 != 0 {
		return 0, fmt.Errorf("amount %d pBTC not expressible in msat",
			p)
	}
	return lnwire.MilliLoki(p / 10), nil
}

// mSatToMBtc converts the given amount in milliloki to milliBTC.
func mSatToMBtc(msat lnwire.MilliLoki) (uint64, error) {
	if msat%100000000 != 0 {
		return 0, fmt.Errorf("%d msat not expressible "+
			"in mBTC", msat)
	}
	return uint64(msat / 100000000), nil
}

// mSatToUBtc converts the given amount in milliloki to microBTC.
func mSatToUBtc(msat lnwire.MilliLoki) (uint64, error) {
	if msat%100000 != 0 {
		return 0, fmt.Errorf("%d msat not expressible "+
			"in uBTC", msat)
	}
	return uint64(msat / 100000), nil
}

// mSatToNBtc converts the given amount in milliloki to nanoBTC.
func mSatToNBtc(msat lnwire.MilliLoki) (uint64, error) {
	if msat%100 != 0 {
		return 0, fmt.Errorf("%d msat not expressible in nBTC", msat)
	}
	return uint64(msat / 100), nil
}

// mSatToPBtc converts the given amount in milliloki to picoBTC.
func mSatToPBtc(msat lnwire.MilliLoki) (uint64, error) {
	return uint64(msat * 10), nil
}

// decodeAmount returns the amount encoded by the provided string in
// millisatoshi.
func decodeAmount(amount string) (lnwire.MilliLoki, error) {
	if len(amount) < 1 {
		return 0, fmt.Errorf("amount must be non-empty")
	}

	// If last character is a digit, then the amount can just be
	// interpreted as FLC.
	char := amount[len(amount)-1]
	digit := char - '0'
	if digit >= 0 && digit <= 9 {
		btc, err := strconv.ParseUint(amount, 10, 64)
		if err != nil {
			return 0, err
		}
		return lnwire.MilliLoki(btc) * mSatPerBtc, nil
	}

	// If not a digit, it must be part of the known units.
	conv, ok := toMSat[char]
	if !ok {
		return 0, fmt.Errorf("unknown multiplier %c", char)
	}

	// Known unit.
	num := amount[:len(amount)-1]
	if len(num) < 1 {
		return 0, fmt.Errorf("number must be non-empty")
	}

	am, err := strconv.ParseUint(num, 10, 64)
	if err != nil {
		return 0, err
	}

	return conv(am)
}

// encodeAmount encodes the provided millisatoshi amount using as few characters
// as possible.
func encodeAmount(msat lnwire.MilliLoki) (string, error) {
	// If possible to express in FLC, that will always be the shortest
	// representation.
	if msat%mSatPerBtc == 0 {
		return strconv.FormatInt(int64(msat/mSatPerBtc), 10), nil
	}

	// Should always be expressible in pico FLC.
	pico, err := fromMSat['p'](msat)
	if err != nil {
		return "", fmt.Errorf("unable to express %d msat as pBTC: %w",
			msat, err)
	}
	shortened := strconv.FormatUint(pico, 10) + "p"
	for unit, conv := range fromMSat {
		am, err := conv(msat)
		if err != nil {
			// Not expressible using this unit.
			continue
		}

		// Save the shortest found representation.
		str := strconv.FormatUint(am, 10) + string(unit)
		if len(str) < len(shortened) {
			shortened = str
		}
	}

	return shortened, nil
}
