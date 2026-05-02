package zpay32

// We use package `zpay32` rather than `zpay32_test` in order to share test data
// with the internal tests.

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/flokiorg/flnd/fn"
	"github.com/flokiorg/flnd/lnwire"
	"github.com/flokiorg/go-flokicoin/chaincfg"
	"github.com/flokiorg/go-flokicoin/chaincfg/chainhash"
	"github.com/flokiorg/go-flokicoin/chainutil"
	"github.com/flokiorg/go-flokicoin/crypto"
	"github.com/flokiorg/go-flokicoin/crypto/ecdsa"
	sphinx "github.com/flokiorg/lightning-onion"
	"github.com/stretchr/testify/require"
)

var (
	testMillisat24BTC    = lnwire.MilliLoki(2400000000000)
	testMillisat2500uBTC = lnwire.MilliLoki(250000000)
	testMillisat25mBTC   = lnwire.MilliLoki(2500000000)
	testMillisat20mBTC   = lnwire.MilliLoki(2000000000)
	testMillisat10mBTC   = lnwire.MilliLoki(1000000000)

	testPaymentHash = [32]byte{
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05,
		0x06, 0x07, 0x08, 0x09, 0x00, 0x01, 0x02, 0x03,
		0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x01, 0x02,
	}

	testPaymentAddr = [32]byte{
		0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x01, 0x02,
		0x06, 0x07, 0x08, 0x09, 0x00, 0x01, 0x02, 0x03,
		0x08, 0x09, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05,
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
	}

	specPaymentAddr = [32]byte{
		0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11,
		0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11,
		0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11,
		0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11,
	}

	testNonUTF8Str      = "1 cup coffee\xff\xfe\xfd"
	testEmptyString     = ""
	testCupOfCoffee     = "1 cup coffee"
	testCoffeeBeans     = "coffee beans"
	testCupOfNonsense   = "ナンセンス 1杯"
	testPleaseConsider  = "Please consider supporting this project"
	testPaymentMetadata = "payment metadata inside"

	testPrivKeyBytes, _     = hex.DecodeString("e126f68f7eafcc8b74f54d269fe206be715000f94dac067d1c04a8ca3b2db734")
	testPrivKey, testPubKey = crypto.PrivKeyFromBytes(testPrivKeyBytes)

	testDescriptionHashSlice = chainhash.HashB([]byte("One piece of chocolate cake, one icecream cone, one pickle, one slice of swiss cheese, one slice of salami, one lollypop, one piece of cherry pie, one sausage, one cupcake, and one slice of watermelon"))

	testExpiry0  = time.Duration(0) * time.Second
	testExpiry60 = time.Duration(60) * time.Second

	testAddrTestnet, _       = chainutil.DecodeAddress("mk2QpYatsKicvFVuTAQLBryyccRXMUaGHP", &chaincfg.TestNet3Params)
	testRustyAddr, _         = chainutil.DecodeAddress("F6G2LhQWNM2FLE7aPuUZEuSmMk3C4CgeS8", &chaincfg.MainNetParams)
	testAddrMainnetP2SH, _   = chainutil.DecodeAddress("3GkKDgJAWJnizg6Tz7DBM8uDtdHtrUkQ2X", &chaincfg.MainNetParams)
	testAddrMainnetP2WPKH, _ = chainutil.DecodeAddress("fc1qw508d6qejxtdg4y5r3zarvary0c5xw7kkrmtt5", &chaincfg.MainNetParams)
	testAddrMainnetP2WSH, _  = chainutil.DecodeAddress("fc1qrp33g0q5c5txsp9arysrx4k6zdkfs4nce4xj0gdcccefvpysxf3qu35lxt", &chaincfg.MainNetParams)
	testAddrMainnetP2TR, _   = chainutil.DecodeAddress("fc1pptdvg0d2nj99568"+
		"qn6ssdy4cygnwuxgw2ukmnwgwz7jpqjz2kszsard44p",
		&chaincfg.MainNetParams)

	testHopHintPubkeyBytes1, _ = hex.DecodeString("029e03a901b85534ff1e92c43c74431f7ce72046060fcf7a95c37e148f78c77255")
	testHopHintPubkey1, _      = crypto.ParsePubKey(testHopHintPubkeyBytes1)
	testHopHintPubkeyBytes2, _ = hex.DecodeString("039e03a901b85534ff1e92c43c74431f7ce72046060fcf7a95c37e148f78c77255")
	testHopHintPubkey2, _      = crypto.ParsePubKey(testHopHintPubkeyBytes2)

	testSingleHop = []HopHint{
		{
			NodeID:                    testHopHintPubkey1,
			ChannelID:                 0x0102030405060708,
			FeeBaseMSat:               0,
			FeeProportionalMillionths: 20,
			CLTVExpiryDelta:           3,
		},
	}
	testDoubleHop = []HopHint{
		{
			NodeID:                    testHopHintPubkey1,
			ChannelID:                 0x0102030405060708,
			FeeBaseMSat:               1,
			FeeProportionalMillionths: 20,
			CLTVExpiryDelta:           3,
		},
		{
			NodeID:                    testHopHintPubkey2,
			ChannelID:                 0x030405060708090a,
			FeeBaseMSat:               2,
			FeeProportionalMillionths: 30,
			CLTVExpiryDelta:           4,
		},
	}

	testMessageSigner = MessageSigner{
		SignCompact: func(msg []byte) ([]byte, error) {
			hash := chainhash.HashB(msg)

			return ecdsa.SignCompact(testPrivKey, hash, true), nil
		},
	}

	emptyFeatures = lnwire.NewFeatureVector(nil, lnwire.Features)

	// Must be initialized in init().
	testDescriptionHash [32]byte

	testBlindedPK1Bytes, _ = hex.DecodeString("03f3311e948feb5115242c4e39" +
		"6c81c448ab7ee5fd24c4e24e66c73533cc4f98b8")
	testBlindedHopPK1, _   = crypto.ParsePubKey(testBlindedPK1Bytes)
	testBlindedPK2Bytes, _ = hex.DecodeString("03a8c97ed5cd40d474e4ef18c8" +
		"99854b25e5070106504cb225e6d2c112d61a805e")
	testBlindedHopPK2, _   = crypto.ParsePubKey(testBlindedPK2Bytes)
	testBlindedPK3Bytes, _ = hex.DecodeString("0220293926219d8efe733336e2" +
		"b674570dd96aa763acb3564e6e367b384d861a0a")
	testBlindedHopPK3, _   = crypto.ParsePubKey(testBlindedPK3Bytes)
	testBlindedPK4Bytes, _ = hex.DecodeString("02c75eb336a038294eaaf76015" +
		"8b2e851c3c0937262e35401ae64a1bee71a2e40c")
	testBlindedHopPK4, _ = crypto.ParsePubKey(testBlindedPK4Bytes)

	blindedPath1 = &BlindedPaymentPath{
		FeeBaseMsat:                 40,
		FeeRate:                     20,
		CltvExpiryDelta:             130,
		HTLCMinMsat:                 2,
		HTLCMaxMsat:                 100,
		Features:                    lnwire.EmptyFeatureVector(),
		FirstEphemeralBlindingPoint: testBlindedHopPK1,
		Hops: []*sphinx.BlindedHopInfo{
			{
				BlindedNodePub: testBlindedHopPK2,
				CipherText:     []byte{1, 2, 3, 4, 5},
			},
			{
				BlindedNodePub: testBlindedHopPK3,
				CipherText:     []byte{5, 4, 3, 2, 1},
			},
			{
				BlindedNodePub: testBlindedHopPK4,
				CipherText: []byte{
					1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
					11, 12, 13, 14,
				},
			},
		},
	}

	blindedPath2 = &BlindedPaymentPath{
		FeeBaseMsat:                 4,
		FeeRate:                     2,
		CltvExpiryDelta:             10,
		HTLCMinMsat:                 0,
		HTLCMaxMsat:                 10,
		Features:                    lnwire.EmptyFeatureVector(),
		FirstEphemeralBlindingPoint: testBlindedHopPK4,
		Hops: []*sphinx.BlindedHopInfo{
			{
				BlindedNodePub: testBlindedHopPK3,
				CipherText:     []byte{1, 2, 3, 4, 5},
			},
		},
	}
)

