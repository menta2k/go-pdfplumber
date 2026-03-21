package plumber

import (
	"strings"
	"testing"
)

func TestCrop(t *testing.T) {
	doc, err := Open(testSimplePDF)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer doc.Close()

	page, err := doc.Page(1)
	if err != nil {
		t.Fatalf("Page(1): %v", err)
	}

	allChars := page.Chars()
	t.Logf("Total chars on page: %d", len(allChars))

	// Crop to the area around the first line only
	// First line is at Y≈755.7, so crop Y from 750 to 775
	cropBox := BBox{X0: 70, Y0: 750, X1: 200, Y1: 775}
	cropped, err := page.Crop(cropBox)
	if err != nil {
		t.Fatalf("Crop: %v", err)
	}

	croppedChars := cropped.Chars()
	t.Logf("Cropped chars: %d", len(croppedChars))

	text := cropped.ExtractText()
	t.Logf("Cropped text: %q", text)

	if !strings.Contains(text, "Hello") {
		t.Error("cropped text should contain 'Hello'")
	}
	if strings.Contains(text, "test PDF") {
		t.Error("cropped text should NOT contain text from other lines")
	}

	// Verify dimensions match crop box
	if cropped.Width() != cropBox.Width() {
		t.Errorf("cropped width = %f, want %f", cropped.Width(), cropBox.Width())
	}
	if cropped.Height() != cropBox.Height() {
		t.Errorf("cropped height = %f, want %f", cropped.Height(), cropBox.Height())
	}
}

func TestCropImmutability(t *testing.T) {
	doc, err := Open(testSimplePDF)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer doc.Close()

	page, err := doc.Page(1)
	if err != nil {
		t.Fatalf("Page(1): %v", err)
	}

	originalCount := len(page.Chars())

	cropBox := BBox{X0: 70, Y0: 750, X1: 200, Y1: 775}
	_, err = page.Crop(cropBox)
	if err != nil {
		t.Fatalf("Crop: %v", err)
	}

	// Original page should be unchanged
	afterCount := len(page.Chars())
	if afterCount != originalCount {
		t.Errorf("original page char count changed from %d to %d", originalCount, afterCount)
	}
}

func TestCropInvalidBBox(t *testing.T) {
	doc, err := Open(testSimplePDF)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer doc.Close()

	page, err := doc.Page(1)
	if err != nil {
		t.Fatalf("Page(1): %v", err)
	}

	// Zero-area bbox
	_, err = page.Crop(BBox{X0: 10, Y0: 10, X1: 10, Y1: 20})
	if err == nil {
		t.Error("expected error for zero-width crop box")
	}
}

func TestWithinBBox(t *testing.T) {
	doc, err := Open(testSimplePDF)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer doc.Close()

	page, err := doc.Page(1)
	if err != nil {
		t.Fatalf("Page(1): %v", err)
	}

	// Region covering first line area
	region := BBox{X0: 70, Y0: 750, X1: 200, Y1: 775}
	filtered := page.WithinBBox(region)

	chars := filtered.Chars()
	t.Logf("WithinBBox chars: %d", len(chars))

	// All returned chars should be fully inside the region
	for _, c := range chars {
		if !region.ContainsBBox(c.BBox) {
			t.Errorf("char %q at (%.1f,%.1f) bbox not fully inside region",
				c.Text, c.X, c.Y)
		}
	}

	// Original page should be unchanged
	allChars := page.Chars()
	if len(allChars) == len(chars) {
		t.Error("WithinBBox should have fewer chars than full page")
	}
}

func TestOutsideBBox(t *testing.T) {
	doc, err := Open(testSimplePDF)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer doc.Close()

	page, err := doc.Page(1)
	if err != nil {
		t.Fatalf("Page(1): %v", err)
	}

	// Remove the first line area
	region := BBox{X0: 70, Y0: 750, X1: 200, Y1: 775}
	filtered := page.OutsideBBox(region)

	text := filtered.ExtractText()
	t.Logf("Outside text: %q", text)

	if strings.Contains(text, "Hello") {
		t.Error("outside text should NOT contain 'Hello'")
	}
	if !strings.Contains(text, "test PDF") {
		t.Error("outside text should contain 'test PDF'")
	}
}

func TestFilterChars(t *testing.T) {
	doc, err := Open(testSimplePDF)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer doc.Close()

	page, err := doc.Page(1)
	if err != nil {
		t.Fatalf("Page(1): %v", err)
	}

	// Keep only uppercase letters
	filtered := page.FilterChars(func(c Char) bool {
		return len(c.Text) > 0 && c.Text[0] >= 'A' && c.Text[0] <= 'Z'
	})

	chars := filtered.Chars()
	t.Logf("Uppercase chars: %d", len(chars))

	for _, c := range chars {
		if c.Text[0] < 'A' || c.Text[0] > 'Z' {
			t.Errorf("non-uppercase char %q passed filter", c.Text)
		}
	}
}

func TestCropThenExtractText(t *testing.T) {
	doc, err := Open(testTablePDF)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer doc.Close()

	page, err := doc.Page(1)
	if err != nil {
		t.Fatalf("Page(1): %v", err)
	}

	// Crop to first column of the table (x: 72-192)
	// Table rects start at y≈691 and go to y≈761
	cropBox := BBox{X0: 72, Y0: 685, X1: 195, Y1: 765}
	cropped, err := page.Crop(cropBox)
	if err != nil {
		t.Fatalf("Crop: %v", err)
	}

	text := cropped.ExtractText()
	t.Logf("First column text: %q", text)

	if !strings.Contains(text, "Name") {
		t.Error("first column should contain 'Name'")
	}
	if !strings.Contains(text, "Alice") {
		t.Error("first column should contain 'Alice'")
	}
	if strings.Contains(text, "Age") {
		t.Error("first column should NOT contain 'Age' (second column)")
	}
}
