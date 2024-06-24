package config

import (
	"encoding/json"
	"errors"
)

type DayOfWeek string

const (
	Sunday    DayOfWeek = "Sunday"
	Monday    DayOfWeek = "Monday"
	Tuesday   DayOfWeek = "Tuesday"
	Wednesday DayOfWeek = "Wednesday"
	Thursday  DayOfWeek = "Thursday"
	Friday    DayOfWeek = "Friday"
	Saturday  DayOfWeek = "Saturday"
)

var allDays = []DayOfWeek{Sunday, Monday, Tuesday, Wednesday, Thursday, Friday, Saturday}

func (d *DayOfWeek) UnmarshalJSON(data []byte) error {
	var day string
	if err := json.Unmarshal(data, &day); err != nil {
		return err
	}

	for _, validDay := range allDays {
		if DayOfWeek(day) == validDay {
			*d = DayOfWeek(day)
			return nil
		}
	}
	return errors.New("invalid day of the week")
}

func (d DayOfWeek) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(d))
}
