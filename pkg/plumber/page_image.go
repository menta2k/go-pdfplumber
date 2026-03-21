package plumber

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

// PageImageOptions controls page image rendering.
type PageImageOptions struct {
	// Resolution in DPI. Default: 72 (1 point = 1 pixel).
	Resolution float64

	// BackgroundColor for the canvas. Default: white.
	BackgroundColor color.Color
}

// DefaultPageImageOptions returns sensible defaults.
func DefaultPageImageOptions() PageImageOptions {
	return PageImageOptions{
		Resolution:      72,
		BackgroundColor: color.White,
	}
}

// PageImage is a visual debugging canvas for a PDF page.
// Drawing methods return new PageImage instances (immutable).
type PageImage struct {
	page   *Page
	img    *image.RGBA
	scale  float64 // pixels per point
	width  int     // image width in pixels
	height int     // image height in pixels
}

// NewPageImage creates a blank canvas sized to the page dimensions.
func NewPageImage(page *Page, opts ...PageImageOptions) *PageImage {
	o := DefaultPageImageOptions()
	if len(opts) > 0 {
		o = withPageImageDefaults(opts[0])
	}

	scale := o.Resolution / 72.0
	w := int(math.Ceil(page.Width() * scale))
	h := int(math.Ceil(page.Height() * scale))

	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Fill background
	bg := o.BackgroundColor
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bg)
		}
	}

	return &PageImage{
		page:   page,
		img:    img,
		scale:  scale,
		width:  w,
		height: h,
	}
}

// DrawRect draws a rectangle overlay.
func (pi *PageImage) DrawRect(bbox BBox, style DrawStyle) *PageImage {
	out := pi.clone()

	x0, y0 := out.toPixel(bbox.X0, bbox.Y1) // top-left in image coords
	x1, y1 := out.toPixel(bbox.X1, bbox.Y0) // bottom-right in image coords

	if style.FillColor != nil {
		fillRect(out.img, x0, y0, x1, y1, style.FillColor)
	}
	if style.StrokeColor != nil {
		strokeWidth := max(1, int(style.StrokeWidth*out.scale))
		strokeRect(out.img, x0, y0, x1, y1, style.StrokeColor, strokeWidth)
	}

	return out
}

// DrawLine draws a line overlay.
func (pi *PageImage) DrawLine(lx0, ly0, lx1, ly1 float64, style DrawStyle) *PageImage {
	out := pi.clone()

	px0, py0 := out.toPixel(lx0, ly0)
	px1, py1 := out.toPixel(lx1, ly1)

	c := style.StrokeColor
	if c == nil {
		c = color.RGBA{R: 255, A: 255}
	}

	drawLine(out.img, px0, py0, px1, py1, c)
	return out
}

// DrawCircle draws a circle overlay at a point.
func (pi *PageImage) DrawCircle(center Point, radius float64, style DrawStyle) *PageImage {
	out := pi.clone()

	cx, cy := out.toPixel(center.X, center.Y)
	r := int(radius * out.scale)

	if style.FillColor != nil {
		fillCircle(out.img, cx, cy, r, style.FillColor)
	}
	if style.StrokeColor != nil {
		strokeCircle(out.img, cx, cy, r, style.StrokeColor)
	}

	return out
}

// DrawChars draws bounding boxes around all characters.
func (pi *PageImage) DrawChars(chars []Char, style DrawStyle) *PageImage {
	out := pi.clone()
	for _, c := range chars {
		x0, y0 := out.toPixel(c.BBox.X0, c.BBox.Y1)
		x1, y1 := out.toPixel(c.BBox.X1, c.BBox.Y0)
		if style.StrokeColor != nil {
			strokeRect(out.img, x0, y0, x1, y1, style.StrokeColor, 1)
		}
	}
	return out
}

// DrawWords draws bounding boxes around all words.
func (pi *PageImage) DrawWords(words []Word, style DrawStyle) *PageImage {
	out := pi.clone()
	for _, w := range words {
		x0, y0 := out.toPixel(w.BBox.X0, w.BBox.Y1)
		x1, y1 := out.toPixel(w.BBox.X1, w.BBox.Y0)
		if style.FillColor != nil {
			fillRect(out.img, x0, y0, x1, y1, style.FillColor)
		}
		if style.StrokeColor != nil {
			sw := max(1, int(style.StrokeWidth*out.scale))
			strokeRect(out.img, x0, y0, x1, y1, style.StrokeColor, sw)
		}
	}
	return out
}

// DrawTable draws the table grid (cell borders and optional fill).
func (pi *PageImage) DrawTable(table Table, style DrawStyle) *PageImage {
	out := pi.clone()
	for _, row := range table.Cells {
		for _, cell := range row {
			x0, y0 := out.toPixel(cell.BBox.X0, cell.BBox.Y1)
			x1, y1 := out.toPixel(cell.BBox.X1, cell.BBox.Y0)
			if style.FillColor != nil {
				fillRect(out.img, x0, y0, x1, y1, style.FillColor)
			}
			if style.StrokeColor != nil {
				sw := max(1, int(style.StrokeWidth*out.scale))
				strokeRect(out.img, x0, y0, x1, y1, style.StrokeColor, sw)
			}
		}
	}
	return out
}

// DrawPageRects draws the page's detected rectangles.
func (pi *PageImage) DrawPageRects(style DrawStyle) *PageImage {
	out := pi.clone()
	for _, r := range pi.page.Rects() {
		x0, y0 := out.toPixel(r.BBox.X0, r.BBox.Y1)
		x1, y1 := out.toPixel(r.BBox.X1, r.BBox.Y0)
		if style.StrokeColor != nil {
			sw := max(1, int(style.StrokeWidth*out.scale))
			strokeRect(out.img, x0, y0, x1, y1, style.StrokeColor, sw)
		}
	}
	return out
}

