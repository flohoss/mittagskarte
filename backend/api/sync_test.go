package api_test

import (
	"reflect"
	"testing"

	"github.com/flohoss/mittagskarte/api"
	_ "github.com/flohoss/mittagskarte/migrations"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
)

func newSyncTestApp(t *testing.T) *tests.TestApp {
	t.Helper()

	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatalf("failed to create test app: %v", err)
	}
	t.Cleanup(app.Cleanup)
	return app
}

func superusersCollection(t *testing.T, app core.App) *core.Collection {
	t.Helper()

	collection, err := app.FindCachedCollectionByNameOrId(core.CollectionNameSuperusers)
	if err != nil {
		t.Fatalf("failed to find superusers collection: %v", err)
	}
	return collection
}

func TestSyncSuperuserDisablesOTPAndMFAInDev(t *testing.T) {
	t.Parallel()

	app := newSyncTestApp(t)

	superusers := superusersCollection(t, app)
	superusers.OTP.Enabled = true
	superusers.MFA.Enabled = true
	if err := app.Save(superusers); err != nil {
		t.Fatalf("failed to enable OTP/MFA: %v", err)
	}

	if err := api.SyncSuperuser(app, testConfig(t, true)); err != nil {
		t.Fatalf("SyncSuperuser returned error: %v", err)
	}

	reloaded := superusersCollection(t, app)
	if reloaded.OTP.Enabled || reloaded.MFA.Enabled {
		t.Fatalf("expected OTP/MFA disabled in dev, got otp=%v mfa=%v", reloaded.OTP.Enabled, reloaded.MFA.Enabled)
	}
}

func TestSyncSuperuserEnablesOTPAndMFAInProduction(t *testing.T) {
	t.Parallel()

	app := newSyncTestApp(t)

	if err := api.SyncSuperuser(app, testConfig(t, false)); err != nil {
		t.Fatalf("SyncSuperuser returned error: %v", err)
	}

	reloaded := superusersCollection(t, app)
	if !reloaded.OTP.Enabled || !reloaded.MFA.Enabled {
		t.Fatalf("expected OTP/MFA enabled in production, got otp=%v mfa=%v", reloaded.OTP.Enabled, reloaded.MFA.Enabled)
	}
}

func TestSyncSuperuserSkipsWithoutCredentials(t *testing.T) {
	t.Parallel()

	app := newSyncTestApp(t)

	before, err := app.FindAllRecords(core.CollectionNameSuperusers)
	if err != nil {
		t.Fatalf("failed to list superusers: %v", err)
	}

	cfg := testConfig(t, true)
	cfg.SuperuserEmail = ""
	cfg.SuperuserPassword = ""

	if err := api.SyncSuperuser(app, cfg); err != nil {
		t.Fatalf("SyncSuperuser returned error: %v", err)
	}

	after, err := app.FindAllRecords(core.CollectionNameSuperusers)
	if err != nil {
		t.Fatalf("failed to list superusers: %v", err)
	}
	if len(after) != len(before) {
		t.Fatalf("expected superuser count unchanged, got %d want %d", len(after), len(before))
	}
}

func TestSyncSuperuserCreatesAndUpdatesPassword(t *testing.T) {
	t.Parallel()

	app := newSyncTestApp(t)

	cfg := testConfig(t, true)
	cfg.SuperuserEmail = "bootstrap@example.com"
	cfg.SuperuserPassword = "initial-password"

	if err := api.SyncSuperuser(app, cfg); err != nil {
		t.Fatalf("SyncSuperuser returned error on create: %v", err)
	}

	record, err := app.FindAuthRecordByEmail(core.CollectionNameSuperusers, cfg.SuperuserEmail)
	if err != nil {
		t.Fatalf("failed to find created superuser: %v", err)
	}
	if !record.ValidatePassword(cfg.SuperuserPassword) {
		t.Fatal("expected created superuser to have the configured password")
	}

	cfg.SuperuserPassword = "updated-password"
	if err := api.SyncSuperuser(app, cfg); err != nil {
		t.Fatalf("SyncSuperuser returned error on update: %v", err)
	}

	updated, err := app.FindAuthRecordByEmail(core.CollectionNameSuperusers, cfg.SuperuserEmail)
	if err != nil {
		t.Fatalf("failed to reload superuser: %v", err)
	}
	if updated.Id != record.Id {
		t.Fatalf("expected the same superuser record, got %q want %q", updated.Id, record.Id)
	}
	if !updated.ValidatePassword("updated-password") {
		t.Fatal("expected password to be updated")
	}
}

func TestSyncSettingsInDev(t *testing.T) {
	t.Parallel()

	app := newSyncTestApp(t)

	app.Settings().RateLimits.Enabled = true
	app.Settings().TrustedProxy = core.TrustedProxyConfig{Headers: []string{"X-Forwarded-For"}}
	if err := app.Save(app.Settings()); err != nil {
		t.Fatalf("failed to store initial settings: %v", err)
	}

	cfg := testConfig(t, true)
	if err := api.SyncSettings(app, cfg); err != nil {
		t.Fatalf("SyncSettings returned error: %v", err)
	}

	settings := app.Settings()
	if settings.RateLimits.Enabled {
		t.Fatal("expected rate limits disabled in dev")
	}
	if len(settings.TrustedProxy.Headers) != 0 || settings.TrustedProxy.UseLeftmostIP {
		t.Fatalf("expected empty trusted proxy in dev, got %+v", settings.TrustedProxy)
	}
	if !reflect.DeepEqual(settings.Meta, cfg.MetaSettings()) {
		t.Fatalf("unexpected meta settings, got %+v want %+v", settings.Meta, cfg.MetaSettings())
	}
	if !reflect.DeepEqual(settings.SMTP, cfg.SMTPSettings()) {
		t.Fatalf("unexpected smtp settings, got %+v want %+v", settings.SMTP, cfg.SMTPSettings())
	}
}

func TestSyncSettingsInProduction(t *testing.T) {
	t.Parallel()

	app := newSyncTestApp(t)

	cfg := testConfig(t, false)
	if err := api.SyncSettings(app, cfg); err != nil {
		t.Fatalf("SyncSettings returned error: %v", err)
	}

	settings := app.Settings()
	if !settings.RateLimits.Enabled {
		t.Fatal("expected rate limits enabled in production")
	}
	if !reflect.DeepEqual(settings.TrustedProxy, core.TrustedProxyConfig{Headers: []string{"X-Forwarded-For"}}) {
		t.Fatalf("unexpected trusted proxy, got %+v", settings.TrustedProxy)
	}
	if !settings.SMTP.Enabled {
		t.Fatal("expected SMTP enabled in production")
	}
}

func TestSyncSettingsIsIdempotent(t *testing.T) {
	t.Parallel()

	app := newSyncTestApp(t)
	cfg := testConfig(t, true)

	if err := api.SyncSettings(app, cfg); err != nil {
		t.Fatalf("SyncSettings returned error on first run: %v", err)
	}
	first, err := app.Settings().Clone()
	if err != nil {
		t.Fatalf("failed to clone settings: %v", err)
	}

	if err := api.SyncSettings(app, cfg); err != nil {
		t.Fatalf("SyncSettings returned error on second run: %v", err)
	}

	if !reflect.DeepEqual(app.Settings(), first) {
		t.Fatal("expected repeated sync with the same config to keep settings unchanged")
	}
}
