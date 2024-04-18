package mittag

import (
	"errors"
	"strings"
)

type Restaurant struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	PageURL   string  `json:"page_url"`
	Address   string  `json:"address"`
	Selected  bool    `json:"selected"`
	RestDays  []uint8 `json:"rest_day"`
	Phone     string  `json:"phone"`
	Group     Group   `json:"group"`
	Thumbnail string  `json:"thumbnail"`
}

type Group uint8

var Groups = []Group{
	Fasanenhof,
	Esslingen,
	Feuerbach,
}

func StringToGroup(s string) (Group, error) {
	switch strings.ToLower(s) {
	case "fasanenhof":
		return Fasanenhof, nil
	case "esslingen":
		return Esslingen, nil
	case "feuerbach":
		return Feuerbach, nil
	default:
		return Fasanenhof, errors.New("not a valid group")
	}
}

const (
	Fasanenhof Group = iota + 1
	Esslingen
	Feuerbach
)

type Card struct {
	RestaurantID     string `json:"restaurant_id" gorm:"primaryKey"`
	Description      string `json:"description"`
	ImageURL         string `json:"image_url"`
	ExistingFileHash string `json:"existing_file_hash"`
	Refreshed        int64  `json:"refreshed"`
	Map              Map    `json:"map" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Food             []Food `json:"food" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	UpdatedAt        int64  `json:"updated_at" gorm:"autoUpdateTime"`
	CreatedAt        int64  `json:"created_at" gorm:"autoCreateTime"`
}

type Map struct {
	CardID   string `json:"card_id" gorm:"primaryKey"`
	Distance string `json:"distance"`
	Duration string `json:"duration"`
}

type Food struct {
	ID          uint64  `json:"id" gorm:"primaryKey"`
	CardID      string  `json:"card_id"`
	Name        string  `json:"name"`
	Day         string  `json:"day"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
}

type Selector struct {
	Regex     string `json:"regex"`
	JQuery    string `json:"jquery"`
	Attribute string `json:"attribute"`
	Fixed     string `json:"fixed"`
}

type Retrieve struct {
	JQuery    string `json:"jquery"`
	Attribute string `json:"attribute"`
	Prefix    string `json:"prefix"`
	Regex     string `json:"regex"`
}

type Download struct {
	IsFile   bool       `json:"is_file"`
	Cropping []Cropping `json:"cropping"`
}

type Cropping struct {
	Gravity string `json:"gravity"`
	Crop    string `json:"crop"`
	Keep    bool   `json:"keep"`
}

type FoodEntry struct {
	Day         Selector `json:"day"`
	Name        Selector `json:"name"`
	Price       Selector `json:"price"`
	Description Selector `json:"description"`
}

type OneForAll struct {
	FixedPrice          float64 `json:"fixed_price"`
	Regex               string  `json:"regex"`
	PositionDay         uint8   `json:"pos_day"`
	PositionFood        uint8   `json:"pos_food"`
	PositionPrice       uint8   `json:"pos_price"`
	PositionDescription uint8   `json:"pos_description"`
	Insensitive         bool    `json:"insensitive"`
	JQuery              JQuery  `json:"jquery"`
}

type JQuery struct {
	Wrapper     string `json:"wrapper"`
	Day         string `json:"day"`
	Food        string `json:"food"`
	Price       string `json:"price"`
	Description string `json:"description"`
}

type Menu struct {
	Description Selector    `json:"description"`
	Food        []FoodEntry `json:"food"`
	OneForAll   OneForAll   `json:"one_for_all"`
}

type Configuration struct {
	Restaurant          Restaurant `json:"restaurant"`
	HTTPOne             bool       `json:"http_one"`
	RetrieveDownloadUrl []Retrieve `json:"retrieve_download_url"`
	Download            Download   `json:"download"`
	Menu                Menu       `json:"menu"`
}

type ThumbnailItem struct {
	ID        string `json:"id"`
	Thumbnail string `json:"thumbnail"`
}

type ThumbnailData struct {
	Data []ThumbnailItem `json:"data"`
}
