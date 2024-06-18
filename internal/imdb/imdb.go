package imdb

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// in-memory database (IMDb
type IMDb struct {
	Redis *redis.Client
	CTX   context.Context
}

func New(host string, port int) *IMDb {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: "",
		DB:       0,
	})
	return &IMDb{
		Redis: rdb,
		CTX:   context.Background(),
	}
}
