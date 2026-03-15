package database

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

func NewRedis(redisURL string) (*redis.Client, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*1e9) // 5s
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	log.Info().Msg("connected to redis")
	return client, nil
}
