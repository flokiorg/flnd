package bolt12

import (
	"bytes"

	"github.com/flokiorg/go-flokicoin/crypto"
)

// bobKey returns the deterministic spec test key for Bob, whose 32-byte scalar
// is 0x42 repeated. Used across signature and round-trip tests so the same key
// is not reconstructed in every callsite.
func bobKey() (*crypto.PrivateKey, *crypto.PublicKey) {
	priv, pub := crypto.PrivKeyFromBytes(bytes.Repeat([]byte{0x42}, 32))

	return priv, pub
}

// aliceKey returns the deterministic spec test key for Alice, whose 32-byte
// scalar is 0x41 repeated.
func aliceKey() (*crypto.PrivateKey, *crypto.PublicKey) {
	priv, pub := crypto.PrivKeyFromBytes(bytes.Repeat([]byte{0x41}, 32))

	return priv, pub
}
