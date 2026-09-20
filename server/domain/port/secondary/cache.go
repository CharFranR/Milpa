package port

import (
	"context"
	"time"
)

type Invalidator interface {
	InvalidateAll(ctx context.Context) error
}

type Cache interface {
	Get(ctx context.Context, key string, dest any) (bool, error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	DeleteByPrefix(ctx context.Context, prefix string) error
	Remember(ctx context.Context, key string, ttl time.Duration, dest any, loader func() error) (bool, error)
}
