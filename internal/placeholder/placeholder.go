package placeholder

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/goodsign/monday"
	"github.com/snabb/isoweek"
)

func getWeekDayDate(weekDayAbbr string, format string, weekOffset string) (string, error) {
	weekDays := map[string]int{
		"mo": 0,
		"tu": 1,
		"we": 2,
		"th": 3,
		"fr": 4,
		"sa": 5,
		"su": 6,
	}
	targetWeekDay, exists := weekDays[weekDayAbbr]
	if !exists {
		return "", fmt.Errorf("invalid weekday abbreviation: %s", weekDayAbbr)
	}

	offsetInt, err := strconv.Atoi(weekOffset)
	if err != nil {
		return "", fmt.Errorf("invalid offset: %s", weekOffset)
	}
	offsetInt = offsetInt * 7

	year, week := time.Now().ISOWeek()
	date := isoweek.StartTime(year, week, time.UTC)

	return date.AddDate(0, 0, targetWeekDay+offsetInt).Format(format), nil
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
		if strings.Contains(key, "date") {
			reSplit := regexp.MustCompile(`date\((.+?)\,(.+?)\,(.+?)\)`)
			match := reSplit.FindStringSubmatch(key)
			if match == nil || len(match) != 4 {
				return key
			}
			day, err := getWeekDayDate(match[1], match[2], match[3])
			if err != nil {
				return key
			}
			return day
		}
		return key
	})

	return result
}
