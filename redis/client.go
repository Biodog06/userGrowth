package redis

import (
	"context"
	"fmt"
	"time"
	config "usergrowth/configs"

	"github.com/redis/go-redis/v9"
)

var client *redis.Client
var JWTExpireTime time.Duration

func InitRedis(cfg *config.Config, ctx context.Context) {
	client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Pass,
		DB:       0,
	})
	JWTExpireTime = cfg.JWT.Expire
	_, err := client.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}
}

func SetCache(key, value string, expired time.Duration, ctx context.Context) error {
	if err := client.Set(ctx, key, value, expired).Err(); err != nil {
		return err
	}
	return nil
}

func GetCache(key string, ctx context.Context) (string, error) {
	val, err := client.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}
