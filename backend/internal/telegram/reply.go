package telegram

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/flohoss/mittagskarte/internal/restaurant"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const (
	replyMenuUpdated    = "✅ <b>Menü aktualisiert!</b> Die neue Tageskarte ist jetzt live."
	replyMenuUnchanged  = "👌 Keine Änderungen erkannt – das Menü bleibt wie es ist."
	replyMenuFailed     = "⚠️ Das Menü konnte leider nicht verarbeitet werden. Bitte versuche es später noch einmal."
	replyUnsupported    = "📎 Schick mir die Tageskarte als Foto, Bilddatei oder PDF."
	replyNotBound       = "🔒 Dieser Chat ist noch keinem Restaurant zugeordnet. Bitte melde dich beim Betreiber."
	replyNoUploads      = "🔒 Für diesen Chat ist kein Restaurant mit Telegram-Upload freigeschaltet. Bitte melde dich beim Betreiber."
	replyInactive       = "⏸️ Dieser Chat ist gerade deaktiviert. Bitte melde dich beim Betreiber."
	replyPrivateOnly    = "💬 Schreib mir bitte direkt im privaten Chat – in Gruppen oder Kanälen kann ich nicht arbeiten."
	replyIntro          = "👋 <b>Willkommen!</b>\n\nSchick mir die Tageskarte deines Restaurants einfach als Foto, Bilddatei oder PDF. Ich übernehme den Rest und aktualisiere das Menü automatisch."
	replyIDHint         = "Tippe auf „Kopieren“ und schick die ID dem Betreiber, damit er diesen Chat freischalten kann."
	replyCopyID         = "Kopieren"
	replyRegistered     = "🎉 <b>Alles bereit!</b>\n\nSchick mir die Tageskarte deines Restaurants als Foto, Bilddatei oder PDF."
	replyPickRestaurant = "🍽️ Für welches Restaurant gilt diese Tageskarte?"
	replyMediaExpired   = "⌛ Die Auswahl ist abgelaufen. Schick die Tageskarte noch einmal."
	replyProcessingPick = "⏳ Tageskarte wird verarbeitet …"
)

func replyParameters(messageID int) *models.ReplyParameters {
	return &models.ReplyParameters{
		MessageID:                messageID,
		AllowSendingWithoutReply: true,
	}
}

func introHTML(chatID int64) string {
	return fmt.Sprintf("%s\n\nDeine Chat-ID: <code>%d</code>\n\n%s", replyIntro, chatID, replyIDHint)
}

func chatIDHTML(chatID int64) string {
	return fmt.Sprintf("<code>%d</code>", chatID)
}

func copyIDMarkup(chatID int64) models.InlineKeyboardMarkup {
	return models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{{
			{Text: replyCopyID, CopyText: &models.CopyTextButton{Text: strconv.FormatInt(chatID, 10)}},
		}},
	}
}

func sendReply(ctx context.Context, b *bot.Bot, chatID int64, messageID int, text string) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:          chatID,
		Text:            text,
		ReplyParameters: replyParameters(messageID),
	})
	if err != nil {
		slog.Error("telegram: failed to send reply", "chatID", chatID, "error", err)
	}
}

func menuReplyFor(err error) string {
	switch {
	case err == nil:
		return replyMenuUpdated
	case errors.Is(err, restaurant.ErrMenuUnchanged):
		return replyMenuUnchanged
	default:
		return replyMenuFailed
	}
}