func init() {
	copy(testDescriptionHash[:], testDescriptionHashSlice[:])
}

// TestDecodeEncode tests that an encoded invoice gets decoded into the expected
// Invoice object, and that reencoding the decoded invoice gets us back to the
// original encoded string.
//
//nolint:ll
func TestDecodeEncode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		encodedInvoice string
		valid          bool
		decodedInvoice func() *Invoice
		decodeOpts     []DecodeOption
		skipEncoding   bool
		beforeEncoding func(*Invoice)
	}{
		{
			encodedInvoice: "asdsaddnasdnas", // no hrp
			valid:          false,
		},
		{
			encodedInvoice: "lnfc1abcde", // too short
			valid:          false,
		},
		{
			encodedInvoice: "1asdsaddnv4wudz", // empty hrp
			valid:          false,
		},
		{
			encodedInvoice: "ln1asdsaddnv4wudz", // hrp too short
			valid:          false,
		},
		{
			encodedInvoice: "llts1dasdajtkfl6", // no "ln" prefix
			valid:          false,
		},
		{
			encodedInvoice: "lnts1dasdapukz0w", // invalid segwit prefix
			valid:          false,
		},
		{
			encodedInvoice: "lnfcm1aaa2pvx9g", // invalid amount
			valid:          false,
		},
		{
			encodedInvoice: "lnfc1000000000m1", // invalid amount
			valid:          false,
		},
		{
			encodedInvoice: "lnfc20m1pvjluezhp58yjmdan79s6qqdhdzgynm4zwqd5d7xmw5fk98klysy043l2ahrqspp5qqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqypqfqqepvrhrm9s57hejg0p662ur5j5cr03890fa7k2pypgttmh4897d3raaq85a293e9jpuqwl0rnfuwzam7yr8e690nd2ypcq9hlkdwdvycqg6gu0g", // empty fallback address field
			valid:          false,
		},
		{
			encodedInvoice: "lnfc20m1pvjluezpp5qqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqypqhp58yjmdan79s6qqdhdzgynm4zwqd5d7xmw5fk98klysy043l2ahrqsfpp3qjmp7lwpagxun9pygexvgpjdc4jdj85frqg00000000j9n4evl6mr5aj9f58zp6fyjzup6ywn3x6sk8akg5v4tgn2q8g4fhx05wf6juaxu9760yp46454gpg5mtzgerlzezqcqvjnhjh8z3g2qqcc3kc3", // invalid routing info length: not a multiple of 51
			valid:          false,
		},
		{
			// no payment hash set
			encodedInvoice: "lnfc20m1pvjluezhp58yjmdan79s6qqdhdzgynm4zwqd5d7xmw5fk98klysy043l2ahrqsjv38luh6p6s2xrv3mzvlmzaya43376h0twal5ax0k6p47498hp3hnaymzhsn424rxqjs0q7apn26yrhaxltq3vzwpqj9nc2r3kzwccx9v7yf",
			valid:          false,
			decodedInvoice: func() *Invoice {
				return &Invoice{
					Net:             &chaincfg.MainNetParams,
					MilliSat:        &testMillisat20mBTC,
					Timestamp:       time.Unix(1496314658, 0),
					DescriptionHash: &testDescriptionHash,
					Destination:     testPubKey,
					Features:        emptyFeatures,
				}
			},
		},
		{
			// Both Description and DescriptionHash set.
			encodedInvoice: "lnfc20m1pvjluezpp5qqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqypqdpl2pkx2ctnv5sxxmmwwd5kgetjypeh2ursdae8g6twvus8g6rfwvs8qun0dfjkxaqhp58yjmdan79s6qqdhdzgynm4zwqd5d7xmw5fk98klysy043l2ahrqs03vghs8y0kuj4ulrzls8ln7fnm9dk7sjsnqmghql6hd6jut36clkqpyuq0s5m6fhureyz0szx2qjc8hkgf4xc2hpw8jpu26jfeyvf4cp9t5vel",
			valid:          false,
			decodedInvoice: func() *Invoice {
				return &Invoice{
					Net:             &chaincfg.MainNetParams,
					MilliSat:        &testMillisat20mBTC,
					Timestamp:       time.Unix(1496314658, 0),
					PaymentHash:     &testPaymentHash,
					Description:     &testPleaseConsider,
					DescriptionHash: &testDescriptionHash,
					Destination:     testPubKey,
					Features:        emptyFeatures,
				}
			},
		},
		{
			// Neither Description nor DescriptionHash set.
			encodedInvoice: "lnfc20m1pvjluezpp5qqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqypqn2rne0kagfl4e0xag0w6hqeg2dwgc54hrm9m0auw52dhwhwcu559qav309h598pyzn69wh2nqauneyyesnpmaax0g6acr8lh9559jmcqsxtras",
			valid:          false,
			decodedInvoice: func() *Invoice {
				return &Invoice{
					Net:         &chaincfg.MainNetParams,
					MilliSat:    &testMillisat20mBTC,
					Timestamp:   time.Unix(1496314658, 0),
					PaymentHash: &testPaymentHash,
					Destination: testPubKey,
					Features:    emptyFeatures,
				}
			},
		},
		{
			// Has a few unknown fields, should just be ignored.
			encodedInvoice: "lnfc20m1pvjluezpp5qqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqypqdpl2pkx2ctnv5sxxmmwwd5kgetjypeh2ursdae8g6twvus8g6rfwvs8qun0dfjkxaqnp4q0n326hr8v9zprg8gsvezcch06gfaqqhde2aj730yg0durunfhv666zxxede8cdasrxu3xhk7e3alpjf9f2qmuqq6gulxxa8u89mk5pjxwmas2m0xvr9hymftjz27ksq2t9lcy6htq0reknqrl3c8zz85k2cqetk2d7",
			valid:          true,
			decodedInvoice: func() *Invoice {
				return &Invoice{
					Net:         &chaincfg.MainNetParams,
					MilliSat:    &testMillisat20mBTC,
					Timestamp:   time.Unix(1496314658, 0),
					PaymentHash: &testPaymentHash,
					Description: &testPleaseConsider,
					Destination: testPubKey,
					Features:    emptyFeatures,
				}
			},
			skipEncoding: true, // Skip encoding since we don't have the unknown fields to encode.
		},
		{
			// Ignore unknown witness version in fallback address.
			encodedInvoice: "lnfc20m1pvjluezpp5qqqsyqcyq5rqwzqfqq" +
				"qsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqypqhp58yjmdan" +
				"79s6qqdhdzgynm4zwqd5d7xmw5fk98klysy043l2ahrq" +
				"sfp4z6yn92zrp97a6q5hhh8swys7uf4hm9tr8a0xylnk" +
				"26fvkg3jx0sdsxvma0zvf2h0pycyyzdrmjncq6lzrfuw" +
				"xfhv6gzz4q5303n3up6as4ghe5qthg7x20z7vae8w5rq" +
				"u6de3g4jl7kvuap3qedprqsqqmgqqx9v7yf",
			valid: true,
			decodedInvoice: func() *Invoice {
				return &Invoice{
					Net:             &chaincfg.MainNetParams,
					MilliSat:        &testMillisat20mBTC,
					Timestamp:       time.Unix(1496314658, 0),
					PaymentHash:     &testPaymentHash,
					DescriptionHash: &testDescriptionHash,
					Destination:     testPubKey,
					Features:        emptyFeatures,
				}
			},
			skipEncoding: true, // Skip encoding since we don't have the unknown fields to encode.
		},
		{
			// Ignore fields with unknown lengths.
			encodedInvoice: "lnfc241pveeq09pp5qqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqypqhp58yjmdan79s6qqdhdzgynm4zwqd5d7xmw5fk98klysy043l2ahrqsnp4q0n326hr8v9zprg8gsvezcch06gfaqqhde2aj730yg0durunfhv66g58qt0xj5hlj3yhnx2m2gexhj4usagkpzrtzamfu0q4k9ggf5smk47838lk977976ehn0p7rfuu90pkagf22vzdxqny7x552g7frcsqqqzk5rj",
			valid:          true,
			decodedInvoice: func() *Invoice {
				return &Invoice{
					Net:             &chaincfg.MainNetParams,
					MilliSat:        &testMillisat24BTC,
					Timestamp:       time.Unix(1503429093, 0),
					PaymentHash:     &testPaymentHash,
					Destination:     testPubKey,
					DescriptionHash: &testDescriptionHash,
					Features:        emptyFeatures,
				}
			},
			skipEncoding: true, // Skip encoding since we don't have the unknown fields to encode.
		},
		{
			// Invoice with no amount.
			encodedInvoice: "lnfc1pvjluezpp5qqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqypqdq5xysxxatsyp3k7enxv4js7dv5g4nqzg2qaksnfsuptpn3dvglezfc5ejgvs0nuj0wl80t2yr3lvsl838kd8kdcpcv042pdjtrrcg3vzzl2xvdk85qyukhnn6sjyqpwfzl55",
			valid:          true,
			decodedInvoice: func() *Invoice {
				return &Invoice{
					Net:         &chaincfg.MainNetParams,
					Timestamp:   time.Unix(1496314658, 0),
					PaymentHash: &testPaymentHash,
					Description: &testCupOfCoffee,
					Destination: testPubKey,
					Features:    emptyFeatures,
				}
			},
			beforeEncoding: func(i *Invoice) {
				// Since this destination pubkey was recovered
				// from the signature, we must set it nil before
				// encoding to get back the same invoice string.
				i.Destination = nil
			},
		},
		{
			// Please make a donation of any amount using rhash 0001020304050607080900010203040506070809000102030405060708090102 to me @03e7156ae33b0a208d0744199163177e909e80176e55d97a2f221ede0f934dd9ad
			encodedInvoice: "lnfc1pvjluezpp5qqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqypqdpl2pkx2ctnv5sxxmmwwd5kgetjypeh2ursdae8g6twvus8g6rfwvs8qun0dfjkxaqyszdep8j8f66q7xqghad7md35ywdxwhgr9xqr5n9xc72lqkvq6vrs2g98vpcj4zqgnetrfcmnf7dp7njjw8d7hwqjdy2mwc2ukj45vqqhmsdsu",
			valid:          true,
			decodedInvoice: func() *Invoice {
				return &Invoice{
					Net:         &chaincfg.MainNetParams,
					Timestamp:   time.Unix(1496314658, 0),
					PaymentHash: &testPaymentHash,
					Description: &testPleaseConsider,
					Destination: testPubKey,
					Features:    emptyFeatures,
				}
			},
			beforeEncoding: func(i *Invoice) {
				// Since this destination pubkey was recovered
				// from the signature, we must set it nil before
				// encoding to get back the same invoice string.
				i.Destination = nil
			},
		},
		{
			// Same as above, pubkey set in 'n' field.
			encodedInvoice: "lnfc241pveeq09pp5qqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqypqdqqnp4q0n326hr8v9zprg8gsvezcch06gfaqqhde2aj730yg0durunfhv66nncy3vkgm498puq6ndlffxa0z0dpqsa68a6ypvrd0nl4xvv3m4yk6m6qkeck4gr5hu47cmlxhj0943lu2s6ylk0xkzt6u552qdlmv5cqnes55v",
			valid:          true,
			decodedInvoice: func() *Invoice {
				return &Invoice{
					Net:         &chaincfg.MainNetParams,
					MilliSat:    &testMillisat24BTC,
					Timestamp:   time.Unix(1503429093, 0),
					PaymentHash: &testPaymentHash,
					Destination: testPubKey,
					Description: &testEmptyString,
					Features:    emptyFeatures,
				}
			},
		},
		{
			// Please send $3 for a cup of coffee to the same peer, within 1 minute
			encodedInvoice: "lnfc2500u1pvjluezpp5qqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqypqdq5xysxxatsyp3k7enxv4jsxqzpum6elh5cfk2x0p87z08gy4427dtsuxlz3e5lf85zezh7fcf3satq59ut6da4cvrr5n25wyvnyq2en3dxhmdujnvzkkh3044cudlwnwjqp9q6tv3",
			valid:          true,
			decodedInvoice: func() *Invoice {
				i, _ := NewInvoice(
					&chaincfg.MainNetParams,
					testPaymentHash,
					time.Unix(1496314658, 0),
					Amount(testMillisat2500uBTC),
					Description(testCupOfCoffee),
					Destination(testPubKey),
					Expiry(testExpiry60))
				return i
			},
			beforeEncoding: func(i *Invoice) {
				// Since this destination pubkey was recovered
				// from the signature, we must set it nil before
				// encoding to get back the same invoice string.
				i.Destination = nil
			},
		},
		{
			// Please send 0.0025 FLC for a cup of nonsense (ナンセンス 1杯) to the same peer, within 1 minute
			encodedInvoice: "lnfc2500u1pvjluezpp5qqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqypqdpquwpc4curk03c9wlrswe78q4eyqc7d8d0xqzpuefmsqyy7lzzkgkw6sjmz2k6jqettj76up287j0kry0pc29yu4ezrx3v8gumdagd6mdpecv0n63t9p5xhe9ej3mcg7zrwjjdv2fr6vnqpdwuql2",
			valid:          true,
			decodedInvoice: func() *Invoice {
				i, _ := NewInvoice(
					&chaincfg.MainNetParams,
					testPaymentHash,
					time.Unix(1496314658, 0),
					Amount(testMillisat2500uBTC),
					Description(testCupOfNonsense),
					Destination(testPubKey),
					Expiry(testExpiry60))
				return i
			},
			beforeEncoding: func(i *Invoice) {
				// Since this destination pubkey was recovered
				// from the signature, we must set it nil before
				// encoding to get back the same invoice string.
				i.Destination = nil
			},
		},
		{
			// Now send $24 for an entire list of things (hashed)
			encodedInvoice: "lnfc20m1pvjluezpp5qqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqypqhp58yjmdan79s6qqdhdzgynm4zwqd5d7xmw5fk98klysy043l2ahrqshlzqryryxgayg49lssdx6dynzg6gfx4ype9as3lmz7ujxlqqd7lq5wd07q5mljv345up9yuwy9xgwfpj92qeg3gt4jfet4djp3n750sq03hq0q",
			valid:          true,
			decodedInvoice: func() *Invoice {
				return &Invoice{
					Net:             &chaincfg.MainNetParams,
					MilliSat:        &testMillisat20mBTC,
					Timestamp:       time.Unix(1496314658, 0),
					PaymentHash:     &testPaymentHash,
					DescriptionHash: &testDescriptionHash,
					Destination:     testPubKey,
					Features:        emptyFeatures,
				}
			},
			beforeEncoding: func(i *Invoice) {
				// Since this destination pubkey was recovered
				// from the signature, we must set it nil before
				// encoding to get back the same invoice string.
				i.Destination = nil
			},
		},
		{
			// The same, on testnet, with a fallback address mk2QpYatsKicvFVuTAQLBryyccRXMUaGHP
			encodedInvoice: "lntf20m1pvjluezpp5qqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqypqhp58yjmdan79s6qqdhdzgynm4zwqd5d7xmw5fk98klysy043l2ahrqsfpp3x9et2e20v6pu37c5d9vax37wxq72un98y5tp20fupr0a3azt9eda59ctk0ztr0jx5yx4jt7qe2rleev5mpl4tgkq6lha8e37js06tzmqpe5e0q0znal4tdmfgkzex3sxstexmpsp4646r9",
			valid:          true,
			decodedInvoice: func() *Invoice {
				return &Invoice{
					Net:             &chaincfg.TestNet3Params,
					MilliSat:        &testMillisat20mBTC,
					Timestamp:       time.Unix(1496314658, 0),
					PaymentHash:     &testPaymentHash,
					DescriptionHash: &testDescriptionHash,
					Destination:     testPubKey,
					FallbackAddr:    testAddrTestnet,
					Features:        emptyFeatures,
				}
			},
			beforeEncoding: func(i *Invoice) {
				// Since this destination pubkey was recovered
				// from the signature, we must set it nil before
				// encoding to get back the same invoice string.
				i.Destination = nil
			},
		},
		{
			// On mainnet, with fallback address 1RustyRX2oai4EYYDpQGWvEL62BBGqN9T with extra routing info to get to node 029e03a901b85534ff1e92c43c74431f7ce72046060fcf7a95c37e148f78c77255
			encodedInvoice: "lnfc20m1pvjluezpp5qqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqypqhp58yjmdan79s6qqdhdzgynm4zwqd5d7xmw5fk98klysy043l2ahrqsfpp3qjmp7lwpagxun9pygexvgpjdc4jdj85frzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvlmrk3q709eudlymzqesjj6reqkwu7hpl5q537rtye797za0e86250jkc8gx8jcdpsv3pvkj54jkh6p7ny7t63wppmtcvy8y5yhpe99cqe2w3tv",
			valid:          true,
			decodedInvoice: func() *Invoice {
				return &Invoice{
					Net:             &chaincfg.MainNetParams,
					MilliSat:        &testMillisat20mBTC,
					Timestamp:       time.Unix(1496314658, 0),
					PaymentHash:     &testPaymentHash,
					DescriptionHash: &testDescriptionHash,
					Destination:     testPubKey,
					FallbackAddr:    testRustyAddr,
					RouteHints:      [][]HopHint{testSingleHop},
					Features:        emptyFeatures,
				}
			},
			beforeEncoding: func(i *Invoice) {
				// Since this destination pubkey was recovered
				// from the signature, we must set it nil before
				// encoding to get back the same invoice string.
				i.Destination = nil
			},
		},
		{
			// On mainnet, with fallback address 1RustyRX2oai4EYYDpQGWvEL62BBGqN9T with extra routing info to go via nodes 029e03a901b85534ff1e92c43c74431f7ce72046060fcf7a95c37e148f78c77255 then 039e03a901b85534ff1e92c43c74431f7ce72046060fcf7a95c37e148f78c77255
			encodedInvoice: "lnfc20m1pvjluezpp5qqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqypqhp58yjmdan79s6qqdhdzgynm4zwqd5d7xmw5fk98klysy043l2ahrqsfpp3qjmp7lwpagxun9pygexvgpjdc4jdj85fr9yq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqpqqqqq9qqqvpeuqafqxu92d8lr6fvg0r5gv0heeeqgcrqlnm6jhphu9y00rrhy4grqszsvpcgpy9qqqqqqgqqqqq7qqzqp63f3hvc658wm20txhzp4d39ngqvh6tlq76fl3xp66wa7uxz6gxn98mfynese2qqvzusgdyskap60h9cjl3jqvwpu9ql4fvz6qjc7nqqhcdlg4",
			valid:          true,
			decodedInvoice: func() *Invoice {
				return &Invoice{
					Net:             &chaincfg.MainNetParams,
					MilliSat:        &testMillisat20mBTC,
					Timestamp:       time.Unix(1496314658, 0),
					PaymentHash:     &testPaymentHash,
					DescriptionHash: &testDescriptionHash,
					Destination:     testPubKey,
					FallbackAddr:    testRustyAddr,
					RouteHints:      [][]HopHint{testDoubleHop},
					Features:        emptyFeatures,
				}
			},
			beforeEncoding: func(i *Invoice) {
				// Since this destination pubkey was recovered
				// from the signature, we must set it nil before
				// encoding to get back the same invoice string.
				i.Destination = nil
			},
		},
		{
			// On mainnet, with fallback (p2sh) address 3EktnHQD7RiAE6uzMj2ZifT9YgRrkSgzQX
			encodedInvoice: "lnfc20m1pvjluezpp5qqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqypqhp58yjmdan79s6qqdhdzgynm4zwqd5d7xmw5fk98klysy043l2ahrqsfppj555eyamtlzsr8uq75w2txvz5a4klfg5p5w7ts0qx40nggxhujtx29wj8sktpu7l4egsp3xup6rl6khgu0ep99xuxmd5lwfdzmr5pnf9elcmvmwytkwu4xwy65hg4ms2mwp774rspen489n",
			valid:          true,
			decodedInvoice: func() *Invoice {
				return &Invoice{
					Net:             &chaincfg.MainNetParams,
					MilliSat:        &testMillisat20mBTC,
					Timestamp:       time.Unix(1496314658, 0),
					PaymentHash:     &testPaymentHash,
					DescriptionHash: &testDescriptionHash,
					Destination:     testPubKey,
					FallbackAddr:    testAddrMainnetP2SH,
					Features:        emptyFeatures,
				}
			},
			beforeEncoding: func(i *Invoice) {
				// Since this destination pubkey was recovered
				// from the signature, we must set it nil before
				// encoding to get back the same invoice string.
				i.Destination = nil
			},
		},
		{
			// On mainnet, please send $30 coffee beans supporting
			// features 9, 15 and 99, using secret 0x11...
			encodedInvoice: "lnfc25m1pvjluezpp5qqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqypqdq5vdhkven9v5sxyetpdeessp5zyg3zyg3zyg3zyg3zyg3zyg3zyg3zyg3zyg3zyg3zyg3zyg3zygs9q5sqqqqqqqqqqqqqqqpqsq6fj7h6fd7dg0vdv60yez2qd0lfg3jsxhxnzwxyuxn00dqtuz6e0zvs9djavnx3jtnusmx0wjfde3aqxjdcd7tu6yg2knn626pv7y9ggpz70mjz",
			valid:          true,
			decodedInvoice: func() *Invoice {
				return &Invoice{
					Net:         &chaincfg.MainNetParams,
					MilliSat:    &testMillisat25mBTC,
					Timestamp:   time.Unix(1496314658, 0),
					PaymentHash: &testPaymentHash,
					PaymentAddr: fn.Some(specPaymentAddr),
					Description: &testCoffeeBeans,
					Destination: testPubKey,
					Features: lnwire.NewFeatureVector(
						lnwire.NewRawFeatureVector(9, 15, 99),
						lnwire.Features,
					),
				}
			},
			beforeEncoding: func(i *Invoice) {
				// Since this destination pubkey was recovered
				// from the signature, we must set it nil before
				// encoding to get back the same invoice string.
				i.Destination = nil
			},
		},
		{
			// On mainnet, please send $30 coffee beans supporting
			// features 9, 15, 99, and 100, using secret 0x11...
			encodedInvoice: "lnfc25m1pvjluezpp5qqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqypqdq5vdhkven9v5sxyetpdeessp5zyg3zyg3zyg3zyg3zyg3zyg3zyg3zyg3zyg3zyg3zyg3zyg3zygs9q4psqqqqqqqqqqqqqqqpqsqtpqkl24kplknxfsdp7gv4wu6tadsya4ax23rkll7drvqmy0wnhcjsf2w8tc6xd2s0fwpveqhjxjfxqh67yvpts9crq800g384jtey6cpmh9um0",
			valid:          true,
			decodedInvoice: func() *Invoice {
				return &Invoice{
					Net:         &chaincfg.MainNetParams,
					MilliSat:    &testMillisat25mBTC,
					Timestamp:   time.Unix(1496314658, 0),
					PaymentHash: &testPaymentHash,
					PaymentAddr: fn.Some(specPaymentAddr),
					Description: &testCoffeeBeans,
					Destination: testPubKey,
					Features: lnwire.NewFeatureVector(
						lnwire.NewRawFeatureVector(9, 15, 99, 100),
						lnwire.Features,
					),
				}
			},
			beforeEncoding: func(i *Invoice) {
				// Since this destination pubkey was recovered
				// from the signature, we must set it nil before
				// encoding to get back the same invoice string.
				i.Destination = nil
			},
		},
		{
			// On mainnet, with fallback (p2wpkh) address fc1qw508d6qejxtdg4y5r3zarvary0c5xw7kv8f3t4
			encodedInvoice: "lnfc20m1pvjluezpp5qqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqypqhp58yjmdan79s6qqdhdzgynm4zwqd5d7xmw5fk98klysy043l2ahrqsfppqw508d6qejxtdg4y5r3zarvary0c5xw7kdm5q5gn477zresrpjj09tad67qrm0ufq8wkpd540pavmgx0ejsmk9sh9wh4m64xy7dwzwzz3p5f2f8v7psdw0yts4laln8ypuxd5z7sq30nvrz",
			valid:          true,
			decodedInvoice: func() *Invoice {
				return &Invoice{
					Net:             &chaincfg.MainNetParams,
					MilliSat:        &testMillisat20mBTC,
					Timestamp:       time.Unix(1496314658, 0),
					PaymentHash:     &testPaymentHash,
					DescriptionHash: &testDescriptionHash,
					Destination:     testPubKey,
					FallbackAddr:    testAddrMainnetP2WPKH,
					Features:        emptyFeatures,
				}
			},
			beforeEncoding: func(i *Invoice) {
				// Since this destination pubkey was recovered
				// from the signature, we must set it nil before
				// encoding to get back the same invoice string.
				i.Destination = nil
			},
		},
		{
			// On mainnet, with fallback (p2wsh) address fc1qrp33g0q5c5txsp9arysrx4k6zdkfs4nce4xj0gdcccefvpysxf3qccfmv3
			encodedInvoice: "lnfc20m1pvjluezpp5qqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqypqhp58yjmdan79s6qqdhdzgynm4zwqd5d7xmw5fk98klysy043l2ahrqsfp4qrp33g0q5c5txsp9arysrx4k6zdkfs4nce4xj0gdcccefvpysxf3qj5v90wn5p5f6vjgpmh05fer24ncd9l3fsavlp3rwd5gkqzms3gxy6wpa5ddrp3mpe0aqwwq33e5k8ea9a7uud4r2e7vr2act98dhs7sqppltk2",
			valid:          true,
			decodedInvoice: func() *Invoice {
				return &Invoice{
					Net:             &chaincfg.MainNetParams,
					MilliSat:        &testMillisat20mBTC,
					Timestamp:       time.Unix(1496314658, 0),
					PaymentHash:     &testPaymentHash,
					DescriptionHash: &testDescriptionHash,
					Destination:     testPubKey,
					FallbackAddr:    testAddrMainnetP2WSH,
					Features:        emptyFeatures,
				}
			},
			beforeEncoding: func(i *Invoice) {
				// Since this destination pubkey was recovered
				// from the signature, we must set it nil before
				// encoding to get back the same invoice string.
				i.Destination = nil
			},
		},
		{
			// On mainnet, with fallback (p2tr) address "fc1pptdvg0d
			// 2nj99568qn6ssdy4cygnwuxgw2ukmnwgwz7jpqjz2kszse2s3lm"
			encodedInvoice: "lnfc20m1pvjluezpp5qqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqypqhp58yjmdan79s6qqdhdzgynm4zwqd5d7xmw5fk98klysy043l2ahrqsfp4pptdvg0d2nj99568qn6ssdy4cygnwuxgw2ukmnwgwz7jpqjz2kszsnp4q0n326hr8v9zprg8gsvezcch06gfaqqhde2aj730yg0durunfhv666y0grxf3whfctun67saerrechf44400sc2887skxh6ghxccxx2ayl8jx20y6d9tm9ctctqn2a0h0dwm7ez3vyvyqgu07yz7j0s083asppp4xkc",

			valid: true,
			decodedInvoice: func() *Invoice {
				return &Invoice{
					Net:             &chaincfg.MainNetParams,
					MilliSat:        &testMillisat20mBTC,
					Timestamp:       time.Unix(1496314658, 0),
					PaymentHash:     &testPaymentHash,
					DescriptionHash: &testDescriptionHash,
					Destination:     testPubKey,
					FallbackAddr:    testAddrMainnetP2TR,
					Features:        emptyFeatures,
				}
			},
			beforeEncoding: func(i *Invoice) {
				// Since this destination pubkey was recovered
				// from the signature, we must set it nil before
				// encoding to get back the same invoice string.
				i.Destination = nil
			},
		},
		{
			// On mainnet, with fallback (p2tr) address "fc1pptdvg0d
			// 2nj99568qn6ssdy4cygnwuxgw2ukmnwgwz7jpqjz2kszse2s3lm"
			// using the test vector payment from BOLT 11
			encodedInvoice: "lnfc20m1pvjluezpp5qqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqypqhp58yjmdan79s6qqdhdzgynm4zwqd5d7xmw5fk98klysy043l2ahrqsfp4pptdvg0d2nj99568qn6ssdy4cygnwuxgw2ukmnwgwz7jpqjz2kszsnp4q0n326hr8v9zprg8gsvezcch06gfaqqhde2aj730yg0durunfhv66sp5zyg3zyg3zyg3zyg3zyg3zyg3zyg3zyg3zyg3zyg3zyg3zyg3zygs9qrsgql35t9678xjzfhehj5pejwtt55zu680h7df2flqsyx3hlsktj7z8xv25swp9lgdfjmhul5kd4n3h4azy9gfkjzuh83n206qnlrp53lrcp500835",

			valid: true,
			decodedInvoice: func() *Invoice {
				return &Invoice{
					Net:             &chaincfg.MainNetParams,
					MilliSat:        &testMillisat20mBTC,
					Timestamp:       time.Unix(1496314658, 0),
					PaymentHash:     &testPaymentHash,
					DescriptionHash: &testDescriptionHash,
					Destination:     testPubKey,
					FallbackAddr:    testAddrMainnetP2TR,
					Features: lnwire.NewFeatureVector(
						lnwire.NewRawFeatureVector(
							8, 14,
						),
						lnwire.Features,
					),
					PaymentAddr: fn.Some(specPaymentAddr),
				}
			},
			// Skip encoding since LND encode the tagged fields
			// in a different order.
			skipEncoding: true,
		},
		{
			// Send 2500uBTC for a cup of coffee with a custom CLTV
			// expiry value.
			encodedInvoice: "lnfc2500u1pvjluezpp5qqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqypqdq5xysxxatsyp3k7enxv4jscqzysnp4q0n326hr8v9zprg8gsvezcch06gfaqqhde2aj730yg0durunfhv6626nynvqmyhpldqhtng8xuax840xdz69l9ez9cv5pahczk322e8pr3n4xld7kfxh9mv8wg5h8xd65ac50me8qvkf28rjtmlxfhu7eftgpnla9tm",
			valid:          true,
			decodedInvoice: func() *Invoice {
				i, _ := NewInvoice(
					&chaincfg.MainNetParams,
					testPaymentHash,
					time.Unix(1496314658, 0),
					Amount(testMillisat2500uBTC),
					Description(testCupOfCoffee),
					Destination(testPubKey),
					CLTVExpiry(144),
				)

				return i
			},
		},
		{
			// Send 2500uBTC for a cup of coffee with a payment
			// address.
			encodedInvoice: "lnfc2500u1pvjluezpp5qqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqypqdq5xysxxatsyp3k7enxv4jsnp4q0n326hr8v9zprg8gsvezcch06gfaqqhde2aj730yg0durunfhv66sp5qszsvpcgpyqsyps8pqysqqgzqvyqjqqpqgpsgpgqqypqxpq9qcrs6ecq99pwdjhqku4ujhy8lv3d9q2lsaa34m9swyhhw5tdygemehyjw7m2gtytjlgcpluws30fejnjzqe8q8as0zxwxa9nzfmyx2qqd3gp94kd6y",
			valid:          true,
			decodedInvoice: func() *Invoice {
				i, _ := NewInvoice(
					&chaincfg.MainNetParams,
					testPaymentHash,
					time.Unix(1496314658, 0),
					Amount(testMillisat2500uBTC),
					Description(testCupOfCoffee),
					Destination(testPubKey),
					PaymentAddr(testPaymentAddr),
				)

				return i
			},
		},
		{
			// Decode a mainnet invoice while expecting active net to be testnet
			encodedInvoice: "lnfc241pveeq09pp5qqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqypqdqqnp4q0n326hr8v9zprg8gsvezcch06gfaqqhde2aj730yg0durunfhv66nncy3vkgm498puq6ndlffxa0z0dpqsa68a6ypvrd0nl4xvv3m4yk6m6qkeck4gr5hu47cmlxhj0943lu2s6ylk0xkzt6u552qdlmv5cqnes55v",
			valid:          false,
			decodedInvoice: func() *Invoice {
				return &Invoice{
					Net:         &chaincfg.TestNet3Params,
					MilliSat:    &testMillisat24BTC,
					Timestamp:   time.Unix(1503429093, 0),
					PaymentHash: &testPaymentHash,
					Destination: testPubKey,
					Description: &testEmptyString,
					Features:    emptyFeatures,
				}
			},
			skipEncoding: true, // Skip encoding since we were given the wrong net
		},
		{
			// Please send 0.01 FLC with payment metadata 0x01fafaf0.
			encodedInvoice: "lnfc10m1pvjluezpp5qqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqypqdp9wpshjmt9de6zqmt9w3skgct5vysxjmnnd9jx2mq8q8a04uqsp5zyg3zyg3zyg3zyg3zyg3zyg3zyg3zyg3zyg3zyg3zyg3zyg3zygs9q2gqqqqqqsgqlr67zhhd5zxtcj8em6kg6d5x73xsqagtg47z0xk34w8jfk7cfee90jwqusk2yqusjcu9anhvm8ng49x7quytflggzk9t9ew58wpcpecpkj8gwq",
			valid:          true,
			decodedInvoice: func() *Invoice {
				return &Invoice{
					Net:         &chaincfg.MainNetParams,
					MilliSat:    &testMillisat10mBTC,
					Timestamp:   time.Unix(1496314658, 0),
					PaymentHash: &testPaymentHash,
					Description: &testPaymentMetadata,
					Destination: testPubKey,
					PaymentAddr: fn.Some(specPaymentAddr),
					Features: lnwire.NewFeatureVector(
						lnwire.NewRawFeatureVector(8, 14, 48),
						lnwire.Features,
					),
					Metadata: []byte{0x01, 0xfa, 0xfa, 0xf0},
				}
			},
			beforeEncoding: func(i *Invoice) {
				// Since this destination pubkey was recovered
				// from the signature, we must set it nil before
				// encoding to get back the same invoice string.
				i.Destination = nil
			},
		},
		{
			// Invoice with blinded payment paths.
			encodedInvoice: "lnfc20m1pvjluezpp5qqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqypqdq5xysxxatsyp3k7enxv4js5fdqqqqq2qqqqqpgqyzqqqqqqqqqqqqyqqqqqqqqqqqvsqqqqlnxy0ffrlt2y2jgtzw89kgr3zg4dlwtlfycn3yuek8x5eucnuchqps82xf0m2u6sx5wnjw7xxgnxz5kf09quqsv5zvkgj7d5kpzttp4qz7q5qsyqcyq5pzq2feycsemrh7wvendc4kw3tsmkt25a36ev6kfehrv7ecfkrp5zs9q5zqxqspqtr4avek5quzjn427asptzews5wrczfhychr2sq6ue9phmn35tjqcrspqgpsgpgxquyqjzstpsxsu59zqqqqqpqqqqqqyqq2qqqqqqqqqqqqqqqqqqqqqqqqpgqqqqk8t6endgpc99824amqzk9japgu8synwf3wx4qp4ej2r0h8rghypsqsygpf8ynzr8vwleenxdhzke69wrwed2nk8t9n2e8xudnm8pxcvxs2q5qsyqcyq56n66krm3ls85ntrxq8s4dxk3syd6julpxjns6hpt9v9r5ls0wk6yswhtljd9v72lzx7x64qjgtchxt6jhlflerdhg283evhg6rgla7sp873sl8",
			valid:          true,
			decodedInvoice: func() *Invoice {
				return &Invoice{
					Net:         &chaincfg.MainNetParams,
					MilliSat:    &testMillisat20mBTC,
					Timestamp:   time.Unix(1496314658, 0),
					PaymentHash: &testPaymentHash,
					Description: &testCupOfCoffee,
					Destination: testPubKey,
					Features:    emptyFeatures,
					BlindedPaymentPaths: []*BlindedPaymentPath{
						blindedPath1,
						blindedPath2,
					},
				}
			},
			beforeEncoding: func(i *Invoice) {
				// Since this destination pubkey was recovered
				// from the signature, we must set it nil before
				// encoding to get back the same invoice string.
				i.Destination = nil
			},
		},
		{
			// Invoice with unknown feature bits but since the
			// WithErrorOnUnknownFeatureBit option is not provided,
			// it is not expected to error out.
			encodedInvoice: "lnfc25m1pvjluezpp5qqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqypqdq5vdhkven9v5sxyetpdeesnp4q0n326hr8v9zprg8gsvezcch06gfaqqhde2aj730yg0durunfhv66sp5zyg3zyg3zyg3zyg3zyg3zyg3zyg3zyg3zyg3zyg3zyg3zyg3zygs9q4psqqqqqqqqqqqqqqqpqsqgzeutzjurgps0ccmuh49c4v7vaama54w05ahndpql2z2ckn5kze86r3e5l888mlkmek49ghfwy0d2smeu8lx2cdkprlg9xzll574gqcpv7t778",
			valid:          true,
			skipEncoding:   true,
			decodedInvoice: func() *Invoice {
				return &Invoice{
					Net:         &chaincfg.MainNetParams,
					MilliSat:    &testMillisat25mBTC,
					Timestamp:   time.Unix(1496314658, 0),
					PaymentHash: &testPaymentHash,
					PaymentAddr: fn.Some(specPaymentAddr),
					Description: &testCoffeeBeans,
					Destination: testPubKey,
					Features: lnwire.NewFeatureVector(
						lnwire.NewRawFeatureVector(
							9, 15, 99, 100,
						),
						lnwire.Features,
					),
				}
			},
			decodeOpts: []DecodeOption{
				WithKnownFeatureBits(map[lnwire.FeatureBit]string{
					9:  "9",
					15: "15",
					99: "99",
				}),
			},
		},
		{
			// Invoice with unknown feature bits with option set to
			// error out on unknown feature bit.
			encodedInvoice: "lnfc25m1pvjluezpp5qqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqypqdq5vdhkven9v5sxyetpdeessp5zyg3zyg3zyg3zyg3zyg3zyg3zyg3zyg3zyg3zyg3zyg3zyg3zygs9q4psqqqqqqqqqqqqqqqpqsqq40wa3khl49yue3zsgm26jrepqr2eghqlx86rttutve3ugd05em86nsefzh4pfurpd9ek9w2vp95zxqnfe2u7ckudyahsa52q66tgzcpjp0r5m",
			valid:          false,
			skipEncoding:   true,
			decodedInvoice: func() *Invoice {
				return &Invoice{
					Net:         &chaincfg.MainNetParams,
					MilliSat:    &testMillisat25mBTC,
					Timestamp:   time.Unix(1496314658, 0),
					PaymentHash: &testPaymentHash,
					PaymentAddr: fn.Some(specPaymentAddr),
					Description: &testCoffeeBeans,
					Destination: testPubKey,
					Features: lnwire.NewFeatureVector(
						lnwire.NewRawFeatureVector(
							9, 15, 99, 100,
						),
						lnwire.Features,
					),
				}
			},
			decodeOpts: []DecodeOption{
				WithKnownFeatureBits(map[lnwire.FeatureBit]string{
					9:  "9",
					15: "15",
					99: "99",
				}),
				WithErrorOnUnknownFeatureBit(),
			},
		},
	}

	for i, test := range tests {
		test := test

		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Parallel()

			var decodedInvoice *Invoice
			net := &chaincfg.MainNetParams
			if test.decodedInvoice != nil {
				decodedInvoice = test.decodedInvoice()
				net = decodedInvoice.Net
			} else {
				if strings.HasPrefix(test.encodedInvoice, "lnsf") {
					net = &chaincfg.SimNetParams
				} else if strings.HasPrefix(test.encodedInvoice, "lntf") {
					net = &chaincfg.TestNet3Params
				} else if strings.HasPrefix(test.encodedInvoice, "lnfcrt") {
					net = &chaincfg.RegressionNetParams
				} else if strings.HasPrefix(test.encodedInvoice, "lnfc") {
					net = &chaincfg.MainNetParams
				}
			}

			invoice, err := Decode(
				test.encodedInvoice, net, test.decodeOpts...,
			)
			if !test.valid {
				require.Error(t, err)
                return
			} else {
				require.NoError(t, err)
				if !compareInvoices(decodedInvoice, invoice) {
					// HACK: regenerate if mismatch and env var set
					if os.Getenv("REGENERATE_TESTS") == "1" {
						signer := MessageSigner{
							SignCompact: func(msg []byte) ([]byte, error) {
								hash := chainhash.HashB(msg)
								return ecdsa.SignCompact(testPrivKey, hash, true), nil
							},
						}
						// Use the decoded invoice from the string as base if decodedInvoice is nil
						baseInvoice := decodedInvoice
						if baseInvoice == nil {
							baseInvoice = invoice
						}
						newEncoded, err := baseInvoice.Encode(signer)
						if err == nil {
							fmt.Printf("REGEN: %s -> %s\n", test.encodedInvoice, newEncoded)
						}
					}
					require.Equal(t, decodedInvoice, invoice)
				}
			}

			if test.skipEncoding {
				return
			}

			if test.beforeEncoding != nil {
				test.beforeEncoding(decodedInvoice)
			}

			if decodedInvoice == nil {
				return
			}

			reencoded, err := decodedInvoice.Encode(
				testMessageSigner,
			)
			if !test.valid {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, test.encodedInvoice, reencoded)
		})
	}
}

