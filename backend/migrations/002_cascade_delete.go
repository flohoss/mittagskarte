package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func setCascadeDelete(app core.App, collectionName string) error {
	collection, err := app.FindCollectionByNameOrId(collectionName)
	if err != nil {
		return err
	}

	for _, field := range collection.Fields {
		relation, ok := field.(*core.RelationField)
		if ok && relation.Name == "restaurant" {
			relation.CascadeDelete = true
		}
	}

	return app.Save(collection)
}

func init() {
	m.Register(func(app core.App) error {
		if err := setCascadeDelete(app, "menus"); err != nil {
			return err
		}
		if err := setCascadeDelete(app, "selectors"); err != nil {
			return err
		}
		return nil
	}, func(app core.App) error {
		// Optionally revert to false if needed
		return nil
	})
}