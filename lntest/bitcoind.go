//go:build bitcoind && !notxindex && !rpcpolling
// +build bitcoind,!notxindex,!rpcpolling

package lntest

import (
	"github.com/flokiorg/go-flokicoin/chaincfg"
)

// NewBackend starts a bitcoind node with the txindex enabled and returns a
// FlokicoindBackendConfig for that node.
func NewBackend(miner string, netParams *chaincfg.Params) (
	*FlokicoindBackendConfig, func() error, error) {

	extraArgs := []string{
		"-regtest",
		"-txindex",
		"-disablewallet",
	}

	return newBackend(miner, netParams, extraArgs, false)
}
