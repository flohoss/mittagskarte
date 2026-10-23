package mittag

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/flohoss/mittagskarte/internal/restaurant"
	_ "github.com/flohoss/mittagskarte/migrations"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
)

const scrapeRouteSlug = "gasthaus-route"

func newScrapeRouteApp(t testing.TB) (*tests.TestApp, *Mittag) {
	t.Helper()

	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatalf("failed to create test app: %v", err)
	}
	t.Cleanup(app.Cleanup)

	collection, err := app.FindCachedCollectionByNameOrId("restaurants")
	if err != nil {
		t.Fatalf("failed to find restaurants collection: %v", err)
	}

	record := core.NewRecord(collection)
	record.Set("name", "Gasthaus Route")
	record.Set("slug", scrapeRouteSlug)
	record.Set("method", "upload")
	record.Set("latitude", 48.5)
	record.Set("longitude", 9.1)
	if err := app.Save(record); err != nil {
		t.Fatalf("failed to save restaurant: %v", err)
	}

	scraper := NewScraper(app, nil, time.Minute)
	t.Cleanup(scraper.Close)

	m := &Mittag{app: app, logger: app.Logger(), scraper: scraper}
	m.bindHooks()

	return app, m
}

func authToken(t testing.TB, app *tests.TestApp) string {
	t.Helper()

	record, err := app.FindAuthRecordByEmail("users", "test@example.com")
	if err != nil {
		t.Fatalf("failed to find auth record: %v", err)
	}

	token, err := record.NewAuthToken()
	if err != nil {
		t.Fatalf("failed to create auth token: %v", err)
	}
	return token
}

func scrapeRouteScenario(t testing.TB, app *tests.TestApp, name, body string) tests.ApiScenario {
	t.Helper()

	return tests.ApiScenario{
		Name:                  name,
		Method:                http.MethodPost,
		URL:                   "/api/restaurants/scrape",
		Body:                  strings.NewReader(body),
		TestAppFactory:        func(t testing.TB) *tests.TestApp { return app },
		DisableTestAppCleanup: true,
		Headers:               map[string]string{"Authorization": authToken(t, app)},
	}
}

func TestHandleScrapeRequiresAuth(t *testing.T) {
	app, _ := newScrapeRouteApp(t)

	scenario := tests.ApiScenario{
		Name:                  "unauthenticated request is rejected",
		Method:                http.MethodPost,
		URL:                   "/api/restaurants/scrape",
		Body:                  strings.NewReader(`{"id":"any"}`),
		TestAppFactory:        func(t testing.TB) *tests.TestApp { return app },
		DisableTestAppCleanup: true,
		ExpectedStatus:        http.StatusUnauthorized,
		ExpectedContent:       []string{"The request requires valid record authorization token."},
	}

	scenario.Test(t)
}

func TestHandleScrapeRejectsMalformedJSON(t *testing.T) {
	app, _ := newScrapeRouteApp(t)

	scenario := scrapeRouteScenario(t, app, "malformed json", `{`)
	scenario.ExpectedStatus = http.StatusBadRequest
	scenario.ExpectedContent = []string{"Ungültiger JSON-Body"}

	scenario.Test(t)
}

func TestHandleScrapeRequiresRestaurantID(t *testing.T) {
	app, _ := newScrapeRouteApp(t)

	scenario := scrapeRouteScenario(t, app, "missing restaurant id", `{"id":"   "}`)
	scenario.ExpectedStatus = http.StatusBadRequest
	scenario.ExpectedContent = []string{"Restaurant-ID ist erforderlich"}

	scenario.Test(t)
}

func TestHandleScrapeUnknownRestaurant(t *testing.T) {
	app, _ := newScrapeRouteApp(t)

	scenario := scrapeRouteScenario(t, app, "unknown restaurant id", `{"id":"doesnotexist123"}`)
	scenario.ExpectedStatus = http.StatusInternalServerError
	scenario.ExpectedContent = []string{"Restaurant konnte nicht geladen werden"}

	scenario.Test(t)
}

func TestHandleScrapeEnqueuesRestaurant(t *testing.T) {
	app, m := newScrapeRouteApp(t)

	record, err := app.FindFirstRecordByFilter("restaurants", "slug = {:slug}", map[string]any{"slug": scrapeRouteSlug})
	if err != nil {
		t.Fatalf("failed to find restaurant: %v", err)
	}

	scenario := scrapeRouteScenario(t, app, "enqueues scrape for existing restaurant", `{"id":"`+record.Id+`"}`)
	scenario.ExpectedStatus = http.StatusOK
	scenario.ExpectedContent = []string{"Aktualisierung für Restaurant " + record.Id + " gestartet"}
	scenario.AfterTestFunc = func(t testing.TB, app *tests.TestApp, res *http.Response) {
		waitForStatus(t, m.scraper, record.Id, ScrapeStatusCoolDown)

		reloaded, err := app.FindRecordById("restaurants", record.Id)
		if err != nil {
			t.Fatalf("failed to reload restaurant: %v", err)
		}
		checkLastCheckStatus(t, reloaded, restaurant.LastCheckStatusNotChanged)
	}

	scenario.Test(t)
}
