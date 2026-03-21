package textutil

import (
	"testing"

	"github.com/menta2k/go-pdfplumber/pkg/model"
)

func TestGroupIntoLines(t *testing.T) {
	words := []model.Word{
		{Text: "Hello", BBox: model.BBox{X0: 72, Y0: 700, X1: 120, Y1: 714}},
		{Text: "World", BBox: model.BBox{X0: 130, Y0: 700, X1: 180, Y1: 714}},
		{Text: "Second", BBox: model.BBox{X0: 72, Y0: 670, X1: 130, Y1: 684}},
		{Text: "Line", BBox: model.BBox{X0: 140, Y0: 670, X1: 180, Y1: 684}},
	}

	lines := GroupIntoLines(words, 3.0)

	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}

	if lines[0].Text != "Hello World" {
		t.Errorf("first line = %q, want 'Hello World'", lines[0].Text)
	}
	if lines[1].Text != "Second Line" {
		t.Errorf("second line = %q, want 'Second Line'", lines[1].Text)
	}

	// Verify top-to-bottom ordering (higher Y = first line)
	if lines[0].BBox.Y0 < lines[1].BBox.Y0 {
		t.Error("lines should be sorted top-to-bottom (higher Y first)")
	}
}

func TestGroupIntoLinesEmpty(t *testing.T) {
	lines := GroupIntoLines(nil, 3.0)
	if lines != nil {
		t.Error("expected nil for empty input")
	}
}

func TestGroupIntoLinesSingleWord(t *testing.T) {
	words := []model.Word{
		{Text: "Alone", BBox: model.BBox{X0: 72, Y0: 700, X1: 120, Y1: 714}},
	}

	lines := GroupIntoLines(words, 3.0)
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	if lines[0].Text != "Alone" {
		t.Errorf("line text = %q, want 'Alone'", lines[0].Text)
	}
}
