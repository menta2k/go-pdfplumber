package tableutil

import (
	"math"
	"sort"

	"github.com/menta2k/go-pdfplumber/pkg/model"
)

// FindIntersections finds all intersection points between horizontal and vertical edges.
// Points within snapTolerance of each other are merged.
func FindIntersections(edges []Edge, snapTolerance float64) []model.Point {
	var horizontal, vertical []Edge
	for _, e := range edges {
		switch e.Orientation {
		case "horizontal":
			horizontal = append(horizontal, e)
		case "vertical":
			vertical = append(vertical, e)
		}
	}

	var points []model.Point

	for _, h := range horizontal {
		for _, v := range vertical {
			// Check if they cross: v.X within h's X range, h.Y within v's Y range
			if v.X0 >= h.X0-snapTolerance && v.X0 <= h.X1+snapTolerance &&
				h.Y0 >= v.Y0-snapTolerance && h.Y0 <= v.Y1+snapTolerance {
				points = append(points, model.Point{X: v.X0, Y: h.Y0})
			}
		}
	}

	// Deduplicate nearby points
	points = dedupePoints(points, snapTolerance)

	// Sort: bottom-to-top (ascending Y), then left-to-right (ascending X)
	sort.Slice(points, func(i, j int) bool {
		if !nearlyEqual(points[i].Y, points[j].Y, 0.5) {
			return points[i].Y < points[j].Y
		}
		return points[i].X < points[j].X
	})

	return points
}

func dedupePoints(points []model.Point, tolerance float64) []model.Point {
	if len(points) == 0 {
		return nil
	}

	var result []model.Point
	for _, p := range points {
		isDupe := false
		for _, existing := range result {
			if nearlyEqual(p.X, existing.X, tolerance) &&
				nearlyEqual(p.Y, existing.Y, tolerance) {
				isDupe = true
				break
			}
		}
		if !isDupe {
			result = append(result, p)
		}
	}
	return result
}

// UniqueCoords extracts unique sorted X and Y coordinates from intersection points.
func UniqueCoords(points []model.Point, tolerance float64) (xs, ys []float64) {
	xSet := uniqueFloats(tolerance)
	ySet := uniqueFloats(tolerance)

	for _, p := range points {
		xSet.add(p.X)
		ySet.add(p.Y)
	}

	xs = xSet.sorted()
	ys = ySet.sorted()
	return
}

type floatSet struct {
	tolerance float64
	values    []float64
}

func uniqueFloats(tolerance float64) *floatSet {
	return &floatSet{tolerance: tolerance}
}

func (s *floatSet) add(v float64) {
	for _, existing := range s.values {
		if math.Abs(v-existing) <= s.tolerance {
			return
		}
	}
	s.values = append(s.values, v)
}

func (s *floatSet) sorted() []float64 {
	sort.Float64s(s.values)
	return s.values
}

func nearlyEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}
