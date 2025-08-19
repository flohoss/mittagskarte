package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
	"gitlab.unjx.de/flohoss/mittag/config"
	"gitlab.unjx.de/flohoss/mittag/views"
)

func handleIndex(ctx echo.Context) error {
	fav := strings.ToLower(ctx.QueryParam("fav"))

	var favSet map[string]struct{}
	cookie, err := ctx.Cookie("favourites")
	if err == nil {
		decodedValue, _ := url.QueryUnescape(cookie.Value)
		if err := json.Unmarshal([]byte(decodedValue), &favSet); err != nil || favSet == nil {
			favSet = make(map[string]struct{})
		}
	}

	if fav != "" {
		if _, exists := favSet[fav]; exists {
			delete(favSet, fav)
		} else {
			favSet[fav] = struct{}{}
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

	if fav != "" {
		return ctx.Redirect(http.StatusFound, "/")
	}

	restaurants := config.GetGroupedRestaurants(favSet)
	return render(ctx, views.HomeIndex(views.Index(restaurants)))
}
