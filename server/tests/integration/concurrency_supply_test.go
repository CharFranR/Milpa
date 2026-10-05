package integration

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"milpa/aplication/use-cases"
	domain "milpa/domain/entities"
	"milpa/infrastructure/adapters/secondary/repository"
	timepkg "milpa/infrastructure/adapters/secondary/time"
	"milpa/internal/auth"
)

var (
	concBuyerID     = uuid.MustParse("cc000000-0000-4000-8000-000000000001")
	concSupplier1   = uuid.MustParse("cc000000-0000-4000-8000-000000000002")
	concSupplier2   = uuid.MustParse("cc000000-0000-4000-8000-000000000003")
	concSupplier3   = uuid.MustParse("cc000000-0000-4000-8000-000000000004")
	concSupplierIDs = []uuid.UUID{concSupplier1, concSupplier2, concSupplier3}
)

type concFixture struct {
	requestRepo   *repository.SupplyRequestRepositoryImpl
	offerRepo     *repository.SupplyOfferRepositoryImpl
	matchRepo     *repository.MatchRepositoryImpl
	transactionRe *repository.TransactionRepositoryImpl
	matchUC       *usecases.MatchUseCaseImpl
	transactionUC *usecases.TransactionUseCaseImpl
	clock         *timepkg.Clock
}

// setupConcurrencyTestData truncates every table and creates one buyer plus
// three suppliers. Concurrency is only meaningful against a clean schema, and
// no test in this package runs in parallel.
func setupConcurrencyTestData(t *testing.T) {
	t.Helper()
	cleanupTables(t)

	ctx := context.Background()
	userRepo := repository.NewUserRepository(TestPool)

	users := []*domain.User{
		{ID: concBuyerID, FirstName: "Conc", LastName: "Buyer", Role: domain.RoleCompradorMinorista,
			Email: "conc-buyer@example.com", PhoneNumber: "4500-0001", PasswordHash: "hash",
			CreatedAt: fixedTime, UpdatedAt: fixedTime},
		{ID: concSupplier1, FirstName: "Conc", LastName: "SupplierOne", Role: domain.RoleAgricultor,
			Email: "conc-supplier-1@example.com", PhoneNumber: "4500-0002", PasswordHash: "hash",
			CreatedAt: fixedTime, UpdatedAt: fixedTime},
		{ID: concSupplier2, FirstName: "Conc", LastName: "SupplierTwo", Role: domain.RoleAgricultor,
			Email: "conc-supplier-2@example.com", PhoneNumber: "4500-0003", PasswordHash: "hash",
			CreatedAt: fixedTime, UpdatedAt: fixedTime},
		{ID: concSupplier3, FirstName: "Conc", LastName: "SupplierThree", Role: domain.RoleAgricultor,
			Email: "conc-supplier-3@example.com", PhoneNumber: "4500-0004", PasswordHash: "hash",
			CreatedAt: fixedTime, UpdatedAt: fixedTime},
	}

	for _, u := range users {
		if _, err := userRepo.Save(ctx, u); err != nil {
			t.Fatalf("insert fixture user %s: %v", u.Email, err)
		}
	}
}

func newConcFixture(t *testing.T) *concFixture {
	t.Helper()

	clock := timepkg.NewClock()
	requestRepo := repository.NewSupplyRequestRepository(TestPool)
	offerRepo := repository.NewSupplyOfferRepository(TestPool)
	matchRepo := repository.NewMatchRepository(TestPool)
	transactionRepo := repository.NewTransactionRepository(TestPool)
	inventoryRepo := repository.NewSupplierInventoryRepository(TestPool)
	userRepo := repository.NewUserRepository(TestPool)
	unitOfWork := repository.NewUnitOfWork(TestPool)

	recommendationUC := usecases.NewRecommendationUseCase(offerRepo, requestRepo, userRepo, inventoryRepo, matchRepo, nil)

	return &concFixture{
		requestRepo:   requestRepo,
		offerRepo:     offerRepo,
		matchRepo:     matchRepo,
		transactionRe: transactionRepo,
		matchUC:       usecases.NewMatchUseCase(requestRepo, offerRepo, matchRepo, transactionRepo, recommendationUC, unitOfWork),
		transactionUC: usecases.NewTransactionUseCase(transactionRepo, matchRepo, requestRepo, offerRepo, clock, unitOfWork),
		clock:         clock,
	}
}

