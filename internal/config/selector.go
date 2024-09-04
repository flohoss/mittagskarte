package config

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"gitlab.unjx.de/flohoss/mittag/internal/helper"
)

func (s *Selector) RegexResult(content string, doc *goquery.Document) string {
	flags := "(?"
	if s.RegexFlags != "" {
		flags += s.RegexFlags + ")"
	} else {
		flags += "i)"
	}
	reg := regexp.MustCompile(flags + helper.ReplacePlaceholder(s.Regex))
	res := reg.FindStringSubmatch(content)
	if len(res) > 1 {
		return res[1]
	}
	if doc != nil {
		res = reg.FindStringSubmatch(doc.Text())
	}
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
	if len(s.JQuery) == 0 {
		return ""
	}
	sel := doc.Find(helper.ReplacePlaceholder(s.JQuery))
	if sel.Length() == 0 {
		return ""
	}
	return strings.TrimSpace(sel.First().Text())
}
