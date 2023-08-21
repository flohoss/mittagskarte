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