func concBuyerCtx() context.Context {
	return auth.WithPrincipal(context.Background(), auth.Principal{UserID: concBuyerID, Role: domain.RoleCompradorMinorista})
}

func concSupplierCtx(supplierID uuid.UUID) context.Context {
	return auth.WithPrincipal(context.Background(), auth.Principal{UserID: supplierID, Role: domain.RoleAgricultor})
}

// newConcRequest persists an open request whose actual_amount starts at
// total_amount, exactly as domain.NewSupplyRequest leaves it.
func (f *concFixture) newConcRequest(t *testing.T, totalAmount float64, multipleProviders bool) *domain.SupplyRequest {
	t.Helper()

	request := domain.NewSupplyRequest(
		concBuyerID, "Maize", totalAmount, domain.Kg, 10, totalAmount/10, domain.Kg,
		domain.Address{}, fixedTime.Add(24*time.Hour), fixedTime.Add(48*time.Hour),
		"concurrency fixture", multipleProviders,
	)
	if err := f.requestRepo.Create(context.Background(), request); err != nil {
		t.Fatalf("insert fixture request: %v", err)
	}
	return request
}

func (f *concFixture) newConcOffer(t *testing.T, supplierID, requestID uuid.UUID, amount float64) *domain.SupplyOffer {
	t.Helper()

	offer := domain.NewSupplyOffer(supplierID, requestID, amount, domain.Kg, fixedTime.Add(72*time.Hour), true)
	if err := f.offerRepo.Create(context.Background(), offer); err != nil {
		t.Fatalf("insert fixture offer: %v", err)
	}
	return offer
}

// concRunConcurrently releases every worker at the same instant. Callers must
// assert on invariants, never on which worker won: the interleaving is the
// point of the test, the outcome of the race is not.
func concRunConcurrently(workers int, fn func(worker int) error) []error {
	start := make(chan struct{})
	errs := make([]error, workers)

	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func(worker int) {
			defer wg.Done()
			<-start
			errs[worker] = fn(worker)
		}(i)
	}

	close(start)
	wg.Wait()

	return errs
}

func concCountSucceeded(errs []error) int {
	succeeded := 0
	for _, err := range errs {
		if err == nil {
			succeeded++
		}
	}
	return succeeded
}

func concQueryCount(t *testing.T, query string, args ...any) int {
	t.Helper()

	var count int
	if err := TestPool.QueryRow(context.Background(), query, args...).Scan(&count); err != nil {
		t.Fatalf("count query %q: %v", query, err)
	}
	return count
}

func concQueryAmount(t *testing.T, query string, args ...any) float64 {
	t.Helper()

	var amount float64
	if err := TestPool.QueryRow(context.Background(), query, args...).Scan(&amount); err != nil {
		t.Fatalf("amount query %q: %v", query, err)
	}
	return float64(amount)
}

func concActiveMatchedAmount(t *testing.T, requestID uuid.UUID) float64 {
	t.Helper()

	return concQueryAmount(t,
		`SELECT COALESCE(SUM(matched_amount), 0) FROM matches WHERE supply_request_id = $1 AND status = $2`,
		requestID, domain.MatchActive)
}

func concActualAmount(t *testing.T, requestID uuid.UUID) float64 {
	t.Helper()

	return concQueryAmount(t, `SELECT actual_amount FROM supply_requests WHERE id = $1`, requestID)
}

func concTransactionForOffer(t *testing.T, offerID uuid.UUID) (matchID uuid.UUID, transactionID uuid.UUID) {
	t.Helper()

	err := TestPool.QueryRow(context.Background(),
		`SELECT m.id, t.id FROM matches m JOIN transactions t ON t.match_id = m.id WHERE m.supply_offer_id = $1`,
		offerID).Scan(&matchID, &transactionID)
	if err != nil {
		t.Fatalf("resolve match and transaction for offer %s: %v", offerID, err)
	}
	return matchID, transactionID
}

