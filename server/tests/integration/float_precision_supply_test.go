package integration

import (
	"context"
	"math"
	"testing"

	domain "milpa/domain/entities"
)

// moneyPrecisionResidual is the tolerance used when asserting that a fully
// delivered request reserved everything it was asked for. 2.7 - 0.9 - 0.9 - 0.9
// in IEEE-754 double arithmetic leaves a residue of about 2.2e-16, so an exact
// zero is not assertable against a DOUBLE PRECISION column. The tolerance is
// many orders of magnitude below the smallest quantity that carries meaning in
// this domain and exists only to absorb that residue.
const moneyPrecisionResidual = 1e-9

// TestFullyDeliveredRequestWithFractionalAmountCompletes is the regression test
// for the money type. A request for 2.7 delivered as three transactions of 0.9
// is fully covered, but the completed amount was accumulated in float32, where
// 3 * 0.9f is 2.6999998092651367 while 2.7f is 2.700000047683716. The guard
// `completedAmount < TotalAmount` therefore stayed true forever and the request
// never left the open set even though it had been delivered in full.
//
// Every amount is float64 in Go and DOUBLE PRECISION in PostgreSQL, and 3 * 0.9
// is exactly the same double as 2.7, so the guard is false and the request is
// closed.
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

	actual := concActualAmount(t, request.ID)
	if math.Abs(actual) > moneyPrecisionResidual {
		t.Errorf("actual amount = %v, want ~%v (the request was reserved in full)", actual, moneyPrecisionResidual)
	}

	if got := concQueryCount(t, `SELECT count(*) FROM supply_requests WHERE id = $1 AND status = $2`,
		request.ID, domain.SupplyRequestOpen); got != 0 {
		t.Errorf("the fully delivered request is still open: open rows = %d, want 0", got)
	}
}
