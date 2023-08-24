package restaurant

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/goodsign/monday"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func convertPrice(strPrice string) float64 {
	removeEuro := strings.Replace(strPrice, "€", "", 1)
	replaceComma := strings.Replace(removeEuro, ",", ".", 1)
	trimmed := strings.TrimSpace(replaceComma)
	price, _ := strconv.ParseFloat(trimmed, 64)
	return price
}

func posInArray(str string, arr []string) int {
	for i, s := range arr {
		if strings.ToLower(s) == strings.ToLower(str) {
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
