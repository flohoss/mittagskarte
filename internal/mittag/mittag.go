package mittag

import (
	"context"
	"log/slog"
	"os"

	"gitlab.unjx.de/flohoss/mittag/internal/database"
	"gitlab.unjx.de/flohoss/mittag/internal/env"
	"gitlab.unjx.de/flohoss/mittag/internal/maps"
	"gorm.io/gorm"
)

type Mittag struct {
	Configurations map[string]*Configuration
	env            *env.Env
	orm            *gorm.DB
}

func NewMittag(env *env.Env) *Mittag {
	c, err := parseAllConfigs()
	if err != nil {
		slog.Error("could not parse configurations", "err", err.Error())
		os.Exit(1)
	}
	mittag := Mittag{
		Configurations: c,
		orm:            database.NewDatabaseConnection("sqlite.db"),
		env:            env,
	}
	mittag.migrateModels()
	if !slog.Default().Enabled(context.Background(), slog.LevelDebug) {
		mittag.UpdateMapsInformation("")
	}
	return &mittag
}

func (m *Mittag) GetORM() *gorm.DB {
	return m.orm
}

func (m *Mittag) UpdateRestaurants() {
	for key := range m.Configurations {
		m.Configurations[key].UpdateInformation(m.orm)
	}
}

func (m *Mittag) DoesConfigurationExist(id string) (bool, *Configuration) {
	value, ok := m.Configurations[id]
	return ok, value
}

func (m *Mittag) UpdateMapsInformation(id string) {
	requests := []maps.MapRequest{}
	if id != "" {
		requests = append(requests, maps.MapRequest{
			Identifier: id,
			Address:    m.Configurations[id].Restaurant.Address,
		})
	} else {
		for key, val := range m.Configurations {
			requests = append(requests, maps.MapRequest{
				Identifier: key,
				Address:    val.Restaurant.Address,
			})
		}
	}
	info := maps.GetMapInformation(m.env.GoogleAPIKey, requests)
	for key, val := range info {
		m.orm.Model(&Card{}).Where("restaurant_id = ?", key).Updates(Card{
			Distance: val.Route.Legs[len(val.Route.Legs)-1].HumanReadable,
			Duration: val.Route.Legs[len(val.Route.Legs)-1].Duration.String(),
		})
	}
}
