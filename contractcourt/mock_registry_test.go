package contractcourt

import (
	"context"

	"github.com/flokiorg/flnd/graph/db/models"
	"github.com/flokiorg/flnd/invoices"
	"github.com/flokiorg/flnd/lntypes"
	"github.com/flokiorg/flnd/lnwire"
)

type notifyExitHopData struct {
	payHash       lntypes.Hash
	paidAmount    lnwire.MilliLoki
	hodlChan      chan<- interface{}
	expiry        uint32
	currentHeight int32
}

type mockRegistry struct {
	notifyChan       chan notifyExitHopData
	notifyErr        error
	notifyResolution invoices.HtlcResolution
}

func (r *mockRegistry) NotifyExitHopHtlc(payHash lntypes.Hash,
	paidAmount lnwire.MilliLoki, expiry uint32, currentHeight int32,
	circuitKey models.CircuitKey, hodlChan chan<- interface{},
	wireCustomRecords lnwire.CustomRecords,
	payload invoices.Payload) (invoices.HtlcResolution, error) {

	// Exit early if the notification channel is nil.
	if hodlChan == nil {
		return r.notifyResolution, r.notifyErr
	}

	r.notifyChan <- notifyExitHopData{
		hodlChan:      hodlChan,
		payHash:       payHash,
		paidAmount:    paidAmount,
		expiry:        expiry,
		currentHeight: currentHeight,
	}

	return r.notifyResolution, r.notifyErr
}

func (r *mockRegistry) HodlUnsubscribeAll(subscriber chan<- interface{}) {}

func (r *mockRegistry) LookupInvoice(context.Context, lntypes.Hash) (
	invoices.Invoice, error) {

	return invoices.Invoice{}, invoices.ErrInvoiceNotFound
}
