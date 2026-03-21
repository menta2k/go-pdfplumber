package textutil

import (
	"testing"

	"github.com/menta2k/go-pdfplumber/pkg/model"
)

func TestGroupIntoWords(t *testing.T) {
	// Simulate "Hello World" with a space between "Hello" and "World"
	chars := []model.Char{
		makeChar("H", 72, 700, 10, 14),
		makeChar("e", 82, 700, 7, 14),
		makeChar("l", 89, 700, 3, 14),
		makeChar("l", 92, 700, 3, 14),
		makeChar("o", 95, 700, 8, 14),
		// gap of ~7 points (> xTolerance=3) before "World"
		makeChar("W", 110, 700, 12, 14),
		makeChar("o", 122, 700, 8, 14),
		makeChar("r", 130, 700, 5, 14),
		makeChar("l", 135, 700, 3, 14),
		makeChar("d", 138, 700, 8, 14),
	}

	opts := model.DefaultTextOptions()
	words := GroupIntoWords(chars, opts)

	if len(words) != 2 {
		t.Fatalf("expected 2 words, got %d", len(words))
	}

	if words[0].Text != "Hello" {
		t.Errorf("first word = %q, want 'Hello'", words[0].Text)
	}
	if words[1].Text != "World" {
		t.Errorf("second word = %q, want 'World'", words[1].Text)
	}

	// Verify bbox covers all chars in word
	if words[0].BBox.X0 != 72 {
		t.Errorf("first word X0 = %f, want 72", words[0].BBox.X0)
	}
	if words[0].Chars == nil || len(words[0].Chars) != 5 {
		t.Errorf("first word should have 5 chars, got %d", len(words[0].Chars))
	}
}

func TestGroupIntoWordsMultiLine(t *testing.T) {
	// Two lines of text
	chars := []model.Char{
		makeChar("A", 72, 700, 10, 14),
		makeChar("B", 82, 700, 10, 14),
		makeChar("C", 72, 680, 10, 14), // different line (Y=680 vs 700)
		makeChar("D", 82, 680, 10, 14),
	}

	opts := model.DefaultTextOptions()
	words := GroupIntoWords(chars, opts)

	if len(words) != 2 {
		t.Fatalf("expected 2 words (one per line), got %d", len(words))
	}

	if words[0].Text != "AB" {
		t.Errorf("first word = %q, want 'AB'", words[0].Text)
	}
	if words[1].Text != "CD" {
		t.Errorf("second word = %q, want 'CD'", words[1].Text)
	}
}

func TestGroupIntoWordsEmpty(t *testing.T) {
	words := GroupIntoWords(nil, model.DefaultTextOptions())
	if words != nil {
		t.Error("expected nil for empty input")
	}
}

func TestGroupIntoWordsSplitPunctuation(t *testing.T) {
	chars := []model.Char{
		makeChar("H", 72, 700, 10, 14),
		makeChar("i", 82, 700, 5, 14),
		makeChar("!", 87, 700, 5, 14),
	}

	opts := model.DefaultTextOptions()
	opts.SplitAtPunctuation = true
	words := GroupIntoWords(chars, opts)

	if len(words) != 2 {
		t.Fatalf("expected 2 words with SplitAtPunctuation, got %d", len(words))
	}
	if words[0].Text != "Hi" {
		t.Errorf("first word = %q, want 'Hi'", words[0].Text)
	}
	if words[1].Text != "!" {
		t.Errorf("second word = %q, want '!'", words[1].Text)
	}
}

func TestGroupIntoWordsExtraAttrs(t *testing.T) {
	chars := []model.Char{
		{Text: "A", FontName: "Helvetica", FontSize: 12, X: 72, Y: 700, Width: 8,
			BBox: model.BBox{X0: 72, Y0: 700, X1: 80, Y1: 712}},
		{Text: "B", FontName: "Courier", FontSize: 12, X: 80, Y: 700, Width: 8,
			BBox: model.BBox{X0: 80, Y0: 700, X1: 88, Y1: 712}},
	}

	opts := model.DefaultTextOptions()
	opts.ExtraAttrs = []string{"FontName"}
	words := GroupIntoWords(chars, opts)

	if len(words) != 2 {
		t.Fatalf("expected 2 words (different fonts), got %d", len(words))
	}
}
