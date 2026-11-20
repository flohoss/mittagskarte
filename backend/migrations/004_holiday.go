package migrations

import (
	"ganigeorgiev.com/pocketbase/pocketbase/core"
	m "ganigeorgiev.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("restaurants")
		if err != nil {
			return err
		}

		collection.Fields.Add(&core.TextField{
			Name:    "holiday_until",
			Pattern: `^\d{4}-\d{2}-\d{2}$`,
		})

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("restaurants")
		if err != nil {
			return err
		}

		collection.Fields.RemoveByName("holiday_until")
		return app.Save(collection)
	})
}
