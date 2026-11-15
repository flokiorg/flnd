package bitcoind_test

import (
	"testing"

	lnwallettest "github.com/flokiorg/flnd/lnwallet/test"
)

// TestLightningWalletFlokicoindZMQ tests LightningWallet powered by bitcoind,
// using its ZMQ interface, against our suite of interface tests.
func TestLightningWalletFlokicoindZMQ(t *testing.T) {
	lnwallettest.TestLightningWallet(t, "bitcoind")
}

// TestLightningWalletFlokicoindRPCPolling tests LightningWallet powered by
// bitcoind, using its RPC interface, against our suite of interface tests.
func TestLightningWalletFlokicoindRPCPolling(t *testing.T) {
	lnwallettest.TestLightningWallet(t, "bitcoind-rpc-polling")
}
