package bitcoindnotify

import (
	"errors"
	"fmt"

	"github.com/flokiorg/flnd/blockcache"
	"github.com/flokiorg/flnd/chainntnfs"
	"github.com/flokiorg/go-flokicoin/chaincfg"
	"github.com/flokiorg/walletd/chain"
)

// createNewNotifier creates a new instance of the ChainNotifier interface
// implemented by BitcoindNotifier.
func createNewNotifier(args ...interface{}) (chainntnfs.ChainNotifier, error) {
	if len(args) != 5 {
		return nil, fmt.Errorf("incorrect number of arguments to "+
			".New(...), expected 5, instead passed %v", len(args))
	}

	chainConn, ok := args[0].(*chain.FlokicoindConn)
	if !ok {
		return nil, errors.New("first argument to bitcoindnotify.New " +
			"is incorrect, expected a *chain.FlokicoindConn")
	}

	chainParams, ok := args[1].(*chaincfg.Params)
	if !ok {
		return nil, errors.New("second argument to bitcoindnotify.New " +
			"is incorrect, expected a *chaincfg.Params")
	}

	spendHintCache, ok := args[2].(chainntnfs.SpendHintCache)
	if !ok {
		return nil, errors.New("third argument to bitcoindnotify.New " +
			"is incorrect, expected a chainntnfs.SpendHintCache")
	}

	confirmHintCache, ok := args[3].(chainntnfs.ConfirmHintCache)
	if !ok {
		return nil, errors.New("fourth argument to bitcoindnotify.New " +
			"is incorrect, expected a chainntnfs.ConfirmHintCache")
	}

	blockCache, ok := args[4].(*blockcache.BlockCache)
	if !ok {
		return nil, errors.New("fifth argument to bitcoindnotify.New " +
			"is incorrect, expected a *blockcache.BlockCache")
	}

	return New(chainConn, chainParams, spendHintCache,
		confirmHintCache, blockCache), nil
}

// init registers a driver for the BtcdNotifier concrete implementation of the
// chainntnfs.ChainNotifier interface.
func init() {
	// Register the driver.
	notifierZMQ := &chainntnfs.NotifierDriver{
		NotifierType: notifierTypeZMQ,
		New:          createNewNotifier,
	}
	if err := chainntnfs.RegisterNotifier(notifierZMQ); err != nil {
		panic(fmt.Sprintf("failed to register notifier driver '%s': %v",
			notifierTypeZMQ, err))
	}

	notifierRPC := &chainntnfs.NotifierDriver{
		NotifierType: notifierTypeRPCPolling,
		New:          createNewNotifier,
	}
	if err := chainntnfs.RegisterNotifier(notifierRPC); err != nil {
		panic(fmt.Sprintf("failed to register notifier driver '%s': %v",
			notifierTypeRPCPolling, err))
	}
}
