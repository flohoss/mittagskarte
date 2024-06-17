package config

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
)

const ConfigLocation = "internal/config/restaurants/"

type Config struct {
	Restaurants map[string]*Restaurant
}

func New() *Config {
	return &Config{
		Restaurants: parseConfigFiles(),
	}
}

func parseConfigFile(path string) (Restaurant, error) {
	var restaurant Restaurant

	content, err := os.ReadFile(path)
	if err != nil {
		return restaurant, err
	}
	err = json.Unmarshal(content, &restaurant)
	if err != nil {
		return restaurant, err
	}

	slog.Debug("config successfully parsed", "path", path)
	return restaurant, nil
}

func parseConfigFiles() map[string]*Restaurant {
	restaurants := make(map[string]*Restaurant)

	filepath.WalkDir(ConfigLocation, func(path string, info os.DirEntry, err error) error {
		if info.Type().IsRegular() {
			config, err := parseConfigFile(path)
			if err != nil {
				slog.Debug("error parsing config file", "path", path, "err", err)
				return filepath.SkipDir
			}
			restaurants[config.ID] = &config
		}
		return nil
	})

	return restaurants
}
