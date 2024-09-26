package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"gitlab.unjx.de/flohoss/mittag/services"
)

type MittagHandler struct {
	mittag *services.Mittag
}

func NewMittagHandler(mittag *services.Mittag) *MittagHandler {
	return &MittagHandler{
		mittag: mittag,
	}
}

// @Summary	Get all restaurants
// @Produce	json
// @Tags		restaurants
// @Success	200	{object}	map[string]services.Restaurant	"ok"
// @Router		/restaurants [get]
func (h *MittagHandler) GetAllRestaurants(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, h.mittag.GetAllRestaurants())
}

// @Summary	Get a single restaurant
// @Produce	json
// @Tags		restaurants
// @Param		id	path		string				true	"Restaurant ID"
// @Success	200	{object}	services.Restaurant	"ok"
// @Failure	404	{object}	echo.HTTPError		"Can not find ID"
// @Router		/restaurants/{id} [get]
func (h *MittagHandler) GetRestaurant(ctx echo.Context) error {
	return h.mittag.GetRestaurant(ctx)
}

// @Summary	Upload a menu
// @Accept		multipart/form-data
// @Tags		restaurants
// @Param		Authorization	header		string			true	"Bearer <Add access token here>"
// @Param		id				path		string			true	"Restaurant ID"
// @Param		file			formData	file			true	"Menu File"
// @Param		token			formData	string			true	"API-Token"
// @Success	200				{object}	nil				"ok"
// @Failure	404				{object}	echo.HTTPError	"Can not find ID"
// @Router		/restaurants/{id} [post]
func (h *MittagHandler) UploadMenu(ctx echo.Context) error {
	return h.mittag.UploadMenu(ctx)
}
