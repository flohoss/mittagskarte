package handler

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/labstack/echo/v4"
	"gitlab.unjx.de/flohoss/mittag/internal/config"
	"gitlab.unjx.de/flohoss/mittag/internal/imdb"
	"gitlab.unjx.de/flohoss/mittag/internal/parse"
	"gitlab.unjx.de/flohoss/mittag/pgk/fetch"
)

type RestaurantHandler struct {
	restaurants map[string]*config.Restaurant
	imdb        *imdb.IMDb
}

func New(restaurants map[string]*config.Restaurant, imdb *imdb.IMDb) *RestaurantHandler {
	return &RestaurantHandler{
		restaurants: restaurants,
		imdb:        imdb,
	}
}

// GetAllRestaurants
//
//	@Produce	json
//	@Tags		restaurants
//	@Success	200	{object}	map[string]Restaurant	"ok"
//	@Router		/restaurants [get]
func (h *RestaurantHandler) GetAllRestaurants(ctx echo.Context) error {
	restaurants := make(map[string]Restaurant)
	for _, restaurant := range h.restaurants {
		restaurants[restaurant.ID] = ReduceRestaurant(restaurant)
	}
	return ctx.JSON(http.StatusOK, restaurants)
}

// GetAllRestaurantsGrouped
//
//	@Produce	json
//	@Tags		groups
//	@Success	200	{object}	map[config.Group][]Restaurant	"ok"
//	@Router		/groups [get]
func (h *RestaurantHandler) GetAllRestaurantsGrouped(ctx echo.Context) error {
	groups := make(map[config.Group][]Restaurant)
	for _, group := range config.AllGroups {
		groups[group] = []Restaurant{}
	}
	for _, restaurant := range h.restaurants {
		groups[restaurant.Group] = append(groups[restaurant.Group], ReduceRestaurant(restaurant))
	}
	return ctx.JSON(http.StatusOK, groups)
}

// GetRestaurant
//
//	@Produce	json
//	@Tags		restaurants
//	@Param		id	path		string			true	"Restaurant ID"
//	@Success	200	{object}	Restaurant		"ok"
//	@Failure	404	{object}	echo.HTTPError	"Can not find ID"
//	@Router		/restaurants/{id} [get]
func (h *RestaurantHandler) GetRestaurant(ctx echo.Context) error {
	id := ctx.Param("id")
	restaurant, ok := h.restaurants[id]
	if !ok {
		return echo.NewHTTPError(http.StatusNotFound, "Can not find ID")
	}
	return ctx.JSON(http.StatusOK, ReduceRestaurant(restaurant))
}

// UploadMenu
//
//	@Accept		multipart/form-data
//	@Tags		restaurants
//	@Param		id	path		string			true	"Restaurant ID"
//	@Param		file	mpfd		string			true	"Menu File"
//	@Param		token	mpfd		string			true	"API-Token"
//	@Success	200	{object}	nil		"ok"
//	@Failure	404	{object}	echo.HTTPError	"Can not find ID"
//	@Router		/restaurants/{id} [post]
func (h *RestaurantHandler) UploadMenu(ctx echo.Context) error {
	id := ctx.Param("id")
	restaurant, ok := h.restaurants[id]
	if !ok {
		return echo.NewHTTPError(http.StatusNotFound, "Can not find ID")
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	src, err := file.Open()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	defer src.Close()

	ext := filepath.Ext(file.Filename)
	fileName := filepath.Join(fetch.DownloadLocation, restaurant.ID+ext)
	dst, err := os.Create(fileName)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	defer dst.Close()
	if _, err = io.Copy(dst, src); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	go func() {
		ocr, outputFileLocation := parse.MoveAndParse(fileName, true)
		parser := parse.NewMenuParser(nil, ocr, &restaurant.Parse, outputFileLocation)

		restaurant.Menu = *parser.Menu
		restaurant.SaveMenu(h.imdb)
	}()

	return ctx.NoContent(http.StatusOK)
}
