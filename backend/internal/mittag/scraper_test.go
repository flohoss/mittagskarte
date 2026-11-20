package mittag

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"ganigeorgiev.com/pocketbase/dbx"
	"ganigeorgiev.com/pocketbase/pocketbase/core"
	"ganigeorgiev.com/pocketbase/pocketbase/tests"
	"ganigeorgiev.com/pocketbase/pocketbase/tools/filesystem"
	"ganigeorgiev.com/pocketbase/pocketbase/tools/types"

	"github.com/flohoss/mittagskarte/internal/restaurant"
	_ "github.com/flohoss/mittagskarte/migrations"
	"github.com/flohoss/mittagskarte/pkg/checksum"
)

func newTestScraper(t *testing.T, coolDown time.Duration) (*tests.TestApp, *Scraper) {
	t.Helper()

	app := newTestApp(t)
	scraper := NewScraper(app, nil, coolDown)
	t.Cleanup(scraper.Close)
	return app, scraper
}

func bindTelegramMenuStubs(t *testing.T, app *tests.TestApp) {
	t.Helper()

	app.OnRecordCreate("menus").BindFunc(func(e *core.RecordEvent) error {
		f, ok := e.Record.Get("file").(*filesystem.File)
		if !ok || f == nil {
			return e.Next()
		}
		rc, err := f.Reader.Open()
		if err != nil {
			return e.Next()
		}
		defer rc.Close()

		hash, err := checksum.Reader(rc)
		if err != nil {
			return e.Next()
		}
		e.Record.Set("hash", hash)
		return e.Next()
	})
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
	record.Set("method", restaurant.MethodUpload)
	record.Set("latitude", 48.5)
	record.Set("longitude", 9.1)
	if err := app.Save(record); err != nil {
		t.Fatalf("failed to save restaurant record: %v", err)
	}
	return restaurant.New(record)
}

func waitForStatus(t testing.TB, s *Scraper, restaurantID, want string) {
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

func checkLastCheckStatus(t testing.TB, record *core.Record, want restaurant.LastCheckStatus) {
	t.Helper()

	raw, ok := record.Get("last_check").(types.JSONRaw)
	if !ok {
		t.Fatalf("unexpected last_check type: %T", record.Get("last_check"))
	}
	if len(raw) == 0 {
		t.Fatalf("last_check is empty (status not yet persisted)")
	}

	var entry struct {
		Status string `json:"status"`
		Detail string `json:"detail"`
	}
	if err := json.Unmarshal(raw, &entry); err != nil {
		t.Fatalf("failed to decode last_check %q: %v", string(raw), err)
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

func TestScraperRecoversFromScrapePanic(t *testing.T) {
	t.Parallel()

	app, scraper := newTestScraper(t, time.Minute)
	panicking := createRestaurantWithMethod(t, app, "panic-restaurant", restaurant.MethodScrape)
	survivor := createUploadRestaurant(t, app, "survivor")

	scraper.Enqueue([]*restaurant.Restaurant{panicking})
	waitForStatus(t, scraper, panicking.ID, ScrapeStatusCoolDown)

	scraper.Enqueue([]*restaurant.Restaurant{survivor})
	waitForStatus(t, scraper, survivor.ID, ScrapeStatusCoolDown)
}

func createRestaurantWithMethod(t *testing.T, app core.App, slug, method string) *restaurant.Restaurant {
	t.Helper()

	collection, err := app.FindCachedCollectionByNameOrId("restaurants")
	if err != nil {
		t.Fatalf("failed to find restaurants collection: %v", err)
	}

	record := core.NewRecord(collection)
	record.Set("name", "Gasthaus "+slug)
	record.Set("slug", slug)
	record.Set("method", method)
	record.Set("latitude", 48.5)
	record.Set("longitude", 9.1)
	if err := app.Save(record); err != nil {
		t.Fatalf("failed to save restaurant record: %v", err)
	}
	return restaurant.New(record)
}

func TestScraperTelegramUploadStoresMenu(t *testing.T) {
	t.Parallel()

	app, scraper := newTestScraper(t, time.Minute)
	bindTelegramMenuStubs(t, app)
	r := createUploadRestaurant(t, app, "telegram-upload")

	filePath := filepath.Join(restaurant.DownloadsFolder, "telegram_test.jpg")
	if err := os.WriteFile(filePath, createTestImage(t), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	done := make(chan error, 1)
	scraper.EnqueueTelegramUpload(r, filePath, func(err error, updated *restaurant.Restaurant) {
		done <- err
	})

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("unexpected upload error: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for telegram upload")
	}

	waitForStatus(t, scraper, r.ID, ScrapeStatusCoolDown)

	menus, err := app.FindRecordsByFilter("menus", "restaurant = {:id}", "-created", 1, 0, dbx.Params{"id": r.ID})
	if err != nil {
		t.Fatalf("failed to query menus: %v", err)
	}
	if len(menus) != 1 {
		t.Fatalf("expected 1 menu, got %d", len(menus))
	}

	reloaded, err := app.FindRecordById("restaurants", r.ID)
	if err != nil {
		t.Fatalf("failed to reload restaurant: %v", err)
	}
	checkLastCheckStatus(t, reloaded, restaurant.LastCheckStatusSuccess)
}

func TestScraperTelegramUploadRespectsCooldown(t *testing.T) {
	t.Parallel()

	app, scraper := newTestScraper(t, time.Minute)
	bindTelegramMenuStubs(t, app)
	r := createUploadRestaurant(t, app, "telegram-cooldown")

	filePath := filepath.Join(restaurant.DownloadsFolder, "telegram_cooldown.jpg")
	if err := os.WriteFile(filePath, createTestImage(t), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	called := make(chan struct{}, 2)
	scraper.EnqueueTelegramUpload(r, filePath, func(error, *restaurant.Restaurant) { called <- struct{}{} })
	waitForStatus(t, scraper, r.ID, ScrapeStatusCoolDown)
	select {
	case <-called:
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for first upload callback")
	}

	scraper.EnqueueTelegramUpload(r, filePath, func(error, *restaurant.Restaurant) { called <- struct{}{} })
	time.Sleep(50 * time.Millisecond)
	select {
	case <-called:
		t.Fatal("expected cooldown to suppress second upload")
	default:
	}
}

func createTestImage(t *testing.T) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("failed to encode test image: %v", err)
	}
	return buf.Bytes()
}
