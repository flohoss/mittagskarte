package config

import (
	"fmt"
	"image"
	"log/slog"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/flohoss/mittagskarte/internal/hash"
	"github.com/fsnotify/fsnotify"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type FileType string

const (
	ConfigFolder = "./config/"

	PDF   FileType = "pdf"
	Image FileType = "image"

	Favorites string = "Favoriten"
)

var cfg GlobalConfig

var validate *validator.Validate
var mu sync.RWMutex

type GlobalConfig struct {
	LogLevel           string                 `mapstructure:"log_level" validate:"omitempty,oneof=debug info warn error"`
	TimeZone           string                 `mapstructure:"time_zone" validate:"required"`
	APIToken           string                 `mapstructure:"api_token" validate:"required"`
	Meta               Meta                   `mapstructure:"meta"`
	Impressum          Impressum              `mapstructure:"impressum"`
	UMAMIAnalytics     UMAMIAnalytics         `mapstructure:"umami_analytics"`
	Server             ServerSettings         `mapstructure:"server"`
	Restaurants        map[string]*Restaurant `mapstructure:"restaurants"`
	GroupedRestaurants []GroupedRestaurants   `mapstructure:"-"`
}

type Meta struct {
	Title       string   `mapstructure:"title" validate:"required"`
	Description string   `mapstructure:"description" validate:"required"`
	Social      []Social `mapstructure:"social"`
}

type Impressum struct {
	Enabled     bool   `mapstructure:"enabled"`
	Responsible string `mapstructure:"responsible"`
	Email       string `mapstructure:"email"`
}

type UMAMIAnalytics struct {
	Enabled   bool   `mapstructure:"enabled"`
	Domain    string `mapstructure:"domain"`
	WebsiteID string `mapstructure:"website_id"`
}

type ServerSettings struct {
	Address string `mapstructure:"address" validate:"required,ipv4"`
	Port    int    `mapstructure:"port" validate:"required,gte=1024,lte=65535"`
}

type Restaurant struct {
	ID            string              `mapstructure:"-"`
	Name          string              `mapstructure:"name"`
	Tags          []string            `mapstructure:"tags"`
	PageUrl       string              `mapstructure:"url"`
	Address       string              `mapstructure:"address"`
	RestDaysSlice []string            `mapstructure:"rest_days"`
	RestDays      map[string]struct{} `mapstructure:"-"`
	Phone         string              `mapstructure:"phone"`
	Group         string              `mapstructure:"group"`
	New           bool                `mapstructure:"new"`
	Parse         Parse               `mapstructure:"parse"`
	Menu          Menu                `mapstructure:"-"`
	Loading       bool                `mapstructure:"-"`
}

type Menu struct {
	URL       string     `mapstructure:"-"`
	Modified  *time.Time `mapstructure:"-"`
	Landscape bool       `mapstructure:"-"`
	Width     string     `mapstructure:"-"`
	Height    string     `mapstructure:"-"`
}

type GroupedRestaurants struct {
	Group       string
	Restaurants []*Restaurant
}

type Parse struct {
	UpdateCron     string     `mapstructure:"update_cron"`
	Navigate       []Selector `mapstructure:"navigate"`
	DirectDownload string     `mapstructure:"direct_download"`
	FileType       FileType   `mapstructure:"file_type"`
}

type Selector struct {
	Locator   string `mapstructure:"locator"`
	Attribute string `mapstructure:"attribute"`
	Style     string `mapstructure:"style"`
}

type Social struct {
	Icon        string `mapstructure:"icon"`
	URL         string `mapstructure:"url"`
	Description string `mapstructure:"description"`
}

func matches(q, s string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(q))
}

func init() {
	os.Mkdir(ConfigFolder, os.ModePerm)
	validate = validator.New()
}

func New() {
	viper.SetDefault("log_level", "info")
	viper.SetDefault("time_zone", "Europe/Berlin")
	viper.SetDefault("api_token", "replace-me")
	viper.SetDefault("server", ServerSettings{
		Address: "0.0.0.0",
		Port:    8156,
	})
	viper.SetDefault("impressum", Impressum{})
	viper.SetDefault("umami_analytics", UMAMIAnalytics{})
	viper.SetDefault("meta", Meta{
		Title:       "Schniddzl.de",
		Description: "deine Mittagskarte für die Region Stuttgart",
		Social:      []Social{},
	})
	viper.SetDefault("restaurants", map[string]*Restaurant{
		"sw34": {
			Name:    "SW34",
			Tags:    []string{"Eventlocation", "Modern", "Gehobene Küche", "Crossover", "Stylisch"},
			PageUrl: "https://sw34.restaurant/essen-trinken",
			Address: "Schelmenwasenstraße 34, 70567 Stuttgart-Fasanenhof",
			Phone:   "+49 711 62042252",
			Group:   "Fasanenhof",
			RestDays: map[string]struct{}{
				"Saturday": {},
				"Sunday":   {},
			},
		},
	})

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

	setupViperWatcher()
}

