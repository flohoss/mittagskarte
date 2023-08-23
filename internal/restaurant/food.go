package restaurant

import (
	"regexp"
)

func (s *Selector) regexResult(content *string) string {
	reg := regexp.MustCompile(replacePlaceholder(s.Regex))
	res := reg.FindStringSubmatch(*content)
	if len(res) > 1 {
		return res[1]
	}
	return ""
}

func (s *Selector) jQueryResult(content *string) string {
	return ""
}

func (f *FoodEntry) getDay(content *string) string {
	if f.Day.Fixed != "" {
		return f.Day.Fixed
	} else if f.Day.Regex != "" {
		return f.Day.regexResult(content)
	} else if f.Day.JQuery != "" {
		return f.Day.jQueryResult(content)
	} else {
		return ""
	}
}

func (f *FoodEntry) getName(content *string) string {
	if f.Food.Fixed != "" {
		return f.Food.Fixed
	} else if f.Food.Regex != "" {
		return f.Food.regexResult(content)
	} else if f.Food.JQuery != "" {
		return f.Food.jQueryResult(content)
	} else {
		return ""
	}
}

func (f *FoodEntry) getPrice(content *string) float64 {
	if f.Price.Fixed != "" {
		return convertPrice(f.Price.Fixed)
	} else if f.Price.Regex != "" {
		return convertPrice(f.Price.regexResult(content))
	} else if f.Price.JQuery != "" {
		return convertPrice(f.Price.jQueryResult(content))
	} else {
		return 0
	}
}

func (f *FoodEntry) getDescription(content *string) string {
	if f.Description.Fixed != "" {
		return f.Description.Fixed
	} else if f.Price.Regex != "" {
		return f.Description.regexResult(content)
	} else if f.Price.JQuery != "" {
		return f.Description.jQueryResult(content)
	} else {
		return ""
	}
}

func (c *Configuration) getAllFood(content *string) []Food {
	var allFood []Food
	if c.Menu.OneForAll.Regex != "" {
		foodRegex := regexp.MustCompile(replacePlaceholder(c.Menu.OneForAll.Regex))
		regexResult := foodRegex.FindAllStringSubmatch(*content, -1)
		for _, r := range regexResult {
			var f Food
			if c.Menu.OneForAll.PositionFood > 0 && len(r) > int(c.Menu.OneForAll.PositionFood) {
				f.Name = clearAndTitleString(r[c.Menu.OneForAll.PositionFood])
			}
			if c.Menu.OneForAll.PositionDay > 0 && len(r) > int(c.Menu.OneForAll.PositionDay) {
				f.Day = clearAndTitleString(r[c.Menu.OneForAll.PositionDay])
			}
			if c.Menu.OneForAll.PositionPrice > 0 && len(r) > int(c.Menu.OneForAll.PositionPrice) {
				f.Price = convertPrice(r[c.Menu.OneForAll.PositionPrice])
			}
			if c.Menu.OneForAll.PositionDescription > 0 && len(r) > int(c.Menu.OneForAll.PositionDescription) {
				f.Description = clearString(r[c.Menu.OneForAll.PositionDescription])
			}
			allFood = append(allFood, f)
		}
	} else {
		for _, f := range c.Menu.Food {
			allFood = append(allFood, Food{
				Name:        f.getName(content),
				Day:         f.getDay(content),
				Price:       f.getPrice(content),
				Description: f.getDescription(content),
			})
		}
	}
	return allFood
}