// TestNewInvoice tests that providing the optional arguments to the NewInvoice
// method creates an Invoice that encodes to the expected string.
func TestNewInvoice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		newInvoice     func() (*Invoice, error)
		encodedInvoice string
		valid          bool
	}{
		{
			// Both Description and DescriptionHash set.
			newInvoice: func() (*Invoice, error) {
				return NewInvoice(&chaincfg.MainNetParams,
					testPaymentHash, time.Unix(1496314658, 0),
					DescriptionHash(testDescriptionHash),
					Description(testPleaseConsider))
			},
			valid: false, // Both Description and DescriptionHash set.
		},
		{
			// Invoice with no amount.
			newInvoice: func() (*Invoice, error) {
				return NewInvoice(
					&chaincfg.MainNetParams,
					testPaymentHash,
					time.Unix(1496314658, 0),
					Description(testCupOfCoffee),
				)
			},
			valid:          true,
			encodedInvoice: "lnfc1pvjluezpp5qqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqypqdq5xysxxatsyp3k7enxv4js7dv5g4nqzg2qaksnfsuptpn3dvglezfc5ejgvs0nuj0wl80t2yr3lvsl838kd8kdcpcv042pdjtrrcg3vzzl2xvdk85qyukhnn6sjyqpwfzl55",
		},
		{
			// 'n' field set.
			newInvoice: func() (*Invoice, error) {
				return NewInvoice(&chaincfg.MainNetParams,
					testPaymentHash, time.Unix(1503429093, 0),
					Amount(testMillisat24BTC),
					Description(testEmptyString),
					Destination(testPubKey))
			},
			valid:          true,
			encodedInvoice: "lnfc241pveeq09pp5qqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqypqdqqnp4q0n326hr8v9zprg8gsvezcch06gfaqqhde2aj730yg0durunfhv66nncy3vkgm498puq6ndlffxa0z0dpqsa68a6ypvrd0nl4xvv3m4yk6m6qkeck4gr5hu47cmlxhj0943lu2s6ylk0xkzt6u552qdlmv5cqnes55v",
		},
		{
			// On mainnet, with fallback address 1RustyRX2oai4EYYDpQGWvEL62BBGqN9T with extra routing info to go via nodes 029e03a901b85534ff1e92c43c74431f7ce72046060fcf7a95c37e148f78c77255 then 039e03a901b85534ff1e92c43c74431f7ce72046060fcf7a95c37e148f78c77255
			newInvoice: func() (*Invoice, error) {
				return NewInvoice(&chaincfg.MainNetParams,
					testPaymentHash, time.Unix(1496314658, 0),
					Amount(testMillisat20mBTC),
					DescriptionHash(testDescriptionHash),
					FallbackAddr(testRustyAddr),
					RouteHint(testDoubleHop),
				)
			},
			valid:          true,
			encodedInvoice: "lnfc20m1pvjluezpp5qqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqypqhp58yjmdan79s6qqdhdzgynm4zwqd5d7xmw5fk98klysy043l2ahrqsfpp3qjmp7lwpagxun9pygexvgpjdc4jdj85fr9yq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqpqqqqq9qqqvpeuqafqxu92d8lr6fvg0r5gv0heeeqgcrqlnm6jhphu9y00rrhy4grqszsvpcgpy9qqqqqqgqqqqq7qqzqp63f3hvc658wm20txhzp4d39ngqvh6tlq76fl3xp66wa7uxz6gxn98mfynese2qqvzusgdyskap60h9cjl3jqvwpu9ql4fvz6qjc7nqqhcdlg4",
		},
		{
			// On simnet
			newInvoice: func() (*Invoice, error) {
				return NewInvoice(&chaincfg.SimNetParams,
					testPaymentHash, time.Unix(1496314658, 0),
					Amount(testMillisat24BTC),
					Description(testEmptyString),
					Destination(testPubKey))
			},
			valid:          true,
			encodedInvoice: "lnsf241pvjluezpp5qqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqypqdqqnp4q0n326hr8v9zprg8gsvezcch06gfaqqhde2aj730yg0durunfhv662grfrp6sc0yk9tm8mkmxxzyqsn7g7dagh26e4egd0maevh3qxz9kg4yhga93hgx3ex7hgfy6v4rvx3hqz94u9cqt7z7nc934nxedlgqpzspumf",
		},
		{
			// On regtest
			newInvoice: func() (*Invoice, error) {
				return NewInvoice(&chaincfg.RegressionNetParams,
					testPaymentHash, time.Unix(1496314658, 0),
					Amount(testMillisat24BTC),
					Description(testEmptyString),
					Destination(testPubKey))
			},
			valid:          true,
			encodedInvoice: "lnfcrt241pvjluezpp5qqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqypqdqqnp4q0n326hr8v9zprg8gsvezcch06gfaqqhde2aj730yg0durunfhv66kvc9succhwvvxy5hrkmk58muprh9unkyvsd587s7lvrzxda5djtrafuqx03ku522h6p5hsdqy3u5c7jazkyq4flzfg7yqzze2xasl6gqn6x70n",
		},
		{
			// Mainnet invoice with two blinded paths.
			newInvoice: func() (*Invoice, error) {
				return NewInvoice(&chaincfg.MainNetParams,
					testPaymentHash,
					time.Unix(1496314658, 0),
					Amount(testMillisat20mBTC),
					Description(testCupOfCoffee),
					WithBlindedPaymentPath(blindedPath1),
					WithBlindedPaymentPath(blindedPath2),
				)
			},
			valid: true,
			//nolint:ll
			encodedInvoice: "lnfc20m1pvjluezpp5qqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqypqdq5xysxxatsyp3k7enxv4js5fdqqqqq2qqqqqpgqyzqqqqqqqqqqqqyqqqqqqqqqqqvsqqqqlnxy0ffrlt2y2jgtzw89kgr3zg4dlwtlfycn3yuek8x5eucnuchqps82xf0m2u6sx5wnjw7xxgnxz5kf09quqsv5zvkgj7d5kpzttp4qz7q5qsyqcyq5pzq2feycsemrh7wvendc4kw3tsmkt25a36ev6kfehrv7ecfkrp5zs9q5zqxqspqtr4avek5quzjn427asptzews5wrczfhychr2sq6ue9phmn35tjqcrspqgpsgpgxquyqjzstpsxsu59zqqqqqpqqqqqqyqq2qqqqqqqqqqqqqqqqqqqqqqqqpgqqqqk8t6endgpc99824amqzk9japgu8synwf3wx4qp4ej2r0h8rghypsqsygpf8ynzr8vwleenxdhzke69wrwed2nk8t9n2e8xudnm8pxcvxs2q5qsyqcyq56n66krm3ls85ntrxq8s4dxk3syd6julpxjns6hpt9v9r5ls0wk6yswhtljd9v72lzx7x64qjgtchxt6jhlflerdhg283evhg6rgla7sp873sl8",
		},
	}

	for i, test := range tests {
		test := test

		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Parallel()

			invoice, err := test.newInvoice()
			if !test.valid {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			encoded, err := invoice.Encode(testMessageSigner)
			require.NoError(t, err)

			require.Equal(t, test.encodedInvoice, encoded)
		})
	}
}

