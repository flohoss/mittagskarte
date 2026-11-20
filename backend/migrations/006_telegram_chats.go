package migrations

import (
	"ganigeorgiev.com/pocketbase/pocketbase/core"
	m "ganigeorgiev.com/pocketbase/pocketbase/migrations"
)

const chatIDMin = float64(1)

func chatIDMinPtr() *float64 {
	v := chatIDMin
	return &v
}

func init() {
	m.Register(func(app core.App) error {
		restaurants, err := app.FindCollectionByNameOrId("restaurants")
		if err != nil {
			return err
		}

		chats := core.NewBaseCollection("telegram_chats")

		chats.Fields.Add(&core.NumberField{
			Name:     "chat_id",
			Required: true,
			OnlyInt:  true,
			Min:      chatIDMinPtr(),
		})
		chats.Fields.Add(&core.RelationField{
			Name:          "restaurant",
			CollectionId:  restaurants.Id,
			MaxSelect:     10,
			CascadeDelete: true,
			Required:      true,
		})

		chats.AddIndex("idx_telegram_chats_chat_id", true, "chat_id", "")

		return app.Save(chats)
	}, func(app core.App) error {
		chats, err := app.FindCollectionByNameOrId("telegram_chats")
		if err != nil {
			return nil
		}

		return app.Delete(chats)
	})
}
