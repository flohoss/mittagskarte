package restaurant_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/flohoss/mittagskarte/internal/restaurant"
	_ "github.com/flohoss/mittagskarte/migrations"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
	"github.com/pocketbase/pocketbase/tools/filesystem"
	"github.com/pocketbase/pocketbase/tools/types"
)

func newTestApp(t *testing.T) *tests.TestApp {
	t.Helper()

	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatalf("failed to create test app: %v", err)
	}
	t.Cleanup(app.Cleanup)
	return app
}

func createRestaurantRecord(t *testing.T, app core.App, slug string) *core.Record {
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
	return record
}

func setNavigateSelectors(t *testing.T, app core.App, restaurantID string, selectorIDs ...string) {
	t.Helper()

	record, err := app.FindRecordById("restaurants", restaurantID)
	if err != nil {
		t.Fatalf("failed to load restaurant record: %v", err)
	}
	record.Set("navigate", selectorIDs)
	if err := app.Save(record); err != nil {
		t.Fatalf("failed to save navigate relation: %v", err)
	}
}

func createSelectorRecord(t *testing.T, app core.App, restaurantID string, order int, locator string) *core.Record {
	t.Helper()

	collection, err := app.FindCachedCollectionByNameOrId("selectors")
	if err != nil {
		t.Fatalf("failed to find selectors collection: %v", err)
	}

	record := core.NewRecord(collection)
	record.Set("order", order)
	record.Set("locator", locator)
	record.Set("restaurant", restaurantID)
	if err := app.Save(record); err != nil {
		t.Fatalf("failed to save selector record: %v", err)
	}
	return record
}

func createMenuRecord(t *testing.T, app core.App, restaurantID, hash string) *core.Record {
	t.Helper()

	file, err := pdfFile(t)
	if err != nil {
		t.Fatalf("failed to build menu file: %v", err)
	}

	collection, err := app.FindCachedCollectionByNameOrId("menus")
	if err != nil {
		t.Fatalf("failed to find menus collection: %v", err)
	}

	record := core.NewRecord(collection)
	record.Set("file", file)
	record.Set("hash", hash)
	record.Set("restaurant", restaurantID)
	if err := app.Save(record); err != nil {
		t.Fatalf("failed to save menu record: %v", err)
	}
	return record
}

func pdfFile(t *testing.T) (*filesystem.File, error) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "menu.pdf")
	if err := os.WriteFile(path, []byte("%PDF-1.4\nmenu"), 0o644); err != nil {
		return nil, err
	}
	return filesystem.NewFileFromPath(path)
}

func TestGetRestaurantsWithMenus(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)

	restaurants, err := restaurant.GetRestaurantsWithMenus(app)
	if err != nil {
		t.Fatalf("GetRestaurantsWithMenus returned error: %v", err)
	}
	if len(restaurants) != 0 {
		t.Fatalf("expected no restaurants in fresh app, got %d", len(restaurants))
	}

	first := createRestaurantRecord(t, app, "gasthaus-erster")
	second := createRestaurantRecord(t, app, "gasthaus-zweiter")
	createMenuRecord(t, app, first.Id, "hash-a")
	createMenuRecord(t, app, second.Id, "hash-b")

	restaurants, err = restaurant.GetRestaurantsWithMenus(app)
	if err != nil {
		t.Fatalf("GetRestaurantsWithMenus returned error: %v", err)
	}
	if len(restaurants) != 2 {
		t.Fatalf("expected 2 restaurants, got %d", len(restaurants))
	}
	for _, r := range restaurants {
		if r.Name == "" {
			t.Fatal("expected restaurant name to be mapped")
		}
	}
}

func TestGetRestaurantsOrdersAlphabetically(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)

	createRestaurantRecord(t, app, "b-zweiters")
	createRestaurantRecord(t, app, "a-ersters")

	restaurants, err := restaurant.GetRestaurantsWithMenus(app)
	if err != nil {
		t.Fatalf("GetRestaurantsWithMenus returned error: %v", err)
	}
	if len(restaurants) < 2 {
		t.Fatalf("expected at least 2 restaurants, got %d", len(restaurants))
	}
	if restaurants[0].Slug > restaurants[1].Slug {
		t.Fatalf("expected restaurants sorted by group,name, got %q before %q", restaurants[0].Slug, restaurants[1].Slug)
	}
}

