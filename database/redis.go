package database

import (
	"context"
	"fmt"
	"tenet-server/config"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

func InitRedis(cfg config.RedisConfig) error {
	RDB = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		DB:   cfg.DB,
	})

	// 测试连接
	ctx := context.Background()
	if err := RDB.Ping(ctx).Err(); err != nil {
		return err
	}
	return nil
}
