package env

import (
	"errors"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/gommon/log"
)

type Config struct {
	TimeZone   string `env:"TZ" envDefault:"Etc/UTC" validate:"timezone"`
	Port       int    `env:"PORT" envDefault:"4000" validate:"min=1024,max=49151"`
	PublicUrl  string `env:"PUBLIC_URL" envDefault:"http://localhost:4000/" validate:"url"`
	LogLevel   string `env:"LOG_LEVEL" envDefault:"info" validate:"oneof=debug info warn error"`
	APIToken   string `env:"API_TOKEN,required,unset"`
	APPVersion string `env:"APP_VERSION" envDefault:"dev"`
}

var errParse = errors.New("error parsing environment variables")

var logLevels = map[string]log.Lvl{
	"debug": 1,
	"info":  2,
	"warn":  3,
	"error": 4,
}

func Parse() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return cfg, err
	}
	if err := validateContent(cfg); err != nil {
		return cfg, err
	}
	setTZDefaultEnv(cfg)
	return cfg, nil
}

func (cfg *Config) GetLogLevel() log.Lvl {
	level := logLevels[cfg.LogLevel]
	return level
}

func validateContent(cfg *Config) error {
	validate := validator.New()
	err := validate.Struct(cfg)
	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return err
		} else {
			for _, err := range err.(validator.ValidationErrors) {
				return err
			}
		}
		return errParse
	}
	return nil
}

func setTZDefaultEnv(e *Config) {
	os.Setenv("TZ", e.TimeZone)
}
