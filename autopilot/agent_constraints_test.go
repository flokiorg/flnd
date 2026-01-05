package autopilot

import (
	prand "math/rand"
	"testing"
	"time"

	"github.com/flokiorg/flnd/lnwire"
	"github.com/flokiorg/go-flokicoin/chainutil"
)

func TestConstraintsChannelBudget(t *testing.T) {
	t.Parallel()

	prand.Seed(time.Now().Unix())

	const (
		minChanSize = 0
		maxChanSize = chainutil.Amount(chainutil.LokiPerFlokicoin)

		chanLimit = 3

		threshold = 0.5
	)

	constraints := NewConstraints(
		minChanSize,
		maxChanSize,
		chanLimit,
		0,
		threshold,
	)

	randChanID := func() lnwire.ShortChannelID {
		return lnwire.NewShortChanIDFromInt(uint64(prand.Int63()))
	}

	testCases := []struct {
		channels  []LocalChannel
		walletAmt chainutil.Amount

		needMore     bool
		amtAvailable chainutil.Amount
		numMore      uint32
	}{
		// Many available funds, but already have too many active open
		// channels.
		{
			[]LocalChannel{
				{
					ChanID:  randChanID(),
					Balance: chainutil.Amount(prand.Int31()),
				},
				{
					ChanID:  randChanID(),
					Balance: chainutil.Amount(prand.Int31()),
				},
				{
					ChanID:  randChanID(),
					Balance: chainutil.Amount(prand.Int31()),
				},
			},
			chainutil.Amount(chainutil.LokiPerFlokicoin * 10),
			false,
			0,
			0,
		},

		// Ratio of funds in channels and total funds meets the
		// threshold.
		{
			[]LocalChannel{
				{
					ChanID:  randChanID(),
					Balance: chainutil.Amount(chainutil.LokiPerFlokicoin),
				},
				{
					ChanID:  randChanID(),
					Balance: chainutil.Amount(chainutil.LokiPerFlokicoin),
				},
			},
			chainutil.Amount(chainutil.LokiPerFlokicoin * 2),
			false,
			0,
			0,
		},

		// Ratio of funds in channels and total funds is below the
		// threshold. We have 10 FLC allocated amongst channels and
		// funds, atm. We're targeting 50%, so 5 FLC should be
		// allocated. Only 1 FLC is atm, so 4 FLC should be
		// recommended. We should also request 2 more channels as the
		// limit is 3.
		{
			[]LocalChannel{
				{
					ChanID:  randChanID(),
					Balance: chainutil.Amount(chainutil.LokiPerFlokicoin),
				},
			},
			chainutil.Amount(chainutil.LokiPerFlokicoin * 9),
			true,
			chainutil.Amount(chainutil.LokiPerFlokicoin * 4),
			2,
		},

		// Ratio of funds in channels and total funds is below the
		// threshold. We have 14 FLC total amongst the wallet's
		// balance, and our currently opened channels. Since we're
		// targeting a 50% allocation, we should commit 7 FLC. The
		// current channels commit 4 FLC, so we should expected 3 FLC
		// to be committed. We should only request a single additional
		// channel as the limit is 3.
		{
			[]LocalChannel{
				{
					ChanID:  randChanID(),
					Balance: chainutil.Amount(chainutil.LokiPerFlokicoin),
				},
				{
					ChanID:  randChanID(),
					Balance: chainutil.Amount(chainutil.LokiPerFlokicoin * 3),
				},
			},
			chainutil.Amount(chainutil.LokiPerFlokicoin * 10),
			true,
			chainutil.Amount(chainutil.LokiPerFlokicoin * 3),
			1,
		},

		// Ratio of funds in channels and total funds is above the
		// threshold.
		{
			[]LocalChannel{
				{
					ChanID:  randChanID(),
					Balance: chainutil.Amount(chainutil.LokiPerFlokicoin),
				},
				{
					ChanID:  randChanID(),
					Balance: chainutil.Amount(chainutil.LokiPerFlokicoin),
				},
			},
			chainutil.Amount(chainutil.LokiPerFlokicoin),
			false,
			0,
			0,
		},
	}

	for i, testCase := range testCases {
		amtToAllocate, numMore := constraints.ChannelBudget(
			testCase.channels, testCase.walletAmt,
		)

		if amtToAllocate != testCase.amtAvailable {
			t.Fatalf("test #%v: expected %v, got %v",
				i, testCase.amtAvailable, amtToAllocate)
		}
		if numMore != testCase.numMore {
			t.Fatalf("test #%v: expected %v, got %v",
				i, testCase.numMore, numMore)
		}
	}
}
