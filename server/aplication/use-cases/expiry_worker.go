package usecases

import (
	"context"
	"log"
	"time"

	port "milpa/domain/port/secondary"
)

type ExpiryWorker struct {
	offeringRepo      port.OfferingRepository
	fuzzyRetrival     port.FuzzyRetrival
	cache             port.Cache
	searchInvalidator port.Invalidator
	timer             port.TimeProvider
}

func NewExpiryWorker(offeringRepo port.OfferingRepository, fuzzyRetrival port.FuzzyRetrival, cache port.Cache, searchInvalidator port.Invalidator, timer port.TimeProvider) *ExpiryWorker {
	return &ExpiryWorker{
		offeringRepo:      offeringRepo,
		fuzzyRetrival:     fuzzyRetrival,
		cache:             cache,
		searchInvalidator: searchInvalidator,
		timer:             timer,
	}
}

func (w *ExpiryWorker) Run(ctx context.Context) {
	w.sweep(ctx)

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.sweep(ctx)
		}
	}
}

func (w *ExpiryWorker) sweep(ctx context.Context) {
	expired, err := w.offeringRepo.DeactivateExpired(ctx, w.timer.Now())
	if err != nil {
		log.Printf("expiry worker: deactivate expired offerings: %v", err)
		return
	}

	for _, offering := range expired {
		if err := w.fuzzyRetrival.Delete(ctx, offering.ID.String()); err != nil {
			log.Printf("expiry worker: delete offering %s from search index: %v", offering.ID, err)
		}
		if err := w.cache.Delete(ctx, "offering:"+offering.ID.String()); err != nil {
			log.Printf("expiry worker: purge offering cache for %s: %v", offering.ID, err)
		}
		if err := w.cache.Delete(ctx, "offerings:byuser:"+offering.UserID.String()); err != nil {
			log.Printf("expiry worker: purge catalogue cache for user %s: %v", offering.UserID, err)
		}
	}

	if len(expired) == 0 {
		return
	}

	if err := w.searchInvalidator.InvalidateAll(ctx); err != nil {
		log.Printf("expiry worker: invalidate search cache: %v", err)
	}
}
