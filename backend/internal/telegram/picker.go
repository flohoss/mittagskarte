package telegram

import (
	"fmt"
	"sync"
	"time"

	"github.com/flohoss/mittagskarte/internal/restaurant"

	"github.com/go-telegram/bot/models"
)

const callbackPrefix = "tmenu:"
const pendingTTL = 10 * time.Minute

type messageRef struct {
	ID int
}

func callbackChat(query *models.CallbackQuery) (models.Chat, messageRef, bool) {
	m := query.Message
	if m.Type == models.MaybeInaccessibleMessageTypeMessage && m.Message != nil {
		return m.Message.Chat, messageRef{ID: m.Message.ID}, true
	}
	if m.Type == models.MaybeInaccessibleMessageTypeInaccessibleMessage && m.InaccessibleMessage != nil {
		return m.InaccessibleMessage.Chat, messageRef{}, true
	}
	return models.Chat{}, messageRef{}, false
}

type chatBinding struct {
	restaurants []*restaurant.Restaurant
}

func newChatBinding(restaurants []*restaurant.Restaurant) *chatBinding {
	uploadable := make([]*restaurant.Restaurant, 0, len(restaurants))
	for _, r := range restaurants {
		if r.Method == restaurant.MethodUpload {
			uploadable = append(uploadable, r)
		}
	}
	return &chatBinding{restaurants: uploadable}
}

func (b *chatBinding) canAcceptUploads() bool {
	return len(b.restaurants) > 0
}

type pendingMedia struct {
	mu      sync.Mutex
	entries map[int64]pendingEntry
}

type pendingEntry struct {
	fileID   string
	kind     telegramMediaKind
	expireAt time.Time
}

func newPendingMedia() *pendingMedia {
	return &pendingMedia{entries: map[int64]pendingEntry{}}
}

func (p *pendingMedia) store(chatID int64, kind telegramMediaKind, fileID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.cleanup()
	p.entries[chatID] = pendingEntry{fileID: fileID, kind: kind, expireAt: time.Now().Add(pendingTTL)}
}

func (p *pendingMedia) take(chatID int64) (telegramMediaKind, string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.cleanup()

	entry, ok := p.entries[chatID]
	if !ok {
		return "", "", fmt.Errorf("no pending media for chat %d", chatID)
	}
	delete(p.entries, chatID)
	return entry.kind, entry.fileID, nil
}

func (p *pendingMedia) cleanup() {
	now := time.Now()
	for chatID, entry := range p.entries {
		if now.After(entry.expireAt) {
			delete(p.entries, chatID)
		}
	}
}

func askMarkup(restaurants []*restaurant.Restaurant) models.InlineKeyboardMarkup {
	rows := make([][]models.InlineKeyboardButton, 0, len(restaurants))
	for _, r := range restaurants {
		rows = append(rows, []models.InlineKeyboardButton{{
			Text:         r.Name,
			CallbackData: callbackPrefix + r.ID,
		}})
	}
	return models.InlineKeyboardMarkup{InlineKeyboard: rows}
}
