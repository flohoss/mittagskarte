package maps

import (
	"image/color"

	sm "github.com/flopp/go-staticmaps"
	"github.com/fogleman/gg"
	"github.com/golang/geo/s2"
	"gitlab.unjx.de/flohoss/mittag/internal/convert"
	"go.uber.org/zap"
)

func CreateMap(lat float64, lng float64, folder string) {
	ctx := sm.NewContext()

	ctx.SetSize(896, 150)
	ctx.AddObject(
		sm.NewMarker(
			s2.LatLngFromDegrees(lat, lng),
			color.RGBA{255, 0, 0, 255},
			16.0,
		),
	)

	img, err := ctx.Render()
	if err != nil {
		zap.L().Error(err.Error())
	}

	old := folder + "/map.png"
	if err := gg.SavePNG(old, img); err != nil {
		zap.L().Error(err.Error())
	}
	_, err = convert.CreateWebp(old)
	if err != nil {
		zap.L().Error(err.Error())
	}
}
