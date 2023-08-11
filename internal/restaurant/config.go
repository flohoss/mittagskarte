package restaurant

import (
	"encoding/json"
	"os"
)

type Configuration struct {
	Download          []ClickConfig  `json:"download"`
	Redirect          []ClickConfig  `json:"redirect"`
	DownloadPrefix    string         `json:"download_prefix"`
	RedirectPrefix    string         `json:"redirect_prefix"`
	DescriptionRegex  string         `json:"description_regex"`
	DescriptionInHtml bool           `json:"description_in_html"`
	FoodRegex         string         `json:"food_regex"`
	MaxFood           int            `json:"max_food"`
	FixPrice          float64        `json:"fix_price"`
	Positions         PositionConfig `json:"positions"`
	TrimImageEdges    bool           `json:"trim_image_edges"`
}

type ClickConfig struct {
	JQuery    string `json:"jquery"`
	Attribute string `json:"attribute"`
}

type PositionConfig struct {
	Name        int `json:"name"`
	Day         int `json:"day"`
	Price       int `json:"price"`
	Description int `json:"description"`
}

const ConfigLocation = "configs/restaurants/"

func parseConfig(path string) (Configuration, error) {
	var config Configuration
	content, err := os.ReadFile(path)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(content, &config)
	if err != nil {
		return config, err
	}
	return config, nil
}
