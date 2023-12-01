package maps

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/fogleman/gg"
	"gitlab.unjx.de/flohoss/mittag/internal/convert"
	"googlemaps.github.io/maps"
)

const MapsFolder = "storage/public/maps"
const originMarkerRange = 5000

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

func (m *MapInformation) getPaths() []maps.Path {
	if m.getLeg().Distance.Meters > originMarkerRange {
		return []maps.Path{}
	}
	latLng, err := m.Route.OverviewPolyline.Decode()
	if err != nil {
		return []maps.Path{}
	}
	paths := []maps.Path{{Location: latLng, Color: "0xEB932D"}}
	return paths
}

func (m *MapInformation) getMarkers() []maps.Marker {
	markers := []maps.Marker{
		{LocationAddress: m.getLeg().EndAddress, Size: string(maps.Small), Color: "0xEB932D"},
	}
	if m.getLeg().Distance.Meters < originMarkerRange {
		markers = append(markers, maps.Marker{LocationAddress: m.getLeg().StartAddress, Size: string(maps.Tiny), Color: "0x14468C"})
	}
	return markers
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
		Size:      "250x1000",
		Scale:     4,
		MapType:   maps.Hybrid,
		Markers:   m.getMarkers(),
		MapStyles: []string{"feature:poi|visibility:off"},
		Paths:     m.getPaths(),
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
