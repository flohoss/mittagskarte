package sitemap

import (
	"encoding/xml"
	"fmt"

	"github.com/flohoss/mittagskarte/internal/restaurant"
	"github.com/pocketbase/pocketbase/core"
)

type URLSet struct {
	XMLName xml.Name `xml:"urlset"`
	Xmlns   string   `xml:"xmlns,attr"`
	URLs    []URL    `xml:"url"`
}

type URL struct {
	Loc string `xml:"loc"`
}

func Build(app core.App, baseURL string) (URLSet, error) {
	restaurants, err := restaurant.GetRestaurantsWithMenus(app)
	if err != nil {
		return URLSet{}, err
	}

	urls := make([]URL, 0, len(restaurants)+2)
	urls = append(urls, URL{Loc: baseURL + "/"})
	urls = append(urls, URL{Loc: baseURL + "/datenschutz"})
	for _, r := range restaurants {
		urls = append(urls, URL{Loc: fmt.Sprintf("%s/restaurants/%s", baseURL, r.Slug)})
	}

	return URLSet{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	}, nil
}

func Robots(baseURL string) string {
	return "User-agent: *\nAllow: /\nSitemap: " + baseURL + "/sitemap.xml\n"
}
