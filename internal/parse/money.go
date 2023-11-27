package parse

import (
	"strconv"
	"strings"
)

func ConvertPrice(strPrice string) float64 {
	badStrings := []string{"EUR", "€"}
	result := strPrice
	for _, str := range badStrings {
		result = strings.Replace(result, str, "", 1)
	}
	result = strings.Replace(result, ",", ".", 1)
	result = strings.TrimSpace(result)
	price, _ := strconv.ParseFloat(result, 64)
	return price
}
