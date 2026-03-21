package plumber

import (
	"fmt"

	"github.com/menta2k/go-pdfplumber/internal/geometry"
)

// Crop returns a new Page containing only objects within the given bounding box.
// Characters are included if their midpoint falls within the crop region.
// Lines and rects are included if they overlap with the crop region.
// The returned page has dimensions matching the crop bbox.
func (p *Page) Crop(bbox BBox) (*Page, error) {
	if bbox.IsEmpty() {
		return nil, fmt.Errorf("%w: crop box has zero or negative area", ErrInvalidBBox)
	}

	p.ensureContent()

	chars := geometry.FilterCharsByMidpoint(p.content.chars, bbox)
	lines := geometry.FilterOverlapping(p.content.lines, bbox)
	rects := geometry.FilterOverlapping(p.content.rects, bbox)

	return newFilteredPage(
		p.number,
		bbox.Width(),
		bbox.Height(),
		bbox,
		chars,
		lines,
		rects,
	), nil
}

// WithinBBox returns a new Page containing only objects fully inside the bbox.
func (p *Page) WithinBBox(bbox BBox) *Page {
	p.ensureContent()

	chars := geometry.FilterWithin(p.content.chars, bbox)
	lines := geometry.FilterWithin(p.content.lines, bbox)
	rects := geometry.FilterWithin(p.content.rects, bbox)

	return newFilteredPage(p.number, p.width, p.height, p.bbox, chars, lines, rects)
}

// OutsideBBox returns a new Page containing only objects fully outside the bbox.
func (p *Page) OutsideBBox(bbox BBox) *Page {
	p.ensureContent()

	chars := geometry.FilterOutside(p.content.chars, bbox)
	lines := geometry.FilterOutside(p.content.lines, bbox)
	rects := geometry.FilterOutside(p.content.rects, bbox)

	return newFilteredPage(p.number, p.width, p.height, p.bbox, chars, lines, rects)
}

// FilterChars returns a new Page containing only chars that pass the predicate.
// Lines and rects are preserved unchanged.
func (p *Page) FilterChars(fn func(Char) bool) *Page {
	p.ensureContent()

	chars := geometry.FilterFunc(p.content.chars, fn)

	return newFilteredPage(p.number, p.width, p.height, p.bbox, chars, p.content.lines, p.content.rects)
}