// DrawPageLines draws the page's detected line segments.
func (pi *PageImage) DrawPageLines(style DrawStyle) *PageImage {
	out := pi.clone()
	c := style.StrokeColor
	if c == nil {
		c = color.RGBA{R: 255, A: 255}
	}
	for _, l := range pi.page.Lines() {
		px0, py0 := out.toPixel(l.X0, l.Y0)
		px1, py1 := out.toPixel(l.X1, l.Y1)
		drawLine(out.img, px0, py0, px1, py1, c)
	}
	return out
}

// ToImage returns the underlying RGBA image.
func (pi *PageImage) ToImage() *image.RGBA {
	// Return a copy
	bounds := pi.img.Bounds()
	cp := image.NewRGBA(bounds)
	copy(cp.Pix, pi.img.Pix)
	return cp
}

// Save writes the image to a PNG file.
func (pi *PageImage) Save(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, pi.img)
}

// --- coordinate conversion ---

// toPixel converts PDF coordinates to image pixel coordinates.
// PDF: origin bottom-left, Y up. Image: origin top-left, Y down.
func (pi *PageImage) toPixel(pdfX, pdfY float64) (int, int) {
	px := int(pdfX * pi.scale)
	py := pi.height - int(pdfY*pi.scale)
	return px, py
}

// clone creates a deep copy of the PageImage.
func (pi *PageImage) clone() *PageImage {
	bounds := pi.img.Bounds()
	cp := image.NewRGBA(bounds)
	copy(cp.Pix, pi.img.Pix)
	return &PageImage{
		page:   pi.page,
		img:    cp,
		scale:  pi.scale,
		width:  pi.width,
		height: pi.height,
	}
}

func withPageImageDefaults(o PageImageOptions) PageImageOptions {
	if o.Resolution <= 0 {
		o.Resolution = 72
	}
	if o.BackgroundColor == nil {
		o.BackgroundColor = color.White
	}
	return o
}

// --- drawing primitives ---

func fillRect(img *image.RGBA, x0, y0, x1, y1 int, c color.Color) {
	bounds := img.Bounds()
	x0 = clamp(x0, bounds.Min.X, bounds.Max.X-1)
	x1 = clamp(x1, bounds.Min.X, bounds.Max.X-1)
	y0 = clamp(y0, bounds.Min.Y, bounds.Max.Y-1)
	y1 = clamp(y1, bounds.Min.Y, bounds.Max.Y-1)

	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			blendPixel(img, x, y, c)
		}
	}
}

func strokeRect(img *image.RGBA, x0, y0, x1, y1 int, c color.Color, width int) {
	for w := 0; w < width; w++ {
		// Top and bottom
		for x := x0; x <= x1; x++ {
			safeSet(img, x, y0+w, c)
			safeSet(img, x, y1-w, c)
		}
		// Left and right
		for y := y0; y <= y1; y++ {
			safeSet(img, x0+w, y, c)
			safeSet(img, x1-w, y, c)
		}
	}
}

// drawLine uses Bresenham's algorithm.
func drawLine(img *image.RGBA, x0, y0, x1, y1 int, c color.Color) {
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx, sy := 1, 1
	if x0 > x1 {
		sx = -1
	}
	if y0 > y1 {
		sy = -1
	}
	err := dx - dy

	for {
		safeSet(img, x0, y0, c)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

func fillCircle(img *image.RGBA, cx, cy, r int, c color.Color) {
	for y := -r; y <= r; y++ {
		for x := -r; x <= r; x++ {
			if x*x+y*y <= r*r {
				safeSet(img, cx+x, cy+y, c)
			}
		}
	}
}

func strokeCircle(img *image.RGBA, cx, cy, r int, c color.Color) {
	// Midpoint circle algorithm
	x, y := r, 0
	err := 1 - r

	for x >= y {
		safeSet(img, cx+x, cy+y, c)
		safeSet(img, cx+y, cy+x, c)
		safeSet(img, cx-y, cy+x, c)
		safeSet(img, cx-x, cy+y, c)
		safeSet(img, cx-x, cy-y, c)
		safeSet(img, cx-y, cy-x, c)
		safeSet(img, cx+y, cy-x, c)
		safeSet(img, cx+x, cy-y, c)
		y++
		if err < 0 {
			err += 2*y + 1
		} else {
			x--
			err += 2*(y-x) + 1
		}
	}
}

func blendPixel(img *image.RGBA, x, y int, c color.Color) {
	bounds := img.Bounds()
	if x < bounds.Min.X || x >= bounds.Max.X || y < bounds.Min.Y || y >= bounds.Max.Y {
		return
	}

	r, g, b, a := c.RGBA()
	if a == 0xFFFF {
		img.Set(x, y, c)
		return
	}

	// Alpha blend
	bg := img.RGBAAt(x, y)
	alpha := float64(a) / 0xFFFF
	invA := 1 - alpha

	img.SetRGBA(x, y, color.RGBA{
		R: uint8(float64(r>>8)*alpha + float64(bg.R)*invA),
		G: uint8(float64(g>>8)*alpha + float64(bg.G)*invA),
		B: uint8(float64(b>>8)*alpha + float64(bg.B)*invA),
		A: 255,
	})
}

func safeSet(img *image.RGBA, x, y int, c color.Color) {
	bounds := img.Bounds()
	if x >= bounds.Min.X && x < bounds.Max.X && y >= bounds.Min.Y && y < bounds.Max.Y {
		img.Set(x, y, c)
	}
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
