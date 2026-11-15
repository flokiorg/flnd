package onionmessage

import (
	"testing"

	"github.com/flokiorg/go-flokicoin/crypto"
	sphinx "github.com/flokiorg/lightning-onion"
	"github.com/flokiorg/flnd/fn"
	"github.com/flokiorg/flnd/lnwire"
	"github.com/flokiorg/flnd/record"
	"github.com/flokiorg/flnd/tlv"
	"github.com/stretchr/testify/require"
)

// processOnionMessageTest defines the test parameters for testing
// processOnionMessage with different routing scenarios.
type processOnionMessageTest struct {
	name             string
	hopsToBlind      []*sphinx.HopInfo
	isDeliver        bool
	expectedNextNode *crypto.PublicKey
	expectedOverride *crypto.PublicKey
}

// TestProcessOnionMessage tests the processOnionMessage function with various
// forwarding and delivery scenarios.
func TestProcessOnionMessage(t *testing.T) {
	// Helper to generate keys.
	genKey := func() *crypto.PrivateKey {
		k, err := crypto.NewPrivateKey()
		require.NoError(t, err)
		return k
	}

	// Setup the local node (router).
	nodeKeyA := genKey()
	pubKeyA := nodeKeyA.PubKey()

	router := sphinx.NewRouter(
		&sphinx.PrivKeyECDH{PrivKey: nodeKeyA},
		sphinx.NewNoOpReplayLog(),
	)
	require.NoError(t, router.Start())
	defer router.Stop()

	resolver := newMockNodeIDResolver()

	// Pre-generate keys for test cases.
	nodeKeyB := genKey()
	pubKeyB := nodeKeyB.PubKey()

	overrideKey := genKey()
	pubKeyOverride := overrideKey.PubKey()

	// Helper to encode route data.
	encodeData := func(data *record.BlindedRouteData) []byte {
		b, err := record.EncodeBlindedRouteData(data)
		require.NoError(t, err)
		return b
	}

	// Case 1 Data: Forward Action Success.
	nextNodeByPubKey := fn.NewLeft[*crypto.PublicKey, lnwire.ShortChannelID](
		pubKeyB,
	)
	rd1A := record.NewNonFinalBlindedRouteDataOnionMessage(
		nextNodeByPubKey, nil, nil,
	)
	rd1B := &record.BlindedRouteData{}
	hops1 := []*sphinx.HopInfo{
		{NodePub: pubKeyA, PlainText: encodeData(rd1A)},
		{NodePub: pubKeyB, PlainText: encodeData(rd1B)},
	}

	// Case 2 Data: Forward Action Path Key Override Success.
	nextNodeWithOverride := fn.NewLeft[
		*crypto.PublicKey, lnwire.ShortChannelID,
	](pubKeyB)
	rd2A := record.NewNonFinalBlindedRouteDataOnionMessage(
		nextNodeWithOverride, pubKeyOverride, nil,
	)
	rd2B := &record.BlindedRouteData{}
	hops2 := []*sphinx.HopInfo{
		{NodePub: pubKeyA, PlainText: encodeData(rd2A)},
		{NodePub: pubKeyB, PlainText: encodeData(rd2B)},
	}

	// Case 3 Data: Forward Action Success with SCID resolution.
	scid := lnwire.NewShortChanIDFromInt(12345)
	resolver.addPeer(scid, pubKeyB)

	nextNodeBySCID := fn.NewRight[*crypto.PublicKey](
		scid,
	)
	rd3A := record.NewNonFinalBlindedRouteDataOnionMessage(
		nextNodeBySCID, nil, nil,
	)
	rd3B := &record.BlindedRouteData{}
	hops3 := []*sphinx.HopInfo{
		{NodePub: pubKeyA, PlainText: encodeData(rd3A)},
		{NodePub: pubKeyB, PlainText: encodeData(rd3B)},
	}

	// Case 4 Data: Deliver Action Success.
	rd4 := &record.BlindedRouteData{}
	hops4 := []*sphinx.HopInfo{
		{NodePub: pubKeyA, PlainText: encodeData(rd4)},
	}

	tests := []processOnionMessageTest{
		{
			name:             "Forward Action Success",
			hopsToBlind:      hops1,
			isDeliver:        false,
			expectedNextNode: pubKeyB,
			expectedOverride: nil, // No path key override.
		},
		{
			name: "Forward Action Path Key Override " +
				"Success",
			hopsToBlind:      hops2,
			isDeliver:        false,
			expectedNextNode: pubKeyB,
			expectedOverride: pubKeyOverride,
		},
		{
			name: "Forward Action Success with SCID " +
				"Resolution",
			hopsToBlind:      hops3,
			isDeliver:        false,
			expectedNextNode: pubKeyB,
			expectedOverride: nil,
		},
		{
			name:        "Deliver Action Success",
			hopsToBlind: hops4,
			isDeliver:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testProcessOnionMessageCase(t, router, resolver, tc)
		})
	}
}

