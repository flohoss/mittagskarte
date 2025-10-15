package humanize

import (
	"fmt"
	"time"
)

func Since(t *time.Time) string {
	if t == nil {
		return "Kein Menü vorhanden"
	}
	d := time.Since(*t)
	seconds := int(d.Seconds())

	if seconds < 0 {
		return "in der Zukunft"
	}

	format := func(count int, unitSingular string, unitPlural string) string {
		if count == 1 {
			return fmt.Sprintf("vor 1 %s", unitSingular)
		}
		return fmt.Sprintf("vor %d %s", count, unitPlural)
	}

	switch {
	case seconds < 10:
		return "jetzt"
	case seconds < 60:
		count := seconds
		return format(count, "Sekunde", "Sekunden")
	case seconds < 3600:
		count := seconds / 60
		return format(count, "Minute", "Minuten")
	case seconds < 86400:
		count := seconds / 3600
		return format(count, "Stunde", "Stunden")
	case seconds < 2592000:
		count := seconds / 86400
		return format(count, "Tag", "Tagen")
	case seconds < 31104000:
		count := seconds / 2592000
		return format(count, "Monat", "Monaten")
	default:
		count := seconds / 31104000
		return format(count, "Jahr", "Jahren")
	}
}
