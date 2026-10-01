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
		offer, err := scope.Offers.GetByID(ctx, supplyOfferID)
		if err != nil {
			return err
		}
		request, err := scope.Requests.LockForUpdate(ctx, offer.SupplyRequest)
		if err != nil {
			return err
		}

		if !offer.IsActionable() {
			return domain.ErrInvalidOfferStatus
		}
		if !request.IsOpen() {
			return domain.ErrInvalidRequestStatus
		}
		if request.BuyerID != principal.UserID {
			return domain.ErrForbidden
		}
		offer, err = scope.Offers.LockByIDForUpdate(ctx, supplyOfferID)
		if err != nil {
			return err
		}
		if !offer.IsActionable() {
			return domain.ErrInvalidOfferStatus
		}

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

		if err := request.ReserveAmount(offer.TotalAmount); err != nil {
			return err
		}

		createdMatch = domain.NewMatch(offer.ID, request.ID, offer.TotalAmount, offer.AmountUnit)
		createdTransaction = domain.NewTransaction(createdMatch.ID)

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
