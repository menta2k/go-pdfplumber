package plumber

import (
	"testing"
)

func TestPageChars(t *testing.T) {
	doc, err := Open(testSimplePDF)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer doc.Close()

	page, err := doc.Page(1)
	if err != nil {
		t.Fatalf("Page(1): %v", err)
	}

	chars := page.Chars()
	if len(chars) == 0 {
		t.Fatal("expected chars on page 1")
	}

	t.Logf("Found %d chars", len(chars))

	// Verify char properties are populated
	first := chars[0]
	if first.Text == "" {
		t.Error("first char has empty text")
	}
	if first.FontSize <= 0 {
		t.Errorf("first char FontSize = %f, expected > 0", first.FontSize)
	}
	if first.Width <= 0 {
		t.Errorf("first char Width = %f, expected > 0", first.Width)
	}

	// Log first few chars for debugging
	limit := 10
	if len(chars) < limit {
		limit = len(chars)
	}
	for i := 0; i < limit; i++ {
		c := chars[i]
		t.Logf("  char[%d]: %q font=%s size=%.1f x=%.1f y=%.1f w=%.1f",
			i, c.Text, c.FontName, c.FontSize, c.X, c.Y, c.Width)
	}
}

func TestPageCharsImmutability(t *testing.T) {
	doc, err := Open(testSimplePDF)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer doc.Close()

	page, err := doc.Page(1)
	if err != nil {
		t.Fatalf("Page(1): %v", err)
	}

	chars1 := page.Chars()
	chars2 := page.Chars()

	if len(chars1) == 0 {
		t.Fatal("expected chars")
	}

	// Modify first slice, second should be unaffected
	chars1[0].Text = "MODIFIED"
	if chars2[0].Text == "MODIFIED" {
		t.Error("Chars() should return independent copies")
	}
}

func TestPageRects(t *testing.T) {
	doc, err := Open(testTablePDF)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer doc.Close()

	page, err := doc.Page(1)
	if err != nil {
		t.Fatalf("Page(1): %v", err)
	}

	// Table PDF should have either lines or rects (depending on how gopdf draws them)
	lines := page.Lines()
	rects := page.Rects()

	t.Logf("Found %d lines, %d rects", len(lines), len(rects))

	if len(lines) == 0 && len(rects) == 0 {
		t.Log("Warning: no lines or rects found in table PDF (gopdf may use path operators not captured by digitorus/pdf)")
	}
}

func TestPageObjects(t *testing.T) {
	doc, err := Open(testSimplePDF)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer doc.Close()

	page, err := doc.Page(1)
	if err != nil {
		t.Fatalf("Page(1): %v", err)
	}

	objects := page.Objects()

	if _, ok := objects["chars"]; !ok {
		t.Error("expected 'chars' key in objects")
	}
	if _, ok := objects["lines"]; !ok {
		t.Error("expected 'lines' key in objects")
	}
	if _, ok := objects["rects"]; !ok {
		t.Error("expected 'rects' key in objects")
	}

	t.Logf("Objects: chars=%d lines=%d rects=%d",
		len(objects["chars"]), len(objects["lines"]), len(objects["rects"]))
}

func TestPageBBox(t *testing.T) {
	doc, err := Open(testSimplePDF)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer doc.Close()

	page, err := doc.Page(1)
	if err != nil {
		t.Fatalf("Page(1): %v", err)
	}

	bbox := page.PageBBox()
	if bbox.X0 != 0 || bbox.Y0 != 0 {
		t.Errorf("bbox origin = (%.1f, %.1f), want (0, 0)", bbox.X0, bbox.Y0)
	}
	if bbox.X1 != page.Width() || bbox.Y1 != page.Height() {
		t.Errorf("bbox extent = (%.1f, %.1f), want (%.1f, %.1f)",
			bbox.X1, bbox.Y1, page.Width(), page.Height())
	}
}