func TestGetRestaurant(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	created := createRestaurantRecord(t, app, "get-restaurant")
	first := createSelectorRecord(t, app, created.Id, 2, "locator-b")
	second := createSelectorRecord(t, app, created.Id, 1, "locator-a")
	setNavigateSelectors(t, app, created.Id, first.Id, second.Id)

	got, err := restaurant.GetRestaurant(app, created.Id)
	if err != nil {
		t.Fatalf("GetRestaurant returned error: %v", err)
	}
	if got.ID != created.Id {
		t.Fatalf("unexpected id, got %q want %q", got.ID, created.Id)
	}
	if got.Slug != "get-restaurant" {
		t.Fatalf("unexpected slug, got %q", got.Slug)
	}
	if len(got.Navigate) != 2 {
		t.Fatalf("expected 2 selectors, got %d", len(got.Navigate))
	}
	if got.Navigate[0].Order > got.Navigate[1].Order {
		t.Fatalf("expected navigate sorted by order, got %d before %d", got.Navigate[0].Order, got.Navigate[1].Order)
	}

	if _, err := restaurant.GetRestaurant(app, "missing-id"); err == nil {
		t.Fatal("expected error for missing restaurant, got nil")
	}
}

func TestGetRestaurantBySlug(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	created := createRestaurantRecord(t, app, "by-slug")

	got, err := restaurant.GetRestaurantBySlug(app, "by-slug")
	if err != nil {
		t.Fatalf("GetRestaurantBySlug returned error: %v", err)
	}
	if got.ID != created.Id {
		t.Fatalf("unexpected id, got %q want %q", got.ID, created.Id)
	}

	if _, err := restaurant.GetRestaurantBySlug(app, "fehlt"); err == nil {
		t.Fatal("expected error for unknown slug, got nil")
	}
}

func TestGetCronGroups(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)

	withCron := createRestaurantRecord(t, app, "mit-cron")
	if err := app.Save(withCron); err != nil {
		t.Fatalf("failed to save record: %v", err)
	}
	loaded, err := app.FindRecordById("restaurants", withCron.Id)
	if err != nil {
		t.Fatalf("failed to reload record: %v", err)
	}
	loaded.Set("cron", "0 11 * * 1-5")
	if err := app.Save(loaded); err != nil {
		t.Fatalf("failed to set cron: %v", err)
	}
	createRestaurantRecord(t, app, "ohne-cron")

	groups, err := restaurant.GetCronGroups(app)
	if err != nil {
		t.Fatalf("GetCronGroups returned error: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("expected 1 cron group, got %d", len(groups))
	}
	group, ok := groups["0 11 * * 1-5"]
	if !ok || len(group) != 1 || group[0].Slug != "mit-cron" {
		t.Fatalf("unexpected cron grouping: %+v", groups)
	}
}

func TestClearExpiredHoliday(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	now := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)

	t.Run("clears expired holiday", func(t *testing.T) {
		t.Parallel()

		record := createRestaurantRecord(t, app, "urlaub-vorbei")
		record.Set("holiday_until", "2026-10-16")
		if err := app.Save(record); err != nil {
			t.Fatalf("failed to set holiday: %v", err)
		}

		if err := restaurant.ClearExpiredHoliday(app, record.Id, now); err != nil {
			t.Fatalf("ClearExpiredHoliday returned error: %v", err)
		}

		reloaded, err := app.FindRecordById("restaurants", record.Id)
		if err != nil {
			t.Fatalf("failed to reload record: %v", err)
		}
		if reloaded.GetString("holiday_until") != "" {
			t.Fatalf("expected holiday to be cleared, got %q", reloaded.GetString("holiday_until"))
		}
	})

	t.Run("keeps running holiday", func(t *testing.T) {
		t.Parallel()

		record := createRestaurantRecord(t, app, "urlaub-laeuft")
		record.Set("holiday_until", "2026-10-31")
		if err := app.Save(record); err != nil {
			t.Fatalf("failed to set holiday: %v", err)
		}

		if err := restaurant.ClearExpiredHoliday(app, record.Id, now); err != nil {
			t.Fatalf("ClearExpiredHoliday returned error: %v", err)
		}

		reloaded, err := app.FindRecordById("restaurants", record.Id)
		if err != nil {
			t.Fatalf("failed to reload record: %v", err)
		}
		if reloaded.GetString("holiday_until") != "2026-10-31" {
			t.Fatalf("expected holiday to be kept, got %q", reloaded.GetString("holiday_until"))
		}
	})

	t.Run("no holiday set", func(t *testing.T) {
		t.Parallel()

		record := createRestaurantRecord(t, app, "kein-urlaub")
		if err := restaurant.ClearExpiredHoliday(app, record.Id, now); err != nil {
			t.Fatalf("ClearExpiredHoliday returned error: %v", err)
		}
	})

	t.Run("missing restaurant", func(t *testing.T) {
		t.Parallel()

		if err := restaurant.ClearExpiredHoliday(app, "does-not-exist", now); err == nil {
			t.Fatal("expected error for missing restaurant, got nil")
		}
	})
}

