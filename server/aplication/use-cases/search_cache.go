package usecases

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"milpa/aplication/dto"
	"milpa/domain/port/primary"
	port "milpa/domain/port/secondary"
)

const searchCacheTTL = 5 * time.Minute
const searchCachePrefix = "search:"

type CachedSearchUseCase struct {
	next  primary.FuzzyUseCase
	cache port.Cache
}

func NewCachedSearchUseCase(next primary.FuzzyUseCase, cache port.Cache) *CachedSearchUseCase {
	return &CachedSearchUseCase{
		next:  next,
		cache: cache,
	}
}

func (uc *CachedSearchUseCase) Search(ctx context.Context, req dto.SearchRequest) (*dto.SearchResponse, error) {
	key := searchCachePrefix + SearchCacheKey(req)

	var response *dto.SearchResponse

	_, err := uc.cache.Remember(ctx, key, searchCacheTTL, &response, func() error {
		result, err := uc.next.Search(ctx, req)
		if err != nil {
			return err
		}
		response = result
		return nil
	})

	return response, err
}

func (uc *CachedSearchUseCase) InvalidateAll(ctx context.Context) error {
	return uc.cache.DeleteByPrefix(ctx, searchCachePrefix)
}

func SearchCacheKey(req dto.SearchRequest) string {
	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%+v", req))))
	}
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}

var _ primary.FuzzyUseCase = (*CachedSearchUseCase)(nil)
