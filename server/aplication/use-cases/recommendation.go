package usecases

import (
	"context"
	"errors"
	"sort"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	domain "milpa/domain/entities"
	"milpa/domain/port/primary"
	port "milpa/domain/port/secondary"
	"milpa/internal/auth"
)

type WeightedScoreFactor struct {
	Factor primary.ScoreFactor
	Weight float64
}

type availabilityScoreFactor struct{}

func (availabilityScoreFactor) Name() string {
	return "availability"
}

func (availabilityScoreFactor) Score(_ context.Context, input primary.OfferScoreInput) (float64, error) {
	return input.AvailableQuantity, nil
}

func AvailabilityScoreFactor() primary.ScoreFactor {
	return availabilityScoreFactor{}
}

func DefaultScoreFactors() []WeightedScoreFactor {
	return []WeightedScoreFactor{
		{Factor: availabilityScoreFactor{}, Weight: 1},
	}
}

type RecommendationUseCaseImpl struct {
	offerRepo     port.SupplyOfferRepository
	requestRepo   port.SupplyRequestRepository
	inventoryRepo port.SupplierInventoryRepository
	matchRepo     port.MatchRepository
	factors       []WeightedScoreFactor
}

func NewRecommendationUseCase(
	offerRepo port.SupplyOfferRepository,
	requestRepo port.SupplyRequestRepository,
	inventoryRepo port.SupplierInventoryRepository,
	matchRepo port.MatchRepository,
	factors []WeightedScoreFactor,
) *RecommendationUseCaseImpl {
	if len(factors) == 0 {
		factors = DefaultScoreFactors()
	}
	return &RecommendationUseCaseImpl{
		offerRepo:     offerRepo,
		requestRepo:   requestRepo,
		inventoryRepo: inventoryRepo,
		matchRepo:     matchRepo,
		factors:       factors,
	}
}

func (uc *RecommendationUseCaseImpl) AvailableQuantity(ctx context.Context, supplierID uuid.UUID, productName string) (float64, error) {
	principal, err := auth.RequirePrincipal(ctx)
	if err != nil {
		return 0, err
	}
	// Only the public read is gated. The internal helper stays ungated because
	// RankOffers resolves it on behalf of the BUYER of a request, who is
	// entitled to see a candidate supplier's stock.
	if supplierID != principal.UserID {
		return 0, domain.ErrForbidden
	}

	return uc.availableQuantity(ctx, supplierID, productName, true)
}

func (uc *RecommendationUseCaseImpl) RankOffers(ctx context.Context, supplyRequestID uuid.UUID) ([]*dto.PrioritizedOfferDTO, error) {
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

	offers, err := uc.offerRepo.ListByRequest(ctx, supplyRequestID)
	if err != nil {
		return nil, err
	}

	ranked := make([]*dto.PrioritizedOfferDTO, 0, len(offers))
	for i := range offers {
		if !offers[i].IsActionable() {
			continue
		}

		available, err := uc.availableQuantity(ctx, offers[i].SupplierID, request.ProductName, false)
		if err != nil {
			return nil, err
		}

		input := primary.OfferScoreInput{
			Offer:             offers[i],
			Request:           request,
			AvailableQuantity: available,
		}

		var score float64
		contributions := make([]dto.ScoreContributionDTO, 0, len(uc.factors))
		for _, weighted := range uc.factors {
			factorScore, err := weighted.Factor.Score(ctx, input)
			if err != nil {
				return nil, err
			}
			weightedScore := weighted.Weight * factorScore
			score += weightedScore
			contributions = append(contributions, dto.ScoreContributionDTO{
				Factor:        weighted.Factor.Name(),
				Weight:        weighted.Weight,
				Score:         factorScore,
				WeightedScore: weightedScore,
			})
		}

		ranked = append(ranked, &dto.PrioritizedOfferDTO{
			// supplyOfferToDTO, not a second copy of it. This file used to hold
			// its own matchOfferToDTO with an identical body, and a copy only
			// fails silently: the ranked offer still serialises, it just carries
			// less than the offer the buyer already saw on the listing endpoint.
			// A price dropped here reads as a broken price score factor, not as a
			// missing field, and that is a much more expensive detour.
			Offer:             *supplyOfferToDTO(&offers[i]),
			Score:             score,
			AvailableQuantity: available,
			Contributions:     contributions,
		})
	}

	sort.Slice(ranked, func(a, b int) bool {
		if ranked[a].Score != ranked[b].Score {
			return ranked[a].Score > ranked[b].Score
		}
		if !ranked[a].Offer.CreatedAt.Equal(ranked[b].Offer.CreatedAt) {
			return ranked[a].Offer.CreatedAt.Before(ranked[b].Offer.CreatedAt)
		}
		return ranked[a].Offer.ID.String() < ranked[b].Offer.ID.String()
	})

	return ranked, nil
}

func (uc *RecommendationUseCaseImpl) availableQuantity(ctx context.Context, supplierID uuid.UUID, productName string, strictNotFound bool) (float64, error) {
	inventory, err := uc.inventoryRepo.FindBySupplierAndProduct(ctx, supplierID, productName)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) && !strictNotFound {
			return 0, nil
		}
		return 0, err
	}

	activeMatches, err := uc.matchRepo.ListActiveBySupplier(ctx, supplierID)
	if err != nil {
		return 0, err
	}

	requests := make(map[uuid.UUID]domain.SupplyRequest)
	var reserved float64
	for i := range activeMatches {
		if !activeMatches[i].IsActive() {
			continue
		}
		if activeMatches[i].AmountUnit != inventory.AmountUnit {
			continue
		}

		request, ok := requests[activeMatches[i].SupplyRequest]
		if !ok {
			request, err = uc.requestRepo.GetByID(ctx, activeMatches[i].SupplyRequest)
			if err != nil {
				if errors.Is(err, domain.ErrNotFound) {
					continue
				}
				return 0, err
			}
			requests[activeMatches[i].SupplyRequest] = request
		}
		if request.ProductName != inventory.ProductName {
			continue
		}

		reserved += activeMatches[i].MatchedAmount
	}

	available := inventory.Quantity - reserved
	if available < 0 {
		return 0, nil
	}
	return available, nil
}

var _ primary.RecommendationUseCase = (*RecommendationUseCaseImpl)(nil)
