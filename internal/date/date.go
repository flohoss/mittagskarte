package date

import (
	"strings"
	"time"
)

func GetMonthFromString(month string) int {
	res := 0
	s := []string{"Januar", "Februar", "Marz März", "April", "Mai", "Juni", "July", "August", "September", "Oktober", "November", "Dezember"}
	for index, date := range s {
		if strings.Contains(date, month) || strings.Contains(strings.ToLower(date), month) {
			res = index
			break
		}
	}
	return res + 1
}

func GetMonthFromNumber(month int) string {
	s := []string{"Januar", "Februar", "März", "April", "Mai", "Juni", "July", "August", "September", "Oktober", "November", "Dezember"}
	return s[month-1]
}

func LindeRelevantDates(format string) (weekNr int, year int, weekDay int, start time.Time, begin string, end string) {
	year, weekNr = time.Now().ISOWeek()
	now := time.Now()
	fromDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	toDate := fromDate.AddDate(0, 1, -1)
	weekDay = int(fromDate.Weekday())
	return weekNr, year, weekDay, fromDate, fromDate.Format(format), toDate.Format(format)
}

func ExtractRelevantFromDateRange(from string, to string, format string) (weekNr int, year int, weekDay int, begin string, end string) {
	fromDate, err := time.Parse(format, from)
	if err != nil {
		return weekNr, year, weekDay, begin, end
	}
	toDate, err := time.Parse(format, to)
	if err != nil {
		return weekNr, year, weekDay, begin, end
	}
	year, weekNr = toDate.ISOWeek()
	weekDay = int(fromDate.Weekday())
	begin = fromDate.Format("02.01.06")
	end = toDate.Format("02.01.06")
	return weekNr, year, weekDay, begin, end
}

func ExtractRelevantFromDateString(date string, format string) (weekNr int, year int, weekDay int, dateStr string, goDate time.Time) {
	goDate, err := time.Parse(format, date)
	if err != nil {
		return weekNr, year, weekDay, dateStr, goDate
	}
	year, weekNr = goDate.ISOWeek()
	weekDay = int(goDate.Weekday())
	dateStr = goDate.Format("02.01.06")
	return weekNr, year, weekDay, dateStr, goDate
}
