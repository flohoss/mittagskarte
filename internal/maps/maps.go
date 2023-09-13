package maps

import (
	"context"
	"log/slog"

	"github.com/fogleman/gg"
	"gitlab.unjx.de/flohoss/mittag/internal/convert"
	"googlemaps.github.io/maps"
)

func CreateMap(address string, folder string, key string, brainority bool) {
	c, err := maps.NewClient(maps.WithAPIKey(key))
	if err != nil {
		slog.Error("cannot create map client", "err", err)
		return
	}

	markers := []maps.Marker{{LocationAddress: address, Size: "small", Color: "0xEB932D"}}

	if brainority {
		markers = append(markers, maps.Marker{LocationAddress: "Brainority Software GmbH", Size: "small", Color: "0x14458A"})
	}

	r := &maps.StaticMapRequest{
		Size:      "900x150",
		Scale:     4,
		MapType:   "hybrid",
		Markers:   markers,
		MapStyles: []string{"feature:poi|visibility:off"},
	}
	img, err := c.StaticMap(context.Background(), r)
	if err != nil {
		slog.Error("cannot render map", "err", err)
		return
	}

	old := folder + "/map.png"
	if err := gg.SavePNG(old, img); err != nil {
		slog.Error("cannot save map", "err", err)
	}
	convert.ConvertToWebp(old, "map")
}
