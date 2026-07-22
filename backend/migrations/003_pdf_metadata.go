package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("menus")
		if err != nil {
			return err
		}

		collection.Fields.Add(&core.JSONField{
			Name: "pdf_metadata",
		})

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("menus")
		if err != nil {
			return err
		}

		collection.Fields.RemoveByName("pdf_metadata")
		return app.Save(collection)
	})
}
