package build

import "testing"

// TestParseVersion covers the single place the reported version now comes
// from, including the build-metadata and pre-release forms the release tooling
// produces.
func TestParseVersion(t *testing.T) {
	tests := []struct {
		name       string
		in         string
		major      uint
		minor      uint
		patch      uint
		preRelease string
	}{
		{name: "release", in: "0.2.3", major: 0, minor: 2, patch: 3},
		{name: "pre-release", in: "0.2.2-beta", major: 0, minor: 2, patch: 2, preRelease: "beta"},
		{name: "dotted pre-release", in: "0.3.0-rc.1", major: 0, minor: 3, patch: 0, preRelease: "rc.1"},
		{name: "build metadata", in: "0.2.3+a1b2c3d", major: 0, minor: 2, patch: 3},
		{name: "pre-release and build metadata", in: "0.2.3-beta+a1b2c3d", major: 0, minor: 2, patch: 3, preRelease: "beta"},
		{name: "development placeholder", in: devVersion, preRelease: "dev"},
		{name: "too few components", in: "0.2"},
		{name: "not numeric", in: "notaversion"},
		{name: "empty", in: ""},
		{
			// semanticAlphabet excludes '/', so it is dropped rather
			// than reported. This is what previously made a
			// hand-written label safe to report, and it is why
			// nothing needs to panic over a malformed ldflag.
			name: "pre-release is normalized", in: "0.2.3-be/ta",
			major: 0, minor: 2, patch: 3, preRelease: "beta",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			major, minor, patch, preRelease := parseVersion(test.in)
			if major != test.major || minor != test.minor ||
				patch != test.patch {

				t.Errorf("parseVersion(%q) components = "+
					"%d.%d.%d, want %d.%d.%d", test.in,
					major, minor, patch, test.major,
					test.minor, test.patch)
			}
			if preRelease != test.preRelease {
				t.Errorf("parseVersion(%q) preRelease = %q, "+
					"want %q", test.in, preRelease,
					test.preRelease)
			}
		})
	}
}

// TestVersionReportsInjectedValue asserts that an injected version is what
// gets reported, which is the whole point of injecting it.
func TestVersionReportsInjectedValue(t *testing.T) {
	defer func(v string) { AppVersion = v }(AppVersion)

	AppVersion = "0.2.3"
	if got, want := Version(), "0.2.3"; got != want {
		t.Errorf("Version() = %q, want %q", got, want)
	}

	AppVersion = ""
	if got, want := Version(), devVersion; got != want {
		t.Errorf("Version() with nothing injected = %q, want %q",
			got, want)
	}
}

// TestComponentsDerivedFromInjectedVersion checks the components the version
// RPCs and `flncli version` report individually are the ones parsed out of the
// injected version. itest/lnd_rest_api_test.go additionally asserts the
// reported minor version is greater than zero, so any build reaching an itest
// must have a version injected -- the Makefile does that via APP_VERSION.
func TestComponentsDerivedFromInjectedVersion(t *testing.T) {
	major, minor, patch, preRelease := parseVersion("0.2.3")
	if major != 0 || minor != 2 || patch != 3 || preRelease != "" {
		t.Fatalf("parseVersion(\"0.2.3\") = %d.%d.%d-%q, want 0.2.3",
			major, minor, patch, preRelease)
	}

	if minor == 0 {
		t.Error("a released minor version of 0 would break " +
			"itest/lnd_rest_api_test.go's AppMinor > 0 assertion")
	}
}
