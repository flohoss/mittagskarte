package controller

type CardType uint8

type Restaurant struct {
	ID           string   `json:"id" gorm:"primaryKey"`
	Name         string   `json:"name"`
	PageURL      string   `json:"page_url"`
	Street       string   `json:"street"`
	StreetNumber string   `json:"street_number"`
	ZipCode      string   `json:"zip_code"`
	City         string   `json:"city"`
	Latitude     float64  `json:"latitude"`
	Longitude    float64  `json:"longitude"`
	Selected     bool     `json:"selected"`
	CardType     CardType `json:"card_type"`
	RestDays     []uint8  `json:"rest_day"`
	Phone        string   `json:"phone"`
	Card         Card     `json:"card" gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

type Card struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	RestaurantID string `json:"restaurant_id"`
	Description  string `json:"description"`
	ImageURL     string `json:"image_url"`
	Year         int    `json:"year"`
	Week         int    `json:"week"`
	Day          int    `json:"day"`
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

func (c *Controller) MigrateModels() {
	c.orm.AutoMigrate(&Restaurant{})
	c.orm.AutoMigrate(&Card{})
	c.orm.AutoMigrate(&Food{})

	restaurants := []Restaurant{{
		ID:           "da-peppone",
		Name:         "Da Peppone",
		PageURL:      "https://dapeppone-restaurant.de/",
		Street:       "Bopseräcker",
		StreetNumber: "1",
		ZipCode:      "70597",
		City:         "Stuttgart",
		Latitude:     48.7333860132231,
		Longitude:    9.168711897889501,
		RestDays:     []uint8{0, 1, 6},
		Phone:        "+49 711 78784911",
	}, {
		ID:           "meet-and-eat",
		Name:         "Meet & Eat",
		PageURL:      "http://meetundeat.com/#Speisekarte",
		Street:       "Schelmenwasenstraße",
		StreetNumber: "7",
		ZipCode:      "70567",
		City:         "Stuttgart",
		Latitude:     48.71425688848292,
		Longitude:    9.166203979395853,
		RestDays:     []uint8{0, 6},
		Phone:        "+49 157 30664255",
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
	}, {
		ID:           "picchiorosso",
		Name:         "Picchiorosso",
		PageURL:      "https://picchiorosso.de/tageskarte/",
		Street:       "Zettachring",
		StreetNumber: "12",
		ZipCode:      "70567",
		City:         "Stuttgart",
		Latitude:     48.70824771163106,
		Longitude:    9.171523914268654,
		RestDays:     []uint8{0, 6},
		Phone:        "+49 711 7156767",
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
	}, {
		ID:           "sw34",
		Name:         "SW34",
		PageURL:      "https://sw34.restaurant/essen-trinken/",
		Street:       "Schelmenwasenstraße",
		StreetNumber: "34",
		ZipCode:      "70567",
		City:         "Stuttgart",
		Latitude:     48.70799627023862,
		Longitude:    9.1695723175327,
		RestDays:     []uint8{0, 6},
		Phone:        "+49 711 62042252",
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
	}, {
		ID:           "linde",
		Name:         "Linde",
		PageURL:      "https://www.gasthauslinde.de/Mittagstisch",
		Street:       "Wasserburgstraße",
		StreetNumber: "3",
		ZipCode:      "72622",
		City:         "Nürtingen",
		Latitude:     48.641688982054845,
		Longitude:    9.347466800930759,
		RestDays:     []uint8{0, 1, 6},
		Phone:        "+49 7022 62306 ",
	}}

	for _, restaurant := range restaurants {
		var res Restaurant
		amount := c.orm.Where("name = ?", restaurant.Name).Find(&res).RowsAffected
		if amount == 0 {
			c.orm.Create(&restaurant)
		}
	}
}
