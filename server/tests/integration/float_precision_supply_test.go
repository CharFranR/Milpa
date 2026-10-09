package integration

import (
	"context"
	"math"
	"testing"

	domain "milpa/domain/entities"
)

func TestFullyDeliveredRequestWithFractionalAmountCompletes(t *testing.T) {
	setupConcurrencyTestData(t)
	f := newConcFixture(t)

	const (
		totalAmount   = 2.7
		eachDelivery  = 0.9
		deliveries    = 3
		supplierCount = 3
	)

	if supplierCount < deliveries {
		t.Fatalf("the fixture provides %d suppliers, need %d", supplierCount, deliveries)
	}

	request := f.newConcRequest(t, totalAmount, true)
	if got := concActualAmount(t, request.ID); got != totalAmount {
		t.Fatalf("actual amount before any delivery = %v, want %v", got, totalAmount)
	}

	suppliers := concSupplierIDs[:deliveries]
	for _, supplierID := range suppliers {
		offer := f.newConcOffer(t, supplierID, request.ID, eachDelivery)

		if _, _, err := f.matchUC.Like(concBuyerCtx(), offer.ID); err != nil {
			t.Fatalf("Like() error for supplier %s: %v", supplierID, err)
		}
		_, transactionID := concTransactionForOffer(t, offer.ID)

		if err := f.transactionUC.ConfirmStart(concBuyerCtx(), transactionID); err != nil {
			t.Fatalf("ConfirmStart(buyer) error: %v", err)
		}
		if err := f.transactionUC.ConfirmStart(concSupplierCtx(supplierID), transactionID); err != nil {
			t.Fatalf("ConfirmStart(supplier %s) error: %v", supplierID, err)
		}
		if err := f.transactionUC.ConfirmDelivery(concBuyerCtx(), transactionID); err != nil {
			t.Fatalf("ConfirmDelivery(buyer) error: %v", err)
		}
		if err := f.transactionUC.ConfirmDelivery(concSupplierCtx(supplierID), transactionID); err != nil {
			t.Fatalf("ConfirmDelivery(supplier %s) error: %v", supplierID, err)
		}
	}

	if got := concQueryCount(t,
		`SELECT count(*) FROM transactions WHERE status = $1`, domain.TransactionCompleted); got != deliveries {
		t.Errorf("completed transactions = %d, want %d", got, deliveries)
	}

	var status domain.SupplyRequestStatus
	if err := TestPool.QueryRow(context.Background(),
		`SELECT status FROM supply_requests WHERE id = $1`, request.ID).Scan(&status); err != nil {
		t.Fatalf("read the persisted request status: %v", err)
	}
	if status != domain.SupplyRequestCompleted {
		t.Fatalf("request status = %v, want completed: the %d completed deliveries of %v cover the total of %v exactly, so the request must leave the open set (actual amount = %v)",
			status, deliveries, eachDelivery, totalAmount, concActualAmount(t, request.ID))
	}

	// Exact equality, not a tolerance: this is the assertion that distinguishes
	// "completion zeroed the column" from "completion left 2.220446049250313e-16
	// behind and happened to look close enough".
	actual := concActualAmount(t, request.ID)
	if actual != 0 {
		t.Errorf("actual amount after completion = %v (bits 0x%016x), want exactly 0: a completed request has nothing left to fulfil, "+
			"so completion must SET actual_amount rather than leave the residue of three fractional subtractions",
			actual, math.Float64bits(actual))
	}

	if got := concQueryCount(t, `SELECT count(*) FROM supply_requests WHERE id = $1 AND status = $2`,
		request.ID, domain.SupplyRequestOpen); got != 0 {
		t.Errorf("the fully delivered request is still open: open rows = %d, want 0", got)
	}
}
