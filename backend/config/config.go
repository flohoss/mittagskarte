package config

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/pocketbase/pocketbase/core"
)

type Config struct {
	Dev               bool          `env:"DEV" envDefault:"false"`
	Location          time.Location `env:"TZ,notEmpty" envDefault:"Europe/Berlin"`
	CoolDownDuration  time.Duration `env:"COOL_DOWN_DURATION,notEmpty" envDefault:"5m"`
	ImprintEmail      string        `env:"IMPRINT_EMAIL,notEmpty" envDefault:"contact@example.com"`
	AppName           string        `env:"APP_NAME,notEmpty" envDefault:"Mittagskarte"`
	AppURL            url.URL       `env:"APP_URL,notEmpty" envDefault:"http://localhost:8090"`
	SenderName        string        `env:"SMTP_SENDER_NAME,notEmpty" envDefault:"Mittagskarte"`
	SenderAddress     string        `env:"SMTP_SENDER_ADDRESS,notEmpty" envDefault:"noreply@example.com"`
	SMTPHost          string        `env:"SMTP_HOST"`
	SMTPPort          int           `env:"SMTP_PORT"`
	SMTPUsername      string        `env:"SMTP_USERNAME"`
	SMTPPassword      string        `env:"SMTP_PASSWORD,unset"`
	SuperuserEmail    string        `env:"SUPERUSER_EMAIL,unset"`
	SuperuserPassword string        `env:"SUPERUSER_PASSWORD,unset"`
	SnapOtterURL      url.URL       `env:"SNAPOTTER_URL,notEmpty" envDefault:"http://snapotter:1349"`
}

func parseURL(v string) (any, error) {
	u, err := url.Parse(v)
	if err != nil {
		return nil, fmt.Errorf("unable to parse URL: %w", err)
	}
	if u.Host == "" {
		return nil, fmt.Errorf("host must not be empty")
	}
	return *u, nil
}

func Load() (*Config, error) {
	var cfg Config
	if err := env.ParseWithOptions(&cfg, env.Options{
		FuncMap: map[reflect.Type]env.ParserFunc{
			reflect.TypeFor[url.URL](): parseURL,
		},
	}); err != nil {
		return nil, err
	}
	if err := cfg.ValidateSMTP(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) ValidateSMTP() error {
	if c.Dev {
		return nil
	}
	var errs []error
	if c.SMTPHost == "" {
		errs = append(errs, fmt.Errorf("SMTP_HOST must be set in production"))
	}
	if c.SMTPPort <= 0 || c.SMTPPort > 65535 {
		errs = append(errs, fmt.Errorf("SMTP_PORT must be between 1 and 65535"))
	}
	if c.SMTPUsername == "" {
		errs = append(errs, fmt.Errorf("SMTP_USERNAME must be set in production"))
	}
	if c.SMTPPassword == "" {
		errs = append(errs, fmt.Errorf("SMTP_PASSWORD must be set in production"))
	}
	return errors.Join(errs...)
}

func (c *Config) SMTPSettings() core.SMTPConfig {
	return core.SMTPConfig{
		Enabled:    !c.Dev,
		Host:       c.SMTPHost,
		Port:       c.SMTPPort,
		Username:   c.SMTPUsername,
		Password:   c.SMTPPassword,
		AuthMethod: "PLAIN",
		TLS:        true,
		LocalName:  c.AppURL.Hostname(),
	}
}

func (c *Config) MetaSettings() core.MetaConfig {
	return core.MetaConfig{
		AppName:       c.AppName,
		AppURL:        c.AppURL.String(),
		SenderName:    c.SenderName,
		SenderAddress: c.SenderAddress,
	}
}
