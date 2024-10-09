package handler

import (
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

//	@Summary	Get all restaurants
//	@Produce	json
//	@Tags		restaurants
//	@Success	200	{object}	map[string]services.CleanRestaurant	"ok"
//	@Router		/restaurants [get]
func (h *MittagHandler) GetAllRestaurants(ctx echo.Context) error {
	return h.mittag.GetAllRestaurants(ctx)
}

//	@Summary	Get a single restaurant
//	@Produce	json
//	@Tags		restaurants
//	@Param		id	path		string						true	"Restaurant ID"
//	@Success	200	{object}	services.CleanRestaurant	"ok"
//	@Failure	404	{object}	echo.HTTPError				"Can not find ID"
//	@Router		/restaurants/{id} [get]
func (h *MittagHandler) GetRestaurant(ctx echo.Context) error {
	return h.mittag.GetRestaurant(ctx)
}

//	@Summary	Upload a menu
//	@Accept		multipart/form-data
//	@Tags		restaurants
//	@Param		Authorization	header		string						true	"Bearer <Add access token here>"
//	@Param		id				path		string						true	"Restaurant ID"
//	@Param		file			formData	file						true	"Menu File"
//	@Success	200				{object}	services.CleanRestaurant	"ok"
//	@Failure	401				{object}	echo.HTTPError				"Unauthorized"
//	@Failure	404				{object}	echo.HTTPError				"Can not find ID"
//	@Router		/restaurants/{id} [post]
func (h *MittagHandler) UploadMenu(ctx echo.Context) error {
	return h.mittag.UploadMenu(ctx)
}

//	@Summary	Refresh a menu
//	@Tags		restaurants
//	@Param		id	path		string			true	"Restaurant ID"
//	@Success	200	{object}	nil				"ok"
//	@Failure	401	{object}	echo.HTTPError	"Unauthorized"
//	@Failure	404	{object}	echo.HTTPError	"Can not find ID"
//	@Failure	500	{object}	echo.HTTPError	"Internal Server Error"
//	@Router		/restaurants/{id} [put]
func (h *MittagHandler) RefreshRestaurant(ctx echo.Context) error {
	return h.mittag.UpdateRestaurant(ctx)
}
