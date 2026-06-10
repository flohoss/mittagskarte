package placeholder

import (
	"strings"
	"testing"
	"time"
)

func TestReplaceNonDatePlaceholder(t *testing.T) {
	t.Parallel()

	got := Replace("/menus/{{restaurant}}")
	want := "/menus/restaurant"
	if got != want {
		t.Fatalf("unexpected replacement, got %q want %q", got, want)
	}
}

func TestReplaceDateUppercase(t *testing.T) {
	t.Parallel()

	got := Replace("{{date(format=Jan,upper=true)}}")
	if got == "" {
		t.Fatal("expected non-empty output")
	}
	if got != strings.ToUpper(got) {
		t.Fatalf("expected uppercase output, got %q", got)
	}
}

func TestReplaceDateCharsLimit(t *testing.T) {
	t.Parallel()

	full := Replace("{{date(format=Jan,lang=de,upper=true)}}")
	got := Replace("{{date(format=Jan,lang=de,upper=true,chars=3)}}")

	runes := []rune(full)
	limit := 3
	if len(runes) < limit {
		limit = len(runes)
	}
	want := string(runes[:limit])

	if got != want {
		t.Fatalf("unexpected limited output, got %q want %q", got, want)
	}
}

func TestReplaceDateWeekdayAndOffset(t *testing.T) {
	t.Parallel()

	got := Replace("{{date(format=2006-01-02,day=monday,offset=1)}}")

	now := time.Now()
	diff := int(time.Monday - now.Weekday())
	want := now.AddDate(0, 0, diff).AddDate(0, 0, 7).Format("2006-01-02")

	if got != want {
		t.Fatalf("unexpected date output, got %q want %q", got, want)
	}
}

func TestReplaceDateDefaults(t *testing.T) {
	t.Parallel()

	got := Replace("{{date()}}")
	if got == "" {
		t.Fatal("expected non-empty replacement for date default")
	}
	if strings.Contains(got, "{{") || strings.Contains(got, "}}") {
		t.Fatalf("expected placeholder to be resolved, got %q", got)
	}
}

func TestReplaceDateInvalidArgsFallback(t *testing.T) {
	t.Parallel()

	got := Replace("{{date(format=2006-01-02,offset=abc)}}")
	want := time.Now().Format("2006-01-02")
	if got != want {
		t.Fatalf("unexpected date with invalid offset, got %q want %q", got, want)
	}
}

func TestReplaceDateUnknownLocaleFallsBackToEnglish(t *testing.T) {
	t.Parallel()

	got := Replace("{{date(format=Monday,lang=xx)}}")
	if got == "" {
		t.Fatal("expected non-empty output")
	}

	validWeekdays := map[string]struct{}{
		"Monday": {}, "Tuesday": {}, "Wednesday": {}, "Thursday": {}, "Friday": {}, "Saturday": {}, "Sunday": {},
	}
	if _, ok := validWeekdays[got]; !ok {
		t.Fatalf("expected english weekday fallback, got %q", got)
	}
}

func TestReplaceDateInvalidWeekdayIgnored(t *testing.T) {
	t.Parallel()

	got := Replace("{{date(format=2006-01-02,day=notaday)}}")
	want := time.Now().Format("2006-01-02")
	if got != want {
		t.Fatalf("unexpected date with invalid weekday, got %q want %q", got, want)
	}
}

func TestReplaceMultiplePlaceholders(t *testing.T) {
	t.Parallel()

	got := Replace("menu-{{date(format=2006)}}-{{restaurant}}")
	year := time.Now().Format("2006")
	want := "menu-" + year + "-restaurant"
	if got != want {
		t.Fatalf("unexpected replacement, got %q want %q", got, want)
	}
}
