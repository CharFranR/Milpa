package primary

import (
	"context"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	domain "milpa/domain/entities"
)

type OfferScoreInput struct {
	Offer             domain.SupplyOffer
	Request           domain.SupplyRequest
	AvailableQuantity float64
}

type ScoreFactor interface {
	Name() string
	Score(ctx context.Context, input OfferScoreInput) (float64, error)
}

type RecommendationUseCase interface {
	RankOffers(ctx context.Context, supplyRequestID uuid.UUID) ([]*dto.PrioritizedOfferDTO, error)
	AvailableQuantity(ctx context.Context, supplierID uuid.UUID, productName string) (float64, error)
}
