package api_test

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/flohoss/mittagskarte/api"
	"github.com/flohoss/mittagskarte/config"
	_ "github.com/flohoss/mittagskarte/migrations"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
	"github.com/pocketbase/pocketbase/tools/filesystem"
)

func testConfig(t *testing.T, dev bool) *config.Config {
	t.Helper()

	appURL, err := url.Parse("https://example.com")
	if err != nil {
		t.Fatalf("failed to parse app url: %v", err)
	}

	return &config.Config{
		Dev:           dev,
		AppURL:        *appURL,
		AppName:       "Mittagskarte",
		SenderName:    "Mittagskarte",
		SenderAddress: "noreply@example.com",
		SMTPHost:      "smtp.example.com",
		SMTPPort:      587,
		SMTPUsername:  "smtp-user",
		SMTPPassword:  "smtp-password",
	}
}

func restaurantWithMenuFactory(t testing.TB) *tests.TestApp {
	t.Helper()

	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatalf("failed to create test app: %v", err)
	}

	relaxMenuHash(t, app)
	createRestaurantWithMenu(t, app, "gasthaus-a")

	return app
}

func relaxMenuHash(t testing.TB, app core.App) {
	t.Helper()

	menus, err := app.FindCachedCollectionByNameOrId("menus")
	if err != nil {
		t.Fatalf("failed to find menus collection: %v", err)
	}

	hashField, ok := menus.Fields.GetByName("hash").(*core.TextField)
	if !ok {
		t.Fatalf("unexpected hash field type: %T", menus.Fields.GetByName("hash"))
	}
	hashField.Required = false
	if err := app.Save(menus); err != nil {
		t.Fatalf("failed to relax hash requirement: %v", err)
	}
}

func createRestaurantWithMenu(t testing.TB, app core.App, slug string) {
	t.Helper()

	restaurants, err := app.FindCachedCollectionByNameOrId("restaurants")
	if err != nil {
		t.Fatalf("failed to find restaurants collection: %v", err)
	}

	record := core.NewRecord(restaurants)
	record.Set("name", "Gasthaus "+slug)
	record.Set("slug", slug)
	record.Set("method", "upload")
	record.Set("latitude", 48.5)
	record.Set("longitude", 9.1)
	if err := app.Save(record); err != nil {
		t.Fatalf("failed to save restaurant: %v", err)
	}

	menus, err := app.FindCachedCollectionByNameOrId("menus")
	if err != nil {
		t.Fatalf("failed to find menus collection: %v", err)
	}

	menuPath := filepath.Join(t.TempDir(), "menu.webp")
	if err := os.WriteFile(menuPath, []byte("menu"), 0o644); err != nil {
		t.Fatalf("failed to write menu fixture: %v", err)
	}
	menuFile, err := filesystem.NewFileFromPath(menuPath)
	if err != nil {
		t.Fatalf("failed to build menu file: %v", err)
	}

	menuRecord := core.NewRecord(menus)
	menuRecord.Set("file", menuFile)
	menuRecord.Set("restaurant", record.Id)
	if err := app.Save(menuRecord); err != nil {
		t.Fatalf("failed to save menu record: %v", err)
	}

	loaded, err := app.FindRecordById("restaurants", record.Id)
	if err != nil {
		t.Fatalf("failed to reload restaurant: %v", err)
	}
	loaded.Set("menus", []string{menuRecord.Id})
	if err := app.Save(loaded); err != nil {
		t.Fatalf("failed to link menu: %v", err)
	}
}

func TestRegisterRoutesHealth(t *testing.T) {
	cfg := testConfig(t, true)

	scenarios := []tests.ApiScenario{
		{
			Name:   "health returns dot",
			Method: http.MethodGet,
			URL:    "/health",
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				if err := api.RegisterRoutes(app, e, cfg); err != nil {
					t.Fatalf("failed to register routes: %v", err)
				}
			},
			ExpectedStatus:  200,
			ExpectedContent: []string{"."},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func TestRegisterRoutesRobots(t *testing.T) {
	cfg := testConfig(t, true)

	scenarios := []tests.ApiScenario{
		{
			Name:   "robots lists sitemap",
			Method: http.MethodGet,
			URL:    "/robots.txt",
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				if err := api.RegisterRoutes(app, e, cfg); err != nil {
					t.Fatalf("failed to register routes: %v", err)
				}
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				"User-agent: *",
				"Allow: /",
				"Sitemap: https://example.com/sitemap.xml",
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func TestRegisterRoutesSitemap(t *testing.T) {
	cfg := testConfig(t, true)

	scenarios := []tests.ApiScenario{
		{
			Name:           "sitemap contains static and restaurant urls",
			Method:         http.MethodGet,
			URL:            "/sitemap.xml",
			TestAppFactory: restaurantWithMenuFactory,
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				if err := api.RegisterRoutes(app, e, cfg); err != nil {
					t.Fatalf("failed to register routes: %v", err)
				}
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				"<urlset",
				"http://www.sitemaps.org/schemas/sitemap/0.9",
				"https://example.com/",
				"https://example.com/datenschutz",
				"https://example.com/restaurants/gasthaus-a",
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func TestRegisterRoutesGzipInProduction(t *testing.T) {
	t.Parallel()

	recorder := performRequest(t, testConfig(t, false), "gzip")

	if got := recorder.Header().Get("Content-Encoding"); got != "gzip" {
		t.Fatalf("unexpected Content-Encoding, got %q want gzip", got)
	}

	body, err := gzip.NewReader(recorder.Body)
	if err != nil {
		t.Fatalf("failed to open gzip reader: %v", err)
	}
	defer body.Close()

	decoded, err := io.ReadAll(body)
	if err != nil {
		t.Fatalf("failed to decode gzip body: %v", err)
	}
	if string(decoded) != "." {
		t.Fatalf("unexpected decoded body, got %q want %q", string(decoded), ".")
	}
}

func TestRegisterRoutesNoGzipInDevelopment(t *testing.T) {
	t.Parallel()

	recorder := performRequest(t, testConfig(t, true), "gzip")

	if got := recorder.Header().Get("Content-Encoding"); got != "" {
		t.Fatalf("unexpected Content-Encoding, got %q want empty", got)
	}
	if got := recorder.Body.String(); got != "." {
		t.Fatalf("unexpected body, got %q want %q", got, ".")
	}
}

func performRequest(t *testing.T, cfg *config.Config, acceptEncoding string) *httptest.ResponseRecorder {
	t.Helper()

	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatalf("failed to create test app: %v", err)
	}
	t.Cleanup(app.Cleanup)

	pbRouter, err := apis.NewRouter(app)
	if err != nil {
		t.Fatalf("failed to create router: %v", err)
	}

	serveEvent := new(core.ServeEvent)
	serveEvent.App = app
	serveEvent.Router = pbRouter

	if err := app.OnServe().Trigger(serveEvent, func(e *core.ServeEvent) error {
		return api.RegisterRoutes(app, e, cfg)
	}); err != nil {
		t.Fatalf("failed to trigger serve event: %v", err)
	}

	mux, err := pbRouter.BuildMux()
	if err != nil {
		t.Fatalf("failed to build mux: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("Accept-Encoding", acceptEncoding)

	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, req)

	return recorder
}
