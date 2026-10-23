package api_test

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flohoss/mittagskarte/api"
	_ "github.com/flohoss/mittagskarte/migrations"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
)

const (
	testImprintEmail = "contact@example.com"
	indexTemplate    = `<!doctype html>
<html><body>
<script id="initial-restaurants" type="application/json">{{ toJSON .Restaurants }}</script>
<script id="initial-email" type="application/json">{{ toJSON .Email }}</script>
</body></html>`
)

func writeFrontendDist(t *testing.T) string {
	t.Helper()

	parent := t.TempDir()
	dist := filepath.Join(parent, "dist")
	if err := os.MkdirAll(filepath.Join(dist, "static"), 0o755); err != nil {
		t.Fatalf("failed to create static dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dist, "index.html"), []byte(indexTemplate), 0o644); err != nil {
		t.Fatalf("failed to write index template: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dist, "static", "app.js"), []byte("console.log('menu')"), 0o644); err != nil {
		t.Fatalf("failed to write static asset: %v", err)
	}

	return parent
}

func newFrontendApp(t *testing.T, dir string, dev bool) (*tests.TestApp, http.Handler) {
	t.Helper()

	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatalf("failed to create test app: %v", err)
	}
	t.Cleanup(app.Cleanup)

	t.Chdir(dir)

	pbRouter, err := apis.NewRouter(app)
	if err != nil {
		t.Fatalf("failed to create router: %v", err)
	}

	serveEvent := new(core.ServeEvent)
	serveEvent.App = app
	serveEvent.Router = pbRouter

	if err := app.OnServe().Trigger(serveEvent, func(e *core.ServeEvent) error {
		return api.ServeFrontend(app, e, testImprintEmail, dev)
	}); err != nil {
		t.Fatalf("failed to trigger serve event: %v", err)
	}

	mux, err := pbRouter.BuildMux()
	if err != nil {
		t.Fatalf("failed to build mux: %v", err)
	}

	return app, mux
}

func get(t *testing.T, mux http.Handler, target string) *httptest.ResponseRecorder {
	t.Helper()

	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, target, nil))
	return recorder
}

func TestServeFrontendWithoutDistDir(t *testing.T) {
	dir := t.TempDir()

	_, mux := newFrontendApp(t, dir, true)

	if got := get(t, mux, "/").Code; got != http.StatusNotFound {
		t.Fatalf("unexpected status without dist dir, got %d want %d", got, http.StatusNotFound)
	}
}

func TestServeFrontendInjectsPageData(t *testing.T) {
	dir := writeFrontendDist(t)
	app, mux := newFrontendApp(t, dir, true)

	relaxMenuHash(t, app)
	createRestaurantWithMenu(t, app, "gasthaus-a")

	recorder := get(t, mux, "/")
	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status, got %d want %d", recorder.Code, http.StatusOK)
	}
	if got := recorder.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/html") {
		t.Fatalf("unexpected Content-Type, got %q", got)
	}

	body := recorder.Body.String()
	encodedEmail := base64.StdEncoding.EncodeToString([]byte(testImprintEmail))
	for _, want := range []string{encodedEmail, `"slug":"gasthaus-a"`, "Gasthaus gasthaus-a"} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected %q in rendered page, got:\n%s", want, body)
		}
	}
}

func TestServeFrontendServesStaticAsset(t *testing.T) {
	dir := writeFrontendDist(t)
	_, mux := newFrontendApp(t, dir, true)

	recorder := get(t, mux, "/static/app.js")
	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status, got %d want %d", recorder.Code, http.StatusOK)
	}
	if got := recorder.Body.String(); got != "console.log('menu')" {
		t.Fatalf("unexpected static body, got %q", got)
	}
}

func TestServeFrontendMissingAssetReturnsNotFound(t *testing.T) {
	dir := writeFrontendDist(t)
	_, mux := newFrontendApp(t, dir, true)

	if got := get(t, mux, "/static/missing.png").Code; got != http.StatusNotFound {
		t.Fatalf("unexpected status for missing asset, got %d want %d", got, http.StatusNotFound)
	}
}

func TestServeFrontendFallsBackToIndexForSpaRoutes(t *testing.T) {
	dir := writeFrontendDist(t)
	_, mux := newFrontendApp(t, dir, true)

	recorder := get(t, mux, "/restaurants/gasthaus-a")
	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status, got %d want %d", recorder.Code, http.StatusOK)
	}
	if !strings.Contains(recorder.Body.String(), "initial-restaurants") {
		t.Fatalf("expected index fallback, got:\n%s", recorder.Body.String())
	}
}

func TestServeFrontendRejectsPathTraversal(t *testing.T) {
	dir := t.TempDir()
	frontendDir := filepath.Join(dir, "dist")
	if err := os.MkdirAll(frontendDir, 0o755); err != nil {
		t.Fatalf("failed to create dist dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(frontendDir, "index.html"), []byte(indexTemplate), 0o644); err != nil {
		t.Fatalf("failed to write index template: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "secret.txt"), []byte("top secret"), 0o644); err != nil {
		t.Fatalf("failed to write secret file: %v", err)
	}

	_, mux := newFrontendApp(t, dir, true)

	recorder := get(t, mux, "/%2e%2e/secret.txt")
	if strings.Contains(recorder.Body.String(), "top secret") {
		t.Fatalf("path traversal leaked file content: %s", recorder.Body.String())
	}
	if recorder.Code != http.StatusBadRequest && recorder.Code != http.StatusNotFound {
		t.Fatalf("unexpected status for traversal attempt, got %d", recorder.Code)
	}
}
