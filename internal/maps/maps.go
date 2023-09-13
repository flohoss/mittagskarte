package maps

import (
	"image/color"
	"log/slog"

	sm "github.com/flopp/go-staticmaps"
	"github.com/fogleman/gg"
	"github.com/golang/geo/s2"
)

func CreateMap(lat float64, lng float64, folder string, brainority bool) {
	ctx := sm.NewContext()

	ctx.SetSize(896, 150)
	ctx.AddObject(
		sm.NewMarker(
			s2.LatLngFromDegrees(lat, lng),
			color.RGBA{235, 147, 45, 255},
			16.0,
		),
	)
	if brainority {
		ctx.AddObject(
			sm.NewMarker(
				s2.LatLngFromDegrees(48.70861927536174, 9.168349225546137),
				color.RGBA{20, 69, 138, 255},
				16.0,
			),
		)
	}

	img, err := ctx.Render()
	if err != nil {
		slog.Error("cannot render map", "err", err)
	}

	old := folder + "/map.jpg"
	if err := gg.SaveJPG(old, img, 70); err != nil {
		slog.Error("cannot save map", "err", err)
	}
}
