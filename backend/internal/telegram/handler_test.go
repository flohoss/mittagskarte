package telegram

import (
	"errors"
	"testing"

	"github.com/flohoss/mittagskarte/internal/restaurant"

	"github.com/go-telegram/bot/models"
)

func TestTelegramMedia(t *testing.T) {
	t.Parallel()

	t.Run("largest photo size wins", func(t *testing.T) {
		t.Parallel()

		msg := &models.Message{
			Photo: []models.PhotoSize{
				{FileID: "small", Width: 100, Height: 100},
				{FileID: "medium", Width: 320, Height: 320},
				{FileID: "large", Width: 800, Height: 800},
			},
		}

		kind, fileID, err := telegramMedia(msg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if kind != telegramMediaImage {
			t.Fatalf("unexpected kind, got %q", kind)
		}
		if fileID != "large" {
			t.Fatalf("unexpected fileID, got %q want %q", fileID, "large")
		}
	})

	t.Run("pdf document", func(t *testing.T) {
		t.Parallel()

		msg := &models.Message{
			Document: &models.Document{FileID: "doc1", MimeType: "application/pdf"},
		}

		kind, fileID, err := telegramMedia(msg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if kind != telegramMediaPDF {
			t.Fatalf("unexpected kind, got %q", kind)
		}
		if fileID != "doc1" {
			t.Fatalf("unexpected fileID, got %q", fileID)
		}
	})

	t.Run("image document", func(t *testing.T) {
		t.Parallel()

		msg := &models.Message{
			Document: &models.Document{FileID: "doc2", MimeType: "image/png"},
		}

		kind, fileID, err := telegramMedia(msg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if kind != telegramMediaImage {
			t.Fatalf("unexpected kind, got %q", kind)
		}
		if fileID != "doc2" {
			t.Fatalf("unexpected fileID, got %q", fileID)
		}
	})

	t.Run("unsupported document", func(t *testing.T) {
		t.Parallel()

		msg := &models.Message{
			Document: &models.Document{FileID: "doc3", MimeType: "application/zip"},
		}

		if _, _, err := telegramMedia(msg); err == nil {
			t.Fatal("expected error for unsupported document, got nil")
		}
	})

	t.Run("text only message", func(t *testing.T) {
		t.Parallel()

		msg := &models.Message{Text: "hello"}

		if _, _, err := telegramMedia(msg); err == nil {
			t.Fatal("expected error for text message, got nil")
		}
	})
}

func TestMediaKindDocumentMimeTypePrefix(t *testing.T) {
	t.Parallel()

	msg := &models.Message{
		Document: &models.Document{FileID: "doc4", MimeType: "image/jpeg"},
	}

	kind, fileID, err := telegramMedia(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if kind != telegramMediaImage || fileID != "doc4" {
		t.Fatalf("unexpected result: kind=%q fileID=%q", kind, fileID)
	}
}

func TestMenuReplyFor(t *testing.T) {
	t.Parallel()

	if got := menuReplyFor(nil); got != replyMenuUpdated {
		t.Fatalf("unexpected reply for success, got %q", got)
	}
	if got := menuReplyFor(restaurant.ErrMenuUnchanged); got != replyMenuUnchanged {
		t.Fatalf("unexpected reply for unchanged, got %q", got)
	}
	if got := menuReplyFor(errors.New("boom")); got != replyMenuFailed {
		t.Fatalf("unexpected reply for failure, got %q", got)
	}
}
