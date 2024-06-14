package config

type Restaurant struct {
	ID       string      `json:"id"`
	Icon     string      `json:"icon"`
	Name     string      `json:"name"`
	PageURL  string      `json:"page_url"`
	Address  string      `json:"address"`
	RestDays []DayOfWeek `json:"rest_days"`
	Phone    string      `json:"phone"`
	Group    Group       `json:"group"`
	Parse    Parse       `json:"parse"`
	Menu     Menu
}

type Parse struct {
	HTTPVersion HTTPVersion  `json:"http_version"`
	Description Selector     `json:"description"`
	Navigate    []Selector   `json:"navigate"`
	IsFile      bool         `json:"is_file"`
	Crop        []Crop       `json:"crop"`
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

type Crop struct {
	Crop    string `json:"crop"`
	Gravity string `json:"gravity"`
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
	Day         Selector
	Name        Selector
	Price       Selector
	Description Selector
}

type FoodEntry struct {
	Day         string
	Name        string
	Price       float64
	Description string
}

type Menu struct {
	Description string
	Card        string
	Food        []FoodEntry
}
