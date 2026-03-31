package main

import (
	"log"
	"net/http"

	"github.com/flohoss/mittagskarte/internal/mittag"

	"github.com/caarlos0/env/v10"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type config struct {
	Dev    bool   `env:"DEV" envDefault:"true"`
	Domain string `env:"DOMAIN,required" envDefault:"localhost:5173"`
}

func main() {
	var cfg config
	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}

	app := pocketbase.NewWithConfig(
		pocketbase.Config{
			DefaultDataDir: "data/pb",
			DefaultDev:     cfg.Dev,
		},
	)

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.GET("/health", func(re *core.RequestEvent) error {
			return re.String(http.StatusOK, ".")
		})

		return se.Next()
	})

	var mittagService *mittag.Mittag
	var err error

	app.OnBootstrap().BindFunc(func(e *core.BootstrapEvent) error {
		if err = e.Next(); err != nil {
			return err
		}

		mittagService, err = mittag.New(e.App, cfg.Domain)
		if err != nil {
			return err
		}

		return nil
	})

	app.OnTerminate().BindFunc(func(e *core.TerminateEvent) error {
		if mittagService != nil {
			mittagService.Close()
		}

		return e.Next()
	})

	if err = app.Start(); err != nil {
		log.Fatal(err)
	}
}
