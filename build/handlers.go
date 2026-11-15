package build

import (
	"os"

	flog "github.com/flokiorg/go-flokicoin/log/v2"
)

// NewDefaultLogHandlers returns the standard console logger and rotating log
// writer handlers that we generally want to use. It also applies the various
// config options to the loggers.
func NewDefaultLogHandlers(cfg *LogConfig,
	rotator *RotatingLogWriter) []flog.Handler {

	var handlers []flog.Handler

	consoleLogHandler := flog.NewDefaultHandler(
		os.Stdout, cfg.Console.HandlerOptions()...,
	)
	logFileHandler := flog.NewDefaultHandler(
		rotator, cfg.File.HandlerOptions()...,
	)

	maybeAddLogger := func(cmdOptionDisable bool, handler flog.Handler) {
		if !cmdOptionDisable {
			handlers = append(handlers, handler)
		}
	}
	switch LoggingType {
	case LogTypeStdOut:
		maybeAddLogger(cfg.Console.Disable, consoleLogHandler)
	case LogTypeDefault:
		maybeAddLogger(cfg.Console.Disable, consoleLogHandler)
		maybeAddLogger(cfg.File.Disable, logFileHandler)
	}

	return handlers
}
