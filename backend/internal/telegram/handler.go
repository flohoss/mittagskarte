package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/flohoss/mittagskarte/internal/mittag"
	"github.com/flohoss/mittagskarte/internal/restaurant"
	"github.com/flohoss/mittagskarte/pkg/curl"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type updateHandler struct {
	logger  *slog.Logger
	scraper *mittag.Scraper
	links   *chatLinks
	pending *pendingMedia
}

type telegramMediaKind string

const (
	telegramMediaImage telegramMediaKind = "image"
	telegramMediaPDF   telegramMediaKind = "pdf"

	maxMediaFileSize = 20 << 20
)

func (h *updateHandler) handle(ctx context.Context, b *bot.Bot, update *models.Update) {
	msg := update.Message
	if msg == nil {
		return
	}
	if msg.Chat.Type != models.ChatTypePrivate {
		h.reply(ctx, b, msg.Chat.ID, msg.ID, replyPrivateOnly)
		return
	}

	switch msg.Text {
	case "/id":
		h.sendChatID(ctx, b, msg.Chat.ID)
		return
	case "/start":
		h.greet(ctx, b, msg.Chat.ID, msg.ID)
		return
	}

	kind, fileID, err := telegramMedia(msg)
	if err != nil {
		h.reply(ctx, b, msg.Chat.ID, msg.ID, replyUnsupported)
		return
	}

	binding, err := h.links.bindingForChat(msg.Chat.ID)
	if err != nil {
		h.logger.Debug("No restaurant bound to chat", "chatID", msg.Chat.ID)
		h.reply(ctx, b, msg.Chat.ID, msg.ID, replyNotBound)
		return
	}
	if !binding.isEligible() {
		h.reply(ctx, b, msg.Chat.ID, msg.ID, replyNoUploads)
		return
	}
	binding.filterEligible()

	if len(binding.restaurants) > 1 {
		h.pending.store(msg.Chat.ID, kind, fileID)
		h.askRestaurant(ctx, b, msg.Chat.ID, msg.ID, binding.restaurants)
		return
	}

	h.startUpload(ctx, b, msg.Chat.ID, msg.ID, binding.restaurants[0], kind, fileID)
}

func (h *updateHandler) handleCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	query := update.CallbackQuery
	if query == nil {
		return
	}

	chat, messageRef, ok := callbackChat(query)
	if !ok || chat.Type != models.ChatTypePrivate {
		return
	}

	restaurantID, ok := strings.CutPrefix(query.Data, callbackPrefix)
	if !ok {
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{CallbackQueryID: query.ID})
		return
	}

	r, err := h.links.restaurantForCallback(chat.ID, restaurantID)
	if err != nil {
		h.logger.Warn("Callback addressed an unbound restaurant", "chatID", chat.ID, "restaurantID", restaurantID)
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{CallbackQueryID: query.ID})
		return
	}

	kind, fileID, err := h.pending.take(chat.ID)
	if err != nil {
		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{CallbackQueryID: query.ID, Text: replyMediaExpired, ShowAlert: true})
		return
	}

	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{CallbackQueryID: query.ID})
	h.editMarkupAndStart(ctx, b, chat.ID, messageRef, r, kind, fileID)
}

func telegramMedia(msg *models.Message) (telegramMediaKind, string, error) {
	if len(msg.Photo) > 0 {
		return telegramMediaImage, msg.Photo[len(msg.Photo)-1].FileID, nil
	}
	if d := msg.Document; d != nil {
		if d.MimeType == "application/pdf" {
			return telegramMediaPDF, d.FileID, nil
		}
		for _, prefix := range []string{"image/", "application/pdf"} {
			if len(d.MimeType) >= len(prefix) && d.MimeType[:len(prefix)] == prefix {
				return telegramMediaImage, d.FileID, nil
			}
		}
	}
	return "", "", fmt.Errorf("unsupported message media")
}