func TestUpdateLastCheck(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	record := createRestaurantRecord(t, app, "last-check")

	if err := restaurant.UpdateLastCheck(app, record.Id, restaurant.LastCheckStatusError, "kaputt"); err != nil {
		t.Fatalf("UpdateLastCheck returned error: %v", err)
	}

	reloaded, err := app.FindRecordById("restaurants", record.Id)
	if err != nil {
		t.Fatalf("failed to reload record: %v", err)
	}

	var lastCheck struct {
		At     string `json:"at"`
		Status string `json:"status"`
		Detail string `json:"detail"`
	}
	if err := json.Unmarshal(reloaded.Get("last_check").(types.JSONRaw), &lastCheck); err != nil {
		t.Fatalf("failed to decode last_check: %v", err)
	}
	if lastCheck.Status != string(restaurant.LastCheckStatusError) {
		t.Fatalf("unexpected status, got %q want %q", lastCheck.Status, restaurant.LastCheckStatusError)
	}
	if lastCheck.Detail != "kaputt" {
		t.Fatalf("unexpected detail, got %q", lastCheck.Detail)
	}
	if lastCheck.At == "" {
		t.Fatal("expected at timestamp to be set")
	}

	if err := restaurant.UpdateLastCheck(app, "does-not-exist", restaurant.LastCheckStatusSuccess, ""); err == nil {
		t.Fatal("expected error for missing restaurant, got nil")
	}
}

func TestGetLatestMenuByRestaurantID(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	created := createRestaurantRecord(t, app, "latest-menu")

	if restaurant.GetLatestMenuByRestaurantID(app, created.Id) != nil {
		t.Fatal("expected nil latest menu for restaurant without menus")
	}

	first := createMenuRecord(t, app, created.Id, "hash-first")
	second := createMenuRecord(t, app, created.Id, "hash-second")

	latest := restaurant.GetLatestMenuByRestaurantID(app, created.Id)
	if latest == nil {
		t.Fatal("expected latest menu record, got nil")
	}
	if latest.Id != first.Id && latest.Id != second.Id {
		t.Fatalf("unexpected latest menu id %q", latest.Id)
	}
}

func TestUpdateMenu(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	created := createRestaurantRecord(t, app, "menu-update")

	collection, err := app.FindCachedCollectionByNameOrId("menus")
	if err != nil {
		t.Fatalf("failed to find menus collection: %v", err)
	}
	// UpdateMenu leaves the hash to the mittag pipeline, so the app under
	// test must accept menu records without one.
	hashField := collection.Fields.GetByName("hash").(*core.TextField)
	hashField.Required = false
	if err := app.Save(collection); err != nil {
		t.Fatalf("failed to relax hash requirement: %v", err)
	}

	r := restaurant.New(created)

	menuPath := filepath.Join(t.TempDir(), "menu.webp")
	if err := os.WriteFile(menuPath, []byte("fake-webp-bytes"), 0o644); err != nil {
		t.Fatalf("failed to write menu file: %v", err)
	}

	if err := r.UpdateMenu(menuPath, app); err != nil {
		t.Fatalf("UpdateMenu returned error: %v", err)
	}

	menus, err := app.FindRecordsByFilter("menus", "restaurant = {:id}", "", 0, 0, dbx.Params{"id": created.Id})
	if err != nil {
		t.Fatalf("failed to query menus: %v", err)
	}
	if len(menus) != 1 {
		t.Fatalf("expected exactly 1 menu, got %d", len(menus))
	}

	if err := r.UpdateMenu("/does/not/exist.webp", app); err == nil {
		t.Fatal("expected error for missing menu file, got nil")
	}
}
