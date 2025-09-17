package lnwallet

import (
	"fmt"
	"testing"

	"github.com/flokiorg/flnd/input"
	"github.com/flokiorg/flnd/lnwire"
	"github.com/flokiorg/go-flokicoin/chainutil"
	"github.com/stretchr/testify/require"
)

// TestDefaultRoutingFeeLimitForAmount tests that we use the correct default
// routing fee depending on the amount.
func TestDefaultRoutingFeeLimitForAmount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		amount        lnwire.MilliLoki
		expectedLimit lnwire.MilliLoki
	}{
		{
			amount:        1,
			expectedLimit: 1,
		},
		{
			amount:        lnwire.NewMLokFromLokis(1_000),
			expectedLimit: lnwire.NewMLokFromLokis(1_000),
		},
		{
			amount:        lnwire.NewMLokFromLokis(1_001),
			expectedLimit: 50_050,
		},
		{
			amount:        5_000_000_000,
			expectedLimit: 250_000_000,
		},
	}

	for _, test := range tests {
		test := test

		t.Run(fmt.Sprintf("%d sats", test.amount), func(t *testing.T) {
			feeLimit := DefaultRoutingFeeLimitForAmount(test.amount)
			require.Equal(t, int64(test.expectedLimit), int64(feeLimit))
		})
	}
}

// TestDustLimitForSize tests that we receive the expected dust limits for
// various script types from btcd's GetDustThreshold function.
func TestDustLimitForSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		size          int
		expectedLimit chainutil.Amount
	}{
		{
			name:          "p2pkh dust limit",
			size:          input.P2PKHSize,
			expectedLimit: chainutil.Amount(546),
		},
		{
			name:          "p2sh dust limit",
			size:          input.P2SHSize,
			expectedLimit: chainutil.Amount(540),
		},
		{
			name:          "p2wpkh dust limit",
			size:          input.P2WPKHSize,
			expectedLimit: chainutil.Amount(294),
		},
		{
			name:          "p2wsh dust limit",
			size:          input.P2WSHSize,
			expectedLimit: chainutil.Amount(330),
		},
		{
			name:          "unknown witness limit",
			size:          input.UnknownWitnessSize,
			expectedLimit: chainutil.Amount(354),
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			dustlimit := DustLimitForSize(test.size)
			require.Equal(t, test.expectedLimit, dustlimit)
		})
	}
}
