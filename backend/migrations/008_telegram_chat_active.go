package migrations

import (
	"ganigeorgiev.com/pocketbase/pocketbase/core"
	m "ganigeorgiev.com/pocketbase/pocketbase/migrations"
)

const telegramChatActiveField = "active"

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("telegram_chats")
		if err != nil {
			return err
		}

		collection.Fields.Add(&core.BoolField{
			Name: telegramChatActiveField,
			Help: "Deaktivierte Chats werden ignoriert, die Zuordnung bleibt aber erhalten.",
		})

		if err := app.Save(collection); err != nil {
			return err
		}

		return activateExistingChats(app)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("telegram_chats")
		if err != nil {
			return err
		}

		collection.Fields.RemoveByName(telegramChatActiveField)
		return app.Save(collection)
	})
}

func activateExistingChats(app core.App) error {
	records, err := app.FindAllRecords("telegram_chats")
	if err != nil {
		return err
	}

	for _, record := range records {
		record.Set(telegramChatActiveField, true)
		if err := app.Save(record); err != nil {
			return err
		}
	}

	return nil
}
