package tableutil

import (
	"testing"

	"github.com/menta2k/go-pdfplumber/pkg/model"
)

func TestExtractCellText(t *testing.T) {
	chars := []model.Char{
		{Text: "A", X: 10, Y: 10, Width: 8, FontSize: 12,
			BBox: model.BBox{X0: 10, Y0: 10, X1: 18, Y1: 22}},
		{Text: "B", X: 110, Y: 10, Width: 8, FontSize: 12,
			BBox: model.BBox{X0: 110, Y0: 10, X1: 118, Y1: 22}},
		{Text: "C", X: 10, Y: 110, Width: 8, FontSize: 12,
			BBox: model.BBox{X0: 10, Y0: 110, X1: 18, Y1: 122}},
		{Text: "D", X: 110, Y: 110, Width: 8, FontSize: 12,
			BBox: model.BBox{X0: 110, Y0: 110, X1: 118, Y1: 122}},
	}

	grid := [][]Cell{
		{
			{BBox: model.BBox{X0: 0, Y0: 0, X1: 100, Y1: 100}},
			{BBox: model.BBox{X0: 100, Y0: 0, X1: 200, Y1: 100}},
		},
		{
			{BBox: model.BBox{X0: 0, Y0: 100, X1: 100, Y1: 200}},
			{BBox: model.BBox{X0: 100, Y0: 100, X1: 200, Y1: 200}},
		},
	}

	opts := model.DefaultTextOptions()
	result := ExtractCellText(chars, grid, opts)

	if len(result) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(result))
	}

	expected := [][]string{
		{"A", "B"},
		{"C", "D"},
	}

	for r, row := range expected {
		for c, want := range row {
			if result[r][c] != want {
				t.Errorf("cell[%d][%d] = %q, want %q", r, c, result[r][c], want)
			}
		}
	}
}

func TestExtractCellTextEmpty(t *testing.T) {
	result := ExtractCellText(nil, nil, model.DefaultTextOptions())
	if result != nil {
		t.Error("expected nil for empty input")
	}
}

func TestExtractCellTextMultiCharCell(t *testing.T) {
	chars := []model.Char{
		{Text: "H", X: 10, Y: 50, Width: 8, FontSize: 12,
			BBox: model.BBox{X0: 10, Y0: 50, X1: 18, Y1: 62}},
		{Text: "i", X: 18, Y: 50, Width: 5, FontSize: 12,
			BBox: model.BBox{X0: 18, Y0: 50, X1: 23, Y1: 62}},
	}

	grid := [][]Cell{
		{{BBox: model.BBox{X0: 0, Y0: 0, X1: 100, Y1: 100}}},
	}

	opts := model.DefaultTextOptions()
	result := ExtractCellText(chars, grid, opts)

	if result[0][0] != "Hi" {
		t.Errorf("cell text = %q, want 'Hi'", result[0][0])
	}
}
