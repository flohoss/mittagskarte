package handlers

import (
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/flohoss/mittagskarte/config"
	"github.com/flohoss/mittagskarte/services"
	"github.com/flohoss/mittagskarte/views"
	"github.com/labstack/echo/v4"
)

func contains(haistack []string, needle string) bool {
	for _, hai := range haistack {
		if hai == needle {
			return true
		}
	}
	return false
}

type MittagHandler struct {
	mittag *services.Mittag
}

func NewMittagHandler(mittag *services.Mittag) *MittagHandler {
	return &MittagHandler{
		mittag: mittag,
	}
}

func cookieReader(ctx echo.Context, fav string) map[string]string {
	group := url.QueryEscape(ctx.QueryParam("group"))

	favSet := make(map[string]string)
	if cookie, err := ctx.Cookie("favourites"); err == nil && cookie.Value != "" {
		for _, pair := range strings.Split(cookie.Value, ",") {
			if kv := strings.SplitN(pair, ":", 2); len(kv) == 2 && kv[0] != "" && kv[1] != "" {
				favSet[kv[0]] = kv[1]
			}
		}
	}

	if fav != "" {
		if _, exists := favSet[fav]; exists {
			delete(favSet, fav)
		} else {
			favSet[fav] = group
		}
	}

	if len(favSet) > 0 {
		favPairs := make([]string, 0, len(favSet))
		for k, v := range favSet {
			favPairs = append(favPairs, k+":"+v)
		}
		ctx.SetCookie(&http.Cookie{
			Name:     "favourites",
			Value:    strings.Join(favPairs, ","),
			Path:     "/",
			HttpOnly: true,
			MaxAge:   400 * 24 * 60 * 60,
		})
	} else {
		ctx.SetCookie(&http.Cookie{
			Name:   "favourites",
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})
	}

	return favSet
}

func (m *MittagHandler) handleFilter(ctx echo.Context) error {
	q := ctx.QueryParam("q")
	favSet := cookieReader(ctx, "")
	restaurants := config.GetGroupedRestaurants(favSet, q)

	return render(ctx, views.Index(restaurants))
}

func (m *MittagHandler) handleIndex(ctx echo.Context) error {
	fav := strings.ToLower(ctx.QueryParam("fav"))
	favSet := cookieReader(ctx, fav)

	if fav != "" {
		return ctx.Redirect(http.StatusFound, "/")
	}

	restaurants := config.GetGroupedRestaurants(favSet, "")
	return render(ctx, views.HomeIndex(views.Index(restaurants)))
}

func (m *MittagHandler) handleUpload(ctx echo.Context) error {
	r, err := config.GetRestaurant(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "File upload failed: "+err.Error())
	}

	if err := m.mittag.UploadMenu(ctx, r, file); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return m.handleFilter(ctx)
}

func (m *MittagHandler) handleDownload(ctx echo.Context) error {
	r, err := config.GetRestaurant(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	if ctx.QueryParam("url") == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "URL parameter is required")
	}

	parsedURL, err := url.Parse(ctx.QueryParam("url"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid URL: "+err.Error())
	}

	ext := filepath.Ext(parsedURL.Path)
	if !contains(config.GetAllowedExtensions(), ext) {
		return echo.NewHTTPError(http.StatusBadRequest, config.GetAllowedExtensionsMessage())
	}

	r.Parse.DownloadURL = parsedURL.String()
	r.Parse.FileType = config.FileType(ext)

	if err := m.mittag.GetImageUrl(r, true); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return m.handleFilter(ctx)
}

func (m *MittagHandler) handleUpdate(ctx echo.Context) error {
	r, err := config.GetRestaurant(ctx.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	if err := m.mittag.GetImageUrl(r, true); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return m.handleFilter(ctx)
}
