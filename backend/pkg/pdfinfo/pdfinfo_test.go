package pdfinfo

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const (
	plainPDFFixture = "testdata/plain.pdf"
	imagePDFFixture = "testdata/withimage.pdf"
)

func imageDpiForTest(widthPx int, pageWidthPt float64) int {
	return int(float64(widthPx)/(pageWidthPt/72.0) + 0.5)
}

func TestDPIMetadata(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		meta    Metadata
		wantDpi int
	}{
		{
			name:    "zero metadata falls back to min dpi",
			meta:    Metadata{},
			wantDpi: MinDpi,
		},
		{
			name:    "image dpi wins when higher than target dpi",
			meta:    Metadata{MaxImageWidthPx: 4000, PageWidthPt: 595},
			wantDpi: imageDpiForTest(4000, 595),
		},
		{
			name:    "image dpi below target dpi uses target dpi",
			meta:    Metadata{MaxImageWidthPx: 100, PageWidthPt: 595},
			wantDpi: dpiForTargetWidth(595, TargetWidth, MinDpi, MaxDpi),
		},
		{
			name:    "image dpi clamped to max",
			meta:    Metadata{MaxImageWidthPx: 120000, PageWidthPt: 100},
			wantDpi: MaxDpi,
		},
		{
			name:    "page width missing ignores image dpi",
			meta:    Metadata{MaxImageWidthPx: 4000},
			wantDpi: dpiForTargetWidth(0, TargetWidth, MinDpi, MaxDpi),
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := tc.meta.DPI(); got != tc.wantDpi {
				t.Fatalf("unexpected dpi, got %d want %d", got, tc.wantDpi)
			}
		})
	}
}

func TestDpiForTargetWidthClamps(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		widthPt float64
		want    int
	}{
		{name: "zero width", widthPt: 0, want: MinDpi},
		{name: "below min", widthPt: 10000, want: MinDpi},
		{name: "at target", widthPt: float64(TargetWidth), want: 1},
		{name: "above max", widthPt: 10, want: MaxDpi},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := dpiForTargetWidth(tc.widthPt, TargetWidth, MinDpi, MaxDpi)
			if got > MaxDpi || got < MinDpi {
				t.Fatalf("dpi %d outside bounds [%d, %d]", got, MinDpi, MaxDpi)
			}
			if tc.widthPt == 0 && got != MinDpi {
				t.Fatalf("unexpected dpi for zero width, got %d", got)
			}
		})
	}
}

func TestClampDpi(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		dpi  int
		want int
	}{
		{name: "below min", dpi: 10, want: MinDpi},
		{name: "at min", dpi: MinDpi, want: MinDpi},
		{name: "in range", dpi: 300, want: 300},
		{name: "at max", dpi: MaxDpi, want: MaxDpi},
		{name: "above max", dpi: 5000, want: MaxDpi},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := clampDpi(tc.dpi); got != tc.want {
				t.Fatalf("unexpected clamp result, got %d want %d", got, tc.want)
			}
		})
	}
}

func TestParseOutput(t *testing.T) {
	t.Parallel()

	output := "Title:           Wochenkarte\n" +
		"Author:          Florian\n" +
		"Creator:         mkpdf\n" +
		"Producer:        mkpdf\n" +
		"CreationDate:    D:20260101000000\n" +
		"ModDate:         D:20260102000000\n" +
		"Pages:           3\n" +
		"Page size:       595 x 842 pts (A4)\n" +
		"Encrypted:       no\n" +
		"\n" +
		": empty value line ignored\n"

	meta := parse(output)

	if meta.Title != "Wochenkarte" {
		t.Fatalf("unexpected title, got %q", meta.Title)
	}
	if meta.Author != "Florian" {
		t.Fatalf("unexpected author, got %q", meta.Author)
	}
	if meta.PageCount != 3 {
		t.Fatalf("unexpected page count, got %d", meta.PageCount)
	}
	if meta.PageWidthPt != 595 || meta.PageHeightPt != 842 {
		t.Fatalf("unexpected page size, got %v/%v", meta.PageWidthPt, meta.PageHeightPt)
	}
	if meta.CreationDate != "D:20260101000000" || meta.ModDate != "D:20260102000000" {
		t.Fatalf("unexpected dates, got %q/%q", meta.CreationDate, meta.ModDate)
	}
}

func TestParseIgnoresUnknownKeys(t *testing.T) {
	t.Parallel()

	meta := parse("File size: 488 bytes\nOptimized: no\n")
	if meta != (Metadata{}) {
		t.Fatalf("expected empty metadata, got %+v", meta)
	}
}

func TestParsePageSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		val               string
		wantWidth, wantHt float64
	}{
		{name: "landscape pts", val: "842 x 595 pts", wantWidth: 842, wantHt: 595},
		{name: "missing separator", val: "842 595 pts"},
		{name: "too few fields", val: "595 pts"},
		{name: "not numeric", val: "abc x def pts"},
		{name: "empty", val: ""},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			w, h := parsePageSize(tc.val)
			if w != tc.wantWidth || h != tc.wantHt {
				t.Fatalf("unexpected page size, got %v/%v want %v/%v", w, h, tc.wantWidth, tc.wantHt)
			}
		})
	}
}

func TestEqual(t *testing.T) {
	t.Parallel()

	meta := Metadata{Title: "a", PageCount: 1}

	if Equal(nil, meta) {
		t.Fatal("nil stored metadata must not equal current")
	}
	if !Equal(meta, meta) {
		t.Fatal("identical metadata must be equal")
	}
	if Equal(meta, Metadata{Title: "a", PageCount: 2}) {
		t.Fatal("different metadata must not be equal")
	}
}

func TestIsPDF(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	pdfPath := filepath.Join(dir, "menu.pdf")
	if err := os.WriteFile(pdfPath, []byte("%PDF-1.4\ncontent"), 0o644); err != nil {
		t.Fatalf("failed to write pdf fixture: %v", err)
	}

	textPath := filepath.Join(dir, "menu.txt")
	if err := os.WriteFile(textPath, []byte("not a pdf"), 0o644); err != nil {
		t.Fatalf("failed to write text fixture: %v", err)
	}

	tests := []struct {
		name       string
		sourcePath string
		want       bool
	}{
		{name: "pdf extension", sourcePath: "menu.PDF", want: true},
		{name: "pdf content sniffed", sourcePath: pdfPath, want: true},
		{name: "extension wins over content", sourcePath: textPath + ".pdf", want: true},
		{name: "text file", sourcePath: textPath, want: false},
		{name: "missing file", sourcePath: filepath.Join(dir, "missing.txt"), want: false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := IsPDF(tc.sourcePath); got != tc.want {
				t.Fatalf("unexpected IsPDF result for %q, got %v want %v", tc.sourcePath, got, tc.want)
			}
		})
	}

	if _, err := os.Stat(textPath + ".pdf"); !os.IsNotExist(err) {
		t.Fatalf("IsPDF must not create files, %s exists", textPath+".pdf")
	}
}

func TestRead(t *testing.T) {
	requirePoppler(t)

	t.Run("plain pdf", func(t *testing.T) {
		t.Parallel()

		meta, err := Read(plainPDFFixture)
		if err != nil {
			t.Fatalf("Read returned error: %v", err)
		}
		if meta.PageCount != 1 {
			t.Fatalf("unexpected page count, got %d", meta.PageCount)
		}
		if meta.PageWidthPt != 595 || meta.PageHeightPt != 842 {
			t.Fatalf("unexpected page size, got %v/%v", meta.PageWidthPt, meta.PageHeightPt)
		}
		if meta.Title != "Plain Menu" {
			t.Fatalf("unexpected title, got %q", meta.Title)
		}
		if meta.MaxImageWidthPx != 0 {
			t.Fatalf("expected no image width for plain pdf, got %d", meta.MaxImageWidthPx)
		}
	})

	t.Run("pdf with image", func(t *testing.T) {
		t.Parallel()

		meta, err := Read(imagePDFFixture)
		if err != nil {
			t.Fatalf("Read returned error: %v", err)
		}
		if meta.MaxImageWidthPx != 4 {
			t.Fatalf("unexpected max image width, got %d", meta.MaxImageWidthPx)
		}
	})

	t.Run("not a pdf", func(t *testing.T) {
		t.Parallel()

		notPDF := filepath.Join(t.TempDir(), "not.pdf")
		if err := os.WriteFile(notPDF, []byte("junk"), 0o644); err != nil {
			t.Fatalf("failed to write fixture: %v", err)
		}

		_, err := Read(notPDF)
		if err == nil {
			t.Fatal("expected error for non-pdf file")
		}
		if !strings.Contains(err.Error(), "pdfinfo failed") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func requirePoppler(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("pdfinfo"); err != nil {
		t.Skip("pdfinfo binary not available")
	}
	if _, err := exec.LookPath("pdfimages"); err != nil {
		t.Skip("pdfimages binary not available")
	}
}
