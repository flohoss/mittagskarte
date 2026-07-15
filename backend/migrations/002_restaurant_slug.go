package migrations

import (
	"github.com/flohoss/mittagskarte/internal/restaurant"
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("restaurants")
		if err != nil {
			return err
		}

		collection.Fields.Add(&core.TextField{
			Name:     "slug",
			Required: false,
			Pattern:  `^[a-z0-9]+(?:-[a-z0-9]+)*$`,
		})

		if err := app.Save(collection); err != nil {
			return err
		}

		records, err := app.FindRecordsByFilter("restaurants", "slug = '' || slug = null", "", 0, 0)
		if err != nil {
			return err
		}

		for _, record := range records {
			record.Set("slug", restaurant.Slugify(record.GetString("name")))
			if err := app.Save(record); err != nil {
				return err
			}
		}

		collection.AddIndex("idx_restaurants_slug", true, "slug", "")

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("restaurants")
		if err != nil {
			return err
		}

		collection.Fields.RemoveByName("slug")
		collection.RemoveIndex("idx_restaurants_slug")

		return app.Save(collection)
	})
}