// concWaitForBlockedQuery blocks until some backend is waiting on a row lock
// while running a query containing fragment. It is a real barrier: it observes
// the database instead of sleeping and hoping the goroutine got scheduled.
func concWaitForBlockedQuery(t *testing.T, fragment string) {
	t.Helper()

	ctx := context.Background()
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		var waiting int
		err := TestPool.QueryRow(ctx,
			`SELECT count(*) FROM pg_stat_activity
			 WHERE state = 'active' AND wait_event_type = 'Lock' AND query LIKE '%' || $1 || '%'`,
			fragment).Scan(&waiting)
		if err != nil {
			t.Fatalf("inspect pg_stat_activity: %v", err)
		}
		if waiting > 0 {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}

	t.Fatalf("timed out waiting for a query blocked on a lock containing %q", fragment)
}

// concAssertConnsReturnTo waits for the pool to hand every connection back. A
// transaction that is not rolled back would keep its connection and its row
// locks forever, so this is the observable that proves the rollback ran.
//
// baseline MUST be sampled while the pool is quiescent, before the caller takes
// a connection of its own (before a blocker transaction is opened, before any
// polling starts). A baseline that already counts a connection the test itself
// released again can never be reached again, and the assertion then fails for a
// correct implementation or passes only when an unrelated query happens to hold
// a connection at the sample instant.
func concAssertConnsReturnTo(t *testing.T, baseline int32) {
	t.Helper()

	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		if TestPool.Stat().AcquiredConns() == baseline {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}

	t.Fatalf("acquired connections = %d, want baseline %d", TestPool.Stat().AcquiredConns(), baseline)
}

// TestLikeConcurrentSameRequestMultiProviderDoesNotOverReserve proves that two
// likes racing on the same request cannot both reserve. The request row lock
// serializes them, and the loser is rejected by the state machine because the
// winner already drained the request.
func TestLikeConcurrentSameRequestMultiProviderDoesNotOverReserve(t *testing.T) {
	setupConcurrencyTestData(t)
	f := newConcFixture(t)

	request := f.newConcRequest(t, 100, true)
	offerA := f.newConcOffer(t, concSupplier1, request.ID, 100)
	offerB := f.newConcOffer(t, concSupplier2, request.ID, 100)
	offers := []*domain.SupplyOffer{offerA, offerB}

	errs := concRunConcurrently(len(offers), func(worker int) error {
		_, _, err := f.matchUC.Like(concBuyerCtx(), offers[worker].ID)
		return err
	})

	if succeeded := concCountSucceeded(errs); succeeded != 1 {
		t.Fatalf("successful likes = %d, want exactly 1 (errors: %v)", succeeded, errs)
	}

	activeSum := concActiveMatchedAmount(t, request.ID)
	if activeSum > 100 {
		t.Errorf("active matched amount = %v, want <= 100", activeSum)
	}

	actual := concActualAmount(t, request.ID)
	if want := float64(100) - activeSum; actual != want {
		t.Errorf("actual amount = %v, want %v (total - active matched)", actual, want)
	}

	if got := concQueryCount(t, `SELECT count(*) FROM matches WHERE supply_request_id = $1`, request.ID); got != 1 {
		t.Errorf("match rows = %d, want 1: the losing like must leave no match behind", got)
	}
	if got := concQueryCount(t, `SELECT count(*) FROM transactions`); got != 1 {
		t.Errorf("transaction rows = %d, want 1: the losing like must leave no transaction behind", got)
	}
}

// TestLikeConcurrentSingleProviderAdmitsExactlyOne proves the single-provider
// gate is evaluated under the request lock. Without that lock all three workers
// would read "no active match" and admit all three.
func TestLikeConcurrentSingleProviderAdmitsExactlyOne(t *testing.T) {
	setupConcurrencyTestData(t)
	f := newConcFixture(t)

	request := f.newConcRequest(t, 100, false)
	offers := make([]*domain.SupplyOffer, 0, 3)
	for i, supplierID := range concSupplierIDs {
		offers = append(offers, f.newConcOffer(t, supplierID, request.ID, 100+float64(i)))
	}

	errs := concRunConcurrently(len(offers), func(worker int) error {
		_, _, err := f.matchUC.Like(concBuyerCtx(), offers[worker].ID)
		return err
	})

	if succeeded := concCountSucceeded(errs); succeeded != 1 {
		t.Fatalf("successful likes = %d, want exactly 1 (errors: %v)", succeeded, errs)
	}

	if got := concQueryCount(t,
		`SELECT count(*) FROM matches WHERE supply_request_id = $1 AND status = $2`,
		request.ID, domain.MatchActive); got != 1 {
		t.Errorf("active matches = %d, want 1", got)
	}
	if got := concQueryCount(t, `SELECT count(*) FROM transactions`); got != 1 {
		t.Errorf("transaction rows = %d, want 1", got)
	}
}

