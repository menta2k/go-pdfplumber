package geometry

import "github.com/menta2k/go-pdfplumber/pkg/model"

// Intersect returns the intersection of two bounding boxes.
// The second return value reports whether the intersection is non-empty.
func Intersect(a, b model.BBox) (model.BBox, bool) {
	result := model.BBox{
		X0: max(a.X0, b.X0),
		Y0: max(a.Y0, b.Y0),
		X1: min(a.X1, b.X1),
		Y1: min(a.Y1, b.Y1),
	}
	if result.IsEmpty() {
		return model.BBox{}, false
	}
	return result, true
}

// Union returns the smallest bounding box containing both a and b.
func Union(a, b model.BBox) model.BBox {
	return model.BBox{
		X0: min(a.X0, b.X0),
		Y0: min(a.Y0, b.Y0),
		X1: max(a.X1, b.X1),
		Y1: max(a.Y1, b.Y1),
	}
}

// UnionAll returns the smallest bounding box containing all given boxes.
// Returns an empty BBox if the slice is empty.
func UnionAll(boxes []model.BBox) model.BBox {
	if len(boxes) == 0 {
		return model.BBox{}
	}
	result := boxes[0]
	for _, b := range boxes[1:] {
		result = Union(result, b)
	}
	return result
}

// Normalize ensures X0 <= X1 and Y0 <= Y1.
func Normalize(b model.BBox) model.BBox {
	if b.X0 > b.X1 {
		b.X0, b.X1 = b.X1, b.X0
	}
	if b.Y0 > b.Y1 {
		b.Y0, b.Y1 = b.Y1, b.Y0
	}
	return b
}

// Expand grows a bbox by the given margin on all sides.
func Expand(b model.BBox, margin float64) model.BBox {
	return model.BBox{
		X0: b.X0 - margin,
		Y0: b.Y0 - margin,
		X1: b.X1 + margin,
		Y1: b.Y1 + margin,
	}
}
