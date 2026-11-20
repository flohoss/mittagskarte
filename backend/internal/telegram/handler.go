package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strconv"
	"time"

	"github.com/flohoss/mittagskarte/config"
	"github.com/flohoss/mittagskarte/internal/mittag"
	"github.com/flohoss/mittagskarte/internal/restaurant"
	"github.com/flohoss/mittagskarte/pkg/curl"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"ganigeorgiev.com/pocketbase/pocketbase/core"
)

type updateHandler struct {
	app     core.App
	cfg     *config.Config
	logger  *slog.Logger
	scraper *mittag.Scraper
	links   *chatLinks
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

	if msg.Text == "/id" {
		h.reply(ctx, b, msg.Chat.ID, msg.ID, replyIDPrefix+strconv.FormatInt(msg.Chat.ID, 10))
		return
	}

	kind, fileID, err := telegramMedia(msg)
	if err != nil {
		h.reply(ctx, b, msg.Chat.ID, msg.ID, replyUnsupported)
		return
	}

	r, err := h.links.restaurantForChat(msg.Chat.ID)
	if err != nil {
		h.logger.Debug("No restaurant bound to chat", "chatID", msg.Chat.ID)
		h.reply(ctx, b, msg.Chat.ID, msg.ID, replyNotBound)
		return
	}

	if r.Method != restaurant.MethodUpload {
		h.logger.Warn("Restaurant is not open for Telegram uploads", "name", r.Name, "method", r.Method)
		h.reply(ctx, b, msg.Chat.ID, msg.ID, replyNotBound)
		return
	}

	filePath, err := h.downloadMedia(ctx, b, r.ID, kind, fileID)
	if err != nil {
		h.logger.Error("Failed to download Telegram media", "restaurant", r.Name, "error", err)
		h.notify(ctx, b, "😢 Download für %s fehlgeschlagen: %v", r.Name, err)
		h.reply(ctx, b, msg.Chat.ID, msg.ID, replyMenuFailed)
		return
	}

	h.scraper.EnqueueTelegramUpload(r, filePath, func(err error, _ *restaurant.Restaurant) {
		h.reply(ctx, b, msg.Chat.ID, msg.ID, menuReplyFor(err))
		h.notify(ctx, b, "Telegram-Upload für %s: %s", r.Name, menuReplyFor(err))
	})
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

func adminChatID(app core.App) int64 {
	records, err := app.FindRecordsByFilter("telegram_settings", "", "", 1, 0)
	if err != nil || len(records) == 0 {
		return 0
	}
	return int64(records[0].GetInt("admin_chat_id"))
}

func (h *updateHandler) reply(ctx context.Context, b *bot.Bot, chatID int64, messageID int, text string) {
	sendReply(ctx, b, chatID, messageID, text)
}

func (h *updateHandler) notify(ctx context.Context, b *bot.Bot, format string, args ...any) {
	adminChatID := adminChatID(h.app)
	if adminChatID == 0 {
		return
	}
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: adminChatID,
		Text:   fmt.Sprintf(format, args...),
	})
	if err != nil {
		h.logger.Error("Failed to notify admin chat", "error", err)
	}
}
