package config

type Restaurant struct {
	ID          string      `json:"id"`
	Price       uint8       `json:"price"`
	Icon        string      `json:"icon"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	PageURL     string      `json:"page_url"`
	Address     string      `json:"address"`
	RestDays    []DayOfWeek `json:"rest_days"`
	Phone       string      `json:"phone"`
	Group       Group       `json:"group"`
	Parse       Parse       `json:"parse"`
	Menu        Menu        `json:"menu"`
}

type Parse struct {
	HTTPVersion HTTPVersion  `json:"http_version"`
	Description Selector     `json:"description"`
	Navigate    []Selector   `json:"navigate"`
	IsFile      bool         `json:"is_file"`
	IsSMTP      bool         `json:"is_smtp"`
	OneForAll   OneForAll    `json:"one_for_all"`
	Food        []FoodParser `json:"food"`
}

type Selector struct {
	Fixed     string `json:"fixed"`
	Regex     string `json:"regex"`
	JQuery    string `json:"jquery"`
	Attribute string `json:"attribute"`
	Prefix    string `json:"prefix"`
}

type OneForAll struct {
	FixedPrice          float64 `json:"fixed_price"`
	Regex               string  `json:"regex"`
	PositionDay         uint8   `json:"pos_day"`
	PositionFood        uint8   `json:"pos_food"`
	PositionPrice       uint8   `json:"pos_price"`
	PositionDescription uint8   `json:"pos_description"`
}

type FoodParser struct {
	Day         Selector `json:"day"`
	Name        Selector `json:"name"`
	Price       Selector `json:"price"`
	Description Selector `json:"description"`
}

type FoodEntry struct {
	Day         string  `json:"day"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
}

type Menu struct {
	Description string      `json:"description"`
	Card        string      `json:"card"`
	Food        []FoodEntry `json:"food"`
}
