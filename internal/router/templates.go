package router

import (
	"errors"
	"html/template"
	"io"
	"os"

	"github.com/Masterminds/sprig/v3"
	"github.com/labstack/echo/v4"
	"gitlab.unjx.de/flohoss/mittag/internal/mittag"
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

func posInArray(str string, arr []string) int {
	for i, s := range arr {
		if s == str {
			return i
		}
	}
	return -1
}

func isToday(food mittag.Food) bool {
	if food.Day == "" || posInArray(food.Day, mittag.GetTodayActiveList()) != -1 {
		return true
	}
	return false
}

func nothingFound(card mittag.Card) bool {
	return len(card.Food) == 0 && card.ImageURL == ""
}

func imageSize(image string) []int {
	res := []int{0, 0}
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
		"isToday":      isToday,
		"isRestDay":    mittag.IsRestDay,
		"imageSize":    imageSize,
		"nothingFound": nothingFound,
	}).ParseFiles(templateString(files)...))
}

func initTemplates() *Template {
	templates := make(map[string]*template.Template)

	templates["countdown"] = generateTemplate("countdown/index.html")
	templates["settings"] = generateTemplate("settings/index.html")
	templates["restaurants"] = generateTemplate("restaurants/index.html")
	templates["groups"] = generateTemplate("groups/index.html")

	return &Template{templates: templates}
}
