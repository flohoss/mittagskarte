package telegram

import (
	"testing"

	"github.com/flohoss/mittagskarte/internal/restaurant"
)

func TestPendingMediaRoundtrip(t *testing.T) {
	t.Parallel()

	pending := newPendingMedia()
	pending.store(1, telegramMediaImage, "file-1")

	kind, fileID, err := pending.take(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if kind != telegramMediaImage || fileID != "file-1" {
		t.Fatalf("unexpected entry: kind=%q fileID=%q", kind, fileID)
	}

	if _, _, err := pending.take(1); err == nil {
		t.Fatal("expected error after take, got nil")
	}
}

func TestPendingMediaOverwrite(t *testing.T) {
	t.Parallel()

	pending := newPendingMedia()
	pending.store(1, telegramMediaImage, "first")
	pending.store(1, telegramMediaPDF, "second")

	kind, fileID, err := pending.take(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if kind != telegramMediaPDF || fileID != "second" {
		t.Fatalf("expected latest media, got kind=%q fileID=%q", kind, fileID)
	}
}

func TestChatBindingEligibility(t *testing.T) {
	t.Parallel()

	binding := newChatBinding([]*restaurant.Restaurant{
		{ID: "a", Method: restaurant.MethodUpload},
		{ID: "b", Method: restaurant.MethodScrape},
	})

	if !binding.canAcceptUploads() {
		t.Fatal("expected binding with one upload restaurant to accept uploads")
	}
	if len(binding.restaurants) != 1 || binding.restaurants[0].ID != "a" {
		t.Fatalf("expected only the upload restaurant to remain, got %+v", binding.restaurants)
	}

	noUpload := newChatBinding([]*restaurant.Restaurant{
		{ID: "b", Method: restaurant.MethodScrape},
	})
	if noUpload.canAcceptUploads() {
		t.Fatal("expected binding without upload restaurants to reject uploads")
	}

	if newChatBinding(nil).canAcceptUploads() {
		t.Fatal("expected empty binding to reject uploads")
	}
}

func TestAskMarkup(t *testing.T) {
	t.Parallel()

	markup := askMarkup([]*restaurant.Restaurant{
		{ID: "id-a", Name: "Gasthaus A"},
		{ID: "id-b", Name: "Gasthaus B"},
	})

	if len(markup.InlineKeyboard) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(markup.InlineKeyboard))
	}
	if markup.InlineKeyboard[0][0].Text != "Gasthaus A" {
		t.Fatalf("unexpected button text, got %q", markup.InlineKeyboard[0][0].Text)
	}
	if markup.InlineKeyboard[0][0].CallbackData != callbackPrefix+"id-a" {
		t.Fatalf("unexpected callback data, got %q", markup.InlineKeyboard[0][0].CallbackData)
	}
}
