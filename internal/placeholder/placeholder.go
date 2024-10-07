package placeholder

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/goodsign/monday"
)

func getWeekRange(t time.Time) (time.Time, time.Time) {
	// Find the weekday (0 = Sunday, 1 = Monday, ..., 6 = Saturday)
	weekday := t.Weekday()
	// Calculate the offset to get Monday (weekday == 1)
	offsetToMonday := (int(weekday) + 6) % 7
	// Calculate Monday by subtracting the offset from the current day
	monday := t.AddDate(0, 0, -offsetToMonday)
	// Calculate Friday by adding 4 days to Monday
	friday := monday.AddDate(0, 0, 4)
	return monday, friday
}

func Replace(input string) string {
	re := regexp.MustCompile(`\{\{(.*?)\}\}`)
	result := re.ReplaceAllStringFunc(input, func(placeholder string) string {
		key := strings.Trim(placeholder, "{}")
		if key == "month" {
			currentMonth := monday.Format(time.Now(), "January", monday.LocaleDeDE)
			return currentMonth
		}
		if key == "year" {
			currentYear := monday.Format(time.Now(), "2006", monday.LocaleDeDE)
			return currentYear
		}
		if key == "weekRange" {
			mo, fr := getWeekRange(time.Now())
			// 07. - 11.
			return fmt.Sprintf("%02d. - %02d.", mo.Day(), fr.Day())
		}
		return key
	})

	return result
}
