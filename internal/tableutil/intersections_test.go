package tableutil

import (
	"testing"

	"github.com/menta2k/go-pdfplumber/pkg/model"
)

func TestFindIntersections(t *testing.T) {
	// Simple cross: one H and one V edge
	edges := []Edge{
		{X0: 0, Y0: 50, X1: 100, Y1: 50, Orientation: "horizontal"},
		{X0: 50, Y0: 0, X1: 50, Y1: 100, Orientation: "vertical"},
	}

	points := FindIntersections(edges, 3.0)

	if len(points) != 1 {
		t.Fatalf("expected 1 intersection, got %d", len(points))
	}
	if !nearlyEqual(points[0].X, 50, 1) || !nearlyEqual(points[0].Y, 50, 1) {
		t.Errorf("expected (50,50), got (%.1f, %.1f)", points[0].X, points[0].Y)
	}
}

func TestFindIntersectionsGrid(t *testing.T) {
	// 2x2 grid: 3 horizontal + 3 vertical = 9 intersections
	edges := []Edge{
		{X0: 0, Y0: 0, X1: 200, Y1: 0, Orientation: "horizontal"},
		{X0: 0, Y0: 100, X1: 200, Y1: 100, Orientation: "horizontal"},
		{X0: 0, Y0: 200, X1: 200, Y1: 200, Orientation: "horizontal"},
		{X0: 0, Y0: 0, X1: 0, Y1: 200, Orientation: "vertical"},
		{X0: 100, Y0: 0, X1: 100, Y1: 200, Orientation: "vertical"},
		{X0: 200, Y0: 0, X1: 200, Y1: 200, Orientation: "vertical"},
	}

	points := FindIntersections(edges, 3.0)

	if len(points) != 9 {
		t.Fatalf("expected 9 intersections for 2x2 grid, got %d", len(points))
	}
}

func TestFindIntersectionsEmpty(t *testing.T) {
	points := FindIntersections(nil, 3.0)
	if points != nil {
		t.Error("expected nil for empty input")
	}
}

func TestUniqueCoords(t *testing.T) {
	points := []model.Point{
		{X: 0, Y: 0},
		{X: 100, Y: 0},
		{X: 0, Y: 50},
		{X: 100, Y: 50},
		{X: 100.5, Y: 50.5}, // should merge with (100, 50)
	}

	xs, ys := UniqueCoords(points, 1.0)

	if len(xs) != 2 {
		t.Errorf("expected 2 unique X coords, got %d: %v", len(xs), xs)
	}
	if len(ys) != 2 {
		t.Errorf("expected 2 unique Y coords, got %d: %v", len(ys), ys)
	}
}
