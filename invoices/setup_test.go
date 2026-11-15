package invoices

import (
	"testing"

	"github.com/flokiorg/flnd/kvdb"
)

func TestMain(m *testing.M) {
	kvdb.RunTests(m)
}