// testProcessOnionMessageCase is a helper that executes a single test case for
// processOnionMessage, building the blinded path and verifying the result.
func testProcessOnionMessageCase(t *testing.T, router OnionRouter,
	resolver NodeIDResolver, tc processOnionMessageTest) {

	blindedPath := BuildBlindedPath(t, tc.hopsToBlind)
	msg, expectedCipherTexts := BuildOnionMessage(
		t, blindedPath, nil,
	)

	// Process the message.
	result := processOnionMessage(t.Context(), router, resolver, msg)
	require.True(t, result.IsOk())

	// Verify result.
	if tc.isDeliver {
		result.WhenOk(func(action routingAction) {
			// Should be deliverAction.
			require.True(t, action.IsRight())
			action.WhenRight(func(dlvrAction deliverAction) {
				require.Equal(
					t,
					expectedCipherTexts[0],
					dlvrAction.payload.EncryptedData,
				)
			})
		})
	} else {
		result.WhenOk(func(action routingAction) {
			// Should be forwardAction.
			require.True(t, action.IsLeft())
			action.WhenLeft(func(fwdAction forwardAction) {
				require.Equal(
					t, tc.expectedNextNode,
					fwdAction.nextNodeID,
				)

				if tc.expectedOverride != nil {
					require.Equal(
						t, tc.expectedOverride,
						fwdAction.nextPathKey,
					)
				} else {
					require.NotNil(t, fwdAction.nextPathKey)
				}

				require.NotEmpty(t, fwdAction.nextPacket)
				require.Equal(
					t,
					expectedCipherTexts[0],
					fwdAction.payload.EncryptedData,
				)
			})
		})
	}
}

// TestIsForwarding tests the isForwarding function.
func TestIsForwarding(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		packet   *sphinx.ProcessedPacket
		expected bool
	}{
		{
			name: "forwarding",
			packet: &sphinx.ProcessedPacket{
				Action: sphinx.MoreHops,
			},
			expected: true,
		},
		{
			name: "delivery",
			packet: &sphinx.ProcessedPacket{
				Action: sphinx.ExitNode,
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := isForwarding(test.packet)
			require.Equal(t, test.expected, result)
		})
	}
}

// TestDeriveNextPathKey tests the deriveNextPathKey function.
func TestDeriveNextPathKey(t *testing.T) {
	t.Parallel()

	// create a private key for the router.
	privKey, err := crypto.NewPrivateKey()
	require.NoError(t, err)

	// create a path key.
	sessionKey, err := crypto.NewPrivateKey()
	require.NoError(t, err)
	pathKey := sessionKey.PubKey()

	// Create a router. We don't need a replay log for this test as
	// NextEphemeral doesn't use it.
	router := sphinx.NewRouter(&sphinx.PrivKeyECDH{PrivKey: privKey}, nil)

	t.Run("override present", func(t *testing.T) {
		overrideKey, err := crypto.NewPrivateKey()
		require.NoError(t, err)

		override := tlv.NewPrimitiveRecord[tlv.TlvType8](
			overrideKey.PubKey(),
		)
		optOverride := tlv.SomeRecordT(override)

		// Router can be nil as it shouldn't be used.
		result := deriveNextPathKey(nil, pathKey, optOverride)
		require.Equal(t, overrideKey.PubKey(), result)
	})

	t.Run("derive success", func(t *testing.T) {
		override := tlv.OptionalRecordT[tlv.TlvType8,
			*crypto.PublicKey]{}

		result := deriveNextPathKey(router, pathKey, override)
		require.NotNil(t, result)

		// Verify it matches manual derivation.
		expected, err := router.NextEphemeral(pathKey)
		require.NoError(t, err)
		require.Equal(t, expected, result)
	})

	// It's currently impossible to test derivation failure as there is no
	// way to make the key derivation fail with an error. You can only make
	// it panick by passing in a nil path key. This is due to how
	// PrivKeyECDH.ECDH is implemented in the keychain package.
}
