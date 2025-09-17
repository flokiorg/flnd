package channeldb

import (
	"github.com/flokiorg/flnd/build"
	mig "github.com/flokiorg/flnd/channeldb/migration"
	"github.com/flokiorg/flnd/channeldb/migration12"
	"github.com/flokiorg/flnd/channeldb/migration13"
	"github.com/flokiorg/flnd/channeldb/migration16"
	"github.com/flokiorg/flnd/channeldb/migration24"
	"github.com/flokiorg/flnd/channeldb/migration30"
	"github.com/flokiorg/flnd/channeldb/migration31"
	"github.com/flokiorg/flnd/channeldb/migration32"
	"github.com/flokiorg/flnd/channeldb/migration33"
	"github.com/flokiorg/flnd/channeldb/migration_01_to_11"
	"github.com/flokiorg/flnd/kvdb"
	flog "github.com/flokiorg/go-flokicoin/log/v2"
)

// log is a logger that is initialized with no output filters.  This
// means the package will not perform any logging by default until the caller
// requests it.
var log flog.Logger

func init() {
	UseLogger(build.NewSubLogger("CHDB", nil))
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
	mig.UseLogger(logger)
	migration_01_to_11.UseLogger(logger)
	migration12.UseLogger(logger)
	migration13.UseLogger(logger)
	migration16.UseLogger(logger)
	migration24.UseLogger(logger)
	migration30.UseLogger(logger)
	migration31.UseLogger(logger)
	migration32.UseLogger(logger)
	migration33.UseLogger(logger)
	kvdb.UseLogger(logger)
}
