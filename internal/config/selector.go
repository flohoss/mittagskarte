package config

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"gitlab.unjx.de/flohoss/mittag/internal/helper"
)

func (s *Selector) RegexResult(content string) string {
	reg := regexp.MustCompile("(?i)" + helper.ReplacePlaceholder(s.Regex))
	res := reg.FindStringSubmatch(content)
	if len(res) > 1 {
		return res[1]
	}
	return ""
}

func (s *Selector) JQueryResult(doc *goquery.Document) string {
	if s.Attribute != "" {
		res, present := doc.Find(helper.ReplacePlaceholder(s.JQuery)).First().Attr(s.Attribute)
		if present {
			return strings.TrimSpace(res)
		}
		return ""
	}
	return strings.TrimSpace(doc.Find(s.JQuery).First().Text())
}
