package router

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/Masterminds/sprig/v3"
	"github.com/goodsign/monday"
	"github.com/labstack/echo/v4"
	"gitlab.unjx.de/flohoss/mittag/internal/controller"
	"golang.org/x/image/webp"
)

type Template struct {
	templates map[string]*template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	template, ok := t.templates[name]
	if !ok {
		err := errors.New("Template not found -> " + name)
		return err
	}
	return template.ExecuteTemplate(w, "layout.html", data)
}

func templateString(files []string) []string {
	templatesFolder := "web/templates/"
	baseTemplates := []string{templatesFolder + "layout.html", templatesFolder + "icons.html"}
	for i := 0; i < len(files); i++ {
		files[i] = templatesFolder + files[i]
	}
	combined := append(baseTemplates, files[:]...)
	return combined
}

func isToday(food controller.Food) bool {
	var loc monday.Locale = monday.LocaleDeDE
	shortRegex := regexp.MustCompile(fmt.Sprintf(`%s\.`, monday.Format(time.Now(), "Mon", loc)))
	if food.Day == "" {
		for _, day := range monday.GetShortDays(loc) {
			if strings.Contains(food.Name, day+".") {
				return shortRegex.MatchString(food.Name)
			}
		}
		return true
	}
	longRegex := regexp.MustCompile("(?i)" + fmt.Sprintf(`^%s$|^%s$|^mo\.?\s?-\s?fr\.?$|alternativ|oder`,
		monday.Format(time.Now(), "Monday", monday.LocaleDeDE),
		monday.Format(time.Now(), "Monday, 02.01.", monday.LocaleDeDE),
	))
	return longRegex.MatchString(food.Day)
}

func isRestDay(restaurant controller.Restaurant) bool {
	for _, restDay := range restaurant.RestDays {
		if uint8(time.Now().Weekday()) == restDay {
			return true
		}
	}
	return false
}

func imageSize(image string) []int {
	var res []int
	img, err := os.Open(image)
	if err != nil {
		return res
	}
	defer img.Close()
	decImg, err := webp.DecodeConfig(img)
	if err != nil {
		return res
	}
	return []int{decImg.Height, decImg.Width}
}

func generateTemplate(files ...string) *template.Template {
	return template.Must(template.New("").Funcs(sprig.FuncMap()).Funcs(template.FuncMap{
		"isToday":   isToday,
		"isRestDay": isRestDay,
		"imageSize": imageSize,
	}).ParseFiles(templateString(files)...))
}

func initTemplates() *Template {
	templates := make(map[string]*template.Template)

	templates["countdown"] = generateTemplate("countdown/index.html")
	templates["settings"] = generateTemplate("settings/index.html")
	templates["restaurants"] = generateTemplate("restaurants/index.html")

	return &Template{templates: templates}
}
