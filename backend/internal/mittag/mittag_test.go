package mittag

import (
	"testing"

	"github.com/flohoss/mittagskarte/pkg/pdfinfo"

	"ganigeorgiev.com/pocketbase/pocketbase/core"
	"ganigeorgiev.com/pocketbase/pocketbase/tests"

	_ "github.com/flohoss/mittagskarte/migrations"
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

func TestMenuRetentionLimit(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	m := &Mittag{app: app}

	limit, err := m.menuRetentionLimit()
	if err != nil {
		t.Fatalf("menuRetentionLimit returned error: %v", err)
	}
	if limit != 10 {
		t.Fatalf("expected retention limit 10 from migrations (MaxSelect), got %d", limit)
	}
}

func TestMenuRetentionLimitSingleSelect(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)

	collection, err := app.FindCachedCollectionByNameOrId("restaurants")
	if err != nil {
		t.Fatalf("failed to find restaurants collection: %v", err)
	}
	menusField := collection.Fields.GetByName("menus").(*core.RelationField)
	originalMaxSelect := menusField.MaxSelect
	menusField.MaxSelect = 1
	if err := app.Save(collection); err != nil {
		t.Fatalf("failed to save collection: %v", err)
	}

	m := &Mittag{app: app}
	limit, err := m.menuRetentionLimit()
	if err != nil {
		t.Fatalf("menuRetentionLimit returned error: %v", err)
	}
	if limit != 1 {
		t.Fatalf("expected retention limit 1 for single-select relation, got %d", limit)
	}

	menusField.MaxSelect = originalMaxSelect
	if err := app.Save(collection); err != nil {
		t.Fatalf("failed to restore collection: %v", err)
	}
}

func TestMenuRetentionLimitNonRelation(t *testing.T) {
	t.Parallel()

	app := newTestApp(t)
	m := &Mittag{app: app}

	collection, err := app.FindCachedCollectionByNameOrId("restaurants")
	if err != nil {
		t.Fatalf("failed to find restaurants collection: %v", err)
	}
	collection.Fields.RemoveByName("menus")
	collection.Fields.Add(&core.TextField{Name: "menus"})
	if err := app.SaveNoValidate(collection); err != nil {
		t.Fatalf("failed to replace field: %v", err)
	}

	if _, err := m.menuRetentionLimit(); err == nil {
		t.Fatal("expected error when menus is not a relation field, got nil")
	}
}

func TestPdfMetadataIdentifiesContent(t *testing.T) {
	t.Parallel()

	untimestamped := pdfinfo.Metadata{
		Title:        "Da Mamma - Mittagstisch",
		Author:       "Da Mamma eGbR",
		Producer:     "Da Mamma Kartenredaktion",
		PageCount:    1,
		PageWidthPt:  595,
		PageHeightPt: 842,
	}

	tests := []struct {
		name   string
		stored any
		cur    pdfinfo.Metadata
		want   bool
	}{
		{
			name:   "identical metadata without timestamps is not proof of unchanged content",
			stored: untimestamped,
			cur:    untimestamped,
			want:   false,
		},
		{
			name:   "identical metadata with timestamps identifies content",
			stored: pdfinfo.Metadata{Title: "Wochenkarte", ModDate: "D:20260101"},
			cur:    pdfinfo.Metadata{Title: "Wochenkarte", ModDate: "D:20260101"},
			want:   true,
		},
		{
			name:   "timestamped metadata that differs does not identify content",
			stored: pdfinfo.Metadata{Title: "Wochenkarte", ModDate: "D:20260101"},
			cur:    pdfinfo.Metadata{Title: "Wochenkarte", ModDate: "D:20260102"},
			want:   false,
		},
		{
			name:   "missing stored metadata does not identify content",
			stored: nil,
			cur:    pdfinfo.Metadata{Title: "Wochenkarte", ModDate: "D:20260101"},
			want:   false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := pdfMetadataIdentifiesContent(tc.stored, tc.cur); got != tc.want {
				t.Fatalf("unexpected result, got %v want %v", got, tc.want)
			}
		})
	}
}
