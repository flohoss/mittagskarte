package sitemap

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	_ "github.com/flohoss/mittagskarte/migrations"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
	"github.com/pocketbase/pocketbase/tools/filesystem"
)

func newTestApp(t *testing.T) *tests.TestApp {
	t.Helper()

	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatalf("failed to create test app: %v", err)
	}
	t.Cleanup(app.Cleanup)
	return app
}

func createRestaurantWithMenu(t *testing.T, app core.App, slug string) {
	t.Helper()

	restaurants, err := app.FindCachedCollectionByNameOrId("restaurants")
	if err != nil {
		t.Fatalf("failed to find restaurants collection: %v", err)
	}
	// the menus.hash value is normally filled by the mittag pipeline,
	// relax the requirement so records can be saved directly here.
	menus, err := app.FindCachedCollectionByNameOrId("menus")
	if err != nil {
		t.Fatalf("failed to find menus collection: %v", err)
	}
	hashField := menus.Fields.GetByName("hash").(*core.TextField)
	hashField.Required = false
	if err := app.Save(menus); err != nil {
		t.Fatalf("failed to relax hash requirement: %v", err)
	}

	record := core.NewRecord(restaurants)
	record.Set("name", "Gasthaus "+slug)
	record.Set("slug", slug)
	record.Set("method", "upload")
	record.Set("latitude", 48.5)
	record.Set("longitude", 9.1)
	if err := app.Save(record); err != nil {
		t.Fatalf("failed to save restaurant: %v", err)
	}

	menuPath := filepath.Join(t.TempDir(), "menu.webp")
	if err := os.WriteFile(menuPath, []byte("menu"), 0o644); err != nil {
		t.Fatalf("failed to write menu fixture: %v", err)
	}
	menuFile, err := filesystem.NewFileFromPath(menuPath)
	if err != nil {
		t.Fatalf("failed to build menu file: %v", err)
	}

	menuRecord := core.NewRecord(menus)
	menuRecord.Set("file", menuFile)
	menuRecord.Set("restaurant", record.Id)
	if err := app.Save(menuRecord); err != nil {
		t.Fatalf("failed to save menu record: %v", err)
	}

	loaded, err := app.FindRecordById("restaurants", record.Id)
	if err != nil {
		t.Fatalf("failed to reload restaurant: %v", err)
	}
	loaded.Set("menus", []string{menuRecord.Id})
	if err := app.Save(loaded); err != nil {
		t.Fatalf("failed to link menu: %v", err)
	}
}

func TestBuild(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)

	set, err := Build(app, "https://example.com")
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if len(set.URLs) != 2 {
		t.Fatalf("expected only static urls for empty db, got %d", len(set.URLs))
	}

	createRestaurantWithMenu(t, app, "gasthaus-a")
	createRestaurantWithMenu(t, app, "gasthaus-b")

	set, err = Build(app, "https://example.com")
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if set.Xmlns != "http://www.sitemaps.org/schemas/sitemap/0.9" {
		t.Fatalf("unexpected xmlns, got %q", set.Xmlns)
	}
	if len(set.URLs) != 4 {
		t.Fatalf("expected 4 urls, got %d", len(set.URLs))
	}
	locations := make([]string, 0, len(set.URLs))
	for _, u := range set.URLs {
		locations = append(locations, u.Loc)
	}
	for _, want := range []string{
		"https://example.com/",
		"https://example.com/datenschutz",
		"https://example.com/restaurants/gasthaus-a",
		"https://example.com/restaurants/gasthaus-b",
	} {
		if !slices.Contains(locations, want) {
			t.Fatalf("expected location %q in %v", want, locations)
		}
	}
}

func TestBuildMarshalsToValidXML(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	createRestaurantWithMenu(t, app, "xml-check")

	set, err := Build(app, "https://example.com")
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	data, err := xml.Marshal(set)
	if err != nil {
		t.Fatalf("failed to marshal urlset: %v", err)
	}
	if !strings.Contains(string(data), "<loc>https://example.com/restaurants/xml-check</loc>") {
		t.Fatalf("unexpected xml output: %s", data)
	}
}

func TestRobots(t *testing.T) {
	t.Parallel()

	want := "User-agent: *\nAllow: /\nSitemap: https://example.com/sitemap.xml\n"
	if got := Robots("https://example.com"); got != want {
		t.Fatalf("unexpected robots output, got %q want %q", got, want)
	}
}
