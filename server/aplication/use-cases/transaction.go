package usecases

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	domain "milpa/domain/entities"
	"milpa/domain/port/primary"
	port "milpa/domain/port/secondary"
	"milpa/internal/auth"
)

type transactionSession struct {
	principal   auth.Principal
	transaction domain.Transaction
	match       *domain.Match
	request     domain.SupplyRequest
	offer       domain.SupplyOffer
	participant domain.TransactionParticipant
}

type TransactionUseCaseImpl struct {
	transactionRepo port.TransactionRepository
	matchRepo       port.MatchRepository
	requestRepo     port.SupplyRequestRepository
	offerRepo       port.SupplyOfferRepository
	timer           port.TimeProvider
	tx              port.UnitOfWork
}

func NewTransactionUseCase(
	transactionRepo port.TransactionRepository,
	matchRepo port.MatchRepository,
	requestRepo port.SupplyRequestRepository,
	offerRepo port.SupplyOfferRepository,
	timer port.TimeProvider,
	tx port.UnitOfWork,
) *TransactionUseCaseImpl {
	return &TransactionUseCaseImpl{
		transactionRepo: transactionRepo,
		matchRepo:       matchRepo,
		requestRepo:     requestRepo,
		offerRepo:       offerRepo,
		timer:           timer,
		tx:              tx,
	}
}

var _ primary.TransactionUseCase = (*TransactionUseCaseImpl)(nil)

func (uc *TransactionUseCaseImpl) GetByMatch(ctx context.Context, matchID uuid.UUID) (*dto.TransactionDTO, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, err
	}

	match, err := uc.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		return nil, mapTransactionNotFound(err)
	}

	request, offer, err := uc.parties(ctx, match)
	if err != nil {
		return nil, err
	}

	if _, err := participantFor(principal, &request, &offer); err != nil {
		return nil, err
	}

	transaction, err := uc.transactionRepo.GetByMatch(ctx, matchID)
	if err != nil {
		return nil, mapTransactionNotFound(err)
	}

	return transactionToDTO(&transaction), nil
}

func (uc *TransactionUseCaseImpl) ListByRequest(ctx context.Context, supplyRequestID uuid.UUID) ([]*dto.TransactionDTO, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, err
	}

	request, err := uc.requestRepo.GetByID(ctx, supplyRequestID)
	if err != nil {
		return nil, mapTransactionNotFound(err)
	}

	transactions, err := uc.transactionRepo.ListByRequest(ctx, supplyRequestID)
	if err != nil {
		return nil, err
	}

	if principal.UserID == request.BuyerID {
		return transactionsToDTO(transactions), nil
	}

	matches, err := uc.matchRepo.ListByRequest(ctx, supplyRequestID)
	if err != nil {
		return nil, err
	}
	matchesByID := make(map[uuid.UUID]domain.Match, len(matches))
	for _, match := range matches {
		matchesByID[match.ID] = match
	}

	suppliersByMatch := make(map[uuid.UUID]uuid.UUID, len(matches))
	visible := make([]domain.Transaction, 0, len(transactions))
	for _, transaction := range transactions {
		match, ok := matchesByID[transaction.MatchID]
		if !ok {
			continue
		}
		supplierID, ok := suppliersByMatch[match.ID]
		if !ok {
			offer, err := uc.offerRepo.GetByID(ctx, match.SupplyOffer)
			if err != nil {
				return nil, mapTransactionNotFound(err)
			}
			supplierID = offer.SupplierID
			suppliersByMatch[match.ID] = supplierID
		}
		if supplierID == principal.UserID {
			visible = append(visible, transaction)
		}
	}

	if len(visible) == 0 {
		return nil, domain.ErrForbidden
	}

	return transactionsToDTO(visible), nil
}

func (uc *TransactionUseCaseImpl) ConfirmStart(ctx context.Context, transactionID uuid.UUID) error {
	session, err := uc.authorizeByID(ctx, transactionID)
	if err != nil {
		return err
	}

	if err := session.transaction.ConfirmStart(session.participant, uc.timer.Now()); err != nil {
		return err
	}

	return uc.transactionRepo.Update(ctx, &session.transaction)
}

// ConfirmDelivery confirms one side of a delivery as ONE unit of work. When the
// confirmation completes the transaction and the request is already covered by
// completed transactions, the request is closed inside the same transaction: the
// request row is locked for the whole unit of work, so a concurrent completion
// cannot read a snapshot that excludes the transaction this one just finished.
//
// Lock order: SupplyRequest -> SupplyOffer -> Match -> Transaction.
func (uc *TransactionUseCaseImpl) ConfirmDelivery(ctx context.Context, transactionID uuid.UUID) error {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return err
	}

	return uc.tx.WithinTx(ctx, func(scope port.TxScope) error {
		session, err := uc.lockSession(ctx, principal, scope, transactionID)
		if err != nil {
			return err
		}

		if err := session.transaction.ConfirmDelivery(session.participant, uc.timer.Now()); err != nil {
			return err
		}

		if err := scope.Transactions.Update(ctx, &session.transaction); err != nil {
			return err
		}

		if session.transaction.Status != domain.TransactionCompleted {
			return nil
		}

		return uc.autoCloseRequest(ctx, scope, session)
	})
}

