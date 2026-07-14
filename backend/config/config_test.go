package config

import (
	"errors"
	"net/url"
	"testing"
	"time"
)

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
