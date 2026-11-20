package main

import (
	"context"
	"fmt"
	"log"

	"github.com/flohoss/mittagskarte/api"
	"github.com/flohoss/mittagskarte/config"
	"github.com/flohoss/mittagskarte/internal/mittag"
	"github.com/flohoss/mittagskarte/internal/telegram"
	_ "github.com/flohoss/mittagskarte/migrations"

	"ganigeorgiev.com/pocketbase/pocketbase"
	"ganigeorgiev.com/pocketbase/pocketbase/core"
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

	var (
		mittagService   *mittag.Mittag
		telegramService *telegram.Service
		telegramStop    context.CancelFunc
	)

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

		mittagService, err = mittag.New(e.App, cfg.SnapOtterURL, cfg.CoolDownDuration)
		if err != nil {
			return fmt.Errorf("failed to create mittag service: %w", err)
		}

		if cfg.TelegramEnabled() {
			telegramService, err = telegram.New(e.App, cfg, mittagService.Scraper())
			if err != nil {
				return fmt.Errorf("failed to create telegram service: %w", err)
			}
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

		if telegramService != nil {
			var ctx context.Context
			ctx, telegramStop = context.WithCancel(context.Background())
			telegramService.Start(ctx)

			if err = telegramService.RegisterWebhookRoute(se); err != nil {
				return err
			}
		}

		return api.RegisterRoutes(app, se, cfg)
	})

	app.OnTerminate().BindFunc(func(e *core.TerminateEvent) error {
		if telegramStop != nil {
			telegramStop()
		}
		if mittagService != nil {
			mittagService.Close()
		}

		return e.Next()
	})

	if err = app.Start(); err != nil {
		log.Fatal(err)
	}
}
