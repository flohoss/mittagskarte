package telegram

import (
	"errors"
	"fmt"

	"github.com/flohoss/mittagskarte/internal/restaurant"

	"ganigeorgiev.com/pocketbase/dbx"
	"ganigeorgiev.com/pocketbase/pocketbase/core"
)

type chatLinks struct {
	app core.App
}

var (
	errChatNotBound = errors.New("chat is not bound to a restaurant")
	errChatInactive = errors.New("chat is deactivated")
)

func (c *chatLinks) bindingForChat(chatID int64) (*chatBinding, error) {
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
		return nil, fmt.Errorf("%w: %d", errChatNotBound, chatID)
	}
	if !records[0].GetBool(activeField) {
		return nil, fmt.Errorf("%w: %d", errChatInactive, chatID)
	}
	return newChatBinding(c.expandRestaurants(records[0].Get("restaurant"))), nil
}

func (c *chatLinks) expandRestaurants(restaurantIDs any) []*restaurant.Restaurant {
	ids, ok := restaurantIDs.([]string)
	if !ok {
		return nil
	}

	restaurants := make([]*restaurant.Restaurant, 0, len(ids))
	for _, id := range ids {
		r, err := restaurant.GetRestaurant(c.app, id)
		if err != nil {
			continue
		}
		restaurants = append(restaurants, r)
	}
	return restaurants
}

func (c *chatLinks) restaurantForCallback(chatID int64, restaurantID string) (*restaurant.Restaurant, error) {
	binding, err := c.bindingForChat(chatID)
	if err != nil {
		return nil, err
	}
	for _, r := range binding.restaurants {
		if r.ID == restaurantID {
			return r, nil
		}
	}
	return nil, fmt.Errorf("restaurant %s is not bound to chat %d", restaurantID, chatID)
}
