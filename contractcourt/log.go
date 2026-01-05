package contractcourt

import (
	"github.com/flokiorg/flnd/build"
	flog "github.com/flokiorg/go-flokicoin/log/v2"
)

var (
	// log is a logger that is initialized with no output filters.  This
	// means the package will not perform any logging by default until the caller
	// requests it.
	log flog.Logger

	// brarLog is the logger used by the breach arb.
	brarLog flog.Logger

	// utxnLog is the logger used by the utxo nursary.
	utxnLog flog.Logger
)

// The default amount of logging is none.
func init() {
	UseLogger(build.NewSubLogger("CNCT", nil))
	UseBreachLogger(build.NewSubLogger("BRAR", nil))
	UseNurseryLogger(build.NewSubLogger("UTXN", nil))
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
	log = logger
}

// UseBreachLogger uses a specified Logger to output package logging info.
// This should be used in preference to SetLogWriter if the caller is also
// using flog.
func UseBreachLogger(logger flog.Logger) {
	brarLog = logger
}

// UseNurseryLogger uses a specified Logger to output package logging info.
// This should be used in preference to SetLogWriter if the caller is also
// using flog.
func UseNurseryLogger(logger flog.Logger) {
	utxnLog = logger
}
