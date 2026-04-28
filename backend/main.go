package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flohoss/mittagskarte/internal/mittag"
	"github.com/flohoss/mittagskarte/internal/restaurant"
	_ "github.com/flohoss/mittagskarte/migrations"

	"github.com/caarlos0/env/v11"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

type config struct {
	Dev              bool          `env:"DEV" envDefault:"false"`
	MaxAmountOfMenus int           `env:"MAX_AMOUNT_OF_MENUS" envDefault:"10"`
	CoolDownDuration time.Duration `env:"COOL_DOWN_DURATION" envDefault:"5m"`
	TZ               time.Location `env:"TZ" envDefault:"UTC"`
}

type frontendPageData struct {
	Restaurants []*restaurant.Restaurant
}

func buildFrontendPageData(app core.App) (*frontendPageData, error) {
	restaurants, err := restaurant.GetRestaurantsWithMenus(app)
	if err != nil {
		return nil, err
	}

	return &frontendPageData{
		Restaurants: restaurants,
	}, nil
}

func serveFrontend(app core.App, se *core.ServeEvent) error {
	frontendDist, err := filepath.Abs("dist")
	if err != nil {
		return err
	}

	if _, err := os.Stat(frontendDist); err != nil {
		// ignore if the dist folder doesn't exist, it will be created by the frontend build process
		return se.Next()
	}

	assetsFS := http.FileServer(http.Dir(frontendDist))

	se.Router.GET("/assets/{path...}", func(re *core.RequestEvent) error {
		http.StripPrefix("/", assetsFS).ServeHTTP(re.Response, re.Request)
		return nil
	}).Bind(apis.SkipSuccessActivityLog())

	se.Router.GET("/static/{path...}", func(re *core.RequestEvent) error {
		http.StripPrefix("/", assetsFS).ServeHTTP(re.Response, re.Request)
		return nil
	}).Bind(apis.SkipSuccessActivityLog())

	indexTemplate, err := template.New("index.html").Funcs(template.FuncMap{
		"toJSON": func(v any) template.JS {
			b, err := json.Marshal(v)
			if err != nil {
				log.Printf("failed to marshal frontend restaurants payload: %v", err)
				return template.JS("[]")
			}
			return template.JS(b)
		},
	}).ParseFiles(filepath.Join(frontendDist, "index.html"))
	if err != nil {
		return err
	}

	se.Router.GET("/{path...}", func(re *core.RequestEvent) error {
		pageData, err := buildFrontendPageData(app)
		if err != nil {
			return err
		}

		re.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
		return indexTemplate.Execute(re.Response, pageData)
	}).Bind(apis.SkipSuccessActivityLog())

	return se.Next()
}

func main() {
	var cfg config
	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}

	app := pocketbase.NewWithConfig(
		pocketbase.Config{
			DefaultDataDir: "./data",
			DefaultDev:     cfg.Dev,
		},
	)
	app.Cron().SetTimezone(&cfg.TZ)

	var mittagService *mittag.Mittag
	var err error

	app.OnBootstrap().BindFunc(func(e *core.BootstrapEvent) error {
		if err = e.Next(); err != nil {
			return err
		}

		mittagService, err = mittag.New(e.App, cfg.MaxAmountOfMenus, cfg.CoolDownDuration)
		if err != nil {
			return err
		}

		return nil
	})

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		if mittagService == nil {
			return nil
		}

		// Use gzip for most routes while skipping the PocketBase Admin UI.
		se.Router.BindFunc(func(e *core.RequestEvent) error {
			path := e.Request.URL.Path
			if strings.HasPrefix(path, "/_/") {
				return e.Next()
			}

			return apis.Gzip().Func(e)
		})

		if err = mittagService.Start(); err != nil {
			return err
		}

		se.Router.GET("/health", func(re *core.RequestEvent) error {
			return re.String(http.StatusOK, ".")
		}).Bind(apis.SkipSuccessActivityLog())

		if err = serveFrontend(app, se); err != nil {
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
