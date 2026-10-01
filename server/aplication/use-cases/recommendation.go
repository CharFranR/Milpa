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

// Availability is reported against the largest availability in the same
// candidate set. Returning the raw quantity made this factor incomparable with
// the others: a supplier holding 500 kg scored 500 while every 0..1 factor
// contributed at most its weight, so availability decided the ranking on its
// own and the other weights were decoration.
func (availabilityScoreFactor) Score(_ context.Context, input primary.OfferScoreInput) (float64, error) {
	if input.MaxAvailableQuantity <= 0 {
		return 0, nil
	}
	return input.AvailableQuantity / input.MaxAvailableQuantity, nil
}

// priceScoreFactor ranks the cheapest quoted offer at 1 and halves the score
// for every doubling above it. An offer with no quote scores 0: the column is
// nullable so pre-existing rows survive, and an unquoted offer cannot claim to
// be cheap. That is the opposite of scoring it 0 and having "0 is cheapest"
// break the comparison, which is why the nil is checked before the division.
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

// deliveryTimeScoreFactor scores an offer that lands on or before the request
// deadline at 1, then decays it by one unit per day late. A caller asking for
// a delivery date that has already passed scores 0 and cannot outrank an offer
// that keeps its promise, but a one-day slip still beats a two-day slip.
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

// reputationScoreFactor normalises the 1..5 average onto 0..1. An offer from a
// supplier with no review yet scores the neutral midpoint rather than 0: zero
// is what a genuinely bad rating looks like, and scoring "unknown" the same as
// "terrible" would push every unreviewed farmer below a one-star one.
type reputationScoreFactor struct {
	reviewRepo port.ReviewRepository
}

func (reputationScoreFactor) Name() string {
	return "reputation"
}

func (f reputationScoreFactor) Score(ctx context.Context, input primary.OfferScoreInput) (float64, error) {
	// An unwired repository is the same observable situation as an unreviewed
	// supplier: there is no reputation signal to read. Both score neutral.
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

// DefaultScoreFactors is the RF-12 list: availability, price, delivery time and
// farmer reputation. The weights are a judgement call, not something the brief
// quantifies — availability leads because it is the only factor backed by real
// data today, and reputation is trusted above price because a farmer reputation
// cannot be gamed by a single lowball bid the way price can.
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

	// The reference values every factor normalises against have to be known
	// before the first offer is scored, so the candidates are collected and
	// measured in their own pass. Availability is keyed by supplier because it
	// is a property of (supplier, product name) and every offer in this request
	// carries the same product name: one supplier with two offers is measured
	// once instead of twice.
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
			// supplyOfferToDTO, not a second copy of it. This file used to hold
			// its own matchOfferToDTO with an identical body, and a copy only
			// fails silently: the ranked offer still serialises, it just carries
			// less than the offer the buyer already saw on the listing endpoint.
			// A price dropped here reads as a broken price score factor, not as a
			// missing field, and that is a much more expensive detour.
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
