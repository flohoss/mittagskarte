package placeholder

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/goodsign/monday"
)

var weekDays = map[string]time.Weekday{
	"monday":    time.Monday,
	"tuesday":   time.Tuesday,
	"wednesday": time.Wednesday,
	"thursday":  time.Thursday,
	"friday":    time.Friday,
	"saturday":  time.Saturday,
	"sunday":    time.Sunday,
}

var locales = map[string]monday.Locale{
	"en": monday.LocaleEnUS,
	"de": monday.LocaleDeDE,
}

func parseDatePlaceholder(placeholder string) string {
	key := strings.Trim(placeholder, "{}")
	if !strings.Contains(key, "date") {
		return key
	}

	// Extract content inside parentheses
	re := regexp.MustCompile(`date\((.*)\)`)
	match := re.FindStringSubmatch(key)
	if match == nil || len(match) != 2 {
		return key
	}

	argsStr := match[1]
	args := map[string]string{}

	// Split by comma and parse key=value
	for _, part := range strings.Split(argsStr, ",") {
		if strings.Contains(part, "=") {
			kv := strings.SplitN(part, "=", 2)
			key := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])
			args[key] = value
		}
	}

	// Get arguments with defaults
	formatStr := args["format"]
	if formatStr == "" {
		formatStr = "Monday, 02 January 2006"
	}

	localeStr := args["lang"]
	if localeStr == "" {
		localeStr = "en"
	}
	loc, ok := locales[localeStr]
	if !ok {
		loc = monday.LocaleEnUS
	}

	weekdayStr := args["day"]
	offsetStr := args["offset"]
	upperStr := args["upper"]

	now := time.Now()

	// Apply weekday adjustment
	if weekdayStr != "" {
		if wd, ok := weekDays[strings.ToLower(weekdayStr)]; ok {
			diff := int(wd - now.Weekday())
			now = now.AddDate(0, 0, diff)
		}
	}

	// Apply week offset
	if offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err == nil {
			now = now.AddDate(0, 0, offset*7)
		}
	}

	result := monday.Format(now, formatStr, loc)

	// Apply uppercase if requested
	if strings.ToLower(upperStr) == "true" {
		result = strings.ToUpper(result)
	}

	return result
}

func Replace(input string) string {
	re := regexp.MustCompile(`\{\{(.*?)\}\}`)
	result := re.ReplaceAllStringFunc(input, func(placeholder string) string {
		return parseDatePlaceholder(placeholder)
	})

	return result
}
