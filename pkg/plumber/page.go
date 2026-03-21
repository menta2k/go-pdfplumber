package plumber

import (
	"math"
	"sync"

	"github.com/digitorus/pdf"
	"github.com/menta2k/go-pdfplumber/internal/extract"
)

const lineThickness = 1.0 // rects thinner than this are classified as lines

// Page represents a single PDF page with access to its objects.
type Page struct {
	number int
	width  float64
	height float64
	bbox   BBox

	pdfPage *pdf.Page // nil for filtered pages
	mu      sync.Once
	content *pageContent
}

type pageContent struct {
	chars []Char
	lines []LineSegment
	rects []RectObject
}

// newPageFromPDF creates a Page from a digitorus/pdf Page.
func newPageFromPDF(pdfPage pdf.Page, number int) *Page {
	width, height := extractPageDimensions(pdfPage)
	return &Page{
		number:  number,
		width:   width,
		height:  height,
		bbox:    BBox{X0: 0, Y0: 0, X1: width, Y1: height},
		pdfPage: &pdfPage,
	}
}

// newFilteredPage creates a Page from pre-computed objects (used by Crop/Filter).
func newFilteredPage(number int, width, height float64, bbox BBox, chars []Char, lines []LineSegment, rects []RectObject) *Page {
	p := &Page{
		number: number,
		width:  width,
		height: height,
		bbox:   bbox,
	}
	p.content = &pageContent{
		chars: chars,
		lines: lines,
		rects: rects,
	}
	p.mu.Do(func() {}) // mark as loaded
	return p
}

// Number returns the 1-based page number.
func (p *Page) Number() int { return p.number }

// Width returns the page width in points.
func (p *Page) Width() float64 { return p.width }

// Height returns the page height in points.
func (p *Page) Height() float64 { return p.height }

// PageBBox returns the page bounding box.
func (p *Page) PageBBox() BBox { return p.bbox }

// Chars returns all characters on the page.
func (p *Page) Chars() []Char {
	p.ensureContent()
	result := make([]Char, len(p.content.chars))
	copy(result, p.content.chars)
	return result
}

// Lines returns all line segments on the page.
func (p *Page) Lines() []LineSegment {
	p.ensureContent()
	result := make([]LineSegment, len(p.content.lines))
	copy(result, p.content.lines)
	return result
}

// Rects returns all rectangles on the page.
func (p *Page) Rects() []RectObject {
	p.ensureContent()
	result := make([]RectObject, len(p.content.rects))
	copy(result, p.content.rects)
	return result
}

// Objects returns a map of all object types on the page.
func (p *Page) Objects() map[string][]BBoxer {
	p.ensureContent()
	objects := make(map[string][]BBoxer)

	chars := make([]BBoxer, len(p.content.chars))
	for i, c := range p.content.chars {
		chars[i] = c
	}
	objects["chars"] = chars

	lines := make([]BBoxer, len(p.content.lines))
	for i, l := range p.content.lines {
		lines[i] = l
	}
	objects["lines"] = lines

	rects := make([]BBoxer, len(p.content.rects))
	for i, r := range p.content.rects {
		rects[i] = r
	}
	objects["rects"] = rects

	return objects
}

func (p *Page) ensureContent() {
	p.mu.Do(func() {
		if p.pdfPage == nil {
			p.content = &pageContent{}
			return
		}
		p.content = extractPageContent(*p.pdfPage)
	})
}

func extractPageContent(pdfPage pdf.Page) *pageContent {
	content := extract.ExtractContent(pdfPage)
	pc := &pageContent{}

	// Convert extracted chars
	pc.chars = make([]Char, 0, len(content.Chars))
	for _, c := range content.Chars {
		if c.Text == "" {
			continue
		}
		charHeight := c.FontSize
		if charHeight <= 0 {
			charHeight = 10
		}

		bbox := BBox{
			X0: c.X,
			Y0: c.Y,
			X1: c.X + c.W,
			Y1: c.Y + charHeight,
		}

		pc.chars = append(pc.chars, Char{
			Text:     c.Text,
			FontName: c.FontName,
			FontSize: c.FontSize,
			BBox:     bbox,
			X:        c.X,
			Y:        c.Y,
			Width:    c.W,
			Top:      c.Y + charHeight,
			Bottom:   c.Y,
		})
	}

	// Convert extracted rectangles - classify thin ones as lines
	for _, r := range content.Rects {
		x0 := math.Min(r.X, r.X+r.W)
		y0 := math.Min(r.Y, r.Y+r.H)
		x1 := math.Max(r.X, r.X+r.W)
		y1 := math.Max(r.Y, r.Y+r.H)

		rectBBox := BBox{X0: x0, Y0: y0, X1: x1, Y1: y1}
		w := rectBBox.Width()
		h := rectBBox.Height()

		switch {
		case w < lineThickness && h >= lineThickness:
			midX := (x0 + x1) / 2
			pc.lines = append(pc.lines, LineSegment{
				BBox: rectBBox, X0: midX, Y0: y0, X1: midX, Y1: y1,
				Orientation: "vertical",
			})
		case h < lineThickness && w >= lineThickness:
			midY := (y0 + y1) / 2
			pc.lines = append(pc.lines, LineSegment{
				BBox: rectBBox, X0: x0, Y0: midY, X1: x1, Y1: midY,
				Orientation: "horizontal",
			})
		default:
			pc.rects = append(pc.rects, RectObject{BBox: rectBBox, Stroke: true})
		}
	}

	// Convert extracted line operations
	for _, l := range content.Lines {
		x0 := math.Min(l.X0, l.X1)
		y0 := math.Min(l.Y0, l.Y1)
		x1 := math.Max(l.X0, l.X1)
		y1 := math.Max(l.Y0, l.Y1)

		lineBBox := BBox{X0: x0, Y0: y0, X1: x1, Y1: y1}

		orientation := "diagonal"
		if math.Abs(l.Y0-l.Y1) < 0.5 {
			orientation = "horizontal"
		} else if math.Abs(l.X0-l.X1) < 0.5 {
			orientation = "vertical"
		}

		pc.lines = append(pc.lines, LineSegment{
			BBox: lineBBox, X0: l.X0, Y0: l.Y0, X1: l.X1, Y1: l.Y1,
			Orientation: orientation,
		})
	}

	return pc
}

func extractPageDimensions(pdfPage pdf.Page) (width, height float64) {
	for _, boxName := range []string{"CropBox", "MediaBox"} {
		box := pdfPage.V.Key(boxName)
		if box.Len() == 4 {
			x0 := box.Index(0).Float64()
			y0 := box.Index(1).Float64()
			x1 := box.Index(2).Float64()
			y1 := box.Index(3).Float64()
			w := math.Abs(x1 - x0)
			h := math.Abs(y1 - y0)
			if w > 0 && h > 0 {
				return w, h
			}
		}
	}
	return 612, 792
}