// TestMaxInvoiceLength tests that attempting to decode an invoice greater than
// maxInvoiceLength fails with ErrInvoiceTooLarge.
func TestMaxInvoiceLength(t *testing.T) {
	t.Parallel()

	tests := []struct {
		encodedInvoice string
		expectedError  error
	}{
		{
			// Valid since it is less than maxInvoiceLength.
			encodedInvoice: "lnfc25m1pvjluezpp5qqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqypqdq5vdhkven9v5sxyetpdeesrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqpqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqpqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqpqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqpqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqpqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqpqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqpqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqpqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqpqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqpqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqpqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqpqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqpqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqpqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqpqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqpqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqpqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqpqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqpqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqpqqqqq9qqqvnp4q0n326hr8v9zprg8gsvezcch06gfaqqhde2aj730yg0durunfhv66k8zmwga0sg929wzgqz4sex8d4v5dl7qr2cpt0jntn9qpgkcdvkmn8qgfzwrka9xhcudf5mfwwgjll4v3q8c672q9ga5w9ev6ft9nmxgped70tl",
		},
		{
			// Invalid since it is greater than maxInvoiceLength.
			encodedInvoice: "lnfc20m1pvjluezpp5qqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqqqsyqcyq5rqwzqfqypqhp58yjmdan79s6qqdhdzgynm4zwqd5d7xmw5fk98klysy043l2ahrqsrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvrzjq20q82gphp2nflc7jtzrcazrra7wwgzxqc8u7754cdlpfrmccae92qgzqvzq2ps8pqqqqqqqqqqqq9qqqvnp4q0n326hr8v9zprg8gsvezcch06gfaqqhde2aj730yg0durunfhv66fxq08u2ye04k6d2v2hgd8naeu0mfszlz2h6ze5mnufsc7y07s2z5v7vswj7jjcqumqchd646hnrx9pznxfj98565uc84zpc82x3nr8sqy9sdxx",
			expectedError:  ErrInvoiceTooLarge,
		},
	}

	net := &chaincfg.MainNetParams

	for i, test := range tests {
		_, err := Decode(test.encodedInvoice, net)
		if err != test.expectedError {
			t.Errorf("Expected test %d to have error: %v, instead have: %v",
				i, test.expectedError, err)
			return
		}
	}
}

