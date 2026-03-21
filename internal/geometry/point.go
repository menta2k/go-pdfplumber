package geometry

import (
	"math"

	"github.com/menta2k/go-pdfplumber/pkg/model"
)

// Distance returns the Euclidean distance between two points.
func Distance(a, b model.Point) float64 {
	dx := b.X - a.X
	dy := b.Y - a.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// SegmentIntersection finds the intersection point of two line segments.
// Returns the point and true if they intersect, or zero Point and false otherwise.
func SegmentIntersection(ax0, ay0, ax1, ay1, bx0, by0, bx1, by1 float64) (model.Point, bool) {
	dax := ax1 - ax0
	day := ay1 - ay0
	dbx := bx1 - bx0
	dby := by1 - by0

	denom := dax*dby - day*dbx
	if math.Abs(denom) < 1e-10 {
		return model.Point{}, false // parallel or coincident
	}

	t := ((bx0-ax0)*dby - (by0-ay0)*dbx) / denom
	u := ((bx0-ax0)*day - (by0-ay0)*dax) / denom

	if t < -1e-10 || t > 1+1e-10 || u < -1e-10 || u > 1+1e-10 {
		return model.Point{}, false // intersection outside segments
	}

	return model.Point{
		X: ax0 + t*dax,
		Y: ay0 + t*day,
	}, true
}

// NearlyEqual reports whether two float64 values are within tolerance.
func NearlyEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}
