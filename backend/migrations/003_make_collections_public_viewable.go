package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		for _, name := range []string{"selector", "restaurants"} {
			collection, err := app.FindCollectionByNameOrId(name)
			if err != nil {
				return err
			}

			collection.ListRule = publicRule()
			collection.ViewRule = publicRule()

			if err := app.Save(collection); err != nil {
				return err
			}
		}

		return nil
	}, func(app core.App) error {
		for _, name := range []string{"selector", "restaurants"} {
			collection, err := app.FindCollectionByNameOrId(name)
			if err != nil {
				continue
			}

			collection.ListRule = nil
			collection.ViewRule = nil

			if err := app.Save(collection); err != nil {
				return err
			}
		}

		return nil
	})
}
