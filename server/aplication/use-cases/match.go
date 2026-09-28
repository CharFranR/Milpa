package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	domain "milpa/domain/entities"
	"milpa/domain/port/primary"
	port "milpa/domain/port/secondary"
	"milpa/internal/auth"
)

type MatchUseCaseImpl struct {
	requestRepo     port.SupplyRequestRepository
	offerRepo       port.SupplyOfferRepository
	matchRepo       port.MatchRepository
	transactionRepo port.TransactionRepository
	recommendations primary.RecommendationUseCase
	tx              port.UnitOfWork
}

func NewMatchUseCase(
	requestRepo port.SupplyRequestRepository,
	offerRepo port.SupplyOfferRepository,
	matchRepo port.MatchRepository,
	transactionRepo port.TransactionRepository,
	recommendations primary.RecommendationUseCase,
	tx port.UnitOfWork,
) *MatchUseCaseImpl {
	return &MatchUseCaseImpl{
		requestRepo:     requestRepo,
		offerRepo:       offerRepo,
		matchRepo:       matchRepo,
		transactionRepo: transactionRepo,
		recommendations: recommendations,
		tx:              tx,
	}
}

// Like matches an offer with its request as ONE unit of work: the reservation,
// the match row, the transaction row and the offer status either all become
// visible or none of them do. The ROLLBACK is the compensation; there is no
// manual compensation ladder to get wrong under a partial failure.
//
// Lock order inside the transaction: SupplyRequest -> SupplyOffer. Matches and
// transactions are created rather than locked, so they take no row lock.
func (uc *MatchUseCaseImpl) Like(ctx context.Context, supplyOfferID uuid.UUID) (*dto.MatchDTO, *dto.TransactionDTO, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, nil, err
	}

	var (
		createdMatch       *domain.Match
		createdTransaction *domain.Transaction
	)

	err = uc.tx.WithinTx(ctx, func(scope port.TxScope) error {
		// Two unlocked PK reads, only to resolve the lock keys. The request id
		// is only reachable through the offer row, so the offer is read first.
		offer, err := scope.Offers.GetByID(ctx, supplyOfferID)
		if err != nil {
			return err
		}

		// FIRST LOCK: the request, per the global order.
		request, err := scope.Requests.LockForUpdate(ctx, offer.SupplyRequest)
		if err != nil {
			return err
		}

		// Re-validate against the FRESH locked rows, not the unlocked reads.
		if !offer.IsActionable() {
			return domain.ErrInvalidOfferStatus
		}
		if !request.IsOpen() {
			return domain.ErrInvalidRequestStatus
		}
		if request.BuyerID != principal.UserID {
			return domain.ErrForbidden
		}

		// SECOND LOCK: the offer, now that the request is held.
		offer, err = scope.Offers.LockByIDForUpdate(ctx, supplyOfferID)
		if err != nil {
			return err
		}
		if !offer.IsActionable() {
			return domain.ErrInvalidOfferStatus
		}

		// Both gates are evaluated under the request lock, so two concurrent
		// likes on the same request serialize and the second one sees the first
		// one's committed match.
		if !request.MultipleProviders {
			existsByRequest, err := scope.Matches.ExistsActiveByRequest(ctx, request.ID)
			if err != nil {
				return err
			}
			if existsByRequest {
				return domain.ErrInvalidMatchStatus
			}
		}

		existsByOffer, err := scope.Matches.ExistsActiveByOffer(ctx, offer.ID)
		if err != nil {
			return err
		}
		if existsByOffer {
			return domain.ErrInvalidMatchStatus
		}

		// Go decides: the state machine rejects the over-assignment.
		if err := request.ReserveAmount(offer.TotalAmount); err != nil {
			return err
		}

		createdMatch = domain.NewMatch(offer.ID, request.ID, offer.TotalAmount, offer.AmountUnit)
		createdTransaction = domain.NewTransaction(createdMatch.ID)

		// SQL enforces: actual_amount >= $1 makes over-assignment impossible
		// even if the lock discipline above is broken later.
		if err := scope.Requests.Reserve(ctx, request.ID, offer.TotalAmount, time.Now()); err != nil {
			return err
		}
		if err := scope.Matches.Create(ctx, createdMatch); err != nil {
			return err
		}
		if err := scope.Transactions.Create(ctx, createdTransaction); err != nil {
			return err
		}
		if err := offer.MarkMatched(); err != nil {
			return err
		}
		if err := scope.Offers.Update(ctx, &offer); err != nil {
			return err
		}

		if !request.MultipleProviders {
			otherOffers, err := scope.Offers.ListByRequest(ctx, request.ID)
			if err != nil {
				return err
			}

			for i := range otherOffers {
				other := otherOffers[i]
				if other.ID == offer.ID || !other.IsActionable() {
					continue
				}

				if err := other.Reject(); err != nil {
					return err
				}
				if err := scope.Offers.Update(ctx, &other); err != nil {
					return err
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return matchToDTO(createdMatch), matchTransactionToDTO(createdTransaction), nil
}

func (uc *MatchUseCaseImpl) Pass(ctx context.Context, supplyOfferID uuid.UUID) error {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return err
	}

	offer, err := uc.offerRepo.GetByID(ctx, supplyOfferID)
	if err != nil {
		return err
	}

	if !offer.IsActionable() {
		return domain.ErrInvalidOfferStatus
	}

	request, err := uc.requestRepo.GetByID(ctx, offer.SupplyRequest)
	if err != nil {
		return err
	}

	if request.BuyerID != principal.UserID {
		return domain.ErrForbidden
	}

	if err := offer.Reject(); err != nil {
		return err
	}

	return uc.offerRepo.Update(ctx, &offer)
}

func (uc *MatchUseCaseImpl) GetByID(ctx context.Context, matchID uuid.UUID) (*dto.MatchDTO, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, err
	}

	match, err := uc.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		return nil, err
	}

	request, err := uc.requestRepo.GetByID(ctx, match.SupplyRequest)
	if err != nil {
		return nil, err
	}

	offer, err := uc.offerRepo.GetByID(ctx, match.SupplyOffer)
	if err != nil {
		return nil, err
	}

	if request.BuyerID != principal.UserID && offer.SupplierID != principal.UserID {
		return nil, domain.ErrForbidden
	}

	return matchToDTO(match), nil
}

func (uc *MatchUseCaseImpl) ListByRequest(ctx context.Context, supplyRequestID uuid.UUID) ([]*dto.MatchDTO, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, err
	}

	request, err := uc.requestRepo.GetByID(ctx, supplyRequestID)
	if err != nil {
		return nil, err
	}

	if request.BuyerID != principal.UserID {
		return nil, domain.ErrForbidden
	}

	matches, err := uc.matchRepo.ListByRequest(ctx, supplyRequestID)
	if err != nil {
		return nil, err
	}

	dtos := make([]*dto.MatchDTO, 0, len(matches))
	for i := range matches {
		dtos = append(dtos, matchToDTO(&matches[i]))
	}

	return dtos, nil
}

func (uc *MatchUseCaseImpl) ListPrioritized(ctx context.Context, supplyRequestID uuid.UUID) ([]*dto.PrioritizedOfferDTO, error) {
	return uc.recommendations.RankOffers(ctx, supplyRequestID)
}

var _ primary.MatchUseCase = (*MatchUseCaseImpl)(nil)

func matchToDTO(match *domain.Match) *dto.MatchDTO {
	return &dto.MatchDTO{
		ID:            match.ID,
		SupplyOffer:   match.SupplyOffer,
		SupplyRequest: match.SupplyRequest,
		Status:        match.Status,
		MatchedAmount: match.MatchedAmount,
		AmountUnit:    match.AmountUnit,
		CreatedAt:     match.CreatedAt,
		UpdatedAt:     match.UpdatedAt,
	}
}

func matchTransactionToDTO(transaction *domain.Transaction) *dto.TransactionDTO {
	id := transaction.ID
	matchID := transaction.MatchID
	return &dto.TransactionDTO{
		ID:      &id,
		MatchID: &matchID,
		Status:  transaction.Status,
	}
}
