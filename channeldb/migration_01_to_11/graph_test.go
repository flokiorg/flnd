package migration_01_to_11

import (
	"encoding/hex"
	"image/color"
	prand "math/rand"
	"net"
	"time"

	lnwire "github.com/flokiorg/flnd/channeldb/migration/lnwire21"
	"github.com/flokiorg/go-flokicoin/crypto"
	"github.com/flokiorg/go-flokicoin/crypto/ecdsa"
)

var (
	testAddr = &net.TCPAddr{IP: (net.IP)([]byte{0xA, 0x0, 0x0, 0x1}),
		Port: 9000}
	anotherAddr, _ = net.ResolveTCPAddr("tcp",
		"[2001:db8:85a3:0:0:8a2e:370:7334]:80")
	testAddrs = []net.Addr{testAddr, anotherAddr}

	testRBytes, _ = hex.DecodeString("8ce2bc69281ce27da07e6683571319d18e949ddfa2965fb6caa1bf0314f882d7")
	testSBytes, _ = hex.DecodeString("299105481d63e0f4bc2a88121167221b6700d72a0ead154c03be696a292d24ae")
	testRScalar   = new(crypto.ModNScalar)
	testSScalar   = new(crypto.ModNScalar)
	_             = testRScalar.SetByteSlice(testRBytes)
	_             = testSScalar.SetByteSlice(testSBytes)
	testSig       = ecdsa.NewSignature(testRScalar, testSScalar)

	testFeatures = lnwire.NewFeatureVector(nil, nil)
)

func createLightningNode(db *DB, priv *crypto.PrivateKey) (*LightningNode, error) {
	updateTime := prand.Int63()

	pub := priv.PubKey().SerializeCompressed()
	n := &LightningNode{
		HaveNodeAnnouncement: true,
		AuthSigBytes:         testSig.Serialize(),
		LastUpdate:           time.Unix(updateTime, 0),
		Color:                color.RGBA{1, 2, 3, 0},
		Alias:                "kek" + string(pub[:]),
		Features:             testFeatures,
		Addresses:            testAddrs,
		db:                   db,
	}
	copy(n.PubKeyBytes[:], priv.PubKey().SerializeCompressed())

	return n, nil
}

func createTestVertex(db *DB) (*LightningNode, error) {
	priv, err := crypto.NewPrivateKey()
	if err != nil {
		return nil, err
	}

	return createLightningNode(db, priv)
}