// TestLikeCancelThenLikeSameOfferSucceeds is the regression test for the
// UNIQUE(supply_offer_id) contradiction: an offer that was matched and then had
// its match cancelled must be matchable again. The original table-level UNIQUE
// made that impossible, which is why migration 000021 replaces it with a partial
// unique index over active matches only.
func TestLikeCancelThenLikeSameOfferSucceeds(t *testing.T) {
	setupConcurrencyTestData(t)
	f := newConcFixture(t)

	request := f.newConcRequest(t, 100, false)
	offer := f.newConcOffer(t, concSupplier1, request.ID, 100)

	if _, _, err := f.matchUC.Like(concBuyerCtx(), offer.ID); err != nil {
		t.Fatalf("first Like() error: %v", err)
	}
	_, firstTransactionID := concTransactionForOffer(t, offer.ID)

	if err := f.transactionUC.Cancel(concBuyerCtx(), firstTransactionID, "buyer changed its mind"); err != nil {
		t.Fatalf("Cancel() error: %v", err)
	}

	if got := concActualAmount(t, request.ID); got != 100 {
		t.Fatalf("actual amount after cancel = %v, want 100 (the release must be conserved)", got)
	}

	// The same offer, a second time. This is the assertion the old UNIQUE
	// constraint made impossible.
	if _, _, err := f.matchUC.Like(concBuyerCtx(), offer.ID); err != nil {
		t.Fatalf("second Like() on the same offer error: %v", err)
	}

	if got := concQueryCount(t, `SELECT count(*) FROM matches WHERE supply_offer_id = $1`, offer.ID); got != 2 {
		t.Errorf("match rows for the offer = %d, want 2 (history is retained)", got)
	}
	if got := concQueryCount(t,
		`SELECT count(*) FROM matches WHERE supply_offer_id = $1 AND status = $2`,
		offer.ID, domain.MatchActive); got != 1 {
		t.Errorf("active matches for the offer = %d, want 1", got)
	}
	if got := concActualAmount(t, request.ID); got != 0 {
		t.Errorf("actual amount = %v, want 0 (amounts must be conserved across the retry)", got)
	}
}

// TestCancelConcurrentSameTransactionReleasesOnce proves the cascade is atomic:
// the second canceller re-reads the committed transaction, finds it terminal and
// releases nothing. The LEAST clamp in Release would otherwise hide a double
// release on the amount, so the transaction status is the real assertion.
func TestCancelConcurrentSameTransactionReleasesOnce(t *testing.T) {
	setupConcurrencyTestData(t)
	f := newConcFixture(t)

	request := f.newConcRequest(t, 100, false)
	offer := f.newConcOffer(t, concSupplier1, request.ID, 100)

	if _, _, err := f.matchUC.Like(concBuyerCtx(), offer.ID); err != nil {
		t.Fatalf("Like() error: %v", err)
	}
	_, transactionID := concTransactionForOffer(t, offer.ID)

	if got := concActualAmount(t, request.ID); got != 0 {
		t.Fatalf("actual amount after like = %v, want 0", got)
	}

	errs := concRunConcurrently(2, func(worker int) error {
		return f.transactionUC.Cancel(concBuyerCtx(), transactionID, fmt.Sprintf("reason-%d", worker))
	})

	if succeeded := concCountSucceeded(errs); succeeded != 1 {
		t.Fatalf("successful cancels = %d, want exactly 1 (errors: %v)", succeeded, errs)
	}
	for _, err := range errs {
		if err == nil {
			continue
		}
		if !errors.Is(err, domain.ErrTerminalState) && !errors.Is(err, domain.ErrInvalidMatchStatus) {
			t.Errorf("loser error = %v, want a terminal-state error", err)
		}
	}

	if got := concActualAmount(t, request.ID); got != 100 {
		t.Errorf("actual amount = %v, want 100 (released exactly once)", got)
	}
	if got := concQueryCount(t, `SELECT count(*) FROM matches WHERE supply_offer_id = $1 AND status = $2`,
		offer.ID, domain.MatchCancelled); got != 1 {
		t.Errorf("cancelled matches = %d, want 1", got)
	}
}

