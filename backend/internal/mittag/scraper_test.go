package mittag

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/flohoss/mittagskarte/internal/restaurant"
	_ "github.com/flohoss/mittagskarte/migrations"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
	"github.com/pocketbase/pocketbase/tools/types"
)

func newTestScraper(t *testing.T, coolDown time.Duration) (*tests.TestApp, *Scraper) {
	t.Helper()

	app := newTestApp(t)
	scraper := NewScraper(app, nil, coolDown)
	t.Cleanup(scraper.Close)
	return app, scraper
}

func createUploadRestaurant(t *testing.T, app core.App, slug string) *restaurant.Restaurant {
	t.Helper()

	collection, err := app.FindCachedCollectionByNameOrId("restaurants")
	if err != nil {
		t.Fatalf("failed to find restaurants collection: %v", err)
	}

	record := core.NewRecord(collection)
	record.Set("name", "Gasthaus "+slug)
	record.Set("slug", slug)
	record.Set("method", "upload")
	record.Set("latitude", 48.5)
	record.Set("longitude", 9.1)
	if err := app.Save(record); err != nil {
		t.Fatalf("failed to save restaurant record: %v", err)
	}
	return restaurant.New(record)
}

func waitForStatus(t *testing.T, s *Scraper, restaurantID, want string) {
	t.Helper()

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if s.StatusForRestaurant(restaurantID) == want {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("restaurant %s never reached status %q", restaurantID, want)
}

func checkLastCheckStatus(t *testing.T, record *core.Record, want restaurant.LastCheckStatus) {
	t.Helper()

	lastCheck, ok := record.Get("last_check").(types.JSONRaw)
	if !ok {
		t.Fatalf("unexpected last_check type: %T", record.Get("last_check"))
	}

	var entry struct {
		Status string `json:"status"`
		Detail string `json:"detail"`
	}
	if err := json.Unmarshal(lastCheck, &entry); err != nil {
		t.Fatalf("failed to decode last_check: %v", err)
	}
	if entry.Status != string(want) {
		t.Fatalf("unexpected last_check status, got %q want %q", entry.Status, want)
	}
}

func TestScraperProcessesUploadMethod(t *testing.T) {
	t.Parallel()

	app, scraper := newTestScraper(t, time.Minute)
	r := createUploadRestaurant(t, app, "upload-processed")

	scraper.Enqueue([]*restaurant.Restaurant{r})
	waitForStatus(t, scraper, r.ID, ScrapeStatusCoolDown)

	reloaded, err := app.FindRecordById("restaurants", r.ID)
	if err != nil {
		t.Fatalf("failed to reload restaurant: %v", err)
	}
	checkLastCheckStatus(t, reloaded, restaurant.LastCheckStatusNotChanged)
}

func TestScraperEnqueueDeduplicatesAndCooldowns(t *testing.T) {
	t.Parallel()

	app, scraper := newTestScraper(t, time.Minute)
	r := createUploadRestaurant(t, app, "deduped")

	scraper.Enqueue([]*restaurant.Restaurant{r})
	waitForStatus(t, scraper, r.ID, ScrapeStatusCoolDown)

	scraper.Enqueue([]*restaurant.Restaurant{r})
	time.Sleep(50 * time.Millisecond)
	if got := scraper.StatusForRestaurant(r.ID); got != ScrapeStatusCoolDown {
		t.Fatalf("expected cooldown after immediate re-enqueue, got %q", got)
	}
}

func TestScraperStatusTransitions(t *testing.T) {
	t.Parallel()

	app, scraper := newTestScraper(t, time.Minute)
	r := createUploadRestaurant(t, app, "transitions")

	if got := scraper.StatusForRestaurant(r.ID); got != ScrapeStatusIdle {
		t.Fatalf("expected idle for unknown restaurant, got %q", got)
	}

	scraper.Enqueue([]*restaurant.Restaurant{r})
	waitForStatus(t, scraper, r.ID, ScrapeStatusCoolDown)
}

func TestScraperEnqueueAfterCloseIsIgnored(t *testing.T) {
	t.Parallel()

	app, scraper := newTestScraper(t, time.Minute)
	r := createUploadRestaurant(t, app, "after-close")

	scraper.Close()
	scraper.Enqueue([]*restaurant.Restaurant{r})

	time.Sleep(50 * time.Millisecond)
	if got := scraper.StatusForRestaurant(r.ID); got != ScrapeStatusIdle {
		t.Fatalf("expected enqueue after close to be ignored, got %q", got)
	}
}

func TestLastCheckFromUploadOnly(t *testing.T) {
	t.Parallel()

	status, detail := restaurant.LastCheckFromError(restaurant.ErrManualUploadOnly)
	if status != restaurant.LastCheckStatusNotChanged {
		t.Fatalf("unexpected status, got %q", status)
	}
	if detail != "" {
		t.Fatalf("unexpected detail, got %q", detail)
	}
}
