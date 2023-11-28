package mittag

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/goodsign/monday"
	"gitlab.unjx.de/flohoss/mittag/internal/helper"
	"gorm.io/gorm"
)

const ConfigLocation = "configs/restaurants/"
const PublicLocation = "storage/public/menus/"

func GetTodayActiveList() []string {
	today := monday.Format(time.Now(), "Monday", monday.LocaleDeDE)
	return []string{today, today + " (Vegetarisch)", "Alternative", "Wochen-Renner", "Veggie-Renner"}
}

func init() {
	os.MkdirAll(PublicLocation, os.ModePerm)
}

func parseConfig(path string) (Configuration, error) {
	slog.Debug("parsing config", "path", path)
	var config Configuration
	content, err := os.ReadFile(path)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(content, &config)
	if err != nil {
		return config, err
	}
	slog.Debug("config successfully parsed", "path", path)
	return config, nil
}

func parseAllConfigs() (map[string]*Configuration, error) {
	configurations := make(map[string]*Configuration)
	err := filepath.WalkDir(ConfigLocation, func(path string, info os.DirEntry, err error) error {
		if info.Type().IsRegular() {
			config, err := parseConfig(path)
			if err != nil {
				return err
			}
			configurations[config.Restaurant.ID] = &config
		}
		return nil
	})
	return configurations, err
}

func (c *Configuration) UpdateInformation(orm *gorm.DB) {
	card := Card{RestaurantID: c.Restaurant.ID}
	orm.FirstOrCreate(&card)
	defer orm.Save(&card)

	l := new(LiveInformation)
	err := l.fetchAndStoreHtmlPage(c.Restaurant.PageURL, c)
	if err != nil {
		return
	}

	for i := 0; i < len(c.RetrieveDownloadUrl); i++ {
		var err error
		err = l.findUrlInPage(&c.RetrieveDownloadUrl[i])
		if err != nil {
			return
		}
		if (i == len(c.RetrieveDownloadUrl)-1) && c.Download.IsFile {
			card.ExistingFileHash, err = l.fetchAndStoreFile(c.Restaurant.ID, l.DownloadUrl, c.HTTPOne, card.ExistingFileHash)
		} else {
			err = l.fetchAndStoreHtmlPage(l.DownloadUrl, c)
		}
		if err != nil {
			return
		}
	}

	if l.FileLocation != "" {
		err := l.parseAndStoreFileText(c)
		if err != nil {
			return
		}
		if slog.Default().Enabled(context.Background(), slog.LevelDebug) {
			helper.SaveContentAsFile(filepath.Dir(l.FileLocation), &l.RawText)
		}
		err = l.prepareFileForPublic(c.Restaurant.ID)
		if err != nil {
			return
		}
		card.Refreshed = time.Now().Unix()
		card.ImageURL = l.FileLocation
	}

	var currentFood []Food
	orm.Where("card_id = ?", c.Restaurant.ID).Find(&currentFood)
	newFood := c.getAllFood(l)

	if !isEqual(currentFood, newFood) {
		if len(currentFood) > 0 {
			orm.Delete(&currentFood)
		}
		card.Food = newFood
		card.Refreshed = time.Now().Unix()
	}

	card.Description = c.getDescription(l)
}

func (c *Configuration) getDescription(l *LiveInformation) string {
	description := ""
	if c.Menu.Description.Fixed != "" {
		description = c.Menu.Description.Fixed
	} else if c.Menu.Description.Regex != "" {
		replaced := helper.ReplacePlaceholder(c.Menu.Description.Regex)
		descriptionExpr := regexp.MustCompile("(?i)" + replaced)
		if len(l.RawText) > 0 {
			description = descriptionExpr.FindString(l.RawText)
		} else {
			description = descriptionExpr.FindString(l.HTMLPages[0].Text())
		}
	} else if c.Menu.Description.JQuery != "" {
		replaced := helper.ReplacePlaceholder(c.Menu.Description.JQuery)
		el := l.HTMLPages[0].Find(replaced).First()
		if c.Menu.Description.Attribute == "" {
			description = el.Text()
		} else {
			description, _ = el.Attr(c.Menu.Description.Attribute)
		}
	}
	return strings.ToValidUTF8(description, "")
}

func (c *Configuration) getAllFood(l *LiveInformation) []Food {
	var allFood []Food
	lastestHtmlPage := l.HTMLPages[len(l.HTMLPages)-1]
	for i := 0; i < len(c.Menu.Food); i++ {
		current := &c.Menu.Food[i]
		food := Food{
			Name:        current.getName(l.RawText, lastestHtmlPage),
			Day:         current.getDay(l.RawText, lastestHtmlPage),
			Price:       current.getPrice(l.RawText, lastestHtmlPage),
			Description: current.getDescription(l.RawText, lastestHtmlPage),
		}
		appendFood(&allFood, &food)
	}
	if c.Menu.OneForAll.Regex != "" {
		regexStr := helper.ReplacePlaceholder(c.Menu.OneForAll.Regex)
		if c.Menu.OneForAll.Insensitive {
			regexStr += "(?i)"
		}
		foodRegex := regexp.MustCompile(regexStr)
		regexResult := foodRegex.FindAllStringSubmatch(l.RawText, -1)
		for _, r := range regexResult {
			var food Food
			if c.Menu.OneForAll.PositionFood > 0 && len(r) > int(c.Menu.OneForAll.PositionFood) {
				food.Name = helper.ClearAndTitleString(r[c.Menu.OneForAll.PositionFood])
			}
			if c.Menu.OneForAll.PositionDay > 0 && len(r) > int(c.Menu.OneForAll.PositionDay) {
				food.Day = helper.ClearAndTitleString(r[c.Menu.OneForAll.PositionDay])
			}
			if c.Menu.OneForAll.FixedPrice != 0 {
				food.Price = c.Menu.OneForAll.FixedPrice
			} else if c.Menu.OneForAll.PositionPrice > 0 && len(r) > int(c.Menu.OneForAll.PositionPrice) {
				food.Price = helper.ConvertPrice(r[c.Menu.OneForAll.PositionPrice])
			}
			if c.Menu.OneForAll.PositionDescription > 0 && len(r) > int(c.Menu.OneForAll.PositionDescription) {
				food.Description = helper.ClearString(r[c.Menu.OneForAll.PositionDescription])
			}
			appendFood(&allFood, &food)
		}
	}
	if c.Menu.OneForAll.JQuery.Wrapper != "" {
		lastestHtmlPage.Find(helper.ReplacePlaceholder(c.Menu.OneForAll.JQuery.Wrapper)).Each(func(i int, s *goquery.Selection) {
			food := Food{
				Name:        strings.TrimSpace(s.Find(c.Menu.OneForAll.JQuery.Food).Text()),
				Day:         strings.TrimSpace(s.Find(c.Menu.OneForAll.JQuery.Day).Text()),
				Price:       helper.ConvertPrice(s.Find(c.Menu.OneForAll.JQuery.Price).Text()),
				Description: strings.TrimSpace(s.Find(c.Menu.OneForAll.JQuery.Description).Text()),
			}
			if c.Menu.OneForAll.FixedPrice != 0 {
				food.Price = c.Menu.OneForAll.FixedPrice
			}
			if foodInAllFood(food, allFood) != -1 {
				return
			}
			appendFood(&allFood, &food)
		})
	}
	return allFood
}
