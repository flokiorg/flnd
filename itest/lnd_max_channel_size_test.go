package itest

import (
	"fmt"

	"github.com/flokiorg/flnd/funding"
	"github.com/flokiorg/flnd/lntest"
	"github.com/flokiorg/flnd/lnwallet"
	"github.com/flokiorg/go-flokicoin/chainutil"
)

// testMaxChannelSize tests that lnd handles --maxchansize parameter correctly.
// Wumbo nodes should enforce a default soft limit of 10 FLC by default. This
// limit can be adjusted with --maxchansize config option.
func testMaxChannelSize(ht *lntest.HarnessTest) {
	// We'll make two new nodes, both wumbo but with the default limit on
	// maximum channel size (10 FLC)
	wumboNode := ht.NewNode(
		"wumbo", []string{"--protocol.wumbo-channels"},
	)
	wumboNode2 := ht.NewNode(
		"wumbo2", []string{"--protocol.wumbo-channels"},
	)

	// We'll send 11 FLC to the wumbo node so it can test the wumbo soft
	// limit.
	ht.FundCoins(11*chainutil.LokiPerFlokicoin, wumboNode)

	// Next we'll connect both nodes, then attempt to make a wumbo channel
	// funding request, which should fail as it exceeds the default wumbo
	// soft limit of 10 FLC.
	ht.EnsureConnected(wumboNode, wumboNode2)

	chanAmt := funding.MaxFlcFundingAmountWumbo + 1
	// The test should show failure due to the channel exceeding our max
	// size.
	expectedErr := lnwallet.ErrChanTooLarge(
		chanAmt, funding.MaxFlcFundingAmountWumbo,
	)
	ht.OpenChannelAssertErr(
		wumboNode, wumboNode2,
		lntest.OpenChannelParams{Amt: chanAmt}, expectedErr,
	)

	// We'll now make another wumbo node with appropriate maximum channel
	// size to accept our wumbo channel funding.
	wumboNode3 := ht.NewNode(
		"wumbo3", []string{
			"--protocol.wumbo-channels",
			fmt.Sprintf(
				"--maxchansize=%v",
				int64(funding.MaxFlcFundingAmountWumbo+1),
			),
		},
	)

	// Creating a wumbo channel between these two nodes should succeed.
	ht.EnsureConnected(wumboNode, wumboNode3)
	ht.OpenChannel(
		wumboNode, wumboNode3, lntest.OpenChannelParams{Amt: chanAmt},
	)
}
