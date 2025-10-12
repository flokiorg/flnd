package flnwallet

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/flokiorg/flnd"
	"github.com/flokiorg/flnd/lncfg"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
}

func createConfig(t *testing.T) flnd.Config {
	conf := flnd.DefaultConfig()
	conf.LndDir = createTestTempDir(t)
	conf.Bitcoin.TestNet3 = true
	conf.Bitcoin.Node = flnd.NeutrinoBackendName
	conf.NeutrinoMode.ConnectPeers = append(conf.NeutrinoMode.ConnectPeers, "lab.in.ionance.com:35212")
	conf.DebugLevel = "debug"
	conf.ProtocolOptions = &lncfg.ProtocolOptions{}
	conf.Pprof = &lncfg.Pprof{}
	conf.LogConfig.Console.Disable = true

	return conf
}

func TestDaemonConnection(t *testing.T) {

	for i := 0; i < 2; i++ {
		t.Run(fmt.Sprintf("run-%d", i), func(t *testing.T) {

			conf := createConfig(t)

			d := newDaemon(context.Background(), &conf)
			if _, err := d.start(); err != nil {
				t.Fatal(err)
			}

			go func() {
				d.stop()
			}()

			d.waitForShutdown()
		})
	}

}

func createTestTempDir(t *testing.T) string {
	t.Helper()

	// tempDir, err := os.MkdirTemp("", "test-temp-*")
	// if err != nil {
	// 	t.Fatalf("failed to create temp dir: %v", err)
	// }

	tempDir := "/u/flzpace/xgit/repos/flokiorg/twallet/test/t0"

	// t.Cleanup(func() {
	// 	if err := os.RemoveAll(tempDir); err != nil {
	// 		t.Fatalf("Failed to remove temp dir: %v", err)
	// 	}
	// })

	return tempDir
}