// TestCancelConcurrentDistinctTransactionsConservesAmount is the lost-update
// test: two transactions on the same request, each releasing its own share at
// the same time. Read-modify-write on the request row drops one of the two
// releases; the row lock plus the arithmetic Release keeps both.
func TestCancelConcurrentDistinctTransactionsConservesAmount(t *testing.T) {
	setupConcurrencyTestData(t)
	f := newConcFixture(t)

	request := f.newConcRequest(t, 100, true)
	offerA := f.newConcOffer(t, concSupplier1, request.ID, 40)
	offerB := f.newConcOffer(t, concSupplier2, request.ID, 60)

	if _, _, err := f.matchUC.Like(concBuyerCtx(), offerA.ID); err != nil {
		t.Fatalf("Like() offer A error: %v", err)
	}
	if _, _, err := f.matchUC.Like(concBuyerCtx(), offerB.ID); err != nil {
		t.Fatalf("Like() offer B error: %v", err)
	}

	_, transactionA := concTransactionForOffer(t, offerA.ID)
	_, transactionB := concTransactionForOffer(t, offerB.ID)

	if got := concActualAmount(t, request.ID); got != 0 {
		t.Fatalf("actual amount after both likes = %v, want 0", got)
	}

	transactions := []uuid.UUID{transactionA, transactionB}
	errs := concRunConcurrently(len(transactions), func(worker int) error {
		return f.transactionUC.Cancel(concBuyerCtx(), transactions[worker], "no longer needed")
	})

	if succeeded := concCountSucceeded(errs); succeeded != len(transactions) {
		t.Fatalf("successful cancels = %d, want %d (errors: %v)", succeeded, len(transactions), errs)
	}

	// 40 + 60 released concurrently onto a request that started at 0.
	if got := concActualAmount(t, request.ID); got != 100 {
		t.Errorf("actual amount = %v, want 100 (every released unit must survive)", got)
	}
	if got := concActiveMatchedAmount(t, request.ID); got != 0 {
		t.Errorf("active matched amount = %v, want 0", got)
	}
}

