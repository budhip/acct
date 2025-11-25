package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	xlog "bitbucket.org/Amartha/go-x/log"

	"github.com/redis/go-redis/v9"
)

type CacheRepository interface {
	Ping(ctx context.Context) error
	SetIfNotExists(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, keys ...string) error
	GetIncrement(ctx context.Context, key string) (int64, error)
	GetAll(ctx context.Context, keyPattern string) (map[string]string, error)
	DeleteKeysWithPrefix(ctx context.Context, prefix string) (err error)
}

type cacheClient struct {
	redis *redis.Client
}

const logMessage = "[CACHE]"

func NewCacheRepository(redis *redis.Client) CacheRepository {
	return &cacheClient{redis: redis}
}

func GetOrSet[T any](cc CacheRepository, opts models.GetOrSetCacheOpts[T]) (value T, err error) {
	if opts.Callback == nil {
		return value, fmt.Errorf("missing callback from options")
	}

	var notExists bool
	ctx := opts.Ctx
	rawVal, err := cc.Get(ctx, opts.Key)
	if err != nil {
		if errors.Is(err, models.GetErrMap(models.ErrKeyDataNotFound)) {
			err = nil
			notExists = true
		} else {
			err = nil
			notExists = true
			xlog.Error(ctx, logMessage, xlog.Err(err))
		}
	}

	if notExists {
		value, err = opts.Callback()
		if err != nil {
			return value, err
		}

		mVal, err := json.Marshal(value)
		if err != nil {
			return value, models.GetErrMap(models.ErrKeyFailedMarshal, err.Error())
		}

		err = cc.Set(ctx, opts.Key, mVal, opts.TTL)
		if err != nil {
			xlog.Warn(ctx, logMessage, xlog.Err(err))
			return value, nil
		}

		return value, nil
	}

	if err = json.Unmarshal([]byte(rawVal), &value); err != nil {
		return value, fmt.Errorf("unable to unmarshal data: %w", err)
	}

	return value, nil
}

func (cc *cacheClient) Ping(ctx context.Context) error {
	_, err := cc.redis.Ping(ctx).Result()
	return err
}

func (cc *cacheClient) SetIfNotExists(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	return cc.redis.SetNX(ctx, key, value, ttl).Result()
}

func (cc *cacheClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return cc.redis.Set(ctx, key, value, ttl).Err()
}

func (cc *cacheClient) Get(ctx context.Context, key string) (string, error) {
	val, err := cc.redis.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return val, models.GetErrMap(models.ErrKeyDataNotFound)
		}
		return val, models.GetErrMap(models.ErrKeyFailedGetFromCache, err.Error())
	}
	val = strings.TrimSpace(val)

	return val, nil
}

func (cc *cacheClient) Del(ctx context.Context, keys ...string) error {
	return cc.redis.Del(ctx, keys...).Err()
}

func (cc *cacheClient) GetIncrement(ctx context.Context, key string) (int64, error) {
	return cc.redis.Incr(ctx, key).Result()
}

func (cc *cacheClient) GetAll(ctx context.Context, keyPattern string) (map[string]string, error) {
	result := make(map[string]string)
	iter := cc.redis.Scan(ctx, 0, keyPattern, 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		val, err := cc.Get(ctx, key)
		if err != nil {
			return nil, err
		}
		result[key] = val
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (cc *cacheClient) DeleteKeysWithPrefix(ctx context.Context, prefix string) (err error) {
	var cursor uint64
	for {
		keys, newCursor, err := cc.redis.Scan(ctx, cursor, prefix, 100).Result()
		if err != nil {
			return err
		}

		if len(keys) > 0 {
			if err := cc.redis.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}

		cursor = newCursor
		if cursor == 0 {
			break
		}
	}

	return nil
}
