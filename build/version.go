// Copyright (c) 2013-2017 The btcsuite developers
// Copyright (c) 2015-2016 The Decred developers
// Heavily inspired by https://github.com/btcsuite/btcd/blob/master/version.go
// Copyright (C) 2015-2022 The Lightning Network Developers

package build

import (
	"context"
	"encoding/hex"
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"

	flog "github.com/flokiorg/go-flokicoin/log/v2"
)

var (
	// Commit stores the current commit of this build, which includes the
	// most recent tag, the number of commits since that tag (if non-zero),
	// the commit hash, and a dirty marker. This should be set using the
	// -ldflags during compilation.
	Commit string

	// CommitHash stores the current commit hash of this build.
	CommitHash string

	// RawTags contains the raw set of build tags, separated by commas.
	RawTags string

	// GoVersion stores the go version that the executable was compiled
	// with.
	GoVersion string
)

// semanticAlphabet is the set of characters that are permitted for use in an
// AppPreRelease.
const semanticAlphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-."

// AppVersion is the application version, set at build time with
// '-ldflags "-X github.com/flokiorg/flnd/build.AppVersion=0.2.3"' from the
// release tag. It is the only place the version is recorded, so what this
// binary reports cannot drift from what was actually published. A plain
// `go build` leaves it empty and devVersion is reported instead; builds made
// through the Makefile get the most recent tag injected.
var AppVersion string

// devVersion is reported by builds that had no version injected. It is
// deliberately not a real version number, so a development build is never
// mistaken for a release.
const devVersion = "0.0.0-dev"

// AppMajor, AppMinor, AppPatch and AppPreRelease are the components of the
// effective version. They are derived from AppVersion rather than maintained
// by hand, and are exported because the version RPCs and `flncli version`
// report them individually.
var AppMajor, AppMinor, AppPatch, AppPreRelease = parseVersion(effectiveVersion())

// effectiveVersion returns the injected version, or the development
// placeholder when nothing was injected.
func effectiveVersion() string {
	if AppVersion != "" {
		return AppVersion
	}

	return devVersion
}

// parseVersion splits a semantic version string into its numeric components
// and pre-release label, discarding any build metadata. The pre-release label
// is stripped of characters outside semanticAlphabet, which is what previously
// made a hand-written label safe to report. A string that does not parse
// yields zeroed components, which is preferable to reporting numbers that were
// never released.
func parseVersion(v string) (major, minor, patch uint, preRelease string) {
	if i := strings.IndexByte(v, '+'); i >= 0 {
		v = v[:i]
	}
	if i := strings.IndexByte(v, '-'); i >= 0 {
		preRelease, v = normalizeVerString(v[i+1:]), v[:i]
	}

	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return 0, 0, 0, preRelease
	}

	var nums [3]uint
	for i, part := range parts {
		n, err := strconv.ParseUint(part, 10, 32)
		if err != nil {
			return 0, 0, 0, preRelease
		}
		nums[i] = uint(n)
	}

	return nums[0], nums[1], nums[2], preRelease
}

// normalizeVerString returns the passed string stripped of all characters
// which are not valid according to the semantic versioning guidelines for
// pre-release version and build metadata strings.
func normalizeVerString(str string) string {
	var result strings.Builder
	for _, r := range str {
		if strings.ContainsRune(semanticAlphabet, r) {
			result.WriteRune(r)
		}
	}

	return result.String()
}

func init() {
	// Get build information from the runtime. AppPreRelease no longer needs
	// asserting here: it is derived from AppVersion and normalized by
	// parseVersion, so a malformed ldflag can no longer panic at startup.
	if info, ok := debug.ReadBuildInfo(); ok {
		GoVersion = info.GoVersion
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				CommitHash = setting.Value

			case "-tags":
				RawTags = setting.Value
			}
		}
	}
}

// Version returns the application version as a properly formed string per the
// semantic versioning 2.0.0 spec (http://semver.org/).
func Version() string {
	return effectiveVersion()
}

// Tags returns the list of build tags that were compiled into the executable.
func Tags() []string {
	if len(RawTags) == 0 {
		return nil
	}

	return strings.Split(RawTags, ",")
}

// WithBuildInfo derives a child context with the build information attached as
// attributes. At the moment, this only includes the current build's commit
// hash.
func WithBuildInfo(ctx context.Context, cfg *LogConfig) (context.Context,
	error) {

	if cfg.NoCommitHash {
		return ctx, nil
	}

	// Convert the commit hash to a byte slice.
	commitHash, err := hex.DecodeString(CommitHash)
	if err != nil {
		return nil, fmt.Errorf("unable to decode commit hash: %w", err)
	}

	return flog.WithCtx(ctx, flog.Hex3("rev", commitHash)), nil
}
