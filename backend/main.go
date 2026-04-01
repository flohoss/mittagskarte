package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/flohoss/mittagskarte/internal/mittag"
	_ "github.com/flohoss/mittagskarte/migrations"

	"github.com/caarlos0/env/v10"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type config struct {
	Dev bool `env:"DEV" envDefault:"false"`
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

		frontendDist, err := filepath.Abs("dist")
		if err != nil {
			return err
		}

		if _, err := os.Stat(frontendDist); err != nil {
			// ignore if the dist folder doesn't exist, it will be created by the frontend build process
			return se.Next()
		}

		assetsFS := http.FileServer(http.Dir(frontendDist))

		se.Router.GET("/", func(re *core.RequestEvent) error {
			http.ServeFile(re.Response, re.Request, filepath.Join(frontendDist, "index.html"))
			return nil
		})

		se.Router.GET("/assets/{path...}", func(re *core.RequestEvent) error {
			http.StripPrefix("/", assetsFS).ServeHTTP(re.Response, re.Request)
			return nil
		})

		se.Router.GET("/static/{path...}", func(re *core.RequestEvent) error {
			http.StripPrefix("/", assetsFS).ServeHTTP(re.Response, re.Request)
			return nil
		})

		return se.Next()
	})

	var mittagService *mittag.Mittag
	var err error

	app.OnBootstrap().BindFunc(func(e *core.BootstrapEvent) error {
		if err = e.Next(); err != nil {
			return err
		}

		mittagService, err = mittag.New(e.App)
		if err != nil {
			return err
		}

		return nil
	})

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		if mittagService == nil {
			return nil
		}

		if err = mittagService.Start(); err != nil {
			return err
		}

		return se.Next()
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
