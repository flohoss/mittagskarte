package config

import (
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type FileType string
type DayOfWeek string
type Group string

const (
	ConfigFolder = "./config/"

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
	Favorites              Group = "Favoriten"
)

func (g Group) String() string {
	return string(g)
}

var Cfg GlobalConfig

var validate *validator.Validate
var mu sync.RWMutex

type GlobalConfig struct {
	LogLevel           string                 `mapstructure:"log_level" validate:"omitempty,oneof=debug info warn error"`
	TimeZone           string                 `mapstructure:"time_zone" validate:"required"`
	Server             ServerSettings         `mapstructure:"server"`
	Restaurants        map[string]*Restaurant `mapstructure:"restaurants"`
	GroupedRestaurants []GroupedRestaurants   `mapstructure:"-"`
}

type ServerSettings struct {
	Address string `mapstructure:"address" validate:"required,ipv4"`
	Port    int    `mapstructure:"port" validate:"required,gte=1024,lte=65535"`
}

type Restaurant struct {
	ID          string      `mapstructure:"-"`
	Name        string      `mapstructure:"name"`
	Description string      `mapstructure:"description"`
	PageUrl     string      `mapstructure:"url"`
	Address     string      `mapstructure:"address"`
	RestDays    []DayOfWeek `mapstructure:"rest_days"`
	Phone       string      `mapstructure:"phone"`
	Group       Group       `mapstructure:"group"`
	Parse       Parse       `mapstructure:"parse"`
	ImageUrl    string      `mapstructure:"-"`
}

type GroupedRestaurants struct {
	Group       Group
	Restaurants []*Restaurant
}

type Parse struct {
	UpdateCron string     `mapstructure:"update_cron"`
	Navigate   []Selector `mapstructure:"navigate"`
	FileType   FileType   `mapstructure:"file_type"`
}

type Selector struct {
	Locator   string `mapstructure:"locator"`
	Attribute string `mapstructure:"attribute"`
	Style     string `mapstructure:"style"`
}

func init() {
	os.Mkdir(ConfigFolder, os.ModePerm)
	validate = validator.New()
}

func New() {
	viper.SetDefault("log_level", "info")
	viper.SetDefault("time_zone", "Etc/UTC")
	viper.SetDefault("server.address", "0.0.0.0")
	viper.SetDefault("server.port", 8156)

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(ConfigFolder)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			err = viper.WriteConfigAs(ConfigFolder + "config.yaml")
			if err != nil {
				slog.Error(err.Error())
				os.Exit(1)
			}
		} else {
			slog.Error("Failed to read configuration file", "error", err)
			os.Exit(1)
		}
	}

	if err := ValidateAndLoadConfig(viper.GetViper()); err != nil {
		slog.Error("Initial configuration validation failed", "error", err)
		os.Exit(1)
	}
}

func ValidateAndLoadConfig(v *viper.Viper) error {
	var tempCfg GlobalConfig
	if err := v.Unmarshal(&tempCfg); err != nil {
		return fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	if err := validate.Struct(tempCfg); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	tempCfg.GroupedRestaurants = computeGroupedRestaurantsForMap(tempCfg.Restaurants)

	mu.Lock()
	Cfg = tempCfg
	mu.Unlock()

	os.Setenv("TZ", Cfg.TimeZone)
	return nil
}

func computeGroupedRestaurantsForMap(restaurants map[string]*Restaurant) []GroupedRestaurants {
	groupMap := make(map[Group][]*Restaurant)
	for id, r := range restaurants {
		r.ID = id
		groupMap[r.Group] = append(groupMap[r.Group], r)
	}

	for _, list := range groupMap {
		sort.Slice(list, func(i, j int) bool {
			return list[i].Name < list[j].Name
		})
	}

	var groups []GroupedRestaurants
	for g, list := range groupMap {
		groups = append(groups, GroupedRestaurants{
			Group:       g,
			Restaurants: list,
		})
	}

	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Group < groups[j].Group
	})

	return groups
}

func ConfigLoaded() bool {
	return viper.ConfigFileUsed() != ""
}

func GetLogLevel() slog.Level {
	mu.RLock()
	defer mu.RUnlock()
	switch strings.ToLower(Cfg.LogLevel) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func GetServer() string {
	mu.RLock()
	defer mu.RUnlock()
	return fmt.Sprintf("%s:%d", Cfg.Server.Address, Cfg.Server.Port)
}

func GetRestaurants(favSet map[string]struct{}) []GroupedRestaurants {
	mu.RLock()
	defer mu.RUnlock()
	r := Cfg.GroupedRestaurants

	if len(favSet) == 0 {
		return r
	}

	var filtered []GroupedRestaurants
	var favourites []*Restaurant

	for _, group := range r {
		var filteredRestaurants []*Restaurant
		for _, restaurant := range group.Restaurants {
			if _, ok := favSet[strings.ToLower(restaurant.ID)]; ok {
				favourites = append(favourites, restaurant)
				continue
			}

			filteredRestaurants = append(filteredRestaurants, restaurant)
		}

		if len(filteredRestaurants) > 0 {
			filtered = append(filtered, GroupedRestaurants{
				Group:       group.Group,
				Restaurants: filteredRestaurants,
			})
		}
	}

	if len(favourites) > 0 {
		filtered = append([]GroupedRestaurants{{
			Group:       Favorites,
			Restaurants: favourites,
		}}, filtered...)
	}

	return filtered
}

func GetRestaurant(name string) (*Restaurant, error) {
	mu.RLock()
	defer mu.RUnlock()
	restaurant, exists := Cfg.Restaurants[name]
	if !exists {
		return nil, fmt.Errorf("restaurant %s not found", name)
	}
	return restaurant, nil
}
