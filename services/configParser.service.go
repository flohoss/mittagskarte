package services

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

type FileType string
type DayOfWeek string
type Group string

const (
	configLocation string = "data/restaurants/"

	PDF   FileType = "pdf"
	Image FileType = "image"

	Sunday    DayOfWeek = "Sunday"
	Monday    DayOfWeek = "Monday"
	Tuesday   DayOfWeek = "Tuesday"
	Wednesday DayOfWeek = "Wednesday"
	Thursday  DayOfWeek = "Thursday"
	Friday    DayOfWeek = "Friday"
	Saturday  DayOfWeek = "Saturday"

	Degerloch              Group = "Degerloch"
	Fasanenhof             Group = "Fasanenhof"
	Feuerbach              Group = "Feuerbach"
	Koengen                Group = "Köngen"
	LeinfeldenEchterdingen Group = "Leinfelden-Echterdingen"
	Nuertingen             Group = "Nürtingen"
)

var allFileTypes = []FileType{PDF, Image}
var allDays = []DayOfWeek{Sunday, Monday, Tuesday, Wednesday, Thursday, Friday, Saturday}
var allGroups = []Group{Degerloch, Fasanenhof, Feuerbach, Koengen, LeinfeldenEchterdingen, Nuertingen}

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

func (r *Restaurant) GetCleanRestaurant() *CleanRestaurant {
	return &CleanRestaurant{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		PageUrl:     r.PageUrl,
		Address:     r.Address,
		RestDays:    r.RestDays,
		Phone:       r.Phone,
		Group:       r.Group,
		UpdateCron:  r.Parse.UpdateCron,
		ImageUrl:    r.ImageUrl,
	}
}

type CleanRestaurant struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	PageUrl     string      `json:"page_url"`
	Address     string      `json:"address"`
	RestDays    []DayOfWeek `json:"rest_days"`
	Phone       string      `json:"phone"`
	Group       Group       `json:"group"`
	UpdateCron  string      `json:"update_cron"`
	ImageUrl    string      `json:"image_url"`
}

type Restaurant struct {
	ID          string      `json:"id"`
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
	UpdateCron string     `json:"update_period"`
	Navigate   []Selector `json:"navigate"`
	FileType   FileType   `json:"file_type"`
	Clip       Clip       `json:"clip"`
}

type Selector struct {
	Locator   string `json:"locator"`
	Attribute string `json:"attribute"`
	Style     string `json:"style"`
}

type Clip struct {
	Width   float64 `json:"width"`
	Height  float64 `json:"height"`
	OffsetX float64 `json:"offset_x"`
	OffsetY float64 `json:"offset_y"`
}

func (f *FileType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	for _, v := range allFileTypes {
		if FileType(s) == v {
			*f = FileType(s)
			return nil
		}
	}

	return fmt.Errorf("invalid file type: %s", s)
}

func (g *Group) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	for _, v := range allGroups {
		if Group(s) == v {
			*g = Group(s)
			return nil
		}
	}

	return fmt.Errorf("invalid group: %s", s)
}

func (d *DayOfWeek) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	for _, v := range allDays {
		if DayOfWeek(s) == v {
			*d = DayOfWeek(s)
			return nil
		}
	}

	return fmt.Errorf("invalid day of the week: %s", s)
}
