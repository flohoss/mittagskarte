package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
	"gitlab.unjx.de/flohoss/mittag/config"
	"gitlab.unjx.de/flohoss/mittag/services"
	"gitlab.unjx.de/flohoss/mittag/views"
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

func (m *MittagHandler) handleIndex(ctx echo.Context) error {
	fav := strings.ToLower(ctx.QueryParam("fav"))
	group := ctx.QueryParam("group")

	var favSet map[string]string
	cookie, err := ctx.Cookie("favourites")
	if err == nil {
		decodedValue, _ := url.QueryUnescape(cookie.Value)
		if err := json.Unmarshal([]byte(decodedValue), &favSet); err != nil || favSet == nil {
			favSet = make(map[string]string)
		}
	}

	if fav != "" {
		if _, exists := favSet[fav]; exists {
			delete(favSet, fav)
		} else {
			favSet[fav] = group
		}
	}

	jsonValue, _ := json.Marshal(favSet)
	encodedValue := url.QueryEscape(string(jsonValue))
	newCookie := &http.Cookie{
		Name:     "favourites",
		Value:    encodedValue,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   30 * 24 * 60 * 60,
	}
	ctx.SetCookie(newCookie)

	if fav != "" && group != "" {
		ctx.SetCookie(&http.Cookie{
			Name:     "lastGroup",
			Value:    url.QueryEscape(group),
			Path:     "/",
			HttpOnly: true,
			MaxAge:   60,
		})
	}

	preselectedGroup := ""
	lastFavCookie, err := ctx.Cookie("lastGroup")
	if err == nil {
		preselectedGroup, _ = url.QueryUnescape(lastFavCookie.Value)

		ctx.SetCookie(&http.Cookie{
			Name:   "lastGroup",
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})
	}

	if fav != "" {
		return ctx.Redirect(http.StatusFound, "/")
	}

	restaurants := config.GetGroupedRestaurants(favSet)
	return render(ctx, views.HomeIndex(views.Index(restaurants, preselectedGroup)))
}

func (m *MittagHandler) handleUpload(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Restaurant ID is required")
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "File upload failed: "+err.Error())
	}

	return m.mittag.UploadMenu(ctx, id, file)
}
