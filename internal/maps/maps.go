package maps

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/fogleman/gg"
	"gitlab.unjx.de/flohoss/mittag/internal/convert"
	"googlemaps.github.io/maps"
)

const MapsFolder = "storage/public/maps"

func init() {
	os.MkdirAll(MapsFolder, os.ModePerm)
}

type MapInformation struct {
	Identifier string
	Route      *maps.Route
}

type MapRequest struct {
	Identifier string
	Address    string
}

func GetMapInformation(key string, mapRequests []MapRequest) map[string]*MapInformation {
	c, err := maps.NewClient(maps.WithAPIKey(key))
	if err != nil {
		slog.Error(err.Error())
	}
	result := make(map[string]*MapInformation)
	for _, request := range mapRequests {
		result[request.Identifier] = &MapInformation{
			Identifier: request.Identifier,
			Route:      getRoute(c, request.Address),
		}
	}
	for key := range result {
		result[key].createMap(c)
	}
	return result
}

func (m *MapInformation) getLeg() *maps.Leg {
	if m.Route != nil && len(m.Route.Legs) > 0 {
		return m.Route.Legs[0]
	}
	return &maps.Leg{}
}

func (m *MapInformation) getMarkers() []maps.Marker {
	return []maps.Marker{
		{LocationAddress: m.getLeg().EndAddress, Label: strings.ToUpper(string(m.Identifier[0])), Size: string(maps.Mid), Color: "0xEB932D"},
	}
}

func getRoute(c *maps.Client, address string) *maps.Route {
	routes, _, err := c.Directions(context.Background(), &maps.DirectionsRequest{
		Origin:      "Brainority Software GmbH, Vor dem Lauch 15, 70567 Stuttgart",
		Destination: address,
		Language:    "de",
		Units:       maps.UnitsMetric,
	})
	if err != nil || len(routes) == 0 {
		return &maps.Route{}
	}
	return &routes[0]
}

func (m *MapInformation) createMap(c *maps.Client) {
	r := &maps.StaticMapRequest{
		Size:      "640x160",
		Zoom:      14,
		Scale:     2,
		MapType:   maps.Hybrid,
		Markers:   m.getMarkers(),
		MapStyles: []string{"feature:all|element:labels|visibility:off"},
	}
	img, err := c.StaticMap(context.Background(), r)
	if err != nil {
		slog.Error(err.Error())
		return
	}

	loc := fmt.Sprintf("%s/%s.png", MapsFolder, m.Identifier)
	if err := gg.SavePNG(loc, img); err != nil {
		slog.Error(err.Error())
	}
	convert.ConvertToWebp(loc, m.Identifier, true)
}