// Cancel cascades the cancellation as ONE unit of work. Every goroutine that
// cancels the same transaction either commits the whole cascade or none of it,
// so a cancellation can no longer release the request amount without releasing
// the match, or vice versa.
//
// Lock order: SupplyRequest -> SupplyOffer -> Match -> Transaction.
func (uc *TransactionUseCaseImpl) Cancel(ctx context.Context, transactionID uuid.UUID, reason string) error {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return err
	}

	return uc.tx.WithinTx(ctx, func(scope port.TxScope) error {
		session, err := uc.lockSession(ctx, principal, scope, transactionID)
		if err != nil {
			return err
		}

		now := uc.timer.Now()

		if err := session.transaction.Cancel(session.principal.UserID, reason, now); err != nil {
			return err
		}
		if err := session.match.Cancel(); err != nil {
			return err
		}
		if err := session.request.ReleaseAmount(session.match.MatchedAmount); err != nil {
			return err
		}

		offerReactivated := false
		if session.offer.Status == domain.OfferMatched {
			session.offer.Status = domain.OfferActive
			session.offer.UpdatedAt = now
			offerReactivated = true
		}

		if err := scope.Transactions.Update(ctx, &session.transaction); err != nil {
			return err
		}
		if err := scope.Matches.Update(ctx, session.match); err != nil {
			return err
		}
		if err := scope.Requests.Release(ctx, session.request.ID, session.match.MatchedAmount, now); err != nil {
			return err
		}
		if offerReactivated {
			if err := scope.Offers.Update(ctx, &session.offer); err != nil {
				return err
			}
		}

		return nil
	})
}

// lockSession resolves the transaction graph inside an already open transaction,
// takes the four row locks in the global order, re-reads every row through the
// locked getters and rebuilds the session from those locked rows.
//
// The unlocked reads exist only to learn the ids to lock: supply_request_id and
// supply_offer_id are only reachable through the match row. Every decision the
// caller makes afterwards is made on the locked rows, never on the reads.
func (uc *TransactionUseCaseImpl) lockSession(
	ctx context.Context,
	principal auth.Principal,
	scope port.TxScope,
	transactionID uuid.UUID,
) (*transactionSession, error) {
	transaction, err := scope.Transactions.GetByID(ctx, transactionID)
	if err != nil {
		return nil, mapTransactionNotFound(err)
	}

	match, err := scope.Matches.GetByID(ctx, transaction.MatchID)
	if err != nil {
		return nil, mapTransactionNotFound(err)
	}

	request, err := scope.Requests.LockForUpdate(ctx, match.SupplyRequest)
	if err != nil {
		return nil, mapTransactionNotFound(err)
	}

	offer, err := scope.Offers.LockByIDForUpdate(ctx, match.SupplyOffer)
	if err != nil {
		return nil, mapTransactionNotFound(err)
	}

	lockedMatch, err := scope.Matches.LockByIDForUpdate(ctx, match.ID)
	if err != nil {
		return nil, mapTransactionNotFound(err)
	}

	lockedTransaction, err := scope.Transactions.LockByIDForUpdate(ctx, transaction.ID)
	if err != nil {
		return nil, mapTransactionNotFound(err)
	}

	participant, err := participantFor(principal, &request, &offer)
	if err != nil {
		return nil, err
	}

	return &transactionSession{
		principal:   principal,
		transaction: lockedTransaction,
		match:       lockedMatch,
		request:     request,
		offer:       offer,
		participant: participant,
	}, nil
}

func (uc *TransactionUseCaseImpl) authorizeByID(ctx context.Context, transactionID uuid.UUID) (*transactionSession, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, err
	}

	transaction, err := uc.transactionRepo.GetByID(ctx, transactionID)
	if err != nil {
		return nil, mapTransactionNotFound(err)
	}

	match, err := uc.matchRepo.GetByID(ctx, transaction.MatchID)
	if err != nil {
		return nil, mapTransactionNotFound(err)
	}

	request, offer, err := uc.parties(ctx, match)
	if err != nil {
		return nil, err
	}

	participant, err := participantFor(principal, &request, &offer)
	if err != nil {
		return nil, err
	}

	return &transactionSession{
		principal:   principal,
		transaction: transaction,
		match:       match,
		request:     request,
		offer:       offer,
		participant: participant,
	}, nil
}

func (uc *TransactionUseCaseImpl) parties(ctx context.Context, match *domain.Match) (domain.SupplyRequest, domain.SupplyOffer, error) {
	request, err := uc.requestRepo.GetByID(ctx, match.SupplyRequest)
	if err != nil {
		return domain.SupplyRequest{}, domain.SupplyOffer{}, mapTransactionNotFound(err)
	}

	offer, err := uc.offerRepo.GetByID(ctx, match.SupplyOffer)
	if err != nil {
		return domain.SupplyRequest{}, domain.SupplyOffer{}, mapTransactionNotFound(err)
	}

	return request, offer, nil
}

