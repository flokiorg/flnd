package keychain

import (
	"testing"

	"github.com/flokiorg/go-flokicoin/crypto"

	"github.com/stretchr/testify/require"
)

func BenchmarkDerivePrivKey(t *testing.B) {
	wallet, err := createTestBtcWallet(t, CoinTypeBitcoin)
	require.NoError(t, err, "unable to create wallet")

	keyRing := NewBtcWalletKeyRing(wallet, CoinTypeBitcoin)

	var (
		privKey *crypto.PrivateKey
	)

	keyDesc := KeyDescriptor{
		KeyLocator: KeyLocator{
			Family: KeyFamilyMultiSig,
			Index:  1,
		},
	}

	t.ReportAllocs()
	t.ResetTimer()

	for i := 0; i < t.N; i++ {
		privKey, err = keyRing.DerivePrivKey(keyDesc)
	}
	require.NoError(t, err)
	require.NotNil(t, privKey)
}
