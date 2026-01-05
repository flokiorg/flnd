package cluster

import (
	"github.com/flokiorg/flnd/build"
	flog "github.com/flokiorg/go-flokicoin/log/v2"
)

// Subsystem defines the logging code for this subsystem.
const Subsystem = "CLUS"

// log is a logger that is initialized with the flog.Disabled logger.
var log flog.Logger

// The default amount of logging is none.
func init() {
	UseLogger(build.NewSubLogger(Subsystem, nil))
}

// DisableLog disables all logging output.
func DisableLog() {
	UseLogger(flog.Disabled)
}

// UseLogger uses a specified Logger to output package logging info.
func UseLogger(logger flog.Logger) {
	log = logger
}
