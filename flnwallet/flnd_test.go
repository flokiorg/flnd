package flnwallet

import (
	"testing"

	"github.com/flokiorg/flnd"
	"github.com/flokiorg/flnd/lncfg"
	"github.com/flokiorg/flnd/signal"
	"github.com/flokiorg/go-flokicoin/chaincfg"
)

func TestMain(t *testing.T) {

	walletdir := "/u/flzpace/xgit/repos/flokiorg/twallet/test/t0"
	network := &chaincfg.TestNet3Params

	interceptor, err := signal.Intercept()
	if err != nil {
		t.Fatal(err)
	}

	conf := flnd.DefaultConfig()
	conf.LndDir = walletdir
	conf.Bitcoin.Node = flnd.NeutrinoBackendName
	conf.NeutrinoMode.ConnectPeers = append(conf.NeutrinoMode.ConnectPeers, "lab.in.ionance.com:35212")
	conf.DebugLevel = "debug"
	conf.ProtocolOptions = &lncfg.ProtocolOptions{}
	conf.Pprof = &lncfg.Pprof{}
	conf.LogConfig.Console.Disable = true
	switch network {
	case &chaincfg.MainNetParams:
		conf.Bitcoin.MainNet = true
	case &chaincfg.TestNet3Params:
		conf.Bitcoin.TestNet3 = true
	case &chaincfg.TestNet4Params:
		conf.Bitcoin.TestNet4 = true
	case &chaincfg.SimNetParams:
		conf.Bitcoin.SigNet = true
	case &chaincfg.RegressionNetParams:
		conf.Bitcoin.RegTest = true
	case &chaincfg.SigNetParams:
		conf.Bitcoin.SigNet = true
	}

	config, err := flnd.ValidateConfig(conf, interceptor, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	impl := config.ImplementationConfig(interceptor)
	flndStarted := make(chan struct{}, 1)
	errCh := make(chan error)

	go func() {
		if err := flnd.Main(config, flnd.ListenerCfg{}, impl, interceptor, flndStarted); err != nil {
			select {
			case errCh <- err:
			default:
			}
		}
	}()

	select {
	case err := <-errCh:
		t.Fatal(err)
	case <-flndStarted:
		t.Logf("started")
	}

}
