package restaurant

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
)

const ConfigLocation = "configs/restaurants/"

func parseConfig(path string) (Configuration, error) {
	slog.Info("parsing config", "path", path)
	var config Configuration
	content, err := os.ReadFile(path)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(content, &config)
	if err != nil {
		return config, err
	}
	slog.Info("config successfully parsed", "path", path)
	return config, nil
}

func parseAllConfigs() ([]Configuration, error) {
	var configurations []Configuration
	err := filepath.WalkDir(ConfigLocation, func(path string, info os.DirEntry, err error) error {
		if info.Type().IsRegular() {
			config, err := parseConfig(path)
			if err != nil {
				return err
			}
			configurations = append(configurations, config)
		}
		return nil
	})
	return configurations, err
}
