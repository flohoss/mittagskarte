package restaurant

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/goodsign/monday"
)

func (s *Selector) regexResult(content *string) string {
	reg := regexp.MustCompile("(?i)" + replacePlaceholder(s.Regex))
	res := reg.FindStringSubmatch(*content)
	if len(res) > 1 {
		return res[1]
	}
	return ""
}

func (s *Selector) jQueryResult(doc *goquery.Document) string {
	return doc.Find(s.JQuery).First().Text()
}

func (f *FoodEntry) getDay(content *string, doc *goquery.Document) string {
	if f.Day.Fixed != "" {
		return f.Day.Fixed
	} else if f.Day.Regex != "" {
		return f.Day.regexResult(content)
	} else if f.Day.JQuery != "" {
		return f.Day.jQueryResult(doc)
	} else {
		return ""
	}
}

func (f *FoodEntry) getName(content *string, doc *goquery.Document) string {
	if f.Food.Fixed != "" {
		return f.Food.Fixed
	} else if f.Food.Regex != "" {
		return f.Food.regexResult(content)
	} else if f.Food.JQuery != "" {
		return f.Food.jQueryResult(doc)
	} else {
		return ""
	}
}

func (f *FoodEntry) getPrice(content *string, doc *goquery.Document) float64 {
	if f.Price.Fixed != "" {
		return convertPrice(f.Price.Fixed)
	} else if f.Price.Regex != "" {
		return convertPrice(f.Price.regexResult(content))
	} else if f.Price.JQuery != "" {
		return convertPrice(f.Price.jQueryResult(doc))
	} else {
		return 0
	}
}

func (f *FoodEntry) getDescription(content *string, doc *goquery.Document) string {
	if f.Description.Fixed != "" {
		return f.Description.Fixed
	} else if f.Price.Regex != "" {
		return f.Description.regexResult(content)
	} else if f.Price.JQuery != "" {
		return strings.TrimSpace(f.Description.jQueryResult(doc))
	} else {
		return ""
	}
}

func appendFood(allFood *[]Food, food *Food) {
	drink, _ := regexp.MatchString("(?i)\\d{1,2},\\d{1,2}\\s?l", food.Description)
	if food.Price != 0.0 && food.Name != "" && !foodExisting(allFood, food) && posInArray(food.Name, monday.GetLongDays(monday.LocaleDeDE)) == -1 && !drink {
		*allFood = append(*allFood, *food)
	}
}

func (c *Configuration) getAllFood(content *string) []Food {
	var allFood []Food
	if len(c.Menu.Food) > 0 {
		for _, f := range c.Menu.Food {
			food := Food{
				Name:        f.getName(content, c.htmlPages[len(c.htmlPages)-1]),
				Day:         f.getDay(content, c.htmlPages[len(c.htmlPages)-1]),
				Price:       f.getPrice(content, c.htmlPages[len(c.htmlPages)-1]),
				Description: f.getDescription(content, c.htmlPages[len(c.htmlPages)-1]),
			}
			appendFood(&allFood, &food)
		}
	}
	if c.Menu.OneForAll.Regex != "" {
		regexStr := replacePlaceholder(c.Menu.OneForAll.Regex)
		if c.Menu.OneForAll.Insensitive {
			regexStr += "(?i)"
		}
		foodRegex := regexp.MustCompile(regexStr)
		regexResult := foodRegex.FindAllStringSubmatch(*content, -1)
		for _, r := range regexResult {
			var food Food
			if c.Menu.OneForAll.PositionFood > 0 && len(r) > int(c.Menu.OneForAll.PositionFood) {
				food.Name = clearAndTitleString(r[c.Menu.OneForAll.PositionFood])
			}
			if c.Menu.OneForAll.PositionDay > 0 && len(r) > int(c.Menu.OneForAll.PositionDay) {
				food.Day = clearAndTitleString(r[c.Menu.OneForAll.PositionDay])
			}
			if c.Menu.OneForAll.FixedPrice != 0 {
				food.Price = c.Menu.OneForAll.FixedPrice
			} else if c.Menu.OneForAll.PositionPrice > 0 && len(r) > int(c.Menu.OneForAll.PositionPrice) {
				food.Price = convertPrice(r[c.Menu.OneForAll.PositionPrice])
			}
			if c.Menu.OneForAll.PositionDescription > 0 && len(r) > int(c.Menu.OneForAll.PositionDescription) {
				food.Description = clearString(r[c.Menu.OneForAll.PositionDescription])
			}
			appendFood(&allFood, &food)
		}
	}
	if c.Menu.OneForAll.JQuery.Wrapper != "" {
		c.htmlPages[len(c.htmlPages)-1].Find(replacePlaceholder(c.Menu.OneForAll.JQuery.Wrapper)).Each(func(i int, s *goquery.Selection) {
			food := Food{
				Name:        strings.TrimSpace(s.Find(c.Menu.OneForAll.JQuery.Food).Text()),
				Day:         strings.TrimSpace(s.Find(c.Menu.OneForAll.JQuery.Day).Text()),
				Price:       convertPrice(s.Find(c.Menu.OneForAll.JQuery.Price).Text()),
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
