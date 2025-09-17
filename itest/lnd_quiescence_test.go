package itest

import (
	"github.com/flokiorg/flnd/lnrpc"
	"github.com/flokiorg/flnd/lnrpc/devrpc"
	"github.com/flokiorg/flnd/lnrpc/routerrpc"
	"github.com/flokiorg/flnd/lntest"
	"github.com/flokiorg/flnd/lntest/node"
	"github.com/flokiorg/go-flokicoin/chainutil"
	"github.com/stretchr/testify/require"
)

// testQuiescence tests whether we can come to agreement on quiescence of a
// channel. We initiate quiescence via RPC and if it succeeds we verify that
// the expected initiator is the resulting initiator.
//
// NOTE FOR REVIEW: this could be improved by blasting the channel with HTLC
// traffic on both sides to increase the surface area of the change under test.
func testQuiescence(ht *lntest.HarnessTest) {
	cfg := node.CfgAnchor
	chanPoints, nodes := ht.CreateSimpleNetwork(
		[][]string{cfg, cfg}, lntest.OpenChannelParams{
			Amt: chainutil.Amount(1000000),
		})

	alice, bob := nodes[0], nodes[1]
	chanPoint := chanPoints[0]

	res := alice.RPC.Quiesce(&devrpc.QuiescenceRequest{
		ChanId: chanPoint,
	})

	require.True(ht, res.Initiator)

	req := &routerrpc.SendPaymentRequest{
		Dest:           alice.PubKey[:],
		Amt:            100,
		PaymentHash:    ht.Random32Bytes(),
		FinalCltvDelta: finalCltvDelta,
		FeeLimitMsat:   noFeeLimitMsat,
	}

	ht.SendPaymentAssertFail(
		bob, req,
		// This fails with insufficient balance because the bandwidth
		// manager reports 0 bandwidth if a link is not eligible for
		// forwarding, which is the case during quiescence.
		lnrpc.PaymentFailureReason_FAILURE_REASON_INSUFFICIENT_BALANCE,
	)
}
