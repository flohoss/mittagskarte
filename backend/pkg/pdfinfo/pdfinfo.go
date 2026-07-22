package pdfinfo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	DefaultDpi  = 300
	MinDpi      = 150
	MaxDpi      = 1200
	TargetWidth = 4000
)

type Metadata struct {
	Title           string  `json:"title"`
	Author          string  `json:"author"`
	Creator         string  `json:"creator"`
	Producer        string  `json:"producer"`
	CreationDate    string  `json:"creationDate"`
	ModDate         string  `json:"modDate"`
	PageCount       int     `json:"pageCount"`
	PageWidthPt     float64 `json:"pageWidthPt"`
	PageHeightPt    float64 `json:"pageHeightPt"`
	MaxImageWidthPx int     `json:"maxImageWidthPx"`
}

func (m Metadata) DPI() int {
	if m.MaxImageWidthPx > 0 && m.PageWidthPt > 0 {
		imageDpi := int(float64(m.MaxImageWidthPx)/(m.PageWidthPt/72.0) + 0.5)
		if imageDpi > dpiForTargetWidth(m.PageWidthPt, TargetWidth, MinDpi, MaxDpi) {
			return clampDpi(imageDpi)
		}
	}
	return dpiForTargetWidth(m.PageWidthPt, TargetWidth, MinDpi, MaxDpi)
}

func Equal(stored any, current Metadata) bool {
	if stored == nil {
		return false
	}
	storedBytes, err := json.Marshal(stored)
	if err != nil {
		return false
	}
	currentBytes, err := json.Marshal(current)
	if err != nil {
		return false
	}
	return string(storedBytes) == string(currentBytes)
}

func IsPDF(sourcePath string) bool {
	if strings.EqualFold(filepath.Ext(sourcePath), ".pdf") {
		return true
	}

	file, err := os.Open(sourcePath)
	if err != nil {
		return false
	}
	defer file.Close()

	header := make([]byte, 512)
	readBytes, err := file.Read(header)
	if err != nil || readBytes == 0 {
		return false
	}

	return http.DetectContentType(header[:readBytes]) == "application/pdf"
}

func Read(path string) (Metadata, error) {
	cmd := exec.Command("pdfinfo", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return Metadata{}, fmt.Errorf("pdfinfo failed: %s", strings.TrimSpace(string(output)))
	}
	meta := parse(string(output))
	if imgWidth, err := maxImageWidthPx(path); err == nil && imgWidth > 0 {
		meta.MaxImageWidthPx = imgWidth
	}
	return meta, nil
}

func parse(output string) Metadata {
	var meta Metadata
	for _, line := range strings.Split(output, "\n") {
		key, val, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		if val == "" {
			continue
		}
		switch key {
		case "Title":
			meta.Title = val
		case "Author":
			meta.Author = val
		case "Creator":
			meta.Creator = val
		case "Producer":
			meta.Producer = val
		case "CreationDate":
			meta.CreationDate = val
		case "ModDate":
			meta.ModDate = val
		case "Page size":
			meta.PageWidthPt, meta.PageHeightPt = parsePageSize(val)
		case "Pages":
			if n, err := strconv.Atoi(val); err == nil {
				meta.PageCount = n
			}
		}
	}
	return meta
}

func parsePageSize(val string) (float64, float64) {
	val = strings.TrimSuffix(val, "pts")
	val = strings.TrimSpace(val)
	parts := strings.Fields(val)
	if len(parts) < 3 || parts[1] != "x" {
		return 0, 0
	}
	w, _ := strconv.ParseFloat(parts[0], 64)
	h, _ := strconv.ParseFloat(parts[2], 64)
	return w, h
}

func dpiForTargetWidth(widthPt float64, targetPx, minDpi, maxDpi int) int {
	if widthPt <= 0 {
		return minDpi
	}
	dpi := int(float64(targetPx)/widthPt + 0.5)
	if dpi < minDpi {
		return minDpi
	}
	if dpi > maxDpi {
		return maxDpi
	}
	return dpi
}

func clampDpi(dpi int) int {
	if dpi < MinDpi {
		return MinDpi
	}
	if dpi > MaxDpi {
		return MaxDpi
	}
	return dpi
}

func maxImageWidthPx(path string) (int, error) {
	cmd := exec.Command("pdfimages", "-list", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("pdfimages failed: %s", strings.TrimSpace(string(output)))
	}
	maxWidth := 0
	for i, line := range strings.Split(string(output), "\n") {
		if i < 2 {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		if fields[2] == "image" || fields[2] == "image-mask" {
			if w, err := strconv.Atoi(fields[3]); err == nil && w > maxWidth {
				maxWidth = w
			}
		}
	}
	return maxWidth, nil
}
