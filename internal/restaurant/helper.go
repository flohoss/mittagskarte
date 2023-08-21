package restaurant

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/goodsign/monday"
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

func posInArray(str string, arr []string) int {
	for i, s := range arr {
		if s == str {
			return i
		}
	}
	return -1
}
