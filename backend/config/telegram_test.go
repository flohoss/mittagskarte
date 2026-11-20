package config

import (
	"testing"
)

func TestValidateTelegram(t *testing.T) {
	t.Parallel()

	t.Run("unset disables the feature", func(t *testing.T) {
		t.Parallel()

		c := &Config{}
		if err := c.ValidateTelegram(); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if c.TelegramEnabled() {
			t.Fatal("expected TelegramEnabled false")
		}
	})

	t.Run("stray keys without token fail", func(t *testing.T) {
		t.Parallel()

		c := &Config{TelegramSecret: "s"}
		if err := c.ValidateTelegram(); err == nil {
			t.Fatal("expected error for secret without token, got nil")
		}

		c = &Config{TelegramWebhookURL: "https://example.com/api/telegram/abc"}
		if err := c.ValidateTelegram(); err == nil {
			t.Fatal("expected error for webhook URL without token, got nil")
		}
	})

	t.Run("token alone enables the feature in dev", func(t *testing.T) {
		t.Parallel()

		c := &Config{Dev: true, TelegramBotToken: "t"}
		if err := c.ValidateTelegram(); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if !c.TelegramEnabled() {
			t.Fatal("expected TelegramEnabled true")
		}
	})

	t.Run("dev does not require webhook URL", func(t *testing.T) {
		t.Parallel()

		c := &Config{
			Dev:              true,
			TelegramBotToken: "t",
			TelegramSecret:   "s",
		}
		if err := c.ValidateTelegram(); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if !c.TelegramEnabled() {
			t.Fatal("expected TelegramEnabled true")
		}
	})

	t.Run("prod requires secret", func(t *testing.T) {
		t.Parallel()

		c := &Config{
			Dev:                false,
			TelegramBotToken:   "t",
			TelegramWebhookURL: "https://example.com/api/telegram/abc",
		}
		if err := c.ValidateTelegram(); err == nil {
			t.Fatal("expected error for missing webhook secret in prod, got nil")
		}
	})

	t.Run("prod requires webhook URL", func(t *testing.T) {
		t.Parallel()

		c := &Config{
			Dev:              false,
			TelegramBotToken: "t",
			TelegramSecret:   "s",
		}
		if err := c.ValidateTelegram(); err == nil {
			t.Fatal("expected error for missing webhook URL, got nil")
		}
	})

	t.Run("prod rejects non-https webhook URL", func(t *testing.T) {
		t.Parallel()

		c := &Config{
			Dev:                false,
			TelegramBotToken:   "t",
			TelegramSecret:     "s",
			TelegramWebhookURL: "http://example.com",
		}
		if err := c.ValidateTelegram(); err == nil {
			t.Fatal("expected error for non-https webhook URL, got nil")
		}
	})

	t.Run("prod accepts https webhook URL", func(t *testing.T) {
		t.Parallel()

		c := &Config{
			Dev:                false,
			TelegramBotToken:   "t",
			TelegramSecret:     "s",
			TelegramWebhookURL: "https://example.com/api/telegram/abc",
		}
		if err := c.ValidateTelegram(); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})
}
