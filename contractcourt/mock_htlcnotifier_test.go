package contractcourt

import (
	"github.com/flokiorg/flnd/channeldb"
	"github.com/flokiorg/flnd/graph/db/models"
)

type mockHTLCNotifier struct {
	HtlcNotifier
}

func (m *mockHTLCNotifier) NotifyFinalHtlcEvent(key models.CircuitKey,
	info channeldb.FinalHtlcInfo) {

}
