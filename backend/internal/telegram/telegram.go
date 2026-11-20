package telegram

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/flohoss/mittagskarte/config"
	"github.com/flohoss/mittagskarte/internal/mittag"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"ganigeorgiev.com/pocketbase/pocketbase/apis"
	"ganigeorgiev.com/pocketbase/pocketbase/core"
)

const webhookPathPrefix = "/api/telegram/"

type Service struct {
	cfg     *config.Config
	logger  *slog.Logger
	bot     *bot.Bot
	handler *updateHandler
}

func New(app core.App, cfg *config.Config, scraper *mittag.Scraper) (*Service, error) {
	handler := &updateHandler{
		app:     app,
		cfg:     cfg,
		logger:  app.Logger().WithGroup("telegram"),
		scraper: scraper,
		links:   &chatLinks{app: app},
	}

	b, err := bot.New(cfg.TelegramBotToken,
		bot.WithDefaultHandler(handler.handle),
		bot.WithWebhookSecretToken(cfg.TelegramSecret),
		bot.WithAllowedUpdates(bot.AllowedUpdates{models.AllowedUpdateMessage}),
	)
	if err != nil {
		return nil, fmt.Errorf("create telegram bot: %w", err)
	}

	return &Service{
		cfg:     cfg,
		logger:  app.Logger().WithGroup("telegram"),
		bot:     b,
		handler: handler,
	}, nil
}

func (s *Service) Start(ctx context.Context) {
	if s.cfg.Dev {
		s.logger.Info("Telegram bot started (long polling)")
		go s.bot.Start(ctx)
		return
	}

	s.logger.Info("Telegram bot started (webhook)", "url", s.cfg.TelegramWebhookURL)
	go s.bot.StartWebhook(ctx)
}

func (s *Service) RegisterWebhookRoute(se *core.ServeEvent) error {
	if s.cfg.Dev {
		return nil
	}

	se.Router.POST(
		webhookPathPrefix+s.cfg.TelegramSecret,
		func(re *core.RequestEvent) error {
			s.bot.WebhookHandler().ServeHTTP(re.Response, re.Request)
			return nil
		},
	).Bind(apis.SkipSuccessActivityLog())

	if _, err := s.bot.SetWebhook(context.Background(), &bot.SetWebhookParams{
		URL:                s.cfg.TelegramWebhookURL,
		SecretToken:        s.cfg.TelegramSecret,
		DropPendingUpdates: true,
	}); err != nil {
		return fmt.Errorf("register telegram webhook: %w", err)
	}

	return nil
}
