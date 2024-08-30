package parse

import (
	"log/slog"
	"regexp"

	"github.com/PuerkitoBio/goquery"
	"gitlab.unjx.de/flohoss/mittag/internal/config"
	"gitlab.unjx.de/flohoss/mittag/internal/helper"
	"gitlab.unjx.de/flohoss/mittag/pgk/convert"
)

type MenuParser struct {
	docStorage  []*goquery.Document
	fileContent string
	parse       *config.Parse
	Menu        *config.Menu
}

func NewMenuParser(docStorage []*goquery.Document, fileContent string, parse *config.Parse, card string) *MenuParser {
	p := &MenuParser{
		docStorage:  docStorage,
		fileContent: fileContent,
		parse:       parse,
		Menu: &config.Menu{
			Card: card,
			Food: []config.FoodEntry{},
		},
	}
	p.Parse()
	return p
}

func (p *MenuParser) Parse() {
	p.ParseDescription()
	p.ParseFood()
}

func (p *MenuParser) ParseDescription() {
	if p.parse.Description.Fixed != "" {
		p.Menu.Description = p.parse.Description.Fixed
	} else if p.parse.Description.JQuery != "" {
		for i := 0; i < len(p.docStorage); i++ {
			p.Menu.Description = p.parse.Description.JQueryResult(p.docStorage[i])
			if p.Menu.Description != "" {
				break
			}
		}
	} else if p.parse.Description.Regex != "" {
		for i := 0; i < len(p.docStorage); i++ {
			p.Menu.Description = p.parse.Description.RegexResult("", p.docStorage[i])
			if p.Menu.Description != "" {
				break
			}
		}
		if p.Menu.Description == "" {
			p.Menu.Description = p.parse.Description.RegexResult(p.fileContent, nil)
		}
	}
	if p.Menu.Description == "" {
		slog.Warn("could not parse description")
		return
	}
	slog.Debug("parsed description", "description", p.Menu.Description)
}

func (p *MenuParser) ParseFood() {
	var lastestHtmlPage *goquery.Document
	if len(p.docStorage) > 0 {
		lastestHtmlPage = p.docStorage[len(p.docStorage)-1]
	}

	if p.parse.OneForAll.Regex != "" {
		regexStr := helper.ReplacePlaceholder(p.parse.OneForAll.Regex) + "(?i)"
		foodRegex := regexp.MustCompile(regexStr)
		regexResult := foodRegex.FindAllStringSubmatch(p.fileContent, -1)
		p.ProcessRegexResult(regexResult)
		if len(p.docStorage) > 0 {
			regexResult = foodRegex.FindAllStringSubmatch(p.docStorage[len(p.docStorage)-1].Text(), -1)
			p.ProcessRegexResult(regexResult)
		}
	}

	for i := 0; i < len(p.parse.Food); i++ {
		f := config.FoodEntry{
			Day:         p.parse.Food[i].GetDay(p.fileContent, lastestHtmlPage),
			Name:        p.parse.Food[i].GetName(p.fileContent, lastestHtmlPage),
			Price:       p.parse.Food[i].GetPrice(p.fileContent, lastestHtmlPage),
			Description: p.parse.Food[i].GetDescription(p.fileContent, lastestHtmlPage),
		}
		if f.Name != "" && f.Price != 0 {
			p.Menu.Food = append(p.Menu.Food, f)
		}
	}
	slog.Debug("parsed food", "entries", len(p.Menu.Food))
}

func (p *MenuParser) ProcessRegexResult(regexResult [][]string) {
	for _, r := range regexResult {
		var f config.FoodEntry
		if p.parse.OneForAll.PositionFood > 0 && len(r) > int(p.parse.OneForAll.PositionFood) {
			f.Name = convert.ClearAndTitleString(r[p.parse.OneForAll.PositionFood])
		}
		if p.parse.OneForAll.PositionDay > 0 && len(r) > int(p.parse.OneForAll.PositionDay) {
			f.Day = convert.ClearAndTitleString(r[p.parse.OneForAll.PositionDay])
		}
		if p.parse.OneForAll.FixedPrice != 0 {
			f.Price = p.parse.OneForAll.FixedPrice
		} else if p.parse.OneForAll.PositionPrice > 0 && len(r) > int(p.parse.OneForAll.PositionPrice) {
			f.Price = convert.ConvertPrice(r[p.parse.OneForAll.PositionPrice])
		}
		if p.parse.OneForAll.PositionDescription > 0 && len(r) > int(p.parse.OneForAll.PositionDescription) {
			f.Description = convert.ClearString(r[p.parse.OneForAll.PositionDescription])
		}
		if f.Name != "" && f.Price != 0 {
			p.Menu.Food = append(p.Menu.Food, f)
		}
	}
}
