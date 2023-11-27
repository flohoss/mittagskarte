package helper

import (
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func ClearAndTitleString(input string) string {
	caser := cases.Title(language.German)
	return caser.String(strings.ReplaceAll(strings.TrimSpace(input), "\n", " "))
}

func ClearString(input string) string {
	return strings.ReplaceAll(strings.TrimSpace(input), "\n", " ")
}
