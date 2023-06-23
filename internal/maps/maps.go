package maps

import (
	"fmt"
	"image/color"
	"os"

	sm "github.com/flopp/go-staticmaps"
	"github.com/fogleman/gg"
	"github.com/golang/geo/s2"
	"gitlab.unjx.de/flohoss/mittag/internal/convert"
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
		fmt.Println(err)
	}

	if err := gg.SavePNG(folder+"/map.png", img); err != nil {
		fmt.Println(err)
	}
	convert.CreateWebp(folder + "/map.png")
	os.Remove(folder + "/map.png")
}
