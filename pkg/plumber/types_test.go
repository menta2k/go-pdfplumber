package plumber

import "testing"

func TestBBoxMethods(t *testing.T) {
	b := BBox{X0: 10, Y0: 20, X1: 50, Y1: 60}

	if w := b.Width(); w != 40 {
		t.Errorf("Width = %v, want 40", w)
	}
	if h := b.Height(); h != 40 {
		t.Errorf("Height = %v, want 40", h)
	}
	if a := b.Area(); a != 1600 {
		t.Errorf("Area = %v, want 1600", a)
	}

	mid := b.Midpoint()
	if mid.X != 30 || mid.Y != 40 {
		t.Errorf("Midpoint = %v, want {30 40}", mid)
	}
}

func TestBBoxContains(t *testing.T) {
	b := BBox{X0: 0, Y0: 0, X1: 10, Y1: 10}

	if !b.Contains(Point{5, 5}) {
		t.Error("expected center point to be contained")
	}
	if !b.Contains(Point{0, 0}) {
		t.Error("expected corner point to be contained (inclusive)")
	}
	if b.Contains(Point{11, 5}) {
		t.Error("expected outside point to not be contained")
	}
}

func TestBBoxContainsBBox(t *testing.T) {
	outer := BBox{X0: 0, Y0: 0, X1: 20, Y1: 20}
	inner := BBox{X0: 5, Y0: 5, X1: 15, Y1: 15}
	partial := BBox{X0: 15, Y0: 15, X1: 25, Y1: 25}

	if !outer.ContainsBBox(inner) {
		t.Error("expected inner to be contained")
	}
	if outer.ContainsBBox(partial) {
		t.Error("expected partial overlap to not be contained")
	}
}

func TestBBoxOverlaps(t *testing.T) {
	a := BBox{X0: 0, Y0: 0, X1: 10, Y1: 10}

	if !a.Overlaps(BBox{X0: 5, Y0: 5, X1: 15, Y1: 15}) {
		t.Error("expected overlap")
	}
	if a.Overlaps(BBox{X0: 20, Y0: 20, X1: 30, Y1: 30}) {
		t.Error("expected no overlap")
	}
}

func TestBBoxIsEmpty(t *testing.T) {
	if !(BBox{X0: 5, Y0: 5, X1: 5, Y1: 10}).IsEmpty() {
		t.Error("zero-width bbox should be empty")
	}
	if !(BBox{X0: 5, Y0: 10, X1: 10, Y1: 5}).IsEmpty() {
		t.Error("negative-height bbox should be empty")
	}
	if (BBox{X0: 0, Y0: 0, X1: 10, Y1: 10}).IsEmpty() {
		t.Error("valid bbox should not be empty")
	}
}

func TestBBoxer(t *testing.T) {
	// Verify all types implement BBoxer
	var _ BBoxer = Char{}
	var _ BBoxer = LineSegment{}
	var _ BBoxer = RectObject{}
	var _ BBoxer = Curve{}
	var _ BBoxer = Word{}
	var _ BBoxer = TextLine{}
}
