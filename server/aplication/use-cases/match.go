package usecases

import (
	"context"
	"errors"

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
}

func NewMatchUseCase(
	requestRepo port.SupplyRequestRepository,
	offerRepo port.SupplyOfferRepository,
	matchRepo port.MatchRepository,
	transactionRepo port.TransactionRepository,
	recommendations primary.RecommendationUseCase,
) *MatchUseCaseImpl {
	return &MatchUseCaseImpl{
		requestRepo:     requestRepo,
		offerRepo:       offerRepo,
		matchRepo:       matchRepo,
		transactionRepo: transactionRepo,
		recommendations: recommendations,
	}
}

func (uc *MatchUseCaseImpl) Like(ctx context.Context, supplyOfferID uuid.UUID) (*dto.MatchDTO, *dto.TransactionDTO, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return nil, nil, err
	}

	offer, err := uc.offerRepo.GetByID(ctx, supplyOfferID)
	if err != nil {
		return nil, nil, err
	}

	if !offer.IsActionable() {
		return nil, nil, domain.ErrInvalidOfferStatus
	}

	request, err := uc.requestRepo.GetByID(ctx, offer.SupplyRequest)
	if err != nil {
		return nil, nil, err
	}

	if !request.IsOpen() {
		return nil, nil, domain.ErrInvalidRequestStatus
	}

	if request.BuyerID != principal.UserID {
		return nil, nil, domain.ErrForbidden
	}

	existsByOffer, err := uc.matchRepo.ExistsActiveByOffer(ctx, offer.ID)
	if err != nil {
		return nil, nil, err
	}
	if existsByOffer {
		return nil, nil, domain.ErrInvalidMatchStatus
	}

	if !request.MultipleProviders {
		existsByRequest, err := uc.matchRepo.ExistsActiveByRequest(ctx, request.ID)
		if err != nil {
			return nil, nil, err
		}
		if existsByRequest {
			return nil, nil, domain.ErrInvalidMatchStatus
		}
	}

	createdMatch := domain.NewMatch(offer.ID, request.ID, offer.TotalAmount, offer.AmountUnit)

	if err := request.ReserveAmount(offer.TotalAmount); err != nil {
		return nil, nil, err
	}

	var rollbacks []func(context.Context) error
	compensate := func(cause error) error {
		for i := len(rollbacks) - 1; i >= 0; i-- {
			if rbErr := rollbacks[i](ctx); rbErr != nil {
				cause = errors.Join(cause, rbErr)
			}
		}
		return cause
	}

	releaseReservation := func(ctx context.Context) error {
		current, err := uc.requestRepo.GetByID(ctx, request.ID)
		if err != nil {
			return err
		}
		if err := current.ReleaseAmount(offer.TotalAmount); err != nil {
			return err
		}
		return uc.requestRepo.Update(ctx, &current)
	}

	if err := uc.requestRepo.Update(ctx, &request); err != nil {
		return nil, nil, err
	}
	rollbacks = append(rollbacks, releaseReservation)

	rollbacks = append(rollbacks, func(ctx context.Context) error {
		return uc.matchRepo.Delete(ctx, createdMatch.ID)
	})
	if err := uc.matchRepo.Create(ctx, createdMatch); err != nil {
		return nil, nil, compensate(err)
	}

	createdTransaction := domain.NewTransaction(createdMatch.ID)
	rollbacks = append(rollbacks, func(ctx context.Context) error {
		return uc.transactionRepo.Delete(ctx, createdTransaction.ID)
	})
	if err := uc.transactionRepo.Create(ctx, createdTransaction); err != nil {
		return nil, nil, compensate(err)
	}

	previousOfferStatus := offer.Status
	previousOfferUpdatedAt := offer.UpdatedAt
	rollbacks = append(rollbacks, func(ctx context.Context) error {
		offer.Status = previousOfferStatus
		offer.UpdatedAt = previousOfferUpdatedAt
		return uc.offerRepo.Update(ctx, &offer)
	})
	if err := offer.MarkMatched(); err != nil {
		return nil, nil, compensate(err)
	}
	if err := uc.offerRepo.Update(ctx, &offer); err != nil {
		return nil, nil, compensate(err)
	}

	if !request.MultipleProviders {
		otherOffers, err := uc.offerRepo.ListByRequest(ctx, request.ID)
		if err != nil {
			return nil, nil, compensate(err)
		}

		for i := range otherOffers {
			other := otherOffers[i]
			if other.ID == offer.ID || !other.IsActionable() {
				continue
			}

			previousStatus := other.Status
			previousUpdatedAt := other.UpdatedAt
			rollbacks = append(rollbacks, func(ctx context.Context) error {
				other.Status = previousStatus
				other.UpdatedAt = previousUpdatedAt
				return uc.offerRepo.Update(ctx, &other)
			})

			if err := other.Reject(); err != nil {
				return nil, nil, compensate(err)
			}
			if err := uc.offerRepo.Update(ctx, &other); err != nil {
				return nil, nil, compensate(err)
			}
		}
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
