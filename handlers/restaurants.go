package handlers

import (
	"net/http"
	"strconv"
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

func cookieReader(ctx echo.Context, fav string) (map[string]string, bool, config.Group) {
	group := ctx.QueryParam("group")

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

	if fav != "" && group != "" {
		ctx.SetCookie(&http.Cookie{
			Name:     "lastGroup",
			Value:    group,
			Path:     "/",
			HttpOnly: true,
			MaxAge:   60,
		})
	}

	var preselectedGroup config.Group
	if lastFavCookie, err := ctx.Cookie("lastGroup"); err == nil {
		if val, err := strconv.Atoi(lastFavCookie.Value); err == nil {
			preselectedGroup = config.Group(uint8(val))
		}
		ctx.SetCookie(&http.Cookie{
			Name:   "lastGroup",
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})
	}

	return favSet, len(favSet) > 0, preselectedGroup
}

func (m *MittagHandler) handleFilter(ctx echo.Context) error {
	q := ctx.QueryParam("q")
	favSet, favApplied, _ := cookieReader(ctx, "")
	restaurants := config.GetGroupedRestaurants(favSet, q)

	return render(ctx, views.Index(restaurants, favApplied, q != "", 0))
}

func (m *MittagHandler) handleIndex(ctx echo.Context) error {
	fav := strings.ToLower(ctx.QueryParam("fav"))
	favSet, favApplied, preselectedGroup := cookieReader(ctx, fav)

	if fav != "" {
		return ctx.Redirect(http.StatusFound, "/")
	}

	restaurants := config.GetGroupedRestaurants(favSet, "")
	return render(ctx, views.HomeIndex(views.Index(restaurants, favApplied, false, preselectedGroup)))
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
