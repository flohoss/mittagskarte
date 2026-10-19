package mittag

import (
	"testing"

	_ "github.com/flohoss/mittagskarte/migrations"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
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
