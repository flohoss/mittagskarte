package telegram

import (
	"context"
	"errors"
	"log/slog"

	"github.com/flohoss/mittagskarte/internal/restaurant"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"ganigeorgiev.com/pocketbase/pocketbase/tools/router"
)

const (
	replyMenuUpdated    = "😋 Menü aktualisiert."
	replyMenuUnchanged  = "🥺 Das Menü hat sich nicht geändert – kein Update nötig."
	replyMenuFailed     = "😖 Das Menü konnte nicht verarbeitet werden. Bitte später erneut versuchen."
	replyUnsupported    = "🫤 Bitte sende mir ein Foto, ein Bild als Datei oder eine PDF-Datei mit der aktuellen Tageskarte."
	replyCooldown       = "🥱 Die Karte wurde gerade erst aktualisiert. Bitte kurz warten und es erneut versuchen."
	replyNotBound       = "🤬 Dieser Chat ist mit keinem Restaurant verknüpft. Bitte wende dich an den Betreiber."
	replyPrivateOnly    = "😶‍🌫️ Bitte sende mir Nachrichten direkt in einem privaten Chat (keine Gruppe, kein Kanal)."
	replyIDPrefix       = "🫣 Deine Chat-ID lautet: "
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
	case isMenuUnchanged(err):
		return replyMenuUnchanged
	default:
		return replyMenuFailed
	}
}

func isMenuUnchanged(err error) bool {
	if errors.Is(err, restaurant.ErrMenuUnchanged) {
		return true
	}
	var apiErr *router.ApiError
	if errors.As(err, &apiErr) {
		return errors.Is(apiErr.RawData().(error), restaurant.ErrMenuUnchanged)
	}
	return false
}
