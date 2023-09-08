package restaurant

type Restaurant struct {
	ID           string  `json:"id" gorm:"primaryKey"`
	Name         string  `json:"name"`
	PageURL      string  `json:"page_url"`
	Street       string  `json:"street"`
	StreetNumber string  `json:"street_number"`
	ZipCode      string  `json:"zip_code"`
	City         string  `json:"city"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	Selected     bool    `json:"selected"`
	RestDays     []uint8 `json:"rest_day"`
	Phone        string  `json:"phone"`
	Group        Group   `json:"group"`
	Card         Card    `json:"card" gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

type Group uint8

var Groups = []Group{
	Fasanenhof,
	Esslingen,
	Feuerbach,
}

const (
	Fasanenhof Group = iota + 1
	Esslingen
	Feuerbach
)

type Card struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	RestaurantID string `json:"restaurant_id"`
	Description  string `json:"description"`
	ImageURL     string `json:"image_url"`
	Food         []Food `json:"food" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	CreatedAt    int64  `json:"created_at" gorm:"autoCreateTime"`
}

type Food struct {
	ID          uint    `json:"id" gorm:"primaryKey"`
	CardID      uint    `json:"card_id"`
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
	Food        Selector `json:"food"`
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
