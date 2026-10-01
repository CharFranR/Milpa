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
	if input.MaxAvailableQuantity <= 0 {
		return 0, nil
	}
	return input.AvailableQuantity / input.MaxAvailableQuantity, nil
}

type priceScoreFactor struct{}

func (priceScoreFactor) Name() string {
	return "price"
}

func (priceScoreFactor) Score(_ context.Context, input primary.OfferScoreInput) (float64, error) {
	if input.Offer.PricePerUnit == nil || input.CheapestPrice == nil || *input.CheapestPrice <= 0 {
		return 0, nil
	}
	price := *input.Offer.PricePerUnit
	if price <= 0 {
		return 0, nil
	}
	return *input.CheapestPrice / price, nil
}

type deliveryTimeScoreFactor struct{}

func (deliveryTimeScoreFactor) Name() string {
	return "delivery_time"
}

func (deliveryTimeScoreFactor) Score(_ context.Context, input primary.OfferScoreInput) (float64, error) {
	if !input.Offer.DeliveryAvailable {
		return 0, nil
	}

	deadline := input.Request.DeliveryDeadline
	if deadline.IsZero() {
		return 0, nil
	}

	daysLate := input.Offer.ProposedDeliveryDay.Sub(deadline).Hours() / 24
	if daysLate <= 0 {
		return 1, nil
	}
	return 1 / (1 + daysLate), nil
}

type reputationScoreFactor struct {
	reviewRepo port.ReviewRepository
}

func (reputationScoreFactor) Name() string {
	return "reputation"
}

func (f reputationScoreFactor) Score(ctx context.Context, input primary.OfferScoreInput) (float64, error) {

	if f.reviewRepo == nil {
		return 0.5, nil
	}

	supplierID := input.Offer.SupplierID
	if supplierID == uuid.Nil {
		return 0.5, nil
	}

	average, count, err := f.reviewRepo.AverageRating(ctx, domain.ReviewTargetUser, supplierID)
	if err != nil {
		return 0, err
	}
	if count == 0 {
		return 0.5, nil
	}
	return (average - 1) / 4, nil
}

func AvailabilityScoreFactor() primary.ScoreFactor {
	return availabilityScoreFactor{}
}

func PriceScoreFactor() primary.ScoreFactor {
	return priceScoreFactor{}
}

func DeliveryTimeScoreFactor() primary.ScoreFactor {
	return deliveryTimeScoreFactor{}
}

func ReputationScoreFactor(reviewRepo port.ReviewRepository) primary.ScoreFactor {
	return reputationScoreFactor{reviewRepo: reviewRepo}
}

func DefaultScoreFactors(reviewRepo port.ReviewRepository) []WeightedScoreFactor {
	return []WeightedScoreFactor{
		{Factor: availabilityScoreFactor{}, Weight: 0.35},
		{Factor: reputationScoreFactor{reviewRepo: reviewRepo}, Weight: 0.25},
		{Factor: priceScoreFactor{}, Weight: 0.2},
		{Factor: deliveryTimeScoreFactor{}, Weight: 0.2},
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
		factors = DefaultScoreFactors(nil)
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

	actionable := make([]domain.SupplyOffer, 0, len(offers))
	for i := range offers {
		if offers[i].IsActionable() {
			actionable = append(actionable, offers[i])
		}
	}

	availabilityBySupplier := make(map[uuid.UUID]float64, len(actionable))
	var maxAvailable float64
	var cheapestPrice *float64

	for _, offer := range actionable {
		available, err := uc.availableQuantity(ctx, offer.SupplierID, request.ProductName, false)
		if err != nil {
			return nil, err
		}
		availabilityBySupplier[offer.SupplierID] = available
		if available > maxAvailable {
			maxAvailable = available
		}
		if offer.PricePerUnit != nil && *offer.PricePerUnit > 0 {
			if cheapestPrice == nil || *offer.PricePerUnit < *cheapestPrice {
				price := *offer.PricePerUnit
				cheapestPrice = &price
			}
		}
	}

	ranked := make([]*dto.PrioritizedOfferDTO, 0, len(actionable))
	for _, offer := range actionable {
		input := primary.OfferScoreInput{
			Offer:                offer,
			Request:              request,
			AvailableQuantity:    availabilityBySupplier[offer.SupplierID],
			CheapestPrice:        cheapestPrice,
			MaxAvailableQuantity: maxAvailable,
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
			Offer:             *supplyOfferToDTO(&offer),
			Score:             score,
			AvailableQuantity: input.AvailableQuantity,
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
