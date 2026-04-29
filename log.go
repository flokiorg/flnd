package flnd

import (
	"github.com/flokiorg/flnd/autopilot"
	"github.com/flokiorg/flnd/build"
	"github.com/flokiorg/flnd/chainio"
	"github.com/flokiorg/flnd/chainntnfs"
	"github.com/flokiorg/flnd/chainreg"
	"github.com/flokiorg/flnd/chanacceptor"
	"github.com/flokiorg/flnd/chanbackup"
	"github.com/flokiorg/flnd/chanfitness"
	"github.com/flokiorg/flnd/channeldb"
	"github.com/flokiorg/flnd/channelnotifier"
	"github.com/flokiorg/flnd/cluster"
	"github.com/flokiorg/flnd/contractcourt"
	"github.com/flokiorg/flnd/discovery"
	"github.com/flokiorg/flnd/funding"
	"github.com/flokiorg/flnd/graph"
	graphdb "github.com/flokiorg/flnd/graph/db"
	"github.com/flokiorg/flnd/healthcheck"
	"github.com/flokiorg/flnd/htlcswitch"
	"github.com/flokiorg/flnd/invoices"
	"github.com/flokiorg/flnd/kvdb/sqlbase"
	"github.com/flokiorg/flnd/lncfg"
	"github.com/flokiorg/flnd/lnrpc/autopilotrpc"
	"github.com/flokiorg/flnd/lnrpc/chainrpc"
	"github.com/flokiorg/flnd/lnrpc/devrpc"
	"github.com/flokiorg/flnd/lnrpc/invoicesrpc"
	"github.com/flokiorg/flnd/lnrpc/neutrinorpc"
	"github.com/flokiorg/flnd/lnrpc/peersrpc"
	"github.com/flokiorg/flnd/lnrpc/routerrpc"
	"github.com/flokiorg/flnd/lnrpc/signrpc"
	"github.com/flokiorg/flnd/lnrpc/verrpc"
	"github.com/flokiorg/flnd/lnrpc/walletrpc"
	"github.com/flokiorg/flnd/lnwallet"
	"github.com/flokiorg/flnd/lnwallet/btcwallet"
	"github.com/flokiorg/flnd/lnwallet/chancloser"
	"github.com/flokiorg/flnd/lnwallet/chanfunding"
	"github.com/flokiorg/flnd/lnwallet/rpcwallet"
	"github.com/flokiorg/flnd/monitoring"
	"github.com/flokiorg/flnd/msgmux"
	"github.com/flokiorg/flnd/netann"
	"github.com/flokiorg/flnd/onionmessage"
	paymentsdb "github.com/flokiorg/flnd/payments/db"
	"github.com/flokiorg/flnd/peer"
	"github.com/flokiorg/flnd/peernotifier"
	"github.com/flokiorg/flnd/protofsm"
	"github.com/flokiorg/flnd/routing"
	"github.com/flokiorg/flnd/routing/blindedpath"
	"github.com/flokiorg/flnd/routing/localchans"
	"github.com/flokiorg/flnd/rpcperms"
	"github.com/flokiorg/flnd/signal"
	"github.com/flokiorg/flnd/sqldb"
	"github.com/flokiorg/flnd/sweep"
	"github.com/flokiorg/flnd/tor"
	"github.com/flokiorg/flnd/watchtower"
	"github.com/flokiorg/flnd/watchtower/wtclient"
	neutrino "github.com/flokiorg/flokicoin-neutrino"
	"github.com/flokiorg/go-flokicoin/connmgr"
	btclogv1 "github.com/flokiorg/go-flokicoin/log"
	flog "github.com/flokiorg/go-flokicoin/log/v2"
	"github.com/flokiorg/go-flokicoin/rpcclient"
)

// replaceableLogger is a thin wrapper around a logger that is used so the
// logger can be replaced easily without some black pointer magic.
type replaceableLogger struct {
	flog.Logger
	subsystem string
}

// Loggers can not be used before the log rotator has been initialized with a
// log file. This must be performed early during application startup by
// calling InitLogRotator() on the main log writer instance in the config.
var (
	// lndPkgLoggers is a list of all lnd package level loggers that are
	// registered. They are tracked here so they can be replaced once the
	// SetupLoggers function is called with the final root logger.
	lndPkgLoggers []*replaceableLogger

	// addLndPkgLogger is a helper function that creates a new replaceable
	// main lnd package level logger and adds it to the list of loggers that
	// are replaced again later, once the final root logger is ready.
	addLndPkgLogger = func(subsystem string) *replaceableLogger {
		l := &replaceableLogger{
			Logger:    build.NewSubLogger(subsystem, nil),
			subsystem: subsystem,
		}
		lndPkgLoggers = append(lndPkgLoggers, l)
		return l
	}

	// Loggers that need to be accessible from the lnd package can be placed
	// here. Loggers that are only used in sub modules can be added directly
	// by using the addSubLogger method. We declare all loggers so we never
	// run into a nil reference if they are used early. But the SetupLoggers
	// function should always be called as soon as possible to finish
	// setting them up properly with a root logger.
	ltndLog = addLndPkgLogger("LTND")
	rpcsLog = addLndPkgLogger("RPCS")
	srvrLog = addLndPkgLogger("SRVR")
	atplLog = addLndPkgLogger("ATPL")
	acsmLog = addLndPkgLogger("ACSM")
)

