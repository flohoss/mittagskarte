package parse

import "github.com/PuerkitoBio/goquery"

type FoodParser struct {
	DocStorage  []*goquery.Document
	FileContent string
}

func NewFoodParser() *FoodParser {
	return &FoodParser{}
}
