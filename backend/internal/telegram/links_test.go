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
	return bindMulti(t, app, chatID, []string{restaurantID})
}

func bindMulti(t *testing.T, app core.App, chatID int64, restaurantIDs []string) *core.Record {
	t.Helper()

	collection, err := app.FindCachedCollectionByNameOrId("telegram_chats")
	if err != nil {
		t.Fatalf("failed to find telegram_chats collection: %v", err)
	}

	record := core.NewRecord(collection)
	record.Set("chat_id", chatID)
	record.Set("restaurant", restaurantIDs)
	if err := app.Save(record); err != nil {
		t.Fatalf("failed to save telegram chat record: %v", err)
	}
	return record
}

func TestChatLinksSingleBindingResolves(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	links := &chatLinks{app: app}
	record := createRestaurant(t, app, "telegram-linked", restaurant.MethodUpload)
	bindChat(t, app, 12345, record.Id)

	binding, err := links.bindingForChat(12345)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(binding.restaurants) != 1 {
		t.Fatalf("expected 1 restaurant, got %d", len(binding.restaurants))
	}
	if binding.restaurants[0].ID != record.Id {
		t.Fatalf("unexpected restaurant, got %q want %q", binding.restaurants[0].ID, record.Id)
	}
	if binding.restaurants[0].Method != restaurant.MethodUpload {
		t.Fatalf("unexpected method, got %q", binding.restaurants[0].Method)
	}
}

func TestChatLinksMultiBindingPreservesOrder(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	links := &chatLinks{app: app}
	first := createRestaurant(t, app, "multi-first", restaurant.MethodUpload)
	second := createRestaurant(t, app, "multi-second", restaurant.MethodScrape)
	third := createRestaurant(t, app, "multi-third", restaurant.MethodUpload)
	bindChat(t, app, 12345, first.Id)
	record := bindMulti(t, app, 12346, []string{first.Id, second.Id, third.Id})

	binding, err := links.bindingForChat(12346)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(binding.restaurants) != 3 {
		t.Fatalf("expected 3 restaurants on record %s, got %d", record.Id, len(binding.restaurants))
	}
	if binding.restaurants[1].ID != second.Id {
		t.Fatalf("unexpected ordering, got %q want %q", binding.restaurants[1].ID, second.Id)
	}
}

func TestChatLinksUnknownChat(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	links := &chatLinks{app: app}

	if _, err := links.bindingForChat(999); err == nil {
		t.Fatal("expected error for unknown chat, got nil")
	}
}

func TestChatLinksCallbackScopedToChat(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	links := &chatLinks{app: app}
	own := createRestaurant(t, app, "callback-own", restaurant.MethodUpload)
	other := createRestaurant(t, app, "callback-other", restaurant.MethodUpload)
	bindChat(t, app, 12345, own.Id)

	if _, err := links.restaurantForCallback(12345, own.Id); err != nil {
		t.Fatalf("unexpected error for own restaurant: %v", err)
	}
	if _, err := links.restaurantForCallback(12345, other.Id); err == nil {
		t.Fatal("expected error for restaurant outside binding, got nil")
	}
}

func TestChatLinksUniqueChatID(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	first := createRestaurant(t, app, "pair-first", restaurant.MethodUpload)
	second := createRestaurant(t, app, "pair-second", restaurant.MethodUpload)
	row := bindMulti(t, app, 777, []string{first.Id, second.Id})
	if saved := row.Get("restaurant").([]string); len(saved) != 2 {
		t.Fatalf("expected both restaurants on the single binding row, got %d", len(saved))
	}

	collection, err := app.FindCollectionByNameOrId("telegram_chats")
	if err != nil {
		t.Fatalf("failed to load collection: %v", err)
	}

	secondRow := core.NewRecord(collection)
	secondRow.Set("chat_id", 777)
	secondRow.Set("restaurant", second.Id)
	if err := app.Save(secondRow); err == nil {
		t.Fatal("expected error for second row with same chat_id, got nil")
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
