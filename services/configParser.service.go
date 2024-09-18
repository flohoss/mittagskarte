package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

type DayOfWeek string
type Group string

const (
	configLocation         string    = "data/restaurants/"
	Sunday                 DayOfWeek = "Sunday"
	Monday                 DayOfWeek = "Monday"
	Tuesday                DayOfWeek = "Tuesday"
	Wednesday              DayOfWeek = "Wednesday"
	Thursday               DayOfWeek = "Thursday"
	Friday                 DayOfWeek = "Friday"
	Saturday               DayOfWeek = "Saturday"
	Degerloch              Group     = "Degerloch"
	Fasanenhof             Group     = "Fasanenhof"
	Feuerbach              Group     = "Feuerbach"
	Koengen                Group     = "Köngen"
	LeinfeldenEchterdingen Group     = "Leinfelden-Echterdingen"
	Nuertingen             Group     = "Nürtingen"
)

var allDays = []DayOfWeek{Sunday, Monday, Tuesday, Wednesday, Thursday, Friday, Saturday}
var AllGroups = []Group{Degerloch, Fasanenhof, Feuerbach, Koengen, LeinfeldenEchterdingen, Nuertingen}

type ConfigParser struct {
	Restaurants map[string]*Restaurant
}

func NewConfigParser() *ConfigParser {
	cp := &ConfigParser{
		Restaurants: make(map[string]*Restaurant),
	}
	if err := cp.parseConfigFiles(); err != nil {
		slog.Error("cannot parse config files", "err", err)
		os.Exit(1)
	}
	return cp
}

func (cp *ConfigParser) parseConfigFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var restaurant Restaurant
	if err = json.Unmarshal(content, &restaurant); err != nil {
		return err
	}
	cp.Restaurants[restaurant.ID] = &restaurant
	return nil
}

func (cp *ConfigParser) parseConfigFiles() error {
	files, err := os.ReadDir(configLocation)
	if err != nil {
		return err
	}
	for _, file := range files {
		if !file.IsDir() {
			err := cp.parseConfigFile(filepath.Join(configLocation, file.Name()))
			if err != nil {
				return fmt.Errorf("file %s: %w", file.Name(), err)
			}
		}
	}
	return nil
}

type Restaurant struct {
	ID          string      `json:"id"`
	Price       uint8       `json:"price"`
	Icon        string      `json:"icon"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	PageUrl     string      `json:"page_url"`
	Address     string      `json:"address"`
	RestDays    []DayOfWeek `json:"rest_days"`
	Phone       string      `json:"phone"`
	Group       Group       `json:"group"`
	Parse       Parse       `json:"parse"`
	ImageUrl    string      `json:"image_url"`
}

type Parse struct {
	Click    []string `json:"click"`
	Navigate []string `json:"navigate"`
	IsFile   bool     `json:"is_file"`
	Scan     Scan     `json:"scan"`
}

type Scan struct {
	Crop   Crop   `json:"crop"`
	Chrome Chrome `json:"chrome"`
}

type Chrome struct {
	Width int `json:"width"`
}

type Crop struct {
	Width   uint `json:"width"`
	Height  uint `json:"height"`
	OffsetX int  `json:"offset_x"`
	OffsetY int  `json:"offset_y"`
}

func (g *Group) UnmarshalJSON(data []byte) error {
	var group string
	if err := json.Unmarshal(data, &group); err != nil {
		return err
	}

	for _, validGroup := range AllGroups {
		if Group(group) == validGroup {
			*g = Group(group)
			return nil
		}
	}
	return errors.New("invalid group")
}

func (g Group) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(g))
}

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
