package chainreg

import (
	"github.com/flokiorg/flnd/keychain"
	"github.com/flokiorg/go-flokicoin/chaincfg"
	"github.com/flokiorg/go-flokicoin/wire"
)

// FlokicoinNetParams couples the p2p parameters of a network with the
// corresponding RPC port of a daemon running on the particular network.
type FlokicoinNetParams struct {
	*chaincfg.Params
	RPCPort  string
	CoinType uint32
}

// FlokicoinTestNetParams contains parameters specific to the 3rd version of the
// test network.
var FlokicoinTestNetParams = FlokicoinNetParams{
	Params:   &chaincfg.TestNet3Params,
	RPCPort:  "35213",
	CoinType: keychain.CoinTypeTestnet,
}

// FlokicoinTestNet4Params contains parameters specific to the 4th version of the
// test network.
var FlokicoinTestNet4Params = FlokicoinNetParams{
	Params:   &chaincfg.TestNet4Params,
	RPCPort:  "65213",
	CoinType: keychain.CoinTypeTestnet,
}

// FlokicoinMainNetParams contains parameters specific to the current Flokicoin
// mainnet.
var FlokicoinMainNetParams = FlokicoinNetParams{
	Params:   &chaincfg.MainNetParams,
	RPCPort:  "15213",
	CoinType: keychain.CoinTypeFlokicoin,
}

// FlokicoinSimNetParams contains parameters specific to the simulation test
// network.
var FlokicoinSimNetParams = FlokicoinNetParams{
	Params:   &chaincfg.SimNetParams,
	RPCPort:  "45213",
	CoinType: keychain.CoinTypeTestnet,
}

// FlokicoinSigNetParams contains parameters specific to the signet test network.
var FlokicoinSigNetParams = FlokicoinNetParams{
	Params:   &chaincfg.SigNetParams,
	RPCPort:  "55213",
	CoinType: keychain.CoinTypeTestnet,
}

// FlokicoinRegTestNetParams contains parameters specific to a local bitcoin
// regtest network.
var FlokicoinRegTestNetParams = FlokicoinNetParams{
	Params:   &chaincfg.RegressionNetParams,
	RPCPort:  "25213",
	CoinType: keychain.CoinTypeTestnet,
}

// IsTestnet tests if the given params correspond to a testnet parameter
// configuration.
func IsTestnet(params *FlokicoinNetParams) bool {
	return params.Params.Net == wire.TestNet3 ||
		params.Params.Net == wire.TestNet4
}
