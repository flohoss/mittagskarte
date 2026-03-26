package config

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Date is a time.Time that unmarshals from a "2006-01-02" JSON string.
type Date struct{ time.Time }

func (d *Date) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		return nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return fmt.Errorf("invalid date %q: %w", s, err)
	}
	d.Time = t
	return nil
}

// MarshalJSON encodes Date back to "2006-01-02" format.
func (d Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Time.Format("2006-01-02"))
}
