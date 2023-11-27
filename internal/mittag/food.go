package mittag

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/goodsign/monday"
	"gitlab.unjx.de/flohoss/mittag/internal/helper"
)

func (s *Selector) regexResult(content *string) string {
	reg := regexp.MustCompile("(?i)" + helper.ReplacePlaceholder(s.Regex))
	res := reg.FindStringSubmatch(*content)
	if len(res) > 1 {
		return res[1]
	}
	return ""
}

func (s *Selector) jQueryResult(doc *goquery.Document) string {
	return strings.TrimSpace(doc.Find(s.JQuery).First().Text())
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
	if f.Name.Fixed != "" {
		return f.Name.Fixed
	} else if f.Name.Regex != "" {
		return f.Name.regexResult(content)
	} else if f.Name.JQuery != "" {
		return f.Name.jQueryResult(doc)
	} else {
		return ""
	}
}

func (f *FoodEntry) getPrice(content *string, doc *goquery.Document) float64 {
	if f.Price.Fixed != "" {
		return helper.ConvertPrice(f.Price.Fixed)
	} else if f.Price.Regex != "" {
		return helper.ConvertPrice(f.Price.regexResult(content))
	} else if f.Price.JQuery != "" {
		return helper.ConvertPrice(f.Price.jQueryResult(doc))
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
		return f.Description.jQueryResult(doc)
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

func foodExisting(allFood *[]Food, food *Food) bool {
	for _, f := range *allFood {
		if f.Name == food.Name {
			return true
		}
	}
	return false
}

func posInArray(str string, arr []string) int {
	for i, s := range arr {
		if strings.EqualFold(s, str) {
			return i
		}
	}
	return -1
}

func foodInAllFood(food Food, arr []Food) int {
	for i, s := range arr {
		if reflect.DeepEqual(food, s) {
			return i
		}
	}
	return -1
}

func (f *Food) equal(other interface{}) bool {
	if otherFood, ok := other.(Food); ok {
		return f.Name == otherFood.Name && f.Price == otherFood.Price && f.Description == otherFood.Description
	}
	return false
}

func isEqual(arr1, arr2 []Food) bool {
	if len(arr1) != len(arr2) {
		return false
	}
	for i := 0; i < len(arr1); i++ {
		if !arr1[i].equal(arr2[i]) {
			return false
		}
	}
	return true
}
