package api

import (
	"net/http"
	"reflect"
	"strings"

	"ganigeorgiev.com/pocketbase/pocketbase/apis"
	"ganigeorgiev.com/pocketbase/pocketbase/core"

	"github.com/flohoss/mittagskarte/config"
	"github.com/flohoss/mittagskarte/internal/sitemap"
)

func SyncSuperuser(app core.App, cfg *config.Config) error {
	superusers, err := app.FindCachedCollectionByNameOrId(core.CollectionNameSuperusers)
	if err != nil {
		return err
	}
	if superusers.OTP.Enabled != !cfg.Dev || superusers.MFA.Enabled != !cfg.Dev {
		superusers.OTP.Enabled = !cfg.Dev
		superusers.MFA.Enabled = !cfg.Dev
		if err := app.Save(superusers); err != nil {
			return err
		}
	}
	if cfg.SuperuserEmail == "" || cfg.SuperuserPassword == "" {
		return nil
	}

	record, err := app.FindAuthRecordByEmail(superusers, cfg.SuperuserEmail)
	if err != nil {
		record = core.NewRecord(superusers)
		record.SetEmail(cfg.SuperuserEmail)
		record.SetPassword(cfg.SuperuserPassword)
		return app.Save(record)
	}

	return updateSuperuserPasswordIfChanged(app, record, cfg.SuperuserPassword)
}

func updateSuperuserPasswordIfChanged(app core.App, record *core.Record, password string) error {
	if record.ValidatePassword(password) {
		return nil
	}

	record.SetPassword(password)
	return app.Save(record)
}

func SyncSettings(app core.App, cfg *config.Config) error {
	settings := app.Settings()
	changed := false

	meta := cfg.MetaSettings()
	if !reflect.DeepEqual(settings.Meta, meta) {
		settings.Meta = meta
		changed = true
	}

	smtp := cfg.SMTPSettings()
	if !reflect.DeepEqual(settings.SMTP, smtp) {
		settings.SMTP = smtp
		changed = true
	}

	var trustedProxy core.TrustedProxyConfig
	if !cfg.Dev {
		trustedProxy = core.TrustedProxyConfig{
			Headers:       []string{"X-Forwarded-For"},
			UseLeftmostIP: false,
		}
	}
	if !reflect.DeepEqual(settings.TrustedProxy, trustedProxy) {
		settings.TrustedProxy = trustedProxy
		changed = true
	}

	wantRateLimits := !cfg.Dev
	if settings.RateLimits.Enabled != wantRateLimits {
		settings.RateLimits.Enabled = wantRateLimits
		changed = true
	}

	if !changed {
		return nil
	}
	return app.Save(settings)
}

func RegisterRoutes(app core.App, se *core.ServeEvent, cfg *config.Config) error {
	if !cfg.Dev {
		gzip := apis.Gzip()
		se.Router.BindFunc(func(e *core.RequestEvent) error {
			path := e.Request.URL.Path
			if strings.HasPrefix(path, "/_/") || strings.HasPrefix(path, "/api/realtime") {
				return e.Next()
			}
			return gzip.Func(e)
		})
	}

	healthRoute := se.Router.GET("/health", func(re *core.RequestEvent) error {
		return re.String(http.StatusOK, ".")
	})
	sitemapRoute := se.Router.GET("/sitemap.xml", func(re *core.RequestEvent) error {
		body, err := sitemap.Build(app, cfg.AppURL.String())
		if err != nil {
			return err
		}
		return re.XML(http.StatusOK, body)
	})
	robotsRoute := se.Router.GET("/robots.txt", func(re *core.RequestEvent) error {
		re.Response.Header().Set("Content-Type", "text/plain; charset=utf-8")
		return re.String(http.StatusOK, sitemap.Robots(cfg.AppURL.String()))
	})
	if !cfg.Dev {
		healthRoute.Bind(apis.SkipSuccessActivityLog())
		sitemapRoute.Bind(apis.SkipSuccessActivityLog())
		robotsRoute.Bind(apis.SkipSuccessActivityLog())
	}

	return ServeFrontend(app, se, FrontendConfig{
		ImprintEmail:    cfg.ImprintEmail,
		BuyMeACoffeeURL: cfg.BuyMeACoffeeURL,
		Dev:             cfg.Dev,
	})
}