func setupViperWatcher() {
	var (
		mu    sync.Mutex
		timer *time.Timer
	)

	debounce := func(d time.Duration, fn func()) {
		mu.Lock()
		defer mu.Unlock()

		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(d, fn)
	}

	viper.OnConfigChange(func(e fsnotify.Event) {
		debounce(2*time.Second, func() {
			slog.Info("Config changed, reloading")
			err := ValidateAndLoadConfig(viper.GetViper())
			if err != nil {
				slog.Error("Failed to reload configuration, keeping old settings", "error", err)
				return
			}
			slog.Info("Config reloaded successfully")
		})
	})

	viper.WatchConfig()
}

func ValidateAndLoadConfig(v *viper.Viper) error {
	var tempCfg GlobalConfig
	if err := v.Unmarshal(&tempCfg); err != nil {
		return fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	if err := validate.Struct(tempCfg); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	normalizeRestaurant(tempCfg.Restaurants)
	tempCfg.GroupedRestaurants = computeGroupedRestaurantsForMap(tempCfg.Restaurants)

	mu.Lock()
	cfg = tempCfg
	mu.Unlock()

	os.Setenv("TZ", cfg.TimeZone)
	return nil
}

func normalizeRestaurant(restaurants map[string]*Restaurant) {
	for _, r := range restaurants {
		r.RestDays = make(map[string]struct{})
		for _, day := range r.RestDaysSlice {
			r.RestDays[day] = struct{}{}
		}
	}
}

func computeGroupedRestaurantsForMap(restaurants map[string]*Restaurant) []GroupedRestaurants {
	groupMap := make(map[string][]*Restaurant)
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
	switch strings.ToLower(cfg.LogLevel) {
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
	return fmt.Sprintf("%s:%d", cfg.Server.Address, cfg.Server.Port)
}

func GetRestaurants() map[string]*Restaurant {
	mu.RLock()
	defer mu.RUnlock()
	return cfg.Restaurants
}

func GetGroupedRestaurants(favSet map[string]string, filter string) []GroupedRestaurants {
	mu.RLock()
	defer mu.RUnlock()
	r := cfg.GroupedRestaurants

	if len(favSet) == 0 && filter == "" {
		return r
	}

	var filtered []GroupedRestaurants
	var favourites []*Restaurant

	for _, group := range r {
		var filteredRestaurants []*Restaurant
		for _, restaurant := range group.Restaurants {
			if filter == "" || matches(filter, restaurant.Name) {
				if _, ok := favSet[strings.ToLower(restaurant.ID)]; ok {
					favourites = append(favourites, restaurant)
					continue
				}

				filteredRestaurants = append(filteredRestaurants, restaurant)
			}
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

func GetRestaurant(id string) (*Restaurant, error) {
	mu.RLock()
	defer mu.RUnlock()
	restaurant, exists := cfg.Restaurants[id]
	if !exists {
		return nil, fmt.Errorf("restaurant %s not found", id)
	}
	return restaurant, nil
}

func SetMenu(filePath string, modTime time.Time, restaurantID string) {
	mu.Lock()
	defer mu.Unlock()

	url := hash.AddHashQueryToFileName(filePath)

	cfg.Restaurants[restaurantID].Menu = Menu{
		URL:      url,
		Modified: &modTime,
	}

	f, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer f.Close()
	image, _, err := image.Decode(f)
	if err != nil {
		return
	}
	cfg.Restaurants[restaurantID].Menu.Landscape = image.Bounds().Dx() > image.Bounds().Dy()
	cfg.Restaurants[restaurantID].Menu.Width = fmt.Sprintf("%dpx", image.Bounds().Dx())
	cfg.Restaurants[restaurantID].Menu.Height = fmt.Sprintf("%dpx", image.Bounds().Dy())

	slog.Debug("Menu updated", "restaurantID", restaurantID, "url", url, "modified", modTime.String())
}

func GetApiToken() string {
	mu.RLock()
	defer mu.RUnlock()
	return cfg.APIToken
}

func (r *Restaurant) SetLoading(loading bool) {
	mu.Lock()
	defer mu.Unlock()
	r.Loading = loading
}

func (r *Restaurant) IsClosed() bool {
	mu.RLock()
	defer mu.RUnlock()
	today := time.Now().Weekday().String()
	_, exists := r.RestDays[today]
	return exists
}

func GetMeta() Meta {
	mu.RLock()
	defer mu.RUnlock()
	return cfg.Meta
}

func GetImpressum() Impressum {
	mu.RLock()
	defer mu.RUnlock()
	return cfg.Impressum
}

func GetAnalytics() UMAMIAnalytics {
	mu.RLock()
	defer mu.RUnlock()
	return cfg.UMAMIAnalytics
}