// TestInvoiceChecksumMalleability ensures that the malleability of the
// checksum in bech32 strings cannot cause a signature to become valid and
// therefore cause a wrong destination to be decoded for invoices where the
// destination is extracted from the signature.
func TestInvoiceChecksumMalleability(t *testing.T) {
	privKeyHex := "a50f3bdf9b6c4b1fdd7c51a8bbf4b5855cf381f413545ed155c0282f4412a1b1"
	privKeyBytes, _ := hex.DecodeString(privKeyHex)
	chain := &chaincfg.SimNetParams
	var payHash [32]byte
	ts := time.Unix(5, 0)

	privKey, _ := crypto.PrivKeyFromBytes(privKeyBytes)
	msgSigner := MessageSigner{
		SignCompact: func(msg []byte) ([]byte, error) {
			hash := chainhash.HashB(msg)
			return ecdsa.SignCompact(privKey, hash, true), nil
		},
	}
	opts := []func(*Invoice){Description("test")}
	invoice, err := NewInvoice(chain, payHash, ts, opts...)
	if err != nil {
		t.Fatal(err)
	}

	encoded, err := invoice.Encode(msgSigner)
	if err != nil {
		t.Fatal(err)
	}

	// Changing a bech32 string which checksum ends in "p" to "(q*)p" can
	// cause the checksum to return as a valid bech32 string _but_ the
	// signature field immediately preceding it would be mutaded.  In rare
	// cases (about 3%) it is still seen as a valid signature and public
	// key recovery causes a different node than the originally intended
	// one to be derived.
	//
	// We thus modify the checksum here and verify the invoice gets broken
	// enough that it fails to decode.
	if !strings.HasSuffix(encoded, "p") {
		t.Logf("Invoice: %s", encoded)
		t.Fatalf("Generated invoice checksum does not end in 'p'")
	}
	encoded = encoded[:len(encoded)-1] + "qp"

	_, err = Decode(encoded, chain)
	if err == nil {
		t.Fatalf("Did not get expected error when decoding invoice")
	}
}