// genSubLogger creates a logger for a subsystem. We provide an instance of
// a signal.Interceptor to be able to shutdown in the case of a critical error.
func genSubLogger(root *build.SubLoggerManager,
	interceptor signal.Interceptor) func(string) flog.Logger {

	// Create a shutdown function which will request shutdown from our
	// interceptor if it is listening.
	shutdown := func() {
		if !interceptor.Listening() {
			return
		}

		interceptor.RequestShutdown()
	}

	// Return a function which will create a sublogger from our root
	// logger without shutdown fn.
	return func(tag string) flog.Logger {
		return root.GenSubLogger(tag, shutdown)
	}
}

// SetupLoggers initializes all package-global logger variables.
//
//nolint:ll
func SetupLoggers(root *build.SubLoggerManager, interceptor signal.Interceptor) {
	genLogger := genSubLogger(root, interceptor)

	// Now that we have the proper root logger, we can replace the
	// placeholder lnd package loggers.
	for _, l := range lndPkgLoggers {
		l.Logger = build.NewSubLogger(l.subsystem, genLogger)
		SetSubLogger(root, l.subsystem, l.Logger)
	}

	// Initialize loggers from packages outside of `lnd` first. The
	// packages below will overwrite the names of the loggers they import.
	// For instance, the logger in `neutrino.query` is overwritten by
	// `btcwallet.chain`, which is overwritten by `lnwallet`. To ensure the
	// overwriting works, we need to initialize the loggers here so they
	// can be overwritten later.
	AddV1SubLogger(root, "FLCN", interceptor, neutrino.UseLogger)
	AddV1SubLogger(root, "CMGR", interceptor, connmgr.UseLogger)
	AddV1SubLogger(root, "RPCC", interceptor, rpcclient.UseLogger)

	// Some of the loggers declared in the main lnd package are also used
	// in sub packages.
	signal.UseLogger(ltndLog)
	autopilot.UseLogger(atplLog)

	AddSubLogger(root, "LNWL", interceptor, lnwallet.UseLogger)
	AddSubLogger(root, "DISC", interceptor, discovery.UseLogger)
	AddSubLogger(root, "NTFN", interceptor, chainntnfs.UseLogger)
	AddSubLogger(root, "CHDB", interceptor, channeldb.UseLogger)
	AddSubLogger(root, "SQLB", interceptor, sqlbase.UseLogger)
	AddSubLogger(root, "HSWC", interceptor, htlcswitch.UseLogger)
	AddSubLogger(root, "CNCT", interceptor, contractcourt.UseLogger)
	AddSubLogger(root, "UTXN", interceptor, contractcourt.UseNurseryLogger)
	AddSubLogger(root, "BRAR", interceptor, contractcourt.UseBreachLogger)
	// AddV1SubLogger(root, "SPHX", interceptor, sphinx.UseLogger) // #FLZ_CHANGE
	AddSubLogger(root, "SWPR", interceptor, sweep.UseLogger)
	AddSubLogger(root, "SGNR", interceptor, signrpc.UseLogger)
	AddSubLogger(root, "WLKT", interceptor, walletrpc.UseLogger)
	AddSubLogger(root, "ARPC", interceptor, autopilotrpc.UseLogger)
	AddSubLogger(root, "NRPC", interceptor, neutrinorpc.UseLogger)
	AddSubLogger(root, "DRPC", interceptor, devrpc.UseLogger)
	AddSubLogger(root, "INVC", interceptor, invoices.UseLogger)
	AddSubLogger(root, "NANN", interceptor, netann.UseLogger)
	AddSubLogger(root, "WTWR", interceptor, watchtower.UseLogger)
	AddSubLogger(root, "NTFR", interceptor, chainrpc.UseLogger)
	AddSubLogger(root, "IRPC", interceptor, invoicesrpc.UseLogger)
	AddSubLogger(root, "CHNF", interceptor, channelnotifier.UseLogger)
	AddSubLogger(root, "CHBU", interceptor, chanbackup.UseLogger)
	AddSubLogger(root, "PROM", interceptor, monitoring.UseLogger)
	AddSubLogger(root, "WTCL", interceptor, wtclient.UseLogger)
	AddSubLogger(root, "PRNF", interceptor, peernotifier.UseLogger)
	AddSubLogger(root, "CHFD", interceptor, chanfunding.UseLogger)
	AddSubLogger(root, "PEER", interceptor, peer.UseLogger)
	AddSubLogger(root, "CHCL", interceptor, chancloser.UseLogger)
	AddSubLogger(root, "LCHN", interceptor, localchans.UseLogger)
	AddSubLogger(root, "PFSM", interceptor, protofsm.UseLogger)

	AddSubLogger(root, routing.Subsystem, interceptor, routing.UseLogger)
	AddSubLogger(root, routerrpc.Subsystem, interceptor, routerrpc.UseLogger)
	AddSubLogger(root, chanfitness.Subsystem, interceptor, chanfitness.UseLogger)
	AddSubLogger(root, verrpc.Subsystem, interceptor, verrpc.UseLogger)
	AddSubLogger(root, healthcheck.Subsystem, interceptor, healthcheck.UseLogger)
	AddSubLogger(root, chainreg.Subsystem, interceptor, chainreg.UseLogger)
	AddSubLogger(root, chanacceptor.Subsystem, interceptor, chanacceptor.UseLogger)
	AddSubLogger(root, funding.Subsystem, interceptor, funding.UseLogger)
	AddSubLogger(root, cluster.Subsystem, interceptor, cluster.UseLogger)
	AddSubLogger(root, rpcperms.Subsystem, interceptor, rpcperms.UseLogger)
	AddSubLogger(root, tor.Subsystem, interceptor, tor.UseLogger)
	AddSubLogger(root, btcwallet.Subsystem, interceptor, btcwallet.UseLogger)
	AddSubLogger(root, rpcwallet.Subsystem, interceptor, rpcwallet.UseLogger)
	AddSubLogger(root, peersrpc.Subsystem, interceptor, peersrpc.UseLogger)
	AddSubLogger(root, graph.Subsystem, interceptor, graph.UseLogger)
	AddSubLogger(root, lncfg.Subsystem, interceptor, lncfg.UseLogger)
	AddSubLogger(
		root, blindedpath.Subsystem, interceptor, blindedpath.UseLogger,
	)
	AddSubLogger(root, graphdb.Subsystem, interceptor, graphdb.UseLogger)
	AddSubLogger(root, chainio.Subsystem, interceptor, chainio.UseLogger)
	AddSubLogger(root, msgmux.Subsystem, interceptor, msgmux.UseLogger)
	AddSubLogger(root, sqldb.Subsystem, interceptor, sqldb.UseLogger)
	AddSubLogger(
		root, paymentsdb.Subsystem, interceptor, paymentsdb.UseLogger,
	)

	AddSubLogger(root, onionmessage.Subsystem, interceptor, onionmessage.UseLogger)
}

