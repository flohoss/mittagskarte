package config

import (
	"errors"
	"net/url"
	"testing"
	"time"
)

func mustParseURL(t *testing.T, raw string) url.URL {
	t.Helper()

	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("failed to parse URL %q: %v", raw, err)
	}
	return *u
}

func TestParseURL(t *testing.T) {
	t.Parallel()

	t.Run("valid url", func(t *testing.T) {
		t.Parallel()

		got, err := parseURL("http://example.com:8080/path")
		if err != nil {
			t.Fatalf("parseURL returned error: %v", err)
		}
		u, ok := got.(url.URL)
		if !ok {
			t.Fatalf("expected url.URL, got %T", got)
		}
		if u.Host != "example.com:8080" {
			t.Fatalf("unexpected host, got %q want %q", u.Host, "example.com:8080")
		}
		if u.Path != "/path" {
			t.Fatalf("unexpected path, got %q want %q", u.Path, "/path")
		}
	})

	t.Run("invalid url", func(t *testing.T) {
		t.Parallel()

		_, err := parseURL("://bad-url")
		if err == nil {
			t.Fatal("expected error for invalid URL, got nil")
		}
		if !errors.Is(err, err) {
			t.Fatalf("expected wrapped error, got %v", err)
		}
	})

	t.Run("missing host", func(t *testing.T) {
		t.Parallel()

		_, err := parseURL("/just/a/path")
		if err == nil {
			t.Fatal("expected error for missing host, got nil")
		}
	})
}

