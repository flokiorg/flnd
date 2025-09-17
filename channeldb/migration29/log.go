package migration29

import flog "github.com/flokiorg/go-flokicoin/log/v2"

// log is a logger that is initialized as disabled. This means the package will
// not perform any logging by default until a logger is set.
var log = flog.Disabled

// UseLogger uses a specific Logger to output package logging info.
func UseLogger(logger flog.Logger) {
	log = logger
}
