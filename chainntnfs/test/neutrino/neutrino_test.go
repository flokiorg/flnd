//go:build dev
// +build dev

package neutrino_test

import (
	"testing"

	chainntnfstest "github.com/flokiorg/flnd/chainntnfs/test"
)

// TestInterfaces executes the generic notifier test suite against a neutrino
// powered chain notifier.
func TestInterfaces(t *testing.T) {
	chainntnfstest.TestInterfaces(t, "neutrino")
}
