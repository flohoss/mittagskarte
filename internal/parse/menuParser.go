package parse

import (
	"log/slog"

	"github.com/PuerkitoBio/goquery"
	"gitlab.unjx.de/flohoss/mittag/internal/config"
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
			p.Menu.Description = p.parse.Description.RegexResult(p.docStorage[i].Text())
			if p.Menu.Description != "" {
				break
			}
		}
		if p.Menu.Description == "" {
			p.Menu.Description = p.parse.Description.RegexResult(p.fileContent)
		}
	}
	if p.Menu.Description == "" {
		slog.Warn("could not parse description")
	}
	slog.Debug("parsed description", "description", p.Menu.Description)
}

func (p *MenuParser) ParseFood() {
	for i := 0; i < len(p.parse.Food); i++ {
		f := config.FoodEntry{
			Day:         p.parse.Food[i].GetDay(p.fileContent, p.docStorage[len(p.docStorage)-1]),
			Name:        p.parse.Food[i].GetName(p.fileContent, p.docStorage[len(p.docStorage)-1]),
			Price:       p.parse.Food[i].GetPrice(p.fileContent, p.docStorage[len(p.docStorage)-1]),
			Description: p.parse.Food[i].GetDescription(p.fileContent, p.docStorage[len(p.docStorage)-1]),
		}
		p.Menu.Food = append(p.Menu.Food, f)
	}
	slog.Debug("parsed food", "entries", len(p.Menu.Food))
}
