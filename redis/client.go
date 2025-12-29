package redis

import (
	"context"
	"fmt"
	"time"
	config "usergrowth/configs"

	"github.com/redis/go-redis/v9"
)

var JWTExpireTime time.Duration

type MyRedis struct {
	*redis.Client
}

type Cache interface {
	SetCache(key, value string, expire time.Duration, ctx context.Context) error
	GetCache(key string, ctx context.Context) (string, error)
	Close() error
}

func NewRedis(cfg *config.Config, ctx context.Context) Cache {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Pass,
		DB:       0,
	})
	JWTExpireTime = cfg.JWT.Expire
	_, err := client.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}
	return &MyRedis{
		client,
	}
}

func (rdb *MyRedis) SetCache(key, value string, expired time.Duration, ctx context.Context) error {
	if err := rdb.Set(ctx, key, value, expired).Err(); err != nil {
		return err
	}
	return nil
}

func (rdb *MyRedis) GetCache(key string, ctx context.Context) (string, error) {
	val, err := rdb.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

func (rdb *MyRedis) Close() error {
	err := rdb.Close()
	if err != nil {
		return err
	}
	return nil
}
