package textutil

import (
	"strings"
	"testing"

	"github.com/menta2k/go-pdfplumber/pkg/model"
)

func TestAssembleText(t *testing.T) {
	lines := []model.TextLine{
		{Text: "Line one", BBox: model.BBox{X0: 72, Y0: 700, X1: 200, Y1: 714}},
		{Text: "Line two", BBox: model.BBox{X0: 72, Y0: 680, X1: 200, Y1: 694}},
		{Text: "Line three", BBox: model.BBox{X0: 72, Y0: 660, X1: 200, Y1: 674}},
	}

	text := AssembleText(lines)

	if !strings.Contains(text, "Line one") {
		t.Error("missing 'Line one'")
	}
	if !strings.Contains(text, "Line two") {
		t.Error("missing 'Line two'")
	}
	if !strings.Contains(text, "Line three") {
		t.Error("missing 'Line three'")
	}

	// Should have newlines between lines
	lineCount := strings.Count(text, "\n")
	if lineCount < 2 {
		t.Errorf("expected at least 2 newlines, got %d", lineCount)
	}
}

func TestAssembleTextParagraphBreak(t *testing.T) {
	// Lines with a large gap in the middle
	lines := []model.TextLine{
		{Text: "Paragraph one", BBox: model.BBox{X0: 72, Y0: 700, X1: 200, Y1: 714}},
		{Text: "Still para one", BBox: model.BBox{X0: 72, Y0: 686, X1: 200, Y1: 700}},
		// Big gap: previous bottom (686) - current top (640) = 46 > 14 * 1.5 = 21
		{Text: "Paragraph two", BBox: model.BBox{X0: 72, Y0: 626, X1: 200, Y1: 640}},
	}

	text := AssembleText(lines)
	if !strings.Contains(text, "\n\n") {
		t.Error("expected double newline for paragraph break")
	}
}

func TestAssembleTextEmpty(t *testing.T) {
	text := AssembleText(nil)
	if text != "" {
		t.Errorf("expected empty string, got %q", text)
	}
}

func TestAssembleTextSingle(t *testing.T) {
	lines := []model.TextLine{
		{Text: "Only line"},
	}
	text := AssembleText(lines)
	if text != "Only line" {
		t.Errorf("expected 'Only line', got %q", text)
	}
}
