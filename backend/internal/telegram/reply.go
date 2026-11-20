package telegram

import (
	"context"
	"errors"
	"log/slog"

	"github.com/flohoss/mittagskarte/internal/restaurant"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const (
	replyMenuUpdated    = "😋 Menü aktualisiert."
	replyMenuUnchanged  = "🥺 Das Menü hat sich nicht geändert – kein Update nötig."
	replyMenuFailed     = "😖 Das Menü konnte nicht verarbeitet werden. Bitte später erneut versuchen."
	replyUnsupported    = "🫤 Bitte sende mir ein Foto, ein Bild als Datei oder eine PDF-Datei mit der aktuellen Tageskarte."
	replyNotBound       = "🤬 Dieser Chat ist mit keinem Restaurant verknüpft. Bitte wende dich an den Betreiber."
	replyPrivateOnly    = "😶‍🌫️ Bitte sende mir Nachrichten direkt in einem privaten Chat (keine Gruppe, kein Kanal)."
	replyIDHint         = "Bitte teile diese ID dem Betreiber, damit er diesen Chat aktivieren kann."
	replyIntro          = "👋 <b>Willkommen!</b>\n\nSende mir die Tageskarte deines Restaurants einfach als Foto, Bild oder PDF. Ich aktualisiere sie dann automatisch."
	replyIDOnly         = "Hier ist deine Chat-ID:"
	replyPickRestaurant = "🤔 Für welches Restaurant ist die Karte?"
	replyMediaExpired   = "🫥 Die Auswahl ist abgelaufen. Bitte sende die Karte erneut."
	replyProcessingPick = "⏳ Karte wird verarbeitet …"
)

func sendReply(ctx context.Context, b *bot.Bot, chatID int64, messageID int, text string) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
		ReplyParameters: &models.ReplyParameters{
			MessageID:                messageID,
			AllowSendingWithoutReply: true,
		},
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
