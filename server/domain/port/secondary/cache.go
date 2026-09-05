package port

import (
	"context"
	"time"
)

type Cache interface {
	Get(ctx context.Context, key string, dest any) (bool, error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Remember(ctx context.Context, key string, ttl time.Duration, dest any, loader func() error) (bool, error)
}