// TestLikeRollbackOnContextCancelLeavesNoOrphan cancels the request context
// while the unit of work waits on a row lock, which is the moment a client
// disconnects mid-transaction. The rollback must run on a context that is not
// the canceled one, or the connection is leaked with its locks held.
func TestLikeRollbackOnContextCancelLeavesNoOrphan(t *testing.T) {
	setupConcurrencyTestData(t)
	f := newConcFixture(t)

	request := f.newConcRequest(t, 100, true)
	offer := f.newConcOffer(t, concSupplier1, request.ID, 50)

	// The pool is quiescent at this point: every fixture write has already
	// returned its connection, and nothing else in the package runs in parallel.
	// This is the only instant where the counter describes the pool's resting
	// state, so it is where the baseline has to be read. Reading it after the
	// blocker transaction below would count the blocker's own connection and make
	// the final assertion demand a connection the rollback legitimately gave back.
	baseline := TestPool.Stat().AcquiredConns()

	blockerCtx := context.Background()
	blocker, err := TestPool.Begin(blockerCtx)
	if err != nil {
		t.Fatalf("begin blocking transaction: %v", err)
	}
	// A barrier timeout below is a t.Fatal, which would otherwise strand this
	// transaction for the rest of the package holding a connection and the
	// request row lock. Rolling back an already-closed transaction is a no-op in
	// pgx, so this is safe after the explicit rollback further down.
	defer func() { _ = blocker.Rollback(blockerCtx) }()

	if _, err := blocker.Exec(blockerCtx,
		`SELECT id FROM supply_requests WHERE id = $1 FOR UPDATE`, request.ID); err != nil {
		t.Fatalf("lock the request row: %v", err)
	}

	likeCtx, cancel := context.WithCancel(concBuyerCtx())
	defer cancel()
	likeDone := make(chan error, 1)
	go func() {
		_, _, err := f.matchUC.Like(likeCtx, offer.ID)
		likeDone <- err
	}()

	// The unit of work is now inside its transaction and waiting on the row
	// lock this test holds. Cancelling here is exactly what a client
	// disconnecting mid-transaction does.
	concWaitForBlockedQuery(t, "FOR UPDATE OF s")
	cancel()

	var likeErr error
	select {
	case likeErr = <-likeDone:
	case <-time.After(30 * time.Second):
		t.Fatal("Like() did not return after the context was cancelled")
	}

	if likeErr == nil {
		t.Fatal("Like() error = nil, want a cancellation error")
	}

	if err := blocker.Rollback(blockerCtx); err != nil {
		t.Fatalf("release the blocking lock: %v", err)
	}

	if got := concQueryCount(t, `SELECT count(*) FROM matches WHERE supply_offer_id = $1`, offer.ID); got != 0 {
		t.Errorf("match rows = %d, want 0: the rolled back unit of work left an orphan", got)
	}
	if got := concQueryCount(t, `SELECT count(*) FROM transactions`); got != 0 {
		t.Errorf("transaction rows = %d, want 0: the rolled back unit of work left an orphan", got)
	}
	if got := concActualAmount(t, request.ID); got != 100 {
		t.Errorf("actual amount = %v, want 100: the reservation must be rolled back", got)
	}
	if got := concQueryCount(t,
		`SELECT count(*) FROM supply_offers WHERE id = $1 AND status = $2`,
		offer.ID, domain.OfferActive); got != 1 {
		t.Errorf("offer still active = %d, want 1: the offer status must be rolled back", got)
	}

	concAssertConnsReturnTo(t, baseline)
}

// TestConfirmDeliveryConcurrentCompletesRequestExactlyOnce races the two
// completions of a request whose two providers together cover its total amount.
// Because the request row is locked for the whole unit of work, the sum each
// completer reads includes the other's finished transaction, so the request is
// closed. Without the lock both could read a sum that excludes both and the
// request would be left open forever.
func TestConfirmDeliveryConcurrentCompletesRequestExactlyOnce(t *testing.T) {
	setupConcurrencyTestData(t)
	f := newConcFixture(t)

	request := f.newConcRequest(t, 100, true)
	offerA := f.newConcOffer(t, concSupplier1, request.ID, 50)
	offerB := f.newConcOffer(t, concSupplier2, request.ID, 50)
	offers := []*domain.SupplyOffer{offerA, offerB}
	suppliers := []uuid.UUID{concSupplier1, concSupplier2}

	for i, offer := range offers {
		if _, _, err := f.matchUC.Like(concBuyerCtx(), offer.ID); err != nil {
			t.Fatalf("Like() offer %d error: %v", i, err)
		}
		_, transactionID := concTransactionForOffer(t, offer.ID)

		if err := f.transactionUC.ConfirmStart(concBuyerCtx(), transactionID); err != nil {
			t.Fatalf("ConfirmStart(buyer) transaction %d error: %v", i, err)
		}
		if err := f.transactionUC.ConfirmStart(concSupplierCtx(suppliers[i]), transactionID); err != nil {
			t.Fatalf("ConfirmStart(supplier) transaction %d error: %v", i, err)
		}
	}

	transactionIDs := make([]uuid.UUID, 0, len(offers))
	for _, offer := range offers {
		_, transactionID := concTransactionForOffer(t, offer.ID)
		transactionIDs = append(transactionIDs, transactionID)
	}

	errs := concRunConcurrently(len(offers), func(worker int) error {
		if err := f.transactionUC.ConfirmDelivery(concBuyerCtx(), transactionIDs[worker]); err != nil {
			return fmt.Errorf("buyer confirm: %w", err)
		}
		if err := f.transactionUC.ConfirmDelivery(concSupplierCtx(suppliers[worker]), transactionIDs[worker]); err != nil {
			return fmt.Errorf("supplier confirm: %w", err)
		}
		return nil
	})

	for i, err := range errs {
		if err != nil {
			t.Fatalf("worker %d error: %v", i, err)
		}
	}

	for i, transactionID := range transactionIDs {
		transaction, err := f.transactionRe.GetByID(context.Background(), transactionID)
		if err != nil {
			t.Fatalf("GetByID() transaction %d: %v", i, err)
		}
		if transaction.Status != domain.TransactionCompleted {
			t.Errorf("transaction %d status = %v, want completed", i, transaction.Status)
		}
	}

	closed := concQueryCount(t, `SELECT count(*) FROM supply_requests WHERE id = $1 AND status = $2`,
		request.ID, domain.SupplyRequestCompleted)
	if closed != 1 {
		t.Errorf("completed requests = %d, want 1: the request must be closed exactly once", closed)
	}
	if got := concActualAmount(t, request.ID); got != 0 {
		t.Errorf("actual amount = %v, want 0: the request must not be corrupted", got)
	}
}