// autoCloseRequest completes the request once its completed transactions cover
// its total amount. It must run inside the same unit of work as the completion
// that triggered it: the request row is already locked by lockSession, so the
// sum read here includes the transaction this unit of work just completed, and a
// concurrent completion reads this one instead of a snapshot that excludes it.
func (uc *TransactionUseCaseImpl) autoCloseRequest(ctx context.Context, scope port.TxScope, session *transactionSession) error {
	if !session.request.IsOpen() {
		return nil
	}

	transactions, err := scope.Transactions.ListByRequest(ctx, session.match.SupplyRequest)
	if err != nil {
		return err
	}

	matches, err := scope.Matches.ListByRequest(ctx, session.match.SupplyRequest)
	if err != nil {
		return err
	}
	matchedAmounts := make(map[uuid.UUID]float64, len(matches))
	for _, match := range matches {
		matchedAmounts[match.ID] = match.MatchedAmount
	}

	var completedAmount float64
	for _, transaction := range transactions {
		if transaction.Status == domain.TransactionCompleted {
			completedAmount += matchedAmounts[transaction.MatchID]
		}
	}

	if completedAmount < session.request.TotalAmount {
		return nil
	}

	if err := session.request.Complete(); err != nil {
		return err
	}

	// Status and the zeroed actual_amount are written together, and nothing
	// else: the request row is locked, so a full-row rewrite would only be able
	// to push a stale snapshot back over the row. Zeroing actual_amount is what
	// makes a fully delivered request land on an exact 0 instead of the
	// IEEE-754 residue of subtracting its fractions one at a time.
	return scope.Requests.UpdateCompletion(ctx, session.request.ID, session.request.Status, session.request.UpdatedAt)
}

func participantFor(principal auth.Principal, request *domain.SupplyRequest, offer *domain.SupplyOffer) (domain.TransactionParticipant, error) {
	switch principal.UserID {
	case request.BuyerID:
		return domain.BuyerParticipant, nil
	case offer.SupplierID:
		return domain.SupplierParticipant, nil
	default:
		return 0, domain.ErrForbidden
	}
}

func mapTransactionNotFound(err error) error {
	if errors.Is(err, domain.ErrNotFound) {
		return domain.ErrNotFound
	}
	return err
}

func transactionToDTO(transaction *domain.Transaction) *dto.TransactionDTO {
	return &dto.TransactionDTO{
		ID:                          &transaction.ID,
		MatchID:                     &transaction.MatchID,
		Status:                      transaction.Status,
		BuyerStartConfirmedAt:       transaction.BuyerStartConfirmedAt,
		SupplierStartConfirmedAt:    transaction.SupplierStartConfirmedAt,
		BuyerDeliveryConfirmedAt:    transaction.BuyerDeliveryConfirmedAt,
		SupplierDeliveryConfirmedAt: transaction.SupplierDeliveryConfirmedAt,
		CancelledBy:                 transaction.CancelledBy,
		CancelReason:                transaction.CancelReason,
		CreatedAt:                   transaction.CreatedAt,
		UpdatedAt:                   transaction.UpdatedAt,
		History:                     transactionHistory(transaction),
	}
}

func transactionsToDTO(transactions []domain.Transaction) []*dto.TransactionDTO {
	dtos := make([]*dto.TransactionDTO, len(transactions))
	for i := range transactions {
		dtos[i] = transactionToDTO(&transactions[i])
	}
	return dtos
}

func transactionHistory(transaction *domain.Transaction) []dto.TransactionHistoryEntryDTO {
	history := []dto.TransactionHistoryEntryDTO{
		{Status: domain.TransactionMatched, At: transaction.CreatedAt},
	}

	if transaction.BuyerStartConfirmedAt != nil && transaction.SupplierStartConfirmedAt != nil {
		history = append(history, dto.TransactionHistoryEntryDTO{
			Status: domain.TransactionInProgress,
			At:     laterTime(*transaction.BuyerStartConfirmedAt, *transaction.SupplierStartConfirmedAt),
		})
	}

	if transaction.Status == domain.TransactionCompleted &&
		transaction.BuyerDeliveryConfirmedAt != nil && transaction.SupplierDeliveryConfirmedAt != nil {
		history = append(history, dto.TransactionHistoryEntryDTO{
			Status: domain.TransactionCompleted,
			At:     laterTime(*transaction.BuyerDeliveryConfirmedAt, *transaction.SupplierDeliveryConfirmedAt),
		})
	}

	if transaction.Status == domain.TransactionCancelled {
		history = append(history, dto.TransactionHistoryEntryDTO{
			Status:       domain.TransactionCancelled,
			At:           transaction.UpdatedAt,
			CancelReason: transaction.CancelReason,
			CancelledBy:  transaction.CancelledBy,
		})
	}

	return history
}

func laterTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}
