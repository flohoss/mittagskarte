package restaurant

import (
	"log/slog"

	"gorm.io/gorm"
)

func MigrateModels(orm *gorm.DB) {
	orm.AutoMigrate(&Restaurant{})
	orm.AutoMigrate(&Card{})
	orm.AutoMigrate(&Food{})

	configs, err := parseAllConfigs()
	if err != nil {
		slog.Error(err.Error())
	}

	for _, c := range configs {
		var res Restaurant
		if c.Restaurant.ID != "" {
			amount := orm.Find(&res, "id = ?", c.Restaurant.ID).RowsAffected
			if amount != 0 {
				orm.Save(&c.Restaurant)
			} else {
				orm.Create(&c.Restaurant)
			}
		}
	}
}
