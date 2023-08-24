package restaurant

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/goodsign/monday"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func convertPrice(strPrice string) float64 {
	price, _ := strconv.ParseFloat(strings.Replace(strPrice, ",", ".", 1), 64)
	return price
}

func replacePlaceholder(input string) string {
	if strings.Contains(input, "%KW%") {
		_, weekNr := time.Now().ISOWeek()
		return strings.Replace(input, "%KW%", fmt.Sprintf("%d", weekNr), -1)
	}
	if strings.Contains(input, "%month%") {
		return strings.Replace(input, "%month%", monday.Format(time.Now(), "January", monday.LocaleDeDE), -1)
	}
	return input
}

func clearAndTitleString(input string) string {
	caser := cases.Title(language.German)
	return caser.String(strings.ReplaceAll(strings.TrimSpace(input), "\n", " "))
}

func clearString(input string) string {
	return strings.ReplaceAll(strings.TrimSpace(input), "\n", " ")
}
