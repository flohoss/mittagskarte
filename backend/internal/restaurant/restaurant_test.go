package restaurant

import (
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLastCheckFromError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		err        error
		wantStatus LastCheckStatus
		wantDetail string
	}{
		{name: "nil", err: nil, wantStatus: LastCheckStatusSuccess, wantDetail: ""},
		{name: "manual upload only", err: ErrManualUploadOnly, wantStatus: LastCheckStatusNotChanged, wantDetail: ""},
		{name: "menu unchanged", err: ErrMenuUnchanged, wantStatus: LastCheckStatusNotChanged, wantDetail: ""},
		{name: "generic error", err: errors.New("boom"), wantStatus: LastCheckStatusError, wantDetail: "boom"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			status, detail := LastCheckFromError(tc.err)
			if status != tc.wantStatus {
				t.Fatalf("unexpected status, got %q want %q", status, tc.wantStatus)
			}
			if detail != tc.wantDetail {
				t.Fatalf("unexpected detail, got %q want %q", detail, tc.wantDetail)
			}
		})
	}
}

func TestIsOnHoliday(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		holidayUntil string
		now          time.Time
		want         bool
	}{
		{name: "empty field", holidayUntil: "", now: time.Date(2026, 8, 14, 12, 0, 0, 0, time.UTC), want: false},
		{name: "before holiday end", holidayUntil: "2026-08-28", now: time.Date(2026, 8, 14, 12, 0, 0, 0, time.UTC), want: true},
		{name: "on holiday end date", holidayUntil: "2026-08-14", now: time.Date(2026, 8, 14, 23, 59, 0, 0, time.UTC), want: true},
		{name: "after holiday end date", holidayUntil: "2026-08-10", now: time.Date(2026, 8, 14, 0, 0, 1, 0, time.UTC), want: false},
		{name: "invalid date format", holidayUntil: "not-a-date", now: time.Date(2026, 8, 14, 12, 0, 0, 0, time.UTC), want: false},
		{name: "holiday end at midnight next day", holidayUntil: "2026-08-14", now: time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC), want: false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := &Restaurant{HolidayUntil: tc.holidayUntil}
			got := r.IsOnHoliday(tc.now)
			if got != tc.want {
				t.Fatalf("unexpected result, got %v want %v", got, tc.want)
			}
		})
	}
}

func TestIsClosedToday(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		restDays []string
		now      time.Time
		want     bool
	}{
		{name: "no rest days", restDays: nil, now: time.Date(2026, 8, 19, 12, 0, 0, 0, time.UTC), want: false},
		{name: "closed on Monday", restDays: []string{"Monday"}, now: time.Date(2026, 8, 17, 12, 0, 0, 0, time.UTC), want: true},
		{name: "open on Tuesday", restDays: []string{"Monday"}, now: time.Date(2026, 8, 18, 12, 0, 0, 0, time.UTC), want: false},
		{name: "closed on Sunday", restDays: []string{"Sunday"}, now: time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC), want: true},
		{name: "multiple rest days includes today", restDays: []string{"Monday", "Tuesday", "Wednesday"}, now: time.Date(2026, 8, 18, 12, 0, 0, 0, time.UTC), want: true},
		{name: "multiple rest days excludes today", restDays: []string{"Monday", "Wednesday"}, now: time.Date(2026, 8, 18, 12, 0, 0, 0, time.UTC), want: false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := &Restaurant{RestDays: tc.restDays}
			got := r.IsClosedToday(tc.now)
			if got != tc.want {
				t.Fatalf("unexpected result, got %v want %v", got, tc.want)
			}
		})
	}
}

func TestDownloadRequiresLocator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		navigate []Selector
	}{
		{name: "no selectors", navigate: nil},
		{name: "empty locator", navigate: []Selector{{Locator: ""}}},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := &Restaurant{Name: "Gasthaus", Navigate: tc.navigate}
			_, err := r.Download(filepath.Join(t.TempDir(), "menu.pdf"), slog.New(slog.DiscardHandler))
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), "no URL defined in first locator") {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestDownloadRejectsInvalidURL(t *testing.T) {
	t.Parallel()

	r := &Restaurant{Name: "Gasthaus", Navigate: []Selector{{Locator: "://bad-url"}}}
	_, err := r.Download(filepath.Join(t.TempDir(), "menu.pdf"), slog.New(slog.DiscardHandler))
	if err == nil {
		t.Fatal("expected error for invalid URL, got nil")
	}
	if !strings.Contains(err.Error(), "invalid URL in first locator") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDownloadWritesServerContent(t *testing.T) {
	requireCurl(t)
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("downloaded menu"))
	}))
	defer server.Close()

	downloadPath := filepath.Join(t.TempDir(), "menu.pdf")
	r := &Restaurant{Name: "Gasthaus", Navigate: []Selector{{Locator: server.URL + "/menu.pdf"}}}

	gotPath, err := r.Download(downloadPath, slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatalf("Download returned error: %v", err)
	}
	if gotPath != downloadPath {
		t.Fatalf("unexpected path, got %q want %q", gotPath, downloadPath)
	}

	data, err := os.ReadFile(gotPath)
	if err != nil {
		t.Fatalf("failed to read downloaded file: %v", err)
	}
	if string(data) != "downloaded menu" {
		t.Fatalf("unexpected file content: %q", string(data))
	}
}

func requireCurl(t *testing.T) {
	t.Helper()

	if _, err := exec.LookPath("curl"); err != nil {
		t.Skip("curl binary not available")
	}
}
