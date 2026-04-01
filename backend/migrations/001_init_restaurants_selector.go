package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func publicRule() *string {
	rule := ""
	return &rule
}

func init() {
	m.Register(func(app core.App) error {
		selector := core.NewBaseCollection("selector")
		selector.ListRule = publicRule()
		selector.ViewRule = publicRule()
		selector.Fields.Add(&core.TextField{
			Name:     "locator",
			Required: true,
		})
		selector.Fields.Add(&core.TextField{
			Name: "attribute",
		})
		selector.Fields.Add(&core.TextField{
			Name: "style",
		})
		selector.Fields.Add(&core.AutodateField{
			Name:     "created",
			OnCreate: true,
		})
		selector.Fields.Add(&core.AutodateField{
			Name:     "updated",
			OnCreate: true,
			OnUpdate: true,
		})

		if err := app.Save(selector); err != nil {
			return err
		}

		restaurants := core.NewBaseCollection("restaurants")
		restaurants.ListRule = publicRule()
		restaurants.ViewRule = publicRule()
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
			Name: "website",
		})
		restaurants.Fields.Add(&core.TextField{
			Name: "phone",
		})
		restaurants.Fields.Add(&core.JSONField{
			Name: "tags",
		})
		restaurants.Fields.Add(&core.JSONField{
			Name: "rest_days",
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
			Name: "cron",
		})
		restaurants.Fields.Add(&core.RelationField{
			Name:         "navigate",
			CollectionId: selector.Id,
			MaxSelect:    20,
		})
		restaurants.Fields.Add(&core.TextField{
			Name: "menu",
		})
		restaurants.Fields.Add(&core.TextField{
			Name: "menu_hash",
		})
		restaurants.Fields.Add(&core.FileField{
			Name: "thumbnail",
		})
		restaurants.Fields.Add(&core.JSONField{
			Name: "menu_dimensions",
		})
		restaurants.Fields.Add(&core.AutodateField{
			Name:     "created",
			OnCreate: true,
		})
		restaurants.Fields.Add(&core.AutodateField{
			Name:     "updated",
			OnCreate: true,
			OnUpdate: true,
		})

		if err := app.Save(restaurants); err != nil {
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

		selector, err := app.FindCollectionByNameOrId("selector")
		if err == nil {
			if deleteErr := app.Delete(selector); deleteErr != nil {
				return deleteErr
			}
		}

		return nil
	})
}