func TestValidateSMTP(t *testing.T) {
	t.Parallel()

	t.Run("dev mode skips validation", func(t *testing.T) {
		t.Parallel()

		c := &Config{Dev: true}
		if err := c.ValidateSMTP(); err != nil {
			t.Fatalf("expected nil error in dev mode, got %v", err)
		}
	})

	t.Run("production requires all SMTP fields", func(t *testing.T) {
		t.Parallel()

		c := &Config{Dev: false}
		err := c.ValidateSMTP()
		if err == nil {
			t.Fatal("expected error for missing SMTP fields, got nil")
		}
	})

	t.Run("production with valid SMTP", func(t *testing.T) {
		t.Parallel()

		c := &Config{
			Dev:          false,
			SMTPHost:     "smtp.example.com",
			SMTPPort:     587,
			SMTPUsername: "user",
			SMTPPassword: "pass",
		}
		if err := c.ValidateSMTP(); err != nil {
			t.Fatalf("expected nil error for valid SMTP, got %v", err)
		}
	})

	t.Run("port out of range low", func(t *testing.T) {
		t.Parallel()

		c := &Config{
			Dev:          false,
			SMTPHost:     "smtp.example.com",
			SMTPPort:     0,
			SMTPUsername: "user",
			SMTPPassword: "pass",
		}
		err := c.ValidateSMTP()
		if err == nil {
			t.Fatal("expected error for port 0, got nil")
		}
	})

	t.Run("port out of range high", func(t *testing.T) {
		t.Parallel()

		c := &Config{
			Dev:          false,
			SMTPHost:     "smtp.example.com",
			SMTPPort:     70000,
			SMTPUsername: "user",
			SMTPPassword: "pass",
		}
		err := c.ValidateSMTP()
		if err == nil {
			t.Fatal("expected error for port 70000, got nil")
		}
	})

	t.Run("missing host only", func(t *testing.T) {
		t.Parallel()

		c := &Config{
			Dev:          false,
			SMTPPort:     587,
			SMTPUsername: "user",
			SMTPPassword: "pass",
		}
		err := c.ValidateSMTP()
		if err == nil {
			t.Fatal("expected error for missing host, got nil")
		}
	})

	t.Run("missing username only", func(t *testing.T) {
		t.Parallel()

		c := &Config{
			Dev:          false,
			SMTPHost:     "smtp.example.com",
			SMTPPort:     587,
			SMTPPassword: "pass",
		}
		err := c.ValidateSMTP()
		if err == nil {
			t.Fatal("expected error for missing username, got nil")
		}
	})

	t.Run("missing password only", func(t *testing.T) {
		t.Parallel()

		c := &Config{
			Dev:          false,
			SMTPHost:     "smtp.example.com",
			SMTPPort:     587,
			SMTPUsername: "user",
		}
		err := c.ValidateSMTP()
		if err == nil {
			t.Fatal("expected error for missing password, got nil")
		}
	})
}

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

	t.Run("secret without token fails", func(t *testing.T) {
		t.Parallel()

		c := &Config{TelegramSecret: "s"}
		if err := c.ValidateTelegram(); err == nil {
			t.Fatal("expected error for secret without token, got nil")
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

	t.Run("dev does not require secret or https app URL", func(t *testing.T) {
		t.Parallel()

		c := &Config{
			Dev:              true,
			TelegramBotToken: "t",
			TelegramSecret:   "s",
			AppURL:           mustParseURL(t, "http://localhost:8090"),
		}
		if err := c.ValidateTelegram(); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("prod requires secret", func(t *testing.T) {
		t.Parallel()

		c := &Config{
			Dev:              false,
			TelegramBotToken: "t",
			AppURL:           mustParseURL(t, "https://example.com"),
		}
		if err := c.ValidateTelegram(); err == nil {
			t.Fatal("expected error for missing webhook secret in prod, got nil")
		}
	})

	t.Run("prod requires https app URL", func(t *testing.T) {
		t.Parallel()

		c := &Config{
			Dev:              false,
			TelegramBotToken: "t",
			TelegramSecret:   "s",
			AppURL:           mustParseURL(t, "http://example.com"),
		}
		if err := c.ValidateTelegram(); err == nil {
			t.Fatal("expected error for non-https app URL, got nil")
		}
	})

	t.Run("prod accepts https app URL", func(t *testing.T) {
		t.Parallel()

		c := &Config{
			Dev:              false,
			TelegramBotToken: "t",
			TelegramSecret:   "s",
			AppURL:           mustParseURL(t, "https://example.com"),
		}
		if err := c.ValidateTelegram(); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})
}

func TestTelegramWebhookURL(t *testing.T) {
	t.Parallel()

	t.Run("derives from app URL and secret", func(t *testing.T) {
		t.Parallel()

		c := &Config{
			TelegramSecret: "secret",
			AppURL:         mustParseURL(t, "https://mittagskarte.example.com"),
		}
		if got, want := c.TelegramWebhookURL(), "https://mittagskarte.example.com/api/telegram/secret"; got != want {
			t.Fatalf("unexpected webhook URL, got %q want %q", got, want)
		}
	})

	t.Run("ignores path, query and fragment of app URL", func(t *testing.T) {
		t.Parallel()

		c := &Config{
			TelegramSecret: "abc",
			AppURL:         mustParseURL(t, "https://example.com/base?foo=bar#frag"),
		}
		if got, want := c.TelegramWebhookURL(), "https://example.com/api/telegram/abc"; got != want {
			t.Fatalf("unexpected webhook URL, got %q want %q", got, want)
		}
	})
}

func TestValidateBuyMeACoffee(t *testing.T) {
	t.Parallel()

	t.Run("empty URL is allowed", func(t *testing.T) {
		t.Parallel()

		c := &Config{}
		if err := c.ValidateBuyMeACoffee(); err != nil {
			t.Fatalf("unexpected error for empty URL: %v", err)
		}
	})

	t.Run("absolute https URL is allowed", func(t *testing.T) {
		t.Parallel()

		c := &Config{BuyMeACoffeeURL: "https://buymeacoffee.com/flohoss"}
		if err := c.ValidateBuyMeACoffee(); err != nil {
			t.Fatalf("unexpected error for valid URL: %v", err)
		}
	})

	t.Run("relative URL is rejected", func(t *testing.T) {
		t.Parallel()

		c := &Config{BuyMeACoffeeURL: "buymeacoffee.com/flohoss"}
		if err := c.ValidateBuyMeACoffee(); err == nil {
			t.Fatal("expected error for relative URL, got nil")
		}
	})

	t.Run("non-http scheme is rejected", func(t *testing.T) {
		t.Parallel()

		c := &Config{BuyMeACoffeeURL: "javascript:alert(1)"}
		if err := c.ValidateBuyMeACoffee(); err == nil {
			t.Fatal("expected error for non-http scheme, got nil")
		}
	})
}

func TestSMTPSettings(t *testing.T) {
	t.Parallel()

	t.Run("production settings", func(t *testing.T) {
		t.Parallel()

		appURL, _ := url.Parse("https://mittagskarte.example.com")
		c := &Config{
			Dev:          false,
			SMTPHost:     "smtp.example.com",
			SMTPPort:     587,
			SMTPUsername: "user",
			SMTPPassword: "pass",
			AppURL:       *appURL,
		}
		s := c.SMTPSettings()
		if s.Enabled != true {
			t.Fatalf("expected Enabled true, got %v", s.Enabled)
		}
		if s.Host != "smtp.example.com" {
			t.Fatalf("unexpected host, got %q", s.Host)
		}
		if s.Port != 587 {
			t.Fatalf("unexpected port, got %d", s.Port)
		}
		if s.Username != "user" {
			t.Fatalf("unexpected username, got %q", s.Username)
		}
		if s.Password != "pass" {
			t.Fatalf("unexpected password, got %q", s.Password)
		}
		if s.AuthMethod != "PLAIN" {
			t.Fatalf("unexpected auth method, got %q", s.AuthMethod)
		}
		if !s.TLS {
			t.Fatal("expected TLS true")
		}
		if s.LocalName != "mittagskarte.example.com" {
			t.Fatalf("unexpected local name, got %q", s.LocalName)
		}
	})

	t.Run("dev mode disabled", func(t *testing.T) {
		t.Parallel()

		c := &Config{Dev: true}
		s := c.SMTPSettings()
		if s.Enabled {
			t.Fatal("expected Enabled false in dev mode")
		}
	})
}

func TestMetaSettings(t *testing.T) {
	t.Parallel()

	appURL, _ := url.Parse("https://mittagskarte.example.com")
	c := &Config{
		AppName:       "Mittagskarte",
		AppURL:        *appURL,
		SenderName:    "Mittagskarte",
		SenderAddress: "noreply@example.com",
	}
	m := c.MetaSettings()
	if m.AppName != "Mittagskarte" {
		t.Fatalf("unexpected app name, got %q", m.AppName)
	}
	if m.AppURL != "https://mittagskarte.example.com" {
		t.Fatalf("unexpected app URL, got %q", m.AppURL)
	}
	if m.SenderName != "Mittagskarte" {
		t.Fatalf("unexpected sender name, got %q", m.SenderName)
	}
	if m.SenderAddress != "noreply@example.com" {
		t.Fatalf("unexpected sender address, got %q", m.SenderAddress)
	}
}

func TestLoad(t *testing.T) {
	t.Run("dev mode with minimal env", func(t *testing.T) {
		t.Setenv("DEV", "true")
		t.Setenv("TZ", "Europe/Berlin")
		t.Setenv("COOL_DOWN_DURATION", "5m")
		t.Setenv("IMPRINT_EMAIL", "contact@example.com")
		t.Setenv("APP_NAME", "Mittagskarte")
		t.Setenv("APP_URL", "http://localhost:8090")
		t.Setenv("SMTP_SENDER_NAME", "Mittagskarte")
		t.Setenv("SMTP_SENDER_ADDRESS", "noreply@example.com")
		t.Setenv("SNAPOTTER_URL", "http://snapotter:1349")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load returned error: %v", err)
		}
		if !cfg.Dev {
			t.Fatal("expected Dev true")
		}
		if cfg.AppName != "Mittagskarte" {
			t.Fatalf("unexpected app name, got %q", cfg.AppName)
		}
		if cfg.AppURL.Host != "localhost:8090" {
			t.Fatalf("unexpected app URL host, got %q", cfg.AppURL.Host)
		}
		if cfg.CoolDownDuration != 5*time.Minute {
			t.Fatalf("unexpected cool down duration, got %v", cfg.CoolDownDuration)
		}
	})

	t.Run("production without SMTP fails", func(t *testing.T) {
		t.Setenv("DEV", "false")
		t.Setenv("TZ", "Europe/Berlin")
		t.Setenv("COOL_DOWN_DURATION", "5m")
		t.Setenv("IMPRINT_EMAIL", "contact@example.com")
		t.Setenv("APP_NAME", "Mittagskarte")
		t.Setenv("APP_URL", "http://localhost:8090")
		t.Setenv("SMTP_SENDER_NAME", "Mittagskarte")
		t.Setenv("SMTP_SENDER_ADDRESS", "noreply@example.com")
		t.Setenv("SNAPOTTER_URL", "http://snapotter:1349")

		_, err := Load()
		if err == nil {
			t.Fatal("expected error for missing SMTP in production, got nil")
		}
	})

	t.Run("invalid URL fails", func(t *testing.T) {
		t.Setenv("DEV", "true")
		t.Setenv("TZ", "Europe/Berlin")
		t.Setenv("COOL_DOWN_DURATION", "5m")
		t.Setenv("IMPRINT_EMAIL", "contact@example.com")
		t.Setenv("APP_NAME", "Mittagskarte")
		t.Setenv("APP_URL", "://bad-url")
		t.Setenv("SMTP_SENDER_NAME", "Mittagskarte")
		t.Setenv("SMTP_SENDER_ADDRESS", "noreply@example.com")
		t.Setenv("SNAPOTTER_URL", "http://snapotter:1349")

		_, err := Load()
		if err == nil {
			t.Fatal("expected error for invalid URL, got nil")
		}
	})
}