// AddSubLogger is a helper method to conveniently create and register the
// logger of one or more sub systems.
func AddSubLogger(root *build.SubLoggerManager, subsystem string,
	interceptor signal.Interceptor, useLoggers ...func(flog.Logger)) {

	// genSubLogger will return a callback for creating a logger instance,
	// which we will give to the root logger.
	genLogger := genSubLogger(root, interceptor)

	// Create and register just a single logger to prevent them from
	// overwriting each other internally.
	logger := build.NewSubLogger(subsystem, genLogger)
	SetSubLogger(root, subsystem, logger, useLoggers...)
}

// SetSubLogger is a helper method to conveniently register the logger of a
// sub system.
func SetSubLogger(root *build.SubLoggerManager, subsystem string,
	logger flog.Logger, useLoggers ...func(flog.Logger)) {

	root.RegisterSubLogger(subsystem, logger)
	for _, useLogger := range useLoggers {
		useLogger(logger)
	}
}

// AddV1SubLogger is a helper method to conveniently create and register the
// logger of one or more sub systems.
func AddV1SubLogger(root *build.SubLoggerManager, subsystem string,
	interceptor signal.Interceptor, useLoggers ...func(btclogv1.Logger)) {

	// genSubLogger will return a callback for creating a logger instance,
	// which we will give to the root logger.
	genLogger := genSubLogger(root, interceptor)

	// Create and register just a single logger to prevent them from
	// overwriting each other internally.
	logger := build.NewSubLogger(subsystem, genLogger)
	SetV1SubLogger(root, subsystem, logger, useLoggers...)
}

// SetV1SubLogger is a helper method to conveniently register the logger of a
// sub system. Note that the btclog v2 logger implements the btclog v1 logger
// which is why we can pass the v2 logger to the UseLogger call-backs that
// expect the v1 logger.
func SetV1SubLogger(root *build.SubLoggerManager, subsystem string,
	logger flog.Logger, useLoggers ...func(btclogv1.Logger)) {

	root.RegisterSubLogger(subsystem, logger)
	for _, useLogger := range useLoggers {
		useLogger(logger)
	}
}
