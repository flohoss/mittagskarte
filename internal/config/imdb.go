package config

import (
	"encoding/json"
	"log/slog"

	"gitlab.unjx.de/flohoss/mittag/internal/imdb"
)

func (r *Restaurant) SaveMenu(imdb *imdb.IMDb) {
	key := r.ID + "-menu"
	value, err := json.Marshal(r.Menu)
	if err != nil {
		slog.Error("could not marshal menu", "restaurant", r.Name, "err", err)
		return
	}
	menu := string(value)
	err = imdb.Redis.Set(imdb.CTX, key, value, 0).Err()
	if err != nil {
		slog.Error("could not save menu", "key", key, "value", menu, "err", err)
	}
}

func (r *Restaurant) RestoreMenu(imdb *imdb.IMDb) {
	key := r.ID + "-menu"
	value, err := imdb.Redis.Get(imdb.CTX, key).Result()
	if err != nil {
		slog.Warn("could not restore menu", "key", key, "err", err)
		return
	}
	err = json.Unmarshal([]byte(value), &r.Menu)
	if err != nil {
		slog.Error("could not unmarshal menu", "restaurant", r.Name, "err", err)
	}
}
