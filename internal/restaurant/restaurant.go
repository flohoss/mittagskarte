package restaurant

import (
	"log/slog"

	_ "github.com/otiai10/gosseract/v2"
	"gorm.io/gorm"
)

func GetNavigation(orm *gorm.DB) [][]Restaurant {
	var navigation [][]Restaurant
	for _, g := range Groups {
		var restaurants []Restaurant
		orm.Where(&Restaurant{Group: g}).Order("Name").Find(&restaurants)
		navigation = append(navigation, restaurants)
	}
	return navigation
}

func GetRestaurants(orm *gorm.DB) []Restaurant {
	var result []Restaurant
	orm.Model(&Restaurant{}).Preload("Card.Food", func(db *gorm.DB) *gorm.DB {
		return db.Order("id")
	}).Order("name").Find(&result)
	return result
}

func (r *Restaurant) Update() (Card, error) {
	slog.Debug("updating restaurant", "name", r.Name)
	config, err := parseConfig(ConfigLocation + r.ID + ".json")
	if err != nil {
		return config.card, err
	}
	config.card = Card{RestaurantID: r.ID}

	if !(len(config.RetrieveDownloadUrl) == 0 && config.Download.IsFile) {
		err = config.getFirstHtmlPage()
		if err != nil {
			return config.card, err
		}
	}

	err = config.getFinalHtmlPage()
	if err != nil {
		return config.card, err
	}

	if config.Download.IsFile {
		err = config.downloadAndParseMenu()
		if err != nil {
			return config.card, err
		}
	} else {
		config.content = []string{config.htmlPages[len(config.htmlPages)-1].Text()}
		config.saveContentAsFile("", config.content[0])
	}

	if len(config.content) > 0 {
		err = config.parseDescription()
		if err != nil {
			return config.card, err
		}
	}

	for _, c := range config.content {
		config.card.Food = append(config.card.Food, config.getAllFood(&c)...)
	}
	return config.card, nil
}
