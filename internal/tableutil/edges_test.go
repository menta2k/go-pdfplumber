package tableutil

import (
	"testing"

	"github.com/menta2k/go-pdfplumber/pkg/model"
)

func TestExtractEdgesFromRects(t *testing.T) {
	// A single rect should produce 4 edges
	rects := []model.RectObject{
		{BBox: model.BBox{X0: 0, Y0: 0, X1: 100, Y1: 50}},
	}

	edges := ExtractEdges(nil, rects, DefaultEdgeOptions())

	hCount, vCount := 0, 0
	for _, e := range edges {
		switch e.Orientation {
		case "horizontal":
			hCount++
		case "vertical":
			vCount++
		}
	}

	if hCount != 2 {
		t.Errorf("expected 2 horizontal edges, got %d", hCount)
	}
	if vCount != 2 {
		t.Errorf("expected 2 vertical edges, got %d", vCount)
	}
}

func TestExtractEdgesFromLines(t *testing.T) {
	lines := []model.LineSegment{
		{X0: 0, Y0: 0, X1: 100, Y1: 0, Orientation: "horizontal"},
		{X0: 0, Y0: 0, X1: 0, Y1: 50, Orientation: "vertical"},
		{X0: 10, Y0: 10, X1: 20, Y1: 20, Orientation: "diagonal"},
	}

	edges := ExtractEdges(lines, nil, DefaultEdgeOptions())

	// Diagonal should be excluded
	for _, e := range edges {
		if e.Orientation == "diagonal" {
			t.Error("diagonal edges should be filtered out")
		}
	}
}

func TestExtractEdgesMinLength(t *testing.T) {
	rects := []model.RectObject{
		{BBox: model.BBox{X0: 0, Y0: 0, X1: 1, Y1: 1}}, // tiny rect, edges < 3
	}

	opts := DefaultEdgeOptions()
	opts.MinEdgeLength = 3.0
	edges := ExtractEdges(nil, rects, opts)

	if len(edges) != 0 {
		t.Errorf("expected 0 edges (all too short), got %d", len(edges))
	}
}

func TestMergeEdges(t *testing.T) {
	edges := []Edge{
		{X0: 0, Y0: 0, X1: 50, Y1: 0, Orientation: "horizontal"},
		{X0: 48, Y0: 0, X1: 100, Y1: 0, Orientation: "horizontal"}, // overlaps
	}

	merged := MergeEdges(edges, 3.0, 3.0)

	hCount := 0
	for _, e := range merged {
		if e.Orientation == "horizontal" {
			hCount++
			if e.X0 != 0 || e.X1 != 100 {
				t.Errorf("merged edge should span 0-100, got %.1f-%.1f", e.X0, e.X1)
			}
		}
	}
	if hCount != 1 {
		t.Errorf("expected 1 merged horizontal edge, got %d", hCount)
	}
}

func TestMergeEdgesNoMerge(t *testing.T) {
	edges := []Edge{
		{X0: 0, Y0: 0, X1: 30, Y1: 0, Orientation: "horizontal"},
		{X0: 50, Y0: 0, X1: 100, Y1: 0, Orientation: "horizontal"}, // gap > joinTolerance
	}

	merged := MergeEdges(edges, 3.0, 3.0)

	hCount := 0
	for _, e := range merged {
		if e.Orientation == "horizontal" {
			hCount++
		}
	}
	if hCount != 2 {
		t.Errorf("expected 2 separate edges, got %d", hCount)
	}
}
