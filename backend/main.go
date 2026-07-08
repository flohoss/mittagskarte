package main

import (
	"fmt"
	"log"

	"github.com/flohoss/mittagskarte/api"
	"github.com/flohoss/mittagskarte/config"
	"github.com/flohoss/mittagskarte/internal/mittag"
	_ "github.com/flohoss/mittagskarte/migrations"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	app := pocketbase.NewWithConfig(
		pocketbase.Config{
			DefaultDataDir: "./data",
			DefaultDev:     cfg.Dev,
		},
	)

	app.Cron().SetTimezone(&cfg.Location)

	var mittagService *mittag.Mittag

	app.OnBootstrap().BindFunc(func(e *core.BootstrapEvent) error {
		if err = e.Next(); err != nil {
			return err
		}

		if err = api.SyncSettings(e.App, cfg); err != nil {
			return fmt.Errorf("failed to sync settings: %w", err)
		}
		if err = api.SyncSuperuser(e.App, cfg); err != nil {
			return fmt.Errorf("failed to sync superuser: %w", err)
		}

		mittagService, err = mittag.New(e.App, cfg.CoolDownDuration)
		if err != nil {
			return fmt.Errorf("failed to create mittag service: %w", err)
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

		return api.RegisterRoutes(app, se, cfg)
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
