package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		restaurants, err := app.FindCollectionByNameOrId("restaurants")
		if err != nil {
			return err
		}

		restaurants.Fields.Add(&core.JSONField{
			Name: "last_check",
		})

		return app.Save(restaurants)
	}, func(app core.App) error {
		restaurants, err := app.FindCollectionByNameOrId("restaurants")
		if err != nil {
			return err
		}

		field := restaurants.Fields.GetByName("last_check")
		if field != nil {
			restaurants.Fields.RemoveById(field.GetId())
		}

		return app.Save(restaurants)
	})
}
