package placeholder

import (
	"regexp"
	"strings"
	"time"

	"github.com/goodsign/monday"
)

func Replace(input string) string {
	currentMonth := monday.Format(time.Now(), "January", monday.LocaleDeDE)
	currentYear := monday.Format(time.Now(), "2006", monday.LocaleDeDE)
	re := regexp.MustCompile(`\{\{(.*?)\}\}`)

	result := re.ReplaceAllStringFunc(input, func(placeholder string) string {
		key := strings.Trim(placeholder, "{}")
		if key == "month" {
			return currentMonth
		}
		if key == "year" {
			return currentYear
		}
		return key
	})

	return result
}