// TestReserveRejectsOverAllocationAtStorageLevel calls the narrow write
// directly, with no state machine in front of it. This is the safety net that
// makes over-assignment impossible even if the lock discipline is broken.
func TestReserveRejectsOverAllocationAtStorageLevel(t *testing.T) {
	setupConcurrencyTestData(t)
	f := newConcFixture(t)

	ctx := context.Background()
	request := f.newConcRequest(t, 100, true)
	now := f.clock.Now()

	err := f.requestRepo.Reserve(ctx, request.ID, 150, now)
	if !errors.Is(err, domain.ErrInsufficientAmount) {
		t.Fatalf("Reserve(150) error = %v, want %v", err, domain.ErrInsufficientAmount)
	}
	if got := concActualAmount(t, request.ID); got != 100 {
		t.Errorf("actual amount after the rejected reserve = %v, want 100 (the row must be unchanged)", got)
	}

	if err := f.requestRepo.Reserve(ctx, request.ID, 100, now); err != nil {
		t.Fatalf("Reserve(100) error: %v", err)
	}
	if got := concActualAmount(t, request.ID); got != 0 {
		t.Errorf("actual amount after the accepted reserve = %v, want 0", got)
	}

	if err := f.requestRepo.Release(ctx, request.ID, 100, now); err != nil {
		t.Fatalf("Release(100) error: %v", err)
	}
	if got := concActualAmount(t, request.ID); got != 100 {
		t.Errorf("actual amount after the release = %v, want 100", got)
	}
}

// TestPartialUniqueIndexRejectsTwoActiveMatchesForOneOffer asserts the
// database-level gate behind the per-offer check, straight through SQL. Without
// it, two concurrent likes on one offer would both pass the application check
// and the second insert would win.
func TestPartialUniqueIndexRejectsTwoActiveMatchesForOneOffer(t *testing.T) {
	setupConcurrencyTestData(t)
	f := newConcFixture(t)

	ctx := context.Background()
	request := f.newConcRequest(t, 100, true)
	offer := f.newConcOffer(t, concSupplier1, request.ID, 100)

	insertMatch := func(id uuid.UUID, status domain.MatchStatus) error {
		_, err := TestPool.Exec(ctx, `
			INSERT INTO matches (id, supply_offer_id, supply_request_id, status, matched_amount, amount_unit, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $7)`,
			id, offer.ID, request.ID, status, 100, domain.Kg, f.clock.Now())
		return err
	}

	first := uuid.New()
	if err := insertMatch(first, domain.MatchActive); err != nil {
		t.Fatalf("first active match insert: %v", err)
	}

	second := uuid.New()
	err := insertMatch(second, domain.MatchActive)
	if err == nil {
		t.Fatal("second active match insert error = nil, want a unique violation")
	}
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		t.Fatalf("second active match insert error = %v, want SQLSTATE 23505", err)
	}

	// A cancelled match on the same offer is history, not a conflict.
	if err := insertMatch(uuid.New(), domain.MatchCancelled); err != nil {
		t.Fatalf("cancelled match insert: %v", err)
	}

	if got := concQueryCount(t, `SELECT count(*) FROM matches WHERE supply_offer_id = $1 AND status = $2`,
		offer.ID, domain.MatchActive); got != 1 {
		t.Errorf("active matches = %d, want 1", got)
	}
}
