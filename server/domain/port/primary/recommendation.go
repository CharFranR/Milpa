package primary

import (
	"context"

	"github.com/google/uuid"

	"milpa/aplication/dto"
	domain "milpa/domain/entities"
)

// CheapestPrice and MaxAvailableQuantity are the reference values for the
// candidate set this offer belongs to. They exist because a score has to be
// comparable across offers, and "price" and "quantity" only mean something
// relative to the other candidates in the same request. Both are nil or zero
// when no candidate qualifies, and every factor that reads them treats that as
// "cannot be compared" rather than as zero.
type OfferScoreInput struct {
	Offer                domain.SupplyOffer
	Request              domain.SupplyRequest
	AvailableQuantity    float64
	CheapestPrice        *float64
	MaxAvailableQuantity float64
}

type ScoreFactor interface {
	Name() string
	Score(ctx context.Context, input OfferScoreInput) (float64, error)
}

type RecommendationUseCase interface {
	RankOffers(ctx context.Context, supplyRequestID uuid.UUID) ([]*dto.PrioritizedOfferDTO, error)
	AvailableQuantity(ctx context.Context, supplierID uuid.UUID, productName string) (float64, error)
}
