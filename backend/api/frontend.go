package api

import (
	"encoding/base64"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/flohoss/mittagskarte/internal/restaurant"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

type frontendPageData struct {
	Restaurants []*restaurant.Restaurant
	Email       string
}

func buildFrontendPageData(app core.App, email string) (*frontendPageData, error) {
	restaurants, err := restaurant.GetRestaurantsWithMenus(app)
	if err != nil {
		return nil, err
	}

	encodedEmail := base64.StdEncoding.EncodeToString([]byte(email))

	return &frontendPageData{
		Restaurants: restaurants,
		Email:       encodedEmail,
	}, nil
}

func ServeFrontend(app core.App, se *core.ServeEvent, email string, dev bool) error {
	frontendDist, err := filepath.Abs("dist")
	if err != nil {
		return err
	}

	if _, err := os.Stat(frontendDist); err != nil {
		return se.Next()
	}

	indexTemplate, err := template.New("index.html").Funcs(template.FuncMap{
		"toJSON": func(v any) template.JS {
			b, err := json.Marshal(v)
			if err != nil {
				log.Printf("failed to marshal frontend restaurants payload: %v", err)
				return template.JS("[]")
			}
			return template.JS(b)
		},
	}).ParseFiles(filepath.Join(frontendDist, "index.html"))
	if err != nil {
		return err
	}

	servePath := func(re *core.RequestEvent) error {
		requestPath := strings.TrimPrefix(re.Request.URL.Path, "/")
		if requestPath != "" {
			filePath := filepath.Join(frontendDist, filepath.Clean(requestPath))
			abs, err := filepath.Abs(filePath)
			if err != nil {
				return re.NotFoundError("not found", nil)
			}
			rel, err := filepath.Rel(frontendDist, abs)
			if err != nil || strings.HasPrefix(rel, "..") || rel == ".." {
				return re.NotFoundError("not found", nil)
			}
			if info, err := os.Stat(abs); err == nil && !info.IsDir() {
				http.ServeFile(re.Response, re.Request, abs)
				return nil
			}
			if filepath.Ext(requestPath) != "" {
				return re.NotFoundError("not found", nil)
			}
		}

		pageData, err := buildFrontendPageData(app, email)
		if err != nil {
			return err
		}

		re.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
		return indexTemplate.Execute(re.Response, pageData)
	}

	route := se.Router.GET("/{path...}", func(re *core.RequestEvent) error {
		return servePath(re)
	})
	if !dev {
		route.Bind(apis.SkipSuccessActivityLog())
	}

	return se.Next()
}
