package telegram

import (
	"testing"

	"ganigeorgiev.com/pocketbase/pocketbase/core"
	"ganigeorgiev.com/pocketbase/pocketbase/tests"

	"github.com/flohoss/mittagskarte/internal/restaurant"
	_ "github.com/flohoss/mittagskarte/migrations"
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

func createRestaurant(t *testing.T, app core.App, slug, method string) *core.Record {
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
	return record
}

func bindChat(t *testing.T, app core.App, chatID int64, restaurantID string) *core.Record {
	t.Helper()

	collection, err := app.FindCachedCollectionByNameOrId("telegram_chats")
	if err != nil {
		t.Fatalf("failed to find telegram_chats collection: %v", err)
	}

	record := core.NewRecord(collection)
	record.Set("chat_id", chatID)
	record.Set("restaurant", restaurantID)
	if err := app.Save(record); err != nil {
		t.Fatalf("failed to save telegram chat record: %v", err)
	}
	return record
}

func TestChatLinksRestaurantForChat(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	links := &chatLinks{app: app}
	record := createRestaurant(t, app, "telegram-linked", restaurant.MethodUpload)
	bindChat(t, app, 12345, record.Id)

	r, err := links.restaurantForChat(12345)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.ID != record.Id {
		t.Fatalf("unexpected restaurant, got %q want %q", r.ID, record.Id)
	}
	if r.Method != restaurant.MethodUpload {
		t.Fatalf("unexpected method, got %q", r.Method)
	}
}

func TestChatLinksUnknownChat(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	links := &chatLinks{app: app}

	if _, err := links.restaurantForChat(999); err == nil {
		t.Fatal("expected error for unknown chat, got nil")
	}
}

func TestChatLinksUniqueChatID(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	first := createRestaurant(t, app, "first", restaurant.MethodUpload)
	second := createRestaurant(t, app, "second", restaurant.MethodUpload)
	bindChat(t, app, 777, first.Id)

	_, err := app.FindCachedCollectionByNameOrId("telegram_chats")
	if err != nil {
		t.Fatalf("failed to find collection: %v", err)
	}

	collection, err := app.FindCollectionByNameOrId("telegram_chats")
	if err != nil {
		t.Fatalf("failed to load collection: %v", err)
	}

	duplicate := core.NewRecord(collection)
	duplicate.Set("chat_id", 777)
	duplicate.Set("restaurant", second.Id)
	if err := app.Save(duplicate); err == nil {
		t.Fatal("expected error for duplicate chat_id, got nil")
	}
}

func TestChatLinksRejectsGroupChat(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	restaurantRecord := createRestaurant(t, app, "group-guard", restaurant.MethodUpload)

	collection, err := app.FindCollectionByNameOrId("telegram_chats")
	if err != nil {
		t.Fatalf("failed to find collection: %v", err)
	}

	record := core.NewRecord(collection)
	record.Set("chat_id", -100200300)
	record.Set("restaurant", restaurantRecord.Id)
	if err := app.Save(record); err == nil {
		t.Fatal("expected error for negative group chat id, got nil")
	}
}

func TestAdminChatIDFromSettings(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)

	if got := adminChatID(app); got != 0 {
		t.Fatalf("expected 0 without settings record, got %d", got)
	}

	collection, err := app.FindCollectionByNameOrId("telegram_settings")
	if err != nil {
		t.Fatalf("failed to find telegram_settings collection: %v", err)
	}

	record := core.NewRecord(collection)
	record.Set("admin_chat_id", 12345)
	if err := app.Save(record); err != nil {
		t.Fatalf("failed to save settings: %v", err)
	}

	if got := adminChatID(app); got != 12345 {
		t.Fatalf("expected admin chat id 12345, got %d", got)
	}

	record.Set("admin_chat_id", 0)
	if err := app.Save(record); err != nil {
		t.Fatalf("failed to clear settings: %v", err)
	}
	if got := adminChatID(app); got != 0 {
		t.Fatalf("expected 0 after clearing, got %d", got)
	}
}

func TestAdminChatIDRejectsNegative(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)

	collection, err := app.FindCollectionByNameOrId("telegram_settings")
	if err != nil {
		t.Fatalf("failed to find telegram_settings collection: %v", err)
	}

	record := core.NewRecord(collection)
	record.Set("admin_chat_id", -100)
	if err := app.Save(record); err == nil {
		t.Fatal("expected error for negative admin chat id, got nil")
	}
}
