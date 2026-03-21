package geometry

import "github.com/menta2k/go-pdfplumber/pkg/model"

// FilterWithin returns objects whose bounding box is fully inside the region.
func FilterWithin[T model.BBoxer](objects []T, region model.BBox) []T {
	var result []T
	for _, obj := range objects {
		if region.ContainsBBox(obj.GetBBox()) {
			result = append(result, obj)
		}
	}
	return result
}

// FilterOutside returns objects whose bounding box is fully outside the region.
func FilterOutside[T model.BBoxer](objects []T, region model.BBox) []T {
	var result []T
	for _, obj := range objects {
		if !obj.GetBBox().Overlaps(region) {
			result = append(result, obj)
		}
	}
	return result
}

// FilterOverlapping returns objects whose bounding box overlaps the region.
func FilterOverlapping[T model.BBoxer](objects []T, region model.BBox) []T {
	var result []T
	for _, obj := range objects {
		if obj.GetBBox().Overlaps(region) {
			result = append(result, obj)
		}
	}
	return result
}

// FilterFunc returns objects for which the predicate returns true.
func FilterFunc[T any](objects []T, fn func(T) bool) []T {
	var result []T
	for _, obj := range objects {
		if fn(obj) {
			result = append(result, obj)
		}
	}
	return result
}

// FilterCharsByMidpoint returns chars whose midpoint is inside the region.
// This is the strategy used for Crop — include a char if its center is within bounds.
func FilterCharsByMidpoint(chars []model.Char, region model.BBox) []model.Char {
	var result []model.Char
	for _, c := range chars {
		mid := c.BBox.Midpoint()
		if region.Contains(mid) {
			result = append(result, c)
		}
	}
	return result
}
