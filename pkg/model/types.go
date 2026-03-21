// Package model provides core value types used across go-pdfplumber.
package model

// Point represents a 2D coordinate in PDF space (origin bottom-left, Y up).
type Point struct {
	X float64
	Y float64
}

// BBox represents an axis-aligned bounding box in PDF coordinates.
// X0,Y0 is the bottom-left corner; X1,Y1 is the top-right corner.
type BBox struct {
	X0 float64
	Y0 float64
	X1 float64
	Y1 float64
}

// Width returns the horizontal extent.
func (b BBox) Width() float64 { return b.X1 - b.X0 }

// Height returns the vertical extent.
func (b BBox) Height() float64 { return b.Y1 - b.Y0 }

// Area returns the area of the bounding box.
func (b BBox) Area() float64 { return b.Width() * b.Height() }

// Contains reports whether point p lies within b (inclusive).
func (b BBox) Contains(p Point) bool {
	return p.X >= b.X0 && p.X <= b.X1 && p.Y >= b.Y0 && p.Y <= b.Y1
}

// ContainsBBox reports whether inner is fully contained within b.
func (b BBox) ContainsBBox(inner BBox) bool {
	return inner.X0 >= b.X0 && inner.Y0 >= b.Y0 &&
		inner.X1 <= b.X1 && inner.Y1 <= b.Y1
}

// Overlaps reports whether b and other share any area.
func (b BBox) Overlaps(other BBox) bool {
	return b.X0 < other.X1 && b.X1 > other.X0 &&
		b.Y0 < other.Y1 && b.Y1 > other.Y0
}

// Midpoint returns the center point.
func (b BBox) Midpoint() Point {
	return Point{X: (b.X0 + b.X1) / 2, Y: (b.Y0 + b.Y1) / 2}
}

// IsEmpty reports whether the bbox has zero or negative area.
func (b BBox) IsEmpty() bool {
	return b.X0 >= b.X1 || b.Y0 >= b.Y1
}

// BBoxer is implemented by any value that has a bounding box.
type BBoxer interface {
	GetBBox() BBox
}

// Char represents a single character extracted from a PDF page.
type Char struct {
	Text     string  // the character(s) as UTF-8
	FontName string  // name of the font
	FontSize float64 // font size in points
	BBox     BBox    // bounding box in page coordinates
	X        float64 // left edge x-coordinate
	Y        float64 // baseline y-coordinate (PDF: bottom-up)
	Width    float64 // advance width
	Top      float64 // top edge (Y1 in PDF coords)
	Bottom   float64 // bottom edge (Y0 in PDF coords)
}

// GetBBox implements BBoxer.
func (c Char) GetBBox() BBox { return c.BBox }

// LineSegment represents a line segment on the page (from explicit PDF drawing).
type LineSegment struct {
	BBox        BBox
	X0, Y0      float64 // start point
	X1, Y1      float64 // end point
	Orientation string  // "horizontal", "vertical", or "diagonal"
}

// GetBBox implements BBoxer.
func (l LineSegment) GetBBox() BBox { return l.BBox }

// RectObject represents a rectangle drawn on the page.
type RectObject struct {
	BBox   BBox
	Stroke bool
	Fill   bool
}

// GetBBox implements BBoxer.
func (r RectObject) GetBBox() BBox { return r.BBox }

// Curve represents a curved path on the page.
type Curve struct {
	Points []Point
	BBox   BBox
}

// GetBBox implements BBoxer.
func (c Curve) GetBBox() BBox { return c.BBox }

// Word represents a group of characters forming a word.
type Word struct {
	Text  string
	BBox  BBox
	Chars []Char
}

// GetBBox implements BBoxer.
func (w Word) GetBBox() BBox { return w.BBox }

// TextLine represents a line of text (group of words on the same baseline).
type TextLine struct {
	Text  string
	BBox  BBox
	Words []Word
}

// GetBBox implements BBoxer.
func (t TextLine) GetBBox() BBox { return t.BBox }

// TextOptions controls text extraction behavior.
type TextOptions struct {
	XTolerance         float64
	YTolerance         float64
	KeepBlankChars     bool
	UseTextFlow        bool
	SplitAtPunctuation bool
	ExtraAttrs         []string
	Layout             bool
	XDensity           float64
	YDensity           float64
}

// DefaultTextOptions returns TextOptions with sensible defaults.
func DefaultTextOptions() TextOptions {
	return TextOptions{
		XTolerance: 3.0,
		YTolerance: 3.0,
		XDensity:   1.0,
		YDensity:   1.0,
	}
}
