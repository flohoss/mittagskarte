package curl

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDownload(t *testing.T) {
	requireCurl(t)

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("today menu"))
		}))
		defer server.Close()

		outPath := filepath.Join(t.TempDir(), "menu.txt")
		gotPath, err := Download(outPath, server.URL+"/menu")
		if err != nil {
			t.Fatalf("Download returned error: %v", err)
		}
		if gotPath != outPath {
			t.Fatalf("unexpected path, got %q want %q", gotPath, outPath)
		}

		data, err := os.ReadFile(outPath)
		if err != nil {
			t.Fatalf("failed to read downloaded file: %v", err)
		}
		if string(data) != "today menu" {
			t.Fatalf("unexpected file content: %q", string(data))
		}
	})

	t.Run("redirect", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/redirect" {
				http.Redirect(w, r, "/final", http.StatusFound)
				return
			}
			_, _ = w.Write([]byte("redirected menu"))
		}))
		defer server.Close()

		outPath := filepath.Join(t.TempDir(), "menu.txt")
		if _, err := Download(outPath, server.URL+"/redirect"); err != nil {
			t.Fatalf("Download returned error for redirect: %v", err)
		}

		data, err := os.ReadFile(outPath)
		if err != nil {
			t.Fatalf("failed to read downloaded file: %v", err)
		}
		if string(data) != "redirected menu" {
			t.Fatalf("unexpected redirected file content: %q", string(data))
		}
	})

	t.Run("invalid url", func(t *testing.T) {
		t.Parallel()

		outPath := filepath.Join(t.TempDir(), "menu.txt")
		_, err := Download(outPath, "://bad-url")
		if err == nil {
			t.Fatal("expected error for invalid URL, got nil")
		}
		if !strings.Contains(err.Error(), "curl failed") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func requireCurl(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("curl"); err != nil {
		t.Skip("curl binary not available")
	}
}
