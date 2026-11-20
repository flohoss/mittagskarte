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
			MaxSelect:     1,
			CascadeDelete: true,
			Required:      true,
		})

		chats.AddIndex("idx_telegram_chats_chat_id", true, "chat_id", "")
		chats.AddIndex("idx_telegram_chats_restaurant", true, "restaurant", "")

		if err := app.Save(chats); err != nil {
			return err
		}

		settings := core.NewBaseCollection("telegram_settings")

		settings.Fields.Add(&core.NumberField{
			Name:    "admin_chat_id",
			OnlyInt: true,
			Min:     chatIDMinPtr(),
		})
		settings.ListRule = nil
		settings.ViewRule = nil
		settings.CreateRule = nil
		settings.UpdateRule = nil
		settings.DeleteRule = nil

		return app.Save(settings)
	}, func(app core.App) error {
		settings, err := app.FindCollectionByNameOrId("telegram_settings")
		if err == nil {
			if err := app.Delete(settings); err != nil {
				return err
			}
		}

		chats, err := app.FindCollectionByNameOrId("telegram_chats")
		if err != nil {
			return nil
		}

		return app.Delete(chats)
	})
}
