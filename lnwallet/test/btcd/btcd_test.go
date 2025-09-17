package btcd_test

import (
	"testing"

	lnwallettest "github.com/flokiorg/flnd/lnwallet/test"
)

// TestLightningWallet tests LightningWallet powered by btcd against our suite
// of interface tests.
func TestLightningWallet(t *testing.T) {
	lnwallettest.TestLightningWallet(t, "btcd")
}
