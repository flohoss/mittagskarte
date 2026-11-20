package mittag

import (
	"testing"
	"time"

	"ganigeorgiev.com/pocketbase/pocketbase/core"
	"github.com/flohoss/mittagskarte/internal/restaurant"
)

func createRestaurantWithCron(t *testing.T, app core.App, name, cronExpr string) {
	t.Helper()

	collection, err := app.FindCachedCollectionByNameOrId("restaurants")
	if err != nil {
		t.Fatalf("failed to find restaurants collection: %v", err)
	}

	record := core.NewRecord(collection)
	record.Set("name", name)
	record.Set("slug", name)
	record.Set("method", restaurant.MethodUpload)
	record.Set("latitude", 48.5)
	record.Set("longitude", 9.1)
	record.Set("cron", cronExpr)
	if err := app.Save(record); err != nil {
		t.Fatalf("failed to save restaurant %q: %v", name, err)
	}
}

func TestInitCronRegistersValidAndSkipsInvalidExpressions(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)

	createRestaurantWithCron(t, app, "invalid-cron", "0 99 * * *")
	createRestaurantWithCron(t, app, "valid-cron", "0 8 * * *")

	scraper := NewScraper(app, nil, time.Minute)
	t.Cleanup(scraper.Close)

	m := &Mittag{app: app, logger: app.Logger(), scraper: scraper}

	if err := m.initCron(); err != nil {
		t.Fatalf("initCron returned error: %v", err)
	}

	ids := make(map[string]struct{}, len(app.Cron().Jobs()))
	for _, job := range app.Cron().Jobs() {
		ids[job.Id()] = struct{}{}
	}

	if _, ok := ids["0 8 * * *"]; !ok {
		t.Fatal("expected the valid cron expression to be registered")
	}
	if _, ok := ids["0 99 * * *"]; ok {
		t.Fatal("expected the invalid cron expression to be skipped instead of panicking")
	}
}