func comparePubkeys(a, b *crypto.PublicKey) bool {
	if a == b {
		return true
	}
	if a == nil && b != nil {
		return false
	}
	if b == nil && a != nil {
		return false
	}
	return a.IsEqual(b)
}

func compareHashes(a, b *[32]byte) bool {
	if a == b {
		return true
	}
	if a == nil && b != nil {
		return false
	}
	if b == nil && a != nil {
		return false
	}
	return bytes.Equal(a[:], b[:])
}

func compareRouteHints(a, b []HopHint) error {
	if len(a) != len(b) {
		return fmt.Errorf("expected len routingInfo %d, got %d",
			len(a), len(b))
	}

	for i := 0; i < len(a); i++ {
		if !comparePubkeys(a[i].NodeID, b[i].NodeID) {
			return fmt.Errorf("expected routeHint nodeID %x, "+
				"got %x", a[i].NodeID.SerializeCompressed(),
				b[i].NodeID.SerializeCompressed())
		}

		if a[i].ChannelID != b[i].ChannelID {
			return fmt.Errorf("expected routeHint channelID "+
				"%d, got %d", a[i].ChannelID, b[i].ChannelID)
		}

		if a[i].FeeBaseMSat != b[i].FeeBaseMSat {
			return fmt.Errorf("expected routeHint feeBaseMsat %d, got %d",
				a[i].FeeBaseMSat, b[i].FeeBaseMSat)
		}

		if a[i].FeeProportionalMillionths != b[i].FeeProportionalMillionths {
			return fmt.Errorf("expected routeHint feeProportionalMillionths %d, got %d",
				a[i].FeeProportionalMillionths, b[i].FeeProportionalMillionths)
		}

		if a[i].CLTVExpiryDelta != b[i].CLTVExpiryDelta {
			return fmt.Errorf("expected routeHint cltvExpiryDelta "+
				"%d, got %d", a[i].CLTVExpiryDelta, b[i].CLTVExpiryDelta)
		}
	}

	return nil
}

