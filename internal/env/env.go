package env

import (
	"errors"
	"os"

	"github.com/caarlos0/env/v8"
	"github.com/go-playground/validator/v10"
)

type Env struct {
	TimeZone     string   `env:"TZ" envDefault:"Etc/UTC" validate:"timezone"`
	Port         int      `env:"PORT" envDefault:"4000" validate:"min=1024,max=49151"`
	PublicUrl    string   `env:"PUBLIC_URL" envDefault:"http://localhost:4000/" validate:"url"`
	LogLevel     string   `env:"LOG_LEVEL" envDefault:"info" validate:"oneof=debug info warn error"`
	AllowedHosts []string `env:"ALLOWED_HOSTS" envDefault:"*" envSeparator:","`
	APIToken     string   `env:"API_TOKEN,required,unset"`
}

var errParse = errors.New("error parsing environment variables")

func Parse() (*Env, error) {
	e := &Env{}
	if err := env.Parse(e); err != nil {
		return e, err
	}
	if err := validateContent(e); err != nil {
		return e, err
	}
	setTZDefaultEnv(e)
	return e, nil
}

func validateContent(e *Env) error {
	validate := validator.New()
	err := validate.Struct(e)
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

func setTZDefaultEnv(e *Env) {
	os.Setenv("TZ", e.TimeZone)
}
