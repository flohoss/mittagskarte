package placeholder

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/goodsign/monday"
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

func TestReplaceDateWithoutParentheses(t *testing.T) {
	t.Parallel()

	got := Replace("{{date}}")
	if got != "date" {
		t.Fatalf("expected raw key fallback, got %q want %q", got, "date")
	}
}

func TestReplaceDateMalformedParentheses(t *testing.T) {
	t.Parallel()

	got := Replace("{{somedate}}")
	if got != "somedate" {
		t.Fatalf("expected raw key fallback, got %q want %q", got, "somedate")
	}
}

func TestReplaceDateCharsZeroReturnsEmpty(t *testing.T) {
	t.Parallel()

	got := Replace("{{date(format=Jan,chars=0)}}")
	if got != "" {
		t.Fatalf("expected empty output for chars=0, got %q", got)
	}
}

func TestReplaceDateCharsNegativeReturnsEmpty(t *testing.T) {
	t.Parallel()

	got := Replace("{{date(format=Jan,chars=-3)}}")
	if got != "" {
		t.Fatalf("expected empty output for negative chars, got %q", got)
	}
}

func TestReplaceDateCharsInvalidIgnored(t *testing.T) {
	t.Parallel()

	got := Replace("{{date(format=Jan,chars=abc)}}")
	if got == "" {
		t.Fatal("expected non-empty output when chars is non-numeric")
	}
}

func TestReplaceDateExplicitLangEn(t *testing.T) {
	t.Parallel()

	got := Replace("{{date(format=Monday,lang=en)}}")
	validWeekdays := map[string]struct{}{
		"Monday": {}, "Tuesday": {}, "Wednesday": {}, "Thursday": {}, "Friday": {}, "Saturday": {}, "Sunday": {},
	}
	if _, ok := validWeekdays[got]; !ok {
		t.Fatalf("expected English weekday, got %q", got)
	}
}

func TestReplaceDateExplicitLangDe(t *testing.T) {
	t.Parallel()

	got := Replace("{{date(format=Monday,lang=de)}}")
	germanWeekdays := map[string]struct{}{
		"Montag": {}, "Dienstag": {}, "Mittwoch": {}, "Donnerstag": {}, "Freitag": {}, "Samstag": {}, "Sonntag": {},
	}
	if _, ok := germanWeekdays[got]; !ok {
		t.Fatalf("expected German weekday, got %q", got)
	}
}

func TestReplaceDateFullDefaultFormat(t *testing.T) {
	t.Parallel()

	got := Replace("{{date()}}")
	want := monday.Format(time.Now(), "Monday, 02 January 2006", monday.LocaleEnUS)
	if got != want {
		t.Fatalf("expected default formatted date, got %q want %q", got, want)
	}
}

func TestReplaceNoPlaceholders(t *testing.T) {
	t.Parallel()

	got := Replace("/menus/restaurant/file.pdf")
	if got != "/menus/restaurant/file.pdf" {
		t.Fatalf("expected unchanged string, got %q", got)
	}
}

func TestReplaceEmptyString(t *testing.T) {
	t.Parallel()

	got := Replace("")
	if got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}

func TestReplaceDateOffsetZero(t *testing.T) {
	t.Parallel()

	got := Replace("{{date(format=2006-01-02,offset=0)}}")
	want := time.Now().Format("2006-01-02")
	if got != want {
		t.Fatalf("expected today's date with offset=0, got %q want %q", got, want)
	}
}

func TestReplaceDateNegativeOffset(t *testing.T) {
	t.Parallel()

	got := Replace("{{date(format=2006-01-02,offset=-2)}}")
	now := time.Now()
	want := now.AddDate(0, 0, -14).Format("2006-01-02")
	if got != want {
		t.Fatalf("expected date with negative offset, got %q want %q", got, want)
	}
}

func TestReplaceDateUppercaseFalseIgnored(t *testing.T) {
	t.Parallel()

	withUpper := Replace("{{date(format=Jan,upper=false)}}")
	without := Replace("{{date(format=Jan)}}")
	if withUpper != without {
		t.Fatalf("expected upper=false to match no-upper output, got %q want %q", withUpper, without)
	}
}

func TestReplaceDateAllWeekdays(t *testing.T) {
	t.Parallel()

	for _, wd := range []struct {
		name string
		day  time.Weekday
	}{
		{"monday", time.Monday},
		{"tuesday", time.Tuesday},
		{"wednesday", time.Wednesday},
		{"thursday", time.Thursday},
		{"friday", time.Friday},
		{"saturday", time.Saturday},
		{"sunday", time.Sunday},
	} {
		wd := wd
		t.Run(wd.name, func(t *testing.T) {
			t.Parallel()

			got := Replace("{{date(format=2006-01-02,day=" + wd.name + ")}}")
			now := time.Now()
			diff := int(wd.day - now.Weekday())
			want := now.AddDate(0, 0, diff).Format("2006-01-02")
			if got != want {
				t.Fatalf("unexpected date for %s, got %q want %q", wd.name, got, want)
			}
		})
	}
}

func TestReplaceDateWeekdayUppercase(t *testing.T) {
	t.Parallel()

	got := Replace("{{date(format=2006-01-02,day=MONDAY)}}")
	now := time.Now()
	diff := int(time.Monday - now.Weekday())
	want := now.AddDate(0, 0, diff).Format("2006-01-02")
	if got != want {
		t.Fatalf("unexpected date for uppercase weekday, got %q want %q", got, want)
	}
}

func TestReplaceDateCharsLargerThanResult(t *testing.T) {
	t.Parallel()

	got := Replace("{{date(format=Jan,chars=100)}}")
	full := Replace("{{date(format=Jan)}}")
	if got != full {
		t.Fatalf("expected full output when chars exceeds length, got %q want %q", got, full)
	}
}

func TestReplaceDateArgWithoutValue(t *testing.T) {
	t.Parallel()

	got := Replace("{{date(format=Jan,nolang)}}")
	if got == "" {
		t.Fatal("expected non-empty output when an arg has no value")
	}
}

func TestReplaceDateCharsExactLength(t *testing.T) {
	t.Parallel()

	full := Replace("{{date(format=Jan,upper=true)}}")
	runes := []rune(full)
	limit := len(runes)
	got := Replace("{{date(format=Jan,upper=true,chars=" + strconv.Itoa(limit) + ")}}")
	if got != full {
		t.Fatalf("expected full output when chars equals length, got %q want %q", got, full)
	}
}
