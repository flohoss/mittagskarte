package restaurant

import "gorm.io/gorm"

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
	Florian,
	Andre,
}

const (
	Fasanenhof Group = iota + 1
	Florian
	Andre
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

func MigrateModels(orm *gorm.DB) {
	orm.AutoMigrate(&Restaurant{})
	orm.AutoMigrate(&Card{})
	orm.AutoMigrate(&Food{})

	restaurants := []Restaurant{{
		ID:           "da-peppone",
		Name:         "Da Peppone",
		PageURL:      "https://dapeppone-restaurant.de/",
		Street:       "Bopseräcker",
		StreetNumber: "1",
		ZipCode:      "70597",
		City:         "Stuttgart-Degerloch",
		Latitude:     48.7333860132231,
		Longitude:    9.168711897889501,
		RestDays:     []uint8{0, 1, 6},
		Phone:        "+49 711 78784911",
		Group:        Fasanenhof,
	}, {
		ID:           "meet-and-eat",
		Name:         "Meet & Eat",
		PageURL:      "http://meetundeat.com/#Speisekarte",
		Street:       "Schelmenwasenstraße",
		StreetNumber: "7",
		ZipCode:      "70567",
		City:         "Stuttgart-Fasanenhof",
		Latitude:     48.71425688848292,
		Longitude:    9.166203979395853,
		RestDays:     []uint8{0, 6},
		Phone:        "+49 157 30664255",
		Group:        Fasanenhof,
	}, {
		ID:           "paulaner",
		Name:         "Paulaner",
		PageURL:      "https://paulaner-le.de/speisen-getraenke/mittagstisch.html",
		Street:       "Burgstraße",
		StreetNumber: "4",
		ZipCode:      "70771",
		City:         "Leinfelden-Echterdingen",
		Latitude:     48.68905406629845,
		Longitude:    9.169774230976934,
		RestDays:     []uint8{0, 1, 6},
		Phone:        "+49 711 7944180",
		Group:        Fasanenhof,
	}, {
		ID:           "picchiorosso",
		Name:         "Picchiorosso",
		PageURL:      "https://picchiorosso.de/tageskarte/",
		Street:       "Zettachring",
		StreetNumber: "12",
		ZipCode:      "70567",
		City:         "Stuttgart-Fasanenhof",
		Latitude:     48.70824771163106,
		Longitude:    9.171523914268654,
		RestDays:     []uint8{0, 6},
		Phone:        "+49 711 7156767",
		Group:        Fasanenhof,
	}, {
		ID:           "ratsstuben",
		Name:         "Ratsstuben",
		PageURL:      "https://ratsstuben.de/wochenkarte/",
		Street:       "Bernhäuser Straße",
		StreetNumber: "16",
		ZipCode:      "70771",
		City:         "Leinfelden-Echterdingen",
		Latitude:     48.68823623924175,
		Longitude:    9.16919593566195,
		RestDays:     []uint8{0, 6},
		Phone:        "+49 711 791725",
		Group:        Fasanenhof,
	}, {
		ID:           "sw34",
		Name:         "SW34",
		PageURL:      "https://sw34.restaurant/essen-trinken/",
		Street:       "Schelmenwasenstraße",
		StreetNumber: "34",
		ZipCode:      "70567",
		City:         "Stuttgart-Fasanenhof",
		Latitude:     48.70799627023862,
		Longitude:    9.1695723175327,
		RestDays:     []uint8{0, 6},
		Phone:        "+49 711 62042252",
		Group:        Fasanenhof,
	}, {
		ID:           "schwedenscheuer",
		Name:         "Schwedenscheuer",
		PageURL:      "https://www.schwedenscheuer.de/restaurant-leinfelden-echterdingen/tageskarte.html",
		Street:       "Hauptstraße",
		StreetNumber: "71/1",
		ZipCode:      "70771",
		City:         "Leinfelden-Echterdingen",
		Latitude:     48.688405113912154,
		Longitude:    9.166638055590445,
		RestDays:     []uint8{0, 6},
		Phone:        "+49 711 7978527",
		Group:        Fasanenhof,
	}, {
		ID:           "koe5",
		Name:         "KÖ5",
		PageURL:      "https://koe5.de/",
		Street:       "Schelmenwasenstraße",
		StreetNumber: "5",
		ZipCode:      "70567",
		City:         "Stuttgart-Fasanenhof",
		Latitude:     48.71539920482492,
		Longitude:    9.165393403838493,
		RestDays:     []uint8{0, 6},
		Phone:        "+49 711 99772624",
		Group:        Fasanenhof,
	}, {
		ID:           "benz",
		Name:         "Benz",
		PageURL:      "https://benz-metzgerei-feinkost.de/schwanen-metzgerei/tagesessen/",
		Street:       "Hirschstraße",
		StreetNumber: "18",
		ZipCode:      "73257",
		City:         "Köngen",
		Latitude:     48.68396263959295,
		Longitude:    9.362366408284398,
		RestDays:     []uint8{0, 6},
		Phone:        "+49 7024 81488",
		Group:        Florian,
	}, {
		ID:           "fass",
		Name:         "Fass",
		PageURL:      "http://www.gasthaus-fass.de/mittagstisch/index.php",
		Street:       "Bahnhofstraße",
		StreetNumber: "10",
		ZipCode:      "72644",
		City:         "Oberboihingen",
		Latitude:     48.64802392359741,
		Longitude:    9.365059991503585,
		RestDays:     []uint8{0, 2, 6},
		Phone:        "+49 7022 61185",
		Group:        Florian,
	}, {
		ID:           "linde",
		Name:         "Linde",
		PageURL:      "https://www.gasthauslinde.de/Mittagstisch",
		Street:       "Wasserburgstraße",
		StreetNumber: "3",
		ZipCode:      "72622",
		City:         "Nürtingen-Zizishausen",
		Latitude:     48.641688982054845,
		Longitude:    9.347466800930759,
		RestDays:     []uint8{0, 1, 6},
		Phone:        "+49 7022 62306",
		Group:        Florian,
	}, {
		ID:           "hoflieferant-munz",
		Name:         "Munz",
		PageURL:      "https://www.hoflieferant-munz.de/aktuelles",
		Street:       "Stuttgarter Straße",
		StreetNumber: "23",
		ZipCode:      "70469",
		City:         "Stuttgart-Feuerbach",
		Latitude:     48.810595643895155,
		Longitude:    9.165437084289135,
		RestDays:     []uint8{0},
		Phone:        "+49 711 365 914 00",
		Group:        Andre,
	}, {
		ID:           "schaible",
		Name:         "Schaible",
		PageURL:      "https://www.feuerbach.de/aktuelles/mittagstisch/",
		Street:       "Staufeneckstraße",
		StreetNumber: "1",
		ZipCode:      "70469",
		City:         "Stuttgart-Feuerbach",
		Latitude:     48.80762311615608,
		Longitude:    9.156540084375933,
		RestDays:     []uint8{0, 6},
		Phone:        "+49 711 8104528",
		Group:        Andre,
	}, {
		ID:           "sg-sushi",
		Name:         "SG Sushi",
		PageURL:      "https://sg-sushi.de/mittagstisch/index.php",
		Street:       "Feuerbacher-Tal-Straße",
		StreetNumber: "1",
		ZipCode:      "70469",
		City:         "Stuttgart-Feuerbach",
		Latitude:     48.80844745582471,
		Longitude:    9.157192535139712,
		RestDays:     []uint8{0, 6},
		Phone:        "+49 711 80674572",
		Group:        Andre,
	}, {
		ID:           "troelsch-feuerbach",
		Name:         "Troelsch",
		PageURL:      "https://www.troelsch.de/ueber-uns/standorte/",
		Street:       "Stuttgarter Straße",
		StreetNumber: "104",
		ZipCode:      "70469",
		City:         "Stuttgart-Feuerbach",
		Latitude:     48.809072665993405,
		Longitude:    9.158019929356717,
		RestDays:     []uint8{0},
		Phone:        "+49 711 85680030",
		Group:        Andre,
	}}

	for _, restaurant := range restaurants {
		var res Restaurant
		amount := orm.Where("name = ?", restaurant.Name).Find(&res).RowsAffected
		if amount == 0 {
			orm.Create(&restaurant)
		}
	}
}
