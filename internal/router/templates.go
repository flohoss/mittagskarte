package router

import (
	"errors"
	"html/template"
	"io"
	"os"
	"regexp"
	"time"

	"github.com/Masterminds/sprig/v3"
	"github.com/goodsign/monday"
	"github.com/labstack/echo/v4"
	"gitlab.unjx.de/flohoss/mittag/internal/restaurant"
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

func isToday(food restaurant.Food) bool {
	expr := regexp.MustCompile(monday.Format(time.Now(), "Monday", monday.LocaleDeDE))
	if food.Day == "" {
		return true
	}
	return expr.MatchString(food.Day)
}

func isRestDay(restaurant restaurant.Restaurant) bool {
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
