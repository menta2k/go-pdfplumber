package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"

	"github.com/menta2k/go-pdfplumber/pkg/plumber"
)

var placeholderDescriptions = map[int]string{
	0: "full_name",              // under "от" (име, презиме, фамилия)
	1: "position",               // after "Длъжност:"
	2: "days_count",             // …… before "ден/дни"
	3: "date_from",              // after "считано от"
	4: "date_to",                // between "до" and "вкл"
	5: "manager_name",           // after "Ръководител: ................"
	6: "director_name",          // after "Съгласен, Управител:"
	7: "date_field",             // after "Дата:"
	8: "applicant_signature",    // after "С уважение:"
}

type placeholder struct {
	ID          int     `json:"id"`
	Label       string  `json:"label"`
	Text        string  `json:"text"`
	Page        int     `json:"page"`
	X           float64 `json:"x"`
	YFromTop    float64 `json:"y_from_top"`
	YFromBottom float64 `json:"y_from_bottom"`
	X1          float64 `json:"x1"`
	Width       float64 `json:"width"`
	Height      float64 `json:"height"`
	Font        string  `json:"font"`
	FontSize    float64 `json:"font_size"`
}

type output struct {
	Document         string            `json:"document"`
	PageCount        int               `json:"page_count"`
	PageWidthPt      float64           `json:"page_width_pt"`
	PageHeightPt     float64           `json:"page_height_pt"`
	CoordinateSystem map[string]string `json:"coordinate_system"`
	Placeholders     []placeholder     `json:"placeholders"`
}

func main() {
	pdfPath := "~/projects/go-pdf-test/testdata/bg_01.pdf"
	if len(os.Args) > 1 {
		pdfPath = os.Args[1]
	}

	doc, err := plumber.Open(pdfPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening PDF: %v\n", err)
		os.Exit(1)
	}
	defer doc.Close()

	page, err := doc.Page(1)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting page: %v\n", err)
		os.Exit(1)
	}

	pageHeight := page.Height()
	pageWidth := page.Width()

	// Find placeholder regions by detecting consecutive runs of dot/ellipsis chars
	// sorted spatially, independent of word grouping (which is unreliable with garbled CIDFont text)
	placeholders := findDotPlaceholders(page, pageHeight)

	result := output{
		Document:     docName(pdfPath),
		PageCount:    doc.NumPages(),
		PageWidthPt:  pageWidth,
		PageHeightPt: pageHeight,
		CoordinateSystem: map[string]string{
			"x":             "left edge of placeholder, measured from left of page (points)",
			"y_from_top":    "top edge of placeholder, measured from top of page (pdfplumber convention)",
			"y_from_bottom": "bottom of placeholder baseline, measured from bottom of page (standard PDF convention)",
		},
		Placeholders: placeholders,
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(result); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		os.Exit(1)
	}

	// Render debug PDF
	if len(placeholders) > 0 {
		renderDebugPDF(pdfPath, placeholders, pageHeight)
	}
}

