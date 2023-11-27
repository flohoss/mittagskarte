package mittag

import (
	"log/slog"
	"os"

	"gitlab.unjx.de/flohoss/mittag/internal/database"
	"gorm.io/gorm"
)

type Mittag struct {
	Configurations map[string]*Configuration
	orm            *gorm.DB
}

func NewMittag() *Mittag {
	c, err := parseAllConfigs()
	if err != nil {
		slog.Error("could not parse configurations", "err", err.Error())
		os.Exit(1)
	}
	mittag := Mittag{
		Configurations: c,
		orm:            database.NewDatabaseConnection("sqlite.db"),
	}
	mittag.migrateModels()
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
