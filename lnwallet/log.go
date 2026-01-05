package lnwallet

import (
	"github.com/flokiorg/flnd/build"
	"github.com/flokiorg/flnd/lnwallet/chainfee"
	flog "github.com/flokiorg/go-flokicoin/log/v2"
)

// walletLog is a logger that is initialized with no output filters.  This
// means the package will not perform any logging by default until the caller
// requests it.
var walletLog flog.Logger

// The default amount of logging is none.
func init() {
	UseLogger(build.NewSubLogger("LNWL", nil))
}

// DisableLog disables all library log output.  Logging output is disabled
// by default until UseLogger is called.
func DisableLog() {
	UseLogger(flog.Disabled)
}

// UseLogger uses a specified Logger to output package logging info.
// This should be used in preference to SetLogWriter if the caller is also
// using flog.
func UseLogger(logger flog.Logger) {
	walletLog = logger

	chainfee.UseLogger(logger)
}
