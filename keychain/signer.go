package keychain

import (
	"github.com/flokiorg/go-flokicoin/crypto"

	"github.com/flokiorg/go-flokicoin/chaincfg/chainhash"
	"github.com/flokiorg/go-flokicoin/crypto/ecdsa"
)

func NewPubKeyMessageSigner(pubKey *crypto.PublicKey, keyLoc KeyLocator,
	signer MessageSignerRing) *PubKeyMessageSigner {

	return &PubKeyMessageSigner{
		pubKey:       pubKey,
		keyLoc:       keyLoc,
		digestSigner: signer,
	}
}

type PubKeyMessageSigner struct {
	pubKey       *crypto.PublicKey
	keyLoc       KeyLocator
	digestSigner MessageSignerRing
}

func (p *PubKeyMessageSigner) PubKey() *crypto.PublicKey {
	return p.pubKey
}

func (p *PubKeyMessageSigner) KeyLocator() KeyLocator {
	return p.keyLoc
}

func (p *PubKeyMessageSigner) SignMessage(message []byte,
	doubleHash bool) (*ecdsa.Signature, error) {

	return p.digestSigner.SignMessage(p.keyLoc, message, doubleHash)
}

func (p *PubKeyMessageSigner) SignMessageCompact(msg []byte,
	doubleHash bool) ([]byte, error) {

	return p.digestSigner.SignMessageCompact(p.keyLoc, msg, doubleHash)
}

func NewPrivKeyMessageSigner(privKey *crypto.PrivateKey,
	keyLoc KeyLocator) *PrivKeyMessageSigner {

	return &PrivKeyMessageSigner{
		privKey: privKey,
		keyLoc:  keyLoc,
	}
}

type PrivKeyMessageSigner struct {
	keyLoc  KeyLocator
	privKey *crypto.PrivateKey
}

func (p *PrivKeyMessageSigner) PubKey() *crypto.PublicKey {
	return p.privKey.PubKey()
}

func (p *PrivKeyMessageSigner) KeyLocator() KeyLocator {
	return p.keyLoc
}

func (p *PrivKeyMessageSigner) SignMessage(msg []byte,
	doubleHash bool) (*ecdsa.Signature, error) {

	var digest []byte
	if doubleHash {
		digest = chainhash.DoubleHashB(msg)
	} else {
		digest = chainhash.HashB(msg)
	}
	return ecdsa.Sign(p.privKey, digest), nil
}

func (p *PrivKeyMessageSigner) SignMessageCompact(msg []byte,
	doubleHash bool) ([]byte, error) {

	var digest []byte
	if doubleHash {
		digest = chainhash.DoubleHashB(msg)
	} else {
		digest = chainhash.HashB(msg)
	}

	return ecdsa.SignCompact(p.privKey, digest, true), nil
}

var _ SingleKeyMessageSigner = (*PubKeyMessageSigner)(nil)
var _ SingleKeyMessageSigner = (*PrivKeyMessageSigner)(nil)
