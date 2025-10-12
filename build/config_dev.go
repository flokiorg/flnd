//go:build dev
// +build dev

package build

import (
	"fmt"
	"strings"

	flogv1 "github.com/flokiorg/go-flokicoin/log"
	flog "github.com/flokiorg/go-flokicoin/log/v2"
)

const (
	resetSeq = "0"
	boldSeq  = "1"
	faintSeq = "2"
	esc      = '\x1b'
	csi      = string(esc) + "["
)

// consoleLoggerCfg extends the LoggerConfig struct by adding a Color option
// which is only available for a console logger.
//
//nolint:ll
type consoleLoggerCfg struct {
	*LoggerConfig `yaml:",inline"`
	Style         bool `long:"style" description:"If set, the output will be styled with color and fonts"`
}

// defaultConsoleLoggerCfg returns the default consoleLoggerCfg for the dev
// console logger.
func defaultConsoleLoggerCfg() *consoleLoggerCfg {
	return &consoleLoggerCfg{
		LoggerConfig: &LoggerConfig{
			CallSite: callSiteShort,
		},
	}
}

// HandlerOptions returns the set of flog.HandlerOptions that the state of the
// config struct translates to.
func (cfg *consoleLoggerCfg) HandlerOptions() []flog.HandlerOption {
	opts := cfg.LoggerConfig.HandlerOptions()

	if !cfg.Style {
		return opts
	}

	return append(
		opts, flog.WithStyledLevel(
			func(l flogv1.Level) string {
				return styleString(
					fmt.Sprintf("[%s]", l),
					boldSeq,
					string(ansiColoSeq(l)),
				)
			},
		),
		flog.WithStyledCallSite(
			func(file string, line int) string {
				str := fmt.Sprintf("%s:%d", file, line)

				return styleString(str, faintSeq)
			},
		),
		flog.WithStyledKeys(func(key string) string {
			return styleString(key, faintSeq)
		}),
	)
}

func styleString(s string, styles ...string) string {
	if len(styles) == 0 {
		return s
	}

	seq := strings.Join(styles, ";")
	if seq == "" {
		return s
	}

	return fmt.Sprintf("%s%sm%s%sm", csi, seq, s, csi+resetSeq)
}

type ansiColorSeq string

const (
	ansiColorSeqDarkTeal  ansiColorSeq = "38;5;30"
	ansiColorSeqDarkBlue  ansiColorSeq = "38;5;63"
	ansiColorSeqLightBlue ansiColorSeq = "38;5;86"
	ansiColorSeqYellow    ansiColorSeq = "38;5;192"
	ansiColorSeqRed       ansiColorSeq = "38;5;204"
	ansiColorSeqPink      ansiColorSeq = "38;5;134"
)

func ansiColoSeq(l flogv1.Level) ansiColorSeq {
	switch l {
	case flog.LevelTrace:
		return ansiColorSeqDarkTeal
	case flog.LevelDebug:
		return ansiColorSeqDarkBlue
	case flog.LevelWarn:
		return ansiColorSeqYellow
	case flog.LevelError:
		return ansiColorSeqRed
	case flog.LevelCritical:
		return ansiColorSeqPink
	default:
		return ansiColorSeqLightBlue
	}
}
