package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheImpl struct {
	client *redis.Client
}

func NewCacheImpl(Addr string, Password string, DB int) *CacheImpl {
	return &CacheImpl{

		client: redis.NewClient(&redis.Options{
			Addr:     Addr,
			Password: Password,
			DB:       DB,
		}),
	}
}

func (c *CacheImpl) Get(ctx context.Context, key string, dest any) (bool, error) {

	val, err := c.client.Get(ctx, key).Result()

	if err == redis.Nil {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	json_err := json.Unmarshal([]byte(val), dest)

	if json_err != nil {
		return false, json_err
	}

	return true, nil
}

func (c *CacheImpl) Set(ctx context.Context, key string, value any, ttl time.Duration) error {

	bytes, err := json.Marshal(value)

	if err != nil {
		return err
	}

	set_err := c.client.Set(ctx, key, bytes, ttl).Err()

	if set_err != nil {
		return set_err
	}

	return nil
}

func (c *CacheImpl) Delete(ctx context.Context, key string) error {

	err := c.client.Del(ctx, key).Err()

	if err != nil {
		return err
	}

	return nil
}

func (c *CacheImpl) Remember(ctx context.Context, key string, ttl time.Duration, dest any, loader func() error) (bool, error) {

	found, get_err := c.Get(ctx, key, dest)

	if get_err != nil {
		return false, get_err
	}

	if found {
		return true, nil
	}

	err := loader()

	if err != nil {
		return false, err
	}

	set_err := c.Set(ctx, key, dest, ttl)

	if set_err != nil {
		return false, set_err
	}

	return false, nil
}
