package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

const (
	restaurantsCollectionName = "restaurants"
	idxRestaurantsGroupName   = "idx_restaurants_group_name"
	idxRestaurantsMethod      = "idx_restaurants_method"
)

func init() {
	m.Register(func(app core.App) error {
		restaurants, err := app.FindCollectionByNameOrId(restaurantsCollectionName)
		if err != nil {
			return err
		}

		if restaurants.Fields.GetByName("created") == nil {
			restaurants.Fields.Add(&core.AutodateField{
				Name:     "created",
				OnCreate: true,
			})
		}

		if restaurants.Fields.GetByName("updated") == nil {
			restaurants.Fields.Add(&core.AutodateField{
				Name:     "updated",
				OnCreate: true,
				OnUpdate: true,
			})
		}

		restaurants.AddIndex(idxRestaurantsGroupName, false, "\"group\", name", "")
		restaurants.AddIndex(idxRestaurantsMethod, false, "method", "")

		return app.Save(restaurants)
	}, func(app core.App) error {
		restaurants, err := app.FindCollectionByNameOrId(restaurantsCollectionName)
		if err != nil {
			return nil
		}

		restaurants.RemoveIndex(idxRestaurantsGroupName)
		restaurants.RemoveIndex(idxRestaurantsMethod)

		return app.Save(restaurants)
	})
}
