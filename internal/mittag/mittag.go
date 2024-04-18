package mittag

import (
	"context"
	"log/slog"
	"os"

	"github.com/vorlif/spreak/humanize"
	"gitlab.unjx.de/flohoss/mittag/internal/database"
	"gitlab.unjx.de/flohoss/mittag/internal/env"
	"gorm.io/gorm"
)

type Mittag struct {
	Configurations map[string]*Configuration
	env            *env.Env
	orm            *gorm.DB
	Humanizer      *humanize.Humanizer
}

func NewMittag(env *env.Env, humanizer *humanize.Humanizer) *Mittag {
	c, err := parseAllConfigs()
	if err != nil {
		slog.Error("could not parse configurations", "err", err.Error())
		os.Exit(1)
	}
	mittag := Mittag{
		Configurations: c,
		orm:            database.NewDatabaseConnection("sqlite.db"),
		env:            env,
		Humanizer:      humanizer,
	}
	mittag.migrateModels()
	if !slog.Default().Enabled(context.Background(), slog.LevelDebug) {
		mittag.UpdateRestaurants()
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
