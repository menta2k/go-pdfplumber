package tableutil

import (
	"testing"

	"github.com/menta2k/go-pdfplumber/pkg/model"
)

func TestBuildCells2x2(t *testing.T) {
	// 2x2 grid
	points := []model.Point{
		{X: 0, Y: 0}, {X: 100, Y: 0}, {X: 200, Y: 0},
		{X: 0, Y: 100}, {X: 100, Y: 100}, {X: 200, Y: 100},
		{X: 0, Y: 200}, {X: 100, Y: 200}, {X: 200, Y: 200},
	}

	edges := []Edge{
		// 3 horizontal lines
		{X0: 0, Y0: 0, X1: 200, Y1: 0, Orientation: "horizontal"},
		{X0: 0, Y0: 100, X1: 200, Y1: 100, Orientation: "horizontal"},
		{X0: 0, Y0: 200, X1: 200, Y1: 200, Orientation: "horizontal"},
		// 3 vertical lines
		{X0: 0, Y0: 0, X1: 0, Y1: 200, Orientation: "vertical"},
		{X0: 100, Y0: 0, X1: 100, Y1: 200, Orientation: "vertical"},
		{X0: 200, Y0: 0, X1: 200, Y1: 200, Orientation: "vertical"},
	}

	grid := BuildCells(points, edges, 3.0)

	if len(grid) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(grid))
	}
	for r, row := range grid {
		if len(row) != 2 {
			t.Errorf("row %d: expected 2 cols, got %d", r, len(row))
		}
	}

	// First cell in reading order is top-left (highest Y row first)
	topLeft := grid[0][0].BBox
	if topLeft.X0 != 0 || topLeft.Y0 != 100 || topLeft.X1 != 100 || topLeft.Y1 != 200 {
		t.Errorf("first cell (top-left) bbox wrong: %+v", topLeft)
	}
}

func TestBuildCellsInsufficientPoints(t *testing.T) {
	points := []model.Point{{X: 0, Y: 0}}
	grid := BuildCells(points, nil, 3.0)
	if grid != nil {
		t.Error("expected nil for insufficient points")
	}
}

func TestBuildCells3x3(t *testing.T) {
	// 3x3 grid = 4x4 intersections = 16 points
	var points []model.Point
	for y := 0; y <= 3; y++ {
		for x := 0; x <= 3; x++ {
			points = append(points, model.Point{X: float64(x * 100), Y: float64(y * 100)})
		}
	}

	edges := []Edge{
		// 4 horizontal
		{X0: 0, Y0: 0, X1: 300, Y1: 0, Orientation: "horizontal"},
		{X0: 0, Y0: 100, X1: 300, Y1: 100, Orientation: "horizontal"},
		{X0: 0, Y0: 200, X1: 300, Y1: 200, Orientation: "horizontal"},
		{X0: 0, Y0: 300, X1: 300, Y1: 300, Orientation: "horizontal"},
		// 4 vertical
		{X0: 0, Y0: 0, X1: 0, Y1: 300, Orientation: "vertical"},
		{X0: 100, Y0: 0, X1: 100, Y1: 300, Orientation: "vertical"},
		{X0: 200, Y0: 0, X1: 200, Y1: 300, Orientation: "vertical"},
		{X0: 300, Y0: 0, X1: 300, Y1: 300, Orientation: "vertical"},
	}

	grid := BuildCells(points, edges, 3.0)

	if len(grid) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(grid))
	}
	for r, row := range grid {
		if len(row) != 3 {
			t.Errorf("row %d: expected 3 cols, got %d", r, len(row))
		}
	}
}

func TestGridBBox(t *testing.T) {
	grid := [][]Cell{
		{
			{BBox: model.BBox{X0: 0, Y0: 0, X1: 100, Y1: 50}},
			{BBox: model.BBox{X0: 100, Y0: 0, X1: 200, Y1: 50}},
		},
		{
			{BBox: model.BBox{X0: 0, Y0: 50, X1: 100, Y1: 100}},
			{BBox: model.BBox{X0: 100, Y0: 50, X1: 200, Y1: 100}},
		},
	}

	bbox := GridBBox(grid)
	if bbox.X0 != 0 || bbox.Y0 != 0 || bbox.X1 != 200 || bbox.Y1 != 100 {
		t.Errorf("GridBBox = %+v, want (0,0)-(200,100)", bbox)
	}
}

func TestGridBBoxEmpty(t *testing.T) {
	bbox := GridBBox(nil)
	if bbox != (model.BBox{}) {
		t.Errorf("GridBBox(nil) = %+v, want zero", bbox)
	}
}