func (h *updateHandler) downloadMedia(ctx context.Context, b *bot.Bot, restaurantID string, kind telegramMediaKind, fileID string) (string, error) {
	file, err := b.GetFile(ctx, &bot.GetFileParams{FileID: fileID})
	if err != nil {
		return "", fmt.Errorf("get file metadata: %w", err)
	}

	if file.FileSize > maxMediaFileSize {
		return "", fmt.Errorf("file too large: %d bytes", file.FileSize)
	}

	downloadURL := b.FileDownloadLink(file)
	ext := filepath.Ext(file.FilePath)
	if ext == "" {
		ext = map[telegramMediaKind]string{
			telegramMediaImage: ".jpg",
			telegramMediaPDF:   ".pdf",
		}[kind]
	}

	downloadPath := filepath.Join(
		restaurant.DownloadsFolder,
		fmt.Sprintf("%d_%s%s", time.Now().Unix(), restaurantID, ext),
	)
	if _, err := curl.Download(downloadPath, downloadURL); err != nil {
		return "", fmt.Errorf("download file: %w", err)
	}

	return downloadPath, nil
}

func (h *updateHandler) reply(ctx context.Context, b *bot.Bot, chatID int64, messageID int, text string) {
	sendReply(ctx, b, chatID, messageID, text)
}

func (h *updateHandler) greet(ctx context.Context, b *bot.Bot, chatID int64, messageID int) {
	if h.links.acceptsUploads(chatID) {
		h.sendHTML(ctx, b, chatID, messageID, replyRegistered)
		return
	}
	h.sendHTML(ctx, b, chatID, messageID, introHTML(chatID), copyIDMarkup(chatID))
}

func (h *updateHandler) sendChatID(ctx context.Context, b *bot.Bot, chatID int64) {
	h.sendHTML(ctx, b, chatID, 0, chatIDHTML(chatID), copyIDMarkup(chatID))
}

func (h *updateHandler) sendHTML(ctx context.Context, b *bot.Bot, chatID int64, messageID int, text string, markup ...models.ReplyMarkup) {
	params := &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      text,
		ParseMode: models.ParseModeHTML,
	}
	if messageID != 0 {
		params.ReplyParameters = replyParameters(messageID)
	}
	if len(markup) > 0 {
		params.ReplyMarkup = markup[0]
	}
	if _, err := b.SendMessage(ctx, params); err != nil {
		h.logger.Error("Failed to send message", "chatID", chatID, "error", err)
	}
}

func (h *updateHandler) askRestaurant(ctx context.Context, b *bot.Bot, chatID int64, messageID int, restaurants []*restaurant.Restaurant) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:          chatID,
		Text:            replyPickRestaurant,
		ReplyParameters: replyParameters(messageID),
		ReplyMarkup:     askMarkup(restaurants),
	})
	if err != nil {
		h.logger.Error("Failed to send restaurant picker", "chatID", chatID, "error", err)
	}
}

func (h *updateHandler) startUpload(ctx context.Context, b *bot.Bot, chatID int64, messageID int, r *restaurant.Restaurant, kind telegramMediaKind, fileID string) {
	filePath, err := h.downloadMedia(ctx, b, r.ID, kind, fileID)
	if err != nil {
		h.logger.Error("Failed to download Telegram media", "restaurant", r.Name, "error", err)
		h.reply(ctx, b, chatID, messageID, replyMenuFailed)
		return
	}

	h.scraper.EnqueueTelegramUpload(r, filePath, func(err error, _ *restaurant.Restaurant) {
		h.reply(ctx, b, chatID, messageID, menuReplyFor(err))
	})
}

func (h *updateHandler) editMarkupAndStart(ctx context.Context, b *bot.Bot, chatID int64, ref messageRef, r *restaurant.Restaurant, kind telegramMediaKind, fileID string) {
	if ref.ID != 0 {
		if _, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:    chatID,
			MessageID: ref.ID,
			Text:      replyProcessingPick,
		}); err != nil {
			h.logger.Debug("Failed to edit picker message", "chatID", chatID, "error", err)
		}
	}
	h.startUpload(ctx, b, chatID, ref.ID, r, kind, fileID)
}
