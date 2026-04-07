package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func publicRule() *string {
	rule := ""
	return &rule
}

func authRule() *string {
	rule := "@request.auth.id != \"\""
	return &rule
}

func init() {
	m.Register(func(app core.App) error {
		minOrder := float64(1)

		selectors := core.NewBaseCollection("selectors")

		selectors.Fields.Add(&core.NumberField{
			Name:     "order",
			Min:      &minOrder,
			OnlyInt:  true,
			Required: true,
		})
		selectors.Fields.Add(&core.TextField{
			Name:     "locator",
			Required: true,
		})
		selectors.Fields.Add(&core.TextField{
			Name: "attribute",
		})
		selectors.Fields.Add(&core.TextField{
			Name: "style",
		})

		selectors.ListRule = publicRule()
		selectors.ViewRule = publicRule()

		selectors.AddIndex("idx_selector_order", false, "order", "")

		if err := app.Save(selectors); err != nil {
			return err
		}

		menus := core.NewBaseCollection("menus")

		menus.Fields.Add(&core.FileField{
			Name:     "file",
			Required: true,
			MaxSize:  25 << 20,
		})
		menus.Fields.Add(&core.TextField{
			Name:     "hash",
			Required: true,
		})
		menus.Fields.Add(&core.JSONField{
			Name: "dimensions",
		})
		menus.Fields.Add(&core.AutodateField{
			Name:     "created",
			OnCreate: true,
		})

		menus.ListRule = publicRule()
		menus.ViewRule = publicRule()
		menus.CreateRule = authRule()

		if err := app.Save(menus); err != nil {
			return err
		}

		restaurants := core.NewBaseCollection("restaurants")

		restaurants.Fields.Add(&core.TextField{
			Name:     "name",
			Required: true,
		})
		restaurants.Fields.Add(&core.TextField{
			Name: "group",
		})
		restaurants.Fields.Add(&core.TextField{
			Name: "address",
		})
		restaurants.Fields.Add(&core.TextField{
			Name:    "website",
			Pattern: `^https?://[^\s/$.?#].[^\s]*$`,
		})
		restaurants.Fields.Add(&core.TextField{
			Name:    "phone",
			Pattern: `^\+?[\d\s\-()/.]+$`,
		})
		restaurants.Fields.Add(&core.JSONField{
			Name: "tags",
		})
		restaurants.Fields.Add(&core.SelectField{
			Name:      "rest_days",
			Values:    []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"},
			MaxSelect: 7,
		})
		restaurants.Fields.Add(&core.SelectField{
			Name:     "method",
			Values:   []string{"scrape", "download", "upload"},
			Required: true,
		})
		restaurants.Fields.Add(&core.SelectField{
			Name:   "content_type",
			Values: []string{"html", "image", "pdf"},
		})
		restaurants.Fields.Add(&core.TextField{
			Name:    "cron",
			Pattern: `^(@(annually|yearly|monthly|weekly|daily|hourly)|((\*|[0-9]+([,\-][0-9]+)*)(/[0-9]+)?)(\s+((\*|[0-9]+([,\-][0-9]+)*)(/[0-9]+)?)){4})$`,
		})
		restaurants.Fields.Add(&core.RelationField{
			Name:         "navigate",
			CollectionId: selectors.Id,
			MaxSelect:    20,
		})
		restaurants.Fields.Add(&core.RelationField{
			Name:         "menus",
			CollectionId: menus.Id,
			MaxSelect:    5,
		})
		restaurants.Fields.Add(&core.FileField{
			Name: "thumbnail",
		})

		restaurants.ListRule = publicRule()
		restaurants.ViewRule = publicRule()

		restaurants.AddIndex("idx_restaurants_group_name", false, "\"group\", name", "")
		restaurants.AddIndex("idx_restaurants_method", false, "method", "")

		if err := app.Save(restaurants); err != nil {
			return err
		}

		menus.Fields.Add(&core.RelationField{
			Name:         "restaurant",
			CollectionId: restaurants.Id,
			MaxSelect:    1,
			Required:     true,
		})

		selectors.Fields.Add(&core.RelationField{
			Name:         "restaurant",
			CollectionId: restaurants.Id,
			MaxSelect:    1,
			Required:     true,
		})

		if err := app.Save(menus); err != nil {
			return err
		}

		if err := app.Save(selectors); err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {
		restaurants, err := app.FindCollectionByNameOrId("restaurants")
		if err == nil {
			if deleteErr := app.Delete(restaurants); deleteErr != nil {
				return deleteErr
			}
		}

		menus, err := app.FindCollectionByNameOrId("menus")
		if err == nil {
			if deleteErr := app.Delete(menus); deleteErr != nil {
				return deleteErr
			}
		}

		selectors, err := app.FindCollectionByNameOrId("selectors")
		if err == nil {
			if deleteErr := app.Delete(selectors); deleteErr != nil {
				return deleteErr
			}
		}

		return nil
	})
}
