package plumber

import (
	"strings"
	"testing"
)

func TestExtractWords(t *testing.T) {
	doc, err := Open(testSimplePDF)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer doc.Close()

	page, err := doc.Page(1)
	if err != nil {
		t.Fatalf("Page(1): %v", err)
	}

	words := page.ExtractWords()
	if len(words) == 0 {
		t.Fatal("expected words")
	}

	t.Logf("Found %d words:", len(words))
	for i, w := range words {
		t.Logf("  word[%d]: %q bbox=(%.1f, %.1f, %.1f, %.1f)",
			i, w.Text, w.BBox.X0, w.BBox.Y0, w.BBox.X1, w.BBox.Y1)
	}

	// Verify "Hello" and "World" are extracted
	found := map[string]bool{}
	for _, w := range words {
		found[w.Text] = true
	}
	for _, want := range []string{"Hello", "World"} {
		if !found[want] {
			t.Errorf("expected word %q not found", want)
		}
	}
}

func TestExtractTextLines(t *testing.T) {
	doc, err := Open(testSimplePDF)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer doc.Close()

	page, err := doc.Page(1)
	if err != nil {
		t.Fatalf("Page(1): %v", err)
	}

	lines := page.ExtractTextLines()
	if len(lines) == 0 {
		t.Fatal("expected text lines")
	}

	t.Logf("Found %d lines:", len(lines))
	for i, l := range lines {
		t.Logf("  line[%d]: %q", i, l.Text)
	}

	// Should have 3 lines
	if len(lines) < 3 {
		t.Errorf("expected at least 3 lines, got %d", len(lines))
	}
}

func TestExtractText(t *testing.T) {
	doc, err := Open(testSimplePDF)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer doc.Close()

	page, err := doc.Page(1)
	if err != nil {
		t.Fatalf("Page(1): %v", err)
	}

	text := page.ExtractText()
	t.Logf("Extracted text:\n%s", text)

	if !strings.Contains(text, "Hello") {
		t.Error("text should contain 'Hello'")
	}
	if !strings.Contains(text, "World") {
		t.Error("text should contain 'World'")
	}
	if !strings.Contains(text, "test PDF document") {
		t.Error("text should contain 'test PDF document'")
	}
	if !strings.Contains(text, "three lines") {
		t.Error("text should contain 'three lines'")
	}
}

func TestExtractTextMultipage(t *testing.T) {
	doc, err := Open(testMultipagePDF)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer doc.Close()

	for i := 1; i <= 3; i++ {
		page, err := doc.Page(i)
		if err != nil {
			t.Fatalf("Page(%d): %v", i, err)
		}

		text := page.ExtractText()
		expected := "Page"
		if !strings.Contains(text, expected) {
			t.Errorf("page %d: text should contain %q, got %q", i, expected, text)
		}
	}
}

func TestExtractWordsCustomTolerance(t *testing.T) {
	doc, err := Open(testSimplePDF)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer doc.Close()

	page, err := doc.Page(1)
	if err != nil {
		t.Fatalf("Page(1): %v", err)
	}

	// Very tight tolerance should produce more words
	tight := TextOptions{XTolerance: 0.1, YTolerance: 3.0}
	tightWords := page.ExtractWords(tight)

	// Very loose tolerance should produce fewer words
	loose := TextOptions{XTolerance: 50.0, YTolerance: 3.0}
	looseWords := page.ExtractWords(loose)

	t.Logf("Tight tolerance: %d words, Loose tolerance: %d words",
		len(tightWords), len(looseWords))

	if len(tightWords) <= len(looseWords) {
		t.Error("tight tolerance should produce more words than loose")
	}
}

func TestSearch(t *testing.T) {
	doc, err := Open(testSimplePDF)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer doc.Close()

	page, err := doc.Page(1)
	if err != nil {
		t.Fatalf("Page(1): %v", err)
	}

	matches := page.Search("Hello")
	if len(matches) == 0 {
		t.Error("expected to find 'Hello'")
	}

	matches = page.Search("xyz")
	if len(matches) != 0 {
		t.Error("expected no matches for 'xyz'")
	}
}

func TestExtractWordsFromTablePDF(t *testing.T) {
	doc, err := Open(testTablePDF)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer doc.Close()

	page, err := doc.Page(1)
	if err != nil {
		t.Fatalf("Page(1): %v", err)
	}

	words := page.ExtractWords()
	t.Logf("Table PDF words: %d", len(words))
	for _, w := range words {
		t.Logf("  %q at (%.1f, %.1f)", w.Text, w.BBox.X0, w.BBox.Y0)
	}

	// Should find table cell contents
	found := map[string]bool{}
	for _, w := range words {
		found[w.Text] = true
	}
	for _, want := range []string{"Name", "Age", "City", "Alice", "Bob"} {
		if !found[want] {
			t.Errorf("expected word %q not found in table PDF", want)
		}
	}
}
