package mock

import (
	"github.com/flokiorg/go-flokicoin/crypto"

	"github.com/flokiorg/flnd/keychain"
	"github.com/flokiorg/go-flokicoin/chaincfg/chainhash"
	"github.com/flokiorg/go-flokicoin/crypto/ecdsa"
	"github.com/flokiorg/go-flokicoin/crypto/schnorr"
	"github.com/flokiorg/go-flokicoin/txscript"
)

// SecretKeyRing is a mock implementation of the SecretKeyRing interface.
type SecretKeyRing struct {
	RootKey *crypto.PrivateKey
}

// DeriveNextKey currently returns dummy values.
func (s *SecretKeyRing) DeriveNextKey(
	_ keychain.KeyFamily) (keychain.KeyDescriptor, error) {

	return keychain.KeyDescriptor{
		PubKey: s.RootKey.PubKey(),
	}, nil
}

// DeriveKey currently returns dummy values.
func (s *SecretKeyRing) DeriveKey(
	_ keychain.KeyLocator) (keychain.KeyDescriptor, error) {

	return keychain.KeyDescriptor{
		PubKey: s.RootKey.PubKey(),
	}, nil
}

// DerivePrivKey currently returns dummy values.
func (s *SecretKeyRing) DerivePrivKey(
	_ keychain.KeyDescriptor) (*crypto.PrivateKey, error) {

	return s.RootKey, nil
}

// ECDH currently returns dummy values.
func (s *SecretKeyRing) ECDH(_ keychain.KeyDescriptor,
	_ *crypto.PublicKey) ([32]byte, error) {

	return [32]byte{}, nil
}

// SignMessage signs the passed message and ignores the KeyDescriptor.
func (s *SecretKeyRing) SignMessage(_ keychain.KeyLocator,
	msg []byte, doubleHash bool) (*ecdsa.Signature, error) {

	var digest []byte
	if doubleHash {
		digest = chainhash.DoubleHashB(msg)
	} else {
		digest = chainhash.HashB(msg)
	}
	return ecdsa.Sign(s.RootKey, digest), nil
}

// SignMessageCompact signs the passed message.
func (s *SecretKeyRing) SignMessageCompact(_ keychain.KeyLocator,
	msg []byte, doubleHash bool) ([]byte, error) {

	var digest []byte
	if doubleHash {
		digest = chainhash.DoubleHashB(msg)
	} else {
		digest = chainhash.HashB(msg)
	}

	return ecdsa.SignCompact(s.RootKey, digest, true), nil
}

// SignMessageSchnorr signs the passed message and ignores the KeyDescriptor.
func (s *SecretKeyRing) SignMessageSchnorr(_ keychain.KeyLocator,
	msg []byte, doubleHash bool, taprootTweak []byte,
	tag []byte) (*schnorr.Signature, error) {

	var digest []byte
	switch {
	case len(tag) > 0:
		taggedHash := chainhash.TaggedHash(tag, msg)
		digest = taggedHash[:]
	case doubleHash:
		digest = chainhash.DoubleHashB(msg)
	default:
		digest = chainhash.HashB(msg)
	}

	privKey := s.RootKey
	if len(taprootTweak) > 0 {
		privKey = txscript.TweakTaprootPrivKey(*privKey, taprootTweak)
	}

	return schnorr.Sign(privKey, digest)
}