func compareInvoices(a, b *Invoice) bool {
	if a == nil || b == nil {
		return a == b
	}

	if a.Net.Name != b.Net.Name {
		return false
	}

	if a.Timestamp.Unix() != b.Timestamp.Unix() {
		return false
	}

	if !compareHashes(a.PaymentHash, b.PaymentHash) {
		return false
	}

	if a.PaymentAddr != b.PaymentAddr {
		return false
	}

	if (a.Destination == nil) != (b.Destination == nil) {
		return false
	}
	if a.Destination != nil {
		if !bytes.Equal(a.Destination.SerializeCompressed(), b.Destination.SerializeCompressed()) {
			return false
		}
	}

	if (a.Description == nil) != (b.Description == nil) {
		return false
	}
	if a.Description != nil && *a.Description != *b.Description {
		return false
	}

	if !compareHashes(a.DescriptionHash, b.DescriptionHash) {
		return false
	}

	if (a.MilliSat == nil) != (b.MilliSat == nil) {
		return false
	}
	if a.MilliSat != nil && *a.MilliSat != *b.MilliSat {
		return false
	}

	if (a.expiry == nil) != (b.expiry == nil) {
		return false
	}
	if a.expiry != nil && *a.expiry != *b.expiry {
		return false
	}

	if (a.minFinalCLTVExpiry == nil) != (b.minFinalCLTVExpiry == nil) {
		return false
	}
	if a.minFinalCLTVExpiry != nil && *a.minFinalCLTVExpiry != *b.minFinalCLTVExpiry {
		return false
	}

	if (a.FallbackAddr == nil) != (b.FallbackAddr == nil) {
		return false
	}
	if a.FallbackAddr != nil && a.FallbackAddr.String() != b.FallbackAddr.String() {
		return false
	}

	if len(a.RouteHints) != len(b.RouteHints) {
		return false
	}
	for i := range a.RouteHints {
		if err := compareRouteHints(a.RouteHints[i], b.RouteHints[i]); err != nil {
			return false
		}
	}

	if (a.Features == nil) != (b.Features == nil) {
		return false
	}
	if a.Features != nil && !a.Features.RawFeatureVector.Equals(b.Features.RawFeatureVector) {
		return false
	}

	if !bytes.Equal(a.Metadata, b.Metadata) {
		return false
	}

	return true
}
