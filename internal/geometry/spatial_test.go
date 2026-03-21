package geometry

import (
	"testing"

	"github.com/menta2k/go-pdfplumber/pkg/model"
)

func TestFilterWithin(t *testing.T) {
	region := model.BBox{X0: 10, Y0: 10, X1: 50, Y1: 50}

	chars := []model.Char{
		{Text: "A", BBox: model.BBox{X0: 15, Y0: 15, X1: 25, Y1: 25}}, // inside
		{Text: "B", BBox: model.BBox{X0: 45, Y0: 45, X1: 55, Y1: 55}}, // partial overlap
		{Text: "C", BBox: model.BBox{X0: 60, Y0: 60, X1: 70, Y1: 70}}, // outside
	}

	result := FilterWithin(chars, region)
	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}
	if result[0].Text != "A" {
		t.Errorf("expected 'A', got %q", result[0].Text)
	}
}

func TestFilterOutside(t *testing.T) {
	region := model.BBox{X0: 10, Y0: 10, X1: 50, Y1: 50}

	chars := []model.Char{
		{Text: "A", BBox: model.BBox{X0: 15, Y0: 15, X1: 25, Y1: 25}}, // inside
		{Text: "B", BBox: model.BBox{X0: 45, Y0: 45, X1: 55, Y1: 55}}, // partial overlap
		{Text: "C", BBox: model.BBox{X0: 60, Y0: 60, X1: 70, Y1: 70}}, // outside
	}

	result := FilterOutside(chars, region)
	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}
	if result[0].Text != "C" {
		t.Errorf("expected 'C', got %q", result[0].Text)
	}
}

func TestFilterOverlapping(t *testing.T) {
	region := model.BBox{X0: 10, Y0: 10, X1: 50, Y1: 50}

	chars := []model.Char{
		{Text: "A", BBox: model.BBox{X0: 15, Y0: 15, X1: 25, Y1: 25}}, // inside
		{Text: "B", BBox: model.BBox{X0: 45, Y0: 45, X1: 55, Y1: 55}}, // partial overlap
		{Text: "C", BBox: model.BBox{X0: 60, Y0: 60, X1: 70, Y1: 70}}, // outside
	}

	result := FilterOverlapping(chars, region)
	if len(result) != 2 {
		t.Fatalf("expected 2 (inside + overlap), got %d", len(result))
	}
}

func TestFilterFunc(t *testing.T) {
	chars := []model.Char{
		{Text: "A", FontSize: 12},
		{Text: "B", FontSize: 24},
		{Text: "C", FontSize: 12},
	}

	large := FilterFunc(chars, func(c model.Char) bool {
		return c.FontSize > 14
	})
	if len(large) != 1 {
		t.Fatalf("expected 1, got %d", len(large))
	}
	if large[0].Text != "B" {
		t.Errorf("expected 'B', got %q", large[0].Text)
	}
}

func TestFilterCharsByMidpoint(t *testing.T) {
	region := model.BBox{X0: 10, Y0: 10, X1: 50, Y1: 50}

	chars := []model.Char{
		{Text: "A", BBox: model.BBox{X0: 20, Y0: 20, X1: 30, Y1: 30}},  // midpoint (25,25) inside
		{Text: "B", BBox: model.BBox{X0: 45, Y0: 45, X1: 55, Y1: 55}},  // midpoint (50,50) on edge
		{Text: "C", BBox: model.BBox{X0: 60, Y0: 60, X1: 70, Y1: 70}},  // midpoint (65,65) outside
	}

	result := FilterCharsByMidpoint(chars, region)
	if len(result) != 2 {
		t.Fatalf("expected 2 (A midpoint inside, B midpoint on edge), got %d", len(result))
	}
}

func TestFilterWithinEmpty(t *testing.T) {
	region := model.BBox{X0: 10, Y0: 10, X1: 50, Y1: 50}
	result := FilterWithin([]model.Char{}, region)
	if result != nil {
		t.Error("expected nil for empty input")
	}
}

func TestFilterWithinLineSegments(t *testing.T) {
	region := model.BBox{X0: 0, Y0: 0, X1: 100, Y1: 100}

	lines := []model.LineSegment{
		{BBox: model.BBox{X0: 10, Y0: 10, X1: 90, Y1: 10}},  // inside
		{BBox: model.BBox{X0: 50, Y0: 50, X1: 150, Y1: 50}},  // partial
	}

	result := FilterWithin(lines, region)
	if len(result) != 1 {
		t.Fatalf("expected 1 line inside, got %d", len(result))
	}
}
