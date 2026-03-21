package textutil

import (
	"testing"

	"github.com/menta2k/go-pdfplumber/pkg/model"
)

func makeChar(text string, x, y, w, fontSize float64) model.Char {
	return model.Char{
		Text:     text,
		FontName: "Helvetica",
		FontSize: fontSize,
		X:        x,
		Y:        y,
		Width:    w,
		BBox:     model.BBox{X0: x, Y0: y, X1: x + w, Y1: y + fontSize},
		Top:      y + fontSize,
		Bottom:   y,
	}
}

func TestSortChars(t *testing.T) {
	chars := []model.Char{
		makeChar("c", 200, 100, 8, 12), // same line, right
		makeChar("a", 100, 100, 8, 12), // same line, left
		makeChar("b", 100, 200, 8, 12), // top line
	}

	sorted := SortChars(chars)
	if sorted[0].Text != "b" {
		t.Errorf("first char should be 'b' (highest Y), got %q", sorted[0].Text)
	}
	if sorted[1].Text != "a" {
		t.Errorf("second char should be 'a' (lower Y, left X), got %q", sorted[1].Text)
	}
	if sorted[2].Text != "c" {
		t.Errorf("third char should be 'c' (lower Y, right X), got %q", sorted[2].Text)
	}

	// Verify original not mutated
	if chars[0].Text != "c" {
		t.Error("SortChars mutated original slice")
	}
}

func TestDedupeChars(t *testing.T) {
	chars := []model.Char{
		makeChar("A", 100, 100, 8, 12),
		makeChar("A", 100.5, 100, 8, 12), // duplicate (within tolerance)
		makeChar("B", 110, 100, 8, 12),
	}

	deduped := DedupeChars(chars, 1.0)
	if len(deduped) != 2 {
		t.Fatalf("expected 2 chars after dedup, got %d", len(deduped))
	}
	if deduped[0].Text != "A" || deduped[1].Text != "B" {
		t.Errorf("unexpected dedup result: %q %q", deduped[0].Text, deduped[1].Text)
	}
}

func TestDedupeCharsEmpty(t *testing.T) {
	result := DedupeChars(nil, 1.0)
	if result != nil {
		t.Error("expected nil for empty input")
	}
}

func TestFilterBlankChars(t *testing.T) {
	chars := []model.Char{
		makeChar("A", 100, 100, 8, 12),
		makeChar(" ", 108, 100, 4, 12),
		makeChar("B", 112, 100, 8, 12),
	}

	filtered := FilterBlankChars(chars)
	if len(filtered) != 2 {
		t.Fatalf("expected 2 non-blank chars, got %d", len(filtered))
	}
	if filtered[0].Text != "A" || filtered[1].Text != "B" {
		t.Errorf("unexpected result: %q %q", filtered[0].Text, filtered[1].Text)
	}
}
