package env

import (
	"log"

	"github.com/caarlos0/env/v8"
)

type Config struct {
	TimeZone string `env:"TZ" envDefault:"Europe/Berlin"`
	Port     int    `env:"PORT" envDefault:"8080"`
	LogLevel string `env:"LOG_LEVEL" envDefault:"info"`
	AdminKey string `env:"ADMIN_KEY" envDefault:"admin"`
}

func Parse() *Config {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		log.Fatalln(err.Error())
	}
	return cfg
}
