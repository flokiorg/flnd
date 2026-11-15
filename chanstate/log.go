package chanstate

import (
	"github.com/flokiorg/flnd/build"
	flog "github.com/flokiorg/go-flokicoin/log/v2"
)

// log is a logger that is initialized with no output filters.  This means the
// package will not perform any logging by default until the caller requests
// it.
//
//nolint:unused
var log flog.Logger

// init initializes the package-global logger instance.
func init() {
	UseLogger(build.NewSubLogger("CHST", nil))
}

// DisableLog disables all library log output.  Logging output is disabled by
// default until UseLogger is called.
func DisableLog() {
	UseLogger(flog.Disabled)
}

// UseLogger uses a specified Logger to output package logging info. This
// should be used in preference to SetLogWriter if the caller is also using
// btclog.
func UseLogger(logger flog.Logger) {
	log = logger
}
