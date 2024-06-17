package imdb

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/v9"
	"gitlab.unjx.de/flohoss/mittag/internal/config"
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

func (imdb *IMDb) SaveMenu(restaurant *config.Restaurant) {
	value, err := json.Marshal(restaurant.Menu)
	if err != nil {
		slog.Error("could not marshal menu", "restaurant", restaurant.Name, "err", err)
		return
	}
	key := restaurant.ID + "-menu"
	menu := string(value)
	slog.Debug("saving menu", "key", key, "value", menu)
	err = imdb.Redis.Set(imdb.CTX, key, value, 0).Err()
	if err != nil {
		slog.Error("could not save menu", "key", key, "value", menu, "err", err)
	}
}

func (imdb *IMDb) RestoreMenu(restaurant *config.Restaurant) {
	key := restaurant.ID + "-menu"
	value, err := imdb.Redis.Get(imdb.CTX, key).Result()
	if err != nil {
		slog.Warn("could not restore menu", "restaurant", restaurant.Name, "err", err)
		return
	}
	err = json.Unmarshal([]byte(value), &restaurant.Menu)
	if err != nil {
		slog.Error("could not unmarshal menu", "restaurant", restaurant.Name, "err", err)
	}
	slog.Debug("menu restored", "key", key, "value", value)
}
