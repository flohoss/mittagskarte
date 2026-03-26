package config

import (
	"errors"
	"fmt"
	"image"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/flohoss/mittagskarte/internal/checksum"
	"github.com/flohoss/mittagskarte/pkg/file"
)

type FileType string
type ParseType string

const (
	ConfigFolder = "./config/"

	PDF   FileType = "pdf"
	Image FileType = "image"

	Download ParseType = "download"
	Scrape   ParseType = "scrape"
	Upload   ParseType = "upload"
)

func GetAllowedExtensions() []string {
	return []string{".pdf", ".jpg", ".jpeg", ".png", ".webp"}
}

func GetAllowedExtensionsMessage() string {
	return fmt.Sprintf("invalid file extension, allowed are %s", strings.Join(GetAllowedExtensions(), ", "))
}

var (
	cfg  GlobalConfig
	once sync.Once
	mu   sync.RWMutex
)

type GlobalConfig struct {
	LogLevel    slog.Level             `json:"log_level"`
	TimeZone    string                 `json:"time_zone"`
	APIToken    string                 `json:"api_token"`
	Server      ServerSettings         `json:"server"`
	Restaurants map[string]*Restaurant `json:"restaurants"`
}

type ServerSettings struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
}

type Restaurant struct {
	ID        string   `json:"-"`
	Name      string   `json:"name"`
	Tags      []string `json:"tags"`
	PageUrl   string   `json:"url"`
	Address   string   `json:"address"`
	RestDays  []string `json:"rest_days"`
	Phone     string   `json:"phone"`
	Group     string   `json:"group"`
	CreatedAt Date     `json:"created_at"`
	Parse     Parse    `json:"parse"`
	Menu      Menu     `json:"-"`
	Loading   bool     `json:"-"`
}

type Menu struct {
	URL       string
	Modified  *time.Time
	Landscape bool
	Width     string
	Height    string
}

type GroupedRestaurants struct {
	Group       string
	Restaurants []*Restaurant
}

type Parse struct {
	Type        ParseType  `json:"type"`
	UpdateCron  string     `json:"update_cron"`
	Navigate    []Selector `json:"navigate"`
	DownloadURL string     `json:"download_url"`
	FileType    FileType   `json:"file_type"`
}

type Selector struct {
	Locator   string `json:"locator"`
	Attribute string `json:"attribute"`
	Style     string `json:"style"`
}

func init() {
	if err := os.MkdirAll(ConfigFolder, os.ModePerm); err != nil {
		slog.Error("Failed to create config directory", "error", err)
		os.Exit(1)
	}
}

func load() {
	once.Do(func() {
		var tempCfg GlobalConfig
		if err := file.Read(ConfigFolder, "config", &tempCfg); err != nil {
			slog.Error("Failed to read configuration file", "error", err)
			os.Exit(1)
		}

		normalizeRestaurant(tempCfg.Restaurants)
		os.Setenv("TZ", tempCfg.TimeZone)

		mu.Lock()
		cfg = tempCfg
		mu.Unlock()
	})
}

func normalizeRestaurant(restaurants map[string]*Restaurant) {
	for id, r := range restaurants {
		r.ID = id
	}
}

func get(getter func() any) any {
	load()
	mu.RLock()
	defer mu.RUnlock()
	return snapshot(getter())
}

func set(setter func()) {
	load()
	mu.Lock()
	defer mu.Unlock()
	setter()
}

func snapshot(v any) any {
	switch x := v.(type) {
	case map[string]*Restaurant:
		return snapshotRestaurants(x)
	case *Restaurant:
		if x == nil {
			return (*Restaurant)(nil)
		}
		r := snapshotRestaurant(*x)
		return &r
	default:
		return v
	}
}

func snapshotRestaurants(in map[string]*Restaurant) map[string]*Restaurant {
	out := make(map[string]*Restaurant, len(in))
	for id, r := range in {
		if r == nil {
			out[id] = nil
			continue
		}
		copyR := snapshotRestaurant(*r)
		out[id] = &copyR
	}
	return out
}

func snapshotRestaurant(r Restaurant) Restaurant {
	r.Tags = append([]string(nil), r.Tags...)
	r.RestDays = append([]string(nil), r.RestDays...)
	r.Parse.Navigate = append([]Selector(nil), r.Parse.Navigate...)
	if r.Menu.Modified != nil {
		modified := *r.Menu.Modified
		r.Menu.Modified = &modified
	}
	return r
}

func GetLogLevel() slog.Level {
	return get(func() any { return cfg.LogLevel }).(slog.Level)
}

func GetServer() string {
	return get(func() any { return fmt.Sprintf("%s:%d", cfg.Server.Address, cfg.Server.Port) }).(string)
}

func GetServerURL() string {
	return get(func() any { return fmt.Sprintf("http://%s", GetServer()) }).(string)
}

func GetRestaurants() map[string]*Restaurant {
	return get(func() any { return cfg.Restaurants }).(map[string]*Restaurant)
}

func GetRestaurant(id string) (*Restaurant, error) {
	res := get(func() any { return cfg.Restaurants[id] }).(*Restaurant)
	if res == nil {
		return nil, errors.New("restaurant not found")
	}
	return res, nil
}

func GetApiToken() string {
	return get(func() any { return cfg.APIToken }).(string)
}

func GetAllCrons() map[string]map[string]*Restaurant {
	restaurants := GetRestaurants()
	cronJobs := make(map[string]map[string]*Restaurant)

	for id, restaurant := range restaurants {
		if restaurant.Parse.UpdateCron == "" {
			continue
		}
		if _, ok := cronJobs[restaurant.Parse.UpdateCron]; !ok {
			cronJobs[restaurant.Parse.UpdateCron] = make(map[string]*Restaurant)
		}
		cronJobs[restaurant.Parse.UpdateCron][id] = restaurant
	}

	return cronJobs
}

func (r *Restaurant) SetLoading(loading bool) {
	set(func() {
		if current := cfg.Restaurants[r.ID]; current != nil {
			current.Loading = loading
			return
		}
		r.Loading = loading
	})
}

func (r *Restaurant) SetMenu(filePath string, modTime time.Time) {
	url := checksum.SuffixQuery(filePath)
	menu := Menu{
		URL:      url,
		Modified: &modTime,
	}

	f, err := os.Open(filePath)
	if err != nil {
		slog.Warn("Failed to open menu file for image decoding", "error", err, "filePath", filePath)
	} else {
		defer f.Close()
		img, _, decodeErr := image.Decode(f)
		if decodeErr == nil {
			menu.Landscape = img.Bounds().Dx() > img.Bounds().Dy()
			menu.Width = fmt.Sprintf("%dpx", img.Bounds().Dx())
			menu.Height = fmt.Sprintf("%dpx", img.Bounds().Dy())
		}
	}

	set(func() {
		if current := cfg.Restaurants[r.ID]; current != nil {
			current.Menu = menu
			return
		}
		r.Menu = menu
	})
}
