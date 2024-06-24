package config

import (
	"github.com/PuerkitoBio/goquery"
	"gitlab.unjx.de/flohoss/mittag/pgk/convert"
)

func (f *FoodParser) GetDay(content string, doc *goquery.Document) string {
	if f.Day.Fixed != "" {
		return f.Day.Fixed
	} else if f.Day.Regex != "" {
		return f.Day.RegexResult(content, doc)
	} else if f.Day.JQuery != "" {
		return f.Day.JQueryResult(doc)
	} else {
		return ""
	}
}

func (f *FoodParser) GetName(content string, doc *goquery.Document) string {
	if f.Name.Fixed != "" {
		return f.Name.Fixed
	} else if f.Name.Regex != "" {
		return f.Name.RegexResult(content, doc)
	} else if f.Name.JQuery != "" {
		return f.Name.JQueryResult(doc)
	} else {
		return ""
	}
}

func (f *FoodParser) GetPrice(content string, doc *goquery.Document) float64 {
	if f.Price.Fixed != "" {
		return convert.ConvertPrice(f.Price.Fixed)
	} else if f.Price.Regex != "" {
		return convert.ConvertPrice(f.Price.RegexResult(content, doc))
	} else if f.Price.JQuery != "" {
		return convert.ConvertPrice(f.Price.JQueryResult(doc))
	} else {
		return 0
	}
}

func (f *FoodParser) GetDescription(content string, doc *goquery.Document) string {
	if f.Description.Fixed != "" {
		return f.Description.Fixed
	} else if f.Description.Regex != "" {
		return f.Description.RegexResult(content, doc)
	} else if f.Description.JQuery != "" {
		return f.Description.JQueryResult(doc)
	} else {
		return ""
	}
}
