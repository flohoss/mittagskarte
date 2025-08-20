package humanize

import (
	"fmt"
	"time"
)

func Since(t time.Time) string {
	d := time.Since(t)
	seconds := int(d.Seconds())

	if seconds < 0 {
		return "in der Zukunft"
	}

	switch {
	case seconds < 10:
		return "jetzt"
	case seconds < 60:
		return fmt.Sprintf("vor %d Sekunden", seconds)
	case seconds < 3600:
		return fmt.Sprintf("vor %d Minuten", seconds/60)
	case seconds < 86400:
		return fmt.Sprintf("vor %d Stunden", seconds/3600)
	case seconds < 2592000:
		return fmt.Sprintf("vor %d Tagen", seconds/86400)
	case seconds < 31104000:
		return fmt.Sprintf("vor %d Monaten", seconds/2592000)
	default:
		return fmt.Sprintf("vor %d Jahren", seconds/31104000)
	}
}
