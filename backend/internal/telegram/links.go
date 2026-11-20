package telegram

import (
	"fmt"

	"github.com/flohoss/mittagskarte/internal/restaurant"

	"ganigeorgiev.com/pocketbase/dbx"
	"ganigeorgiev.com/pocketbase/pocketbase/core"
)

type chatLinks struct {
	app core.App
}

func (c *chatLinks) restaurantForChat(chatID int64) (*restaurant.Restaurant, error) {
	records, err := c.app.FindRecordsByFilter(
		"telegram_chats",
		"chat_id = {:chat_id}",
		"",
		1,
		0,
		dbx.Params{"chat_id": chatID},
	)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("no restaurant bound to chat %d", chatID)
	}

	return restaurant.GetRestaurant(c.app, records[0].GetString("restaurant"))
}