// findDotPlaceholders scans all chars, picks out dots/ellipsis, sorts them
// spatially, and groups consecutive runs on the same line into placeholder regions.
// Uses width-based detection: dots are narrow chars (≤5pt) while Cyrillic letters are 7-14pt.
func findDotPlaceholders(page *plumber.Page, pageHeight float64) []placeholder {
	chars := page.Chars()

	// Collect dot/ellipsis chars using both text and width heuristics.
	// CIDFont garbled text often decodes Cyrillic as ASCII, but real dots
	// have small advance width (3-4pt) while Cyrillic chars are 7-14pt.
	const maxDotWidth = 5.0
	pageW := page.Width()
	var dots []plumber.Char
	for _, c := range chars {
		// Skip chars positioned outside page bounds (garbled CIDFont positioning)
		if c.X < 0 || c.X > pageW || c.Y < 0 || c.Y > pageHeight {
			continue
		}
		// Ellipsis char always counts regardless of width
		if c.Text == "…" {
			dots = append(dots, c)
			continue
		}
		// Period chars with small width (≤5pt) are dot placeholders.
		// The extractor now maps garbled CID period glyphs to '.' so we
		// only need to check for actual period chars.
		if c.Text == "." && c.Width > 0 && c.Width <= maxDotWidth {
			dots = append(dots, c)
		}
	}

	// Sort spatially: top-to-bottom (descending Y), then left-to-right (ascending X)
	sort.Slice(dots, func(i, j int) bool {
		if !nearlyEqual(dots[i].Y, dots[j].Y, 2.0) {
			return dots[i].Y > dots[j].Y
		}
		return dots[i].X < dots[j].X
	})

	// Group consecutive dots on the same line into runs
	// A new run starts when there's a gap > maxGap between consecutive dots on the same line
	const maxGap = 2.0    // max gap between consecutive dots to be same run (real dots are tightly packed)
	const minRunLen = 5   // minimum dots in a run to be a placeholder
	const minEllipsisRun = 2 // but ellipsis runs need only 2 (each … = 3 dots visually)
	const yTolerance = 2.0

	type dotRun struct {
		chars []plumber.Char
	}

	var runs []dotRun
	var current []plumber.Char

	for i, d := range dots {
		if i == 0 {
			current = append(current, d)
			continue
		}

		prev := current[len(current)-1]
		sameLine := nearlyEqual(d.Y, prev.Y, yTolerance)
		gap := d.X - (prev.X + prev.Width)

		if sameLine && gap < maxGap {
			current = append(current, d)
		} else {
			if isValidRun(current, minRunLen, minEllipsisRun) {
				runs = append(runs, dotRun{chars: current})
			}
			current = []plumber.Char{d}
		}
	}
	if isValidRun(current, minRunLen, minEllipsisRun) {
		runs = append(runs, dotRun{chars: current})
	}

	// Convert runs to placeholders
	var placeholders []placeholder
	for _, run := range runs {
		bbox := runBBox(run.chars)
		idx := len(placeholders)

		label := fmt.Sprintf("placeholder_%d", idx)
		if desc, ok := placeholderDescriptions[idx]; ok {
			label = desc
		}

		// Build text from the dots
		var text string
		for _, c := range run.chars {
			text += c.Text
		}

		yFromTop := round3(pageHeight - bbox.Y1)
		yFromBottom := round3(bbox.Y0)

		placeholders = append(placeholders, placeholder{
			ID:          idx,
			Label:       label,
			Text:        text,
			Page:        1,
			X:           round3(bbox.X0),
			YFromTop:    yFromTop,
			YFromBottom: yFromBottom,
			X1:          round3(bbox.X1),
			Width:       round3(bbox.Width()),
			Height:      round3(bbox.Height()),
			Font:        run.chars[0].FontName,
			FontSize:    round3(run.chars[0].FontSize),
		})
	}

	return placeholders
}

// isLikelyDotGlyph returns true for chars that are likely dot/period glyphs.
// Includes actual dots, control chars (garbled CMap), and ellipsis.
// Excludes letters, digits, and common punctuation that isn't dot-like.
func isLikelyDotGlyph(text string) bool {
	for _, r := range text {
		switch {
		case r == '.' || r == '…':
			return true
		case r < 0x20:
			// Control chars from garbled CIDFont CMap (e.g. \x03, \x0f, \x11)
			return true
		}
	}
	return false
}

// isValidRun checks if a run of chars qualifies as a placeholder.
func isValidRun(chars []plumber.Char, minLen, minEllipsisLen int) bool {
	if len(chars) == 0 {
		return false
	}
	// Count ellipsis chars — they represent 3 visible dots each
	hasEllipsis := false
	for _, c := range chars {
		if c.Text == "…" {
			hasEllipsis = true
			break
		}
	}
	if hasEllipsis {
		return len(chars) >= minEllipsisLen
	}
	return len(chars) >= minLen
}

func isDotChar(text string) bool {
	for _, r := range text {
		if r == '.' || r == '…' {
			return true
		}
	}
	return false
}

func runBBox(chars []plumber.Char) plumber.BBox {
	if len(chars) == 0 {
		return plumber.BBox{}
	}
	bbox := chars[0].BBox
	for _, c := range chars[1:] {
		if c.BBox.X0 < bbox.X0 {
			bbox.X0 = c.BBox.X0
		}
		if c.BBox.Y0 < bbox.Y0 {
			bbox.Y0 = c.BBox.Y0
		}
		if c.BBox.X1 > bbox.X1 {
			bbox.X1 = c.BBox.X1
		}
		if c.BBox.Y1 > bbox.Y1 {
			bbox.Y1 = c.BBox.Y1
		}
	}
	return bbox
}

func nearlyEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}

func renderDebugPDF(sourcePDF string, phs []placeholder, pageHeight float64) {
	overlays := make([]plumber.DebugOverlay, len(phs))
	for i, ph := range phs {
		overlays[i] = plumber.DebugOverlay{
			BBox: plumber.BBox{
				X0: ph.X,
				Y0: ph.YFromBottom,
				X1: ph.X1,
				Y1: pageHeight - ph.YFromTop,
			},
			Label: fmt.Sprintf("[%d] %s", ph.ID, ph.Label),
		}
	}

	outPath := "placeholders_debug.pdf"
	err := plumber.SaveDebugPDF(sourcePDF, 1, overlays, outPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saving debug PDF: %v\n", err)
		return
	}
	fmt.Fprintf(os.Stderr, "Saved debug PDF with %d overlays to %s\n", len(phs), outPath)
}

func round3(f float64) float64 {
	return float64(int(f*1000+0.5)) / 1000
}

func docName(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[i+1:]
		}
	}
	return path
}
