package controller

import (
	"fmt"
	"regexp"
	"time"

	"github.com/goodsign/monday"
	"go.uber.org/zap"
)

func (c *Controller) deleteCard(card *Card) {
	if card.ID != 0 {
		zap.L().Debug("deleting card", zap.Any("card", card))
		c.orm.Delete(card)
	}
}

func (c *Controller) createCard(card *Card, text []byte, regex string, format string) error {
	dateRegex := regexp.MustCompile(regex)
	groups := dateRegex.FindAll(text, 2)
	if len(groups) < 2 {
		zap.L().Debug("No dates detected", zap.String("groups", fmt.Sprintf("%s", groups)), zap.String("ocr", string(text)))
		return fmt.Errorf("%s", "No dates detected")
	}
	for i := 0; i < len(groups); i++ {
		groups[i] = WhiteSpaceRegex.ReplaceAllLiteral(groups[i], []byte(""))
	}
	var dates []time.Time
	for i := 0; i < len(groups); i++ {
		d, err := monday.Parse(format, string(groups[i]), monday.LocaleDeDE)
		if err != nil {
			newDate := fmt.Sprintf("%s%d", groups[i], time.Now().Year())
			d, _ = monday.Parse(format, newDate, monday.LocaleDeDE)
		}
		dates = append(dates, d)
	}
	from := dates[0].Format("02.01.06")
	year, weekNr := dates[0].ISOWeek()
	weekDay := int(dates[0].Weekday())
	to := dates[1].Format("02.01.06")
	card.Description = fmt.Sprintf("Woche %d: %s bis %s", weekNr, from, to)
	card.Year = year
	card.Week = weekNr
	card.Day = weekDay
	c.orm.Create(card)
	return nil
}
