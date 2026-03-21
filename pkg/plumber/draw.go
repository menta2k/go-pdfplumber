package plumber

import (
	"image/color"
)

// DrawStyle controls the appearance of drawn overlays.
type DrawStyle struct {
	StrokeColor color.Color
	FillColor   color.Color
	StrokeWidth float64
}

// DefaultDrawStyle returns a visible red stroke with no fill.
func DefaultDrawStyle() DrawStyle {
	return DrawStyle{
		StrokeColor: color.RGBA{R: 255, A: 255},
		StrokeWidth: 1.0,
	}
}

// Predefined styles for common debugging use.
var (
	StyleCharBBox = DrawStyle{
		StrokeColor: color.RGBA{R: 0, G: 150, B: 255, A: 180},
		StrokeWidth: 0.5,
	}
	StyleWordBBox = DrawStyle{
		StrokeColor: color.RGBA{R: 0, G: 200, B: 0, A: 200},
		StrokeWidth: 1.0,
	}
	StyleTableEdge = DrawStyle{
		StrokeColor: color.RGBA{R: 255, G: 0, B: 0, A: 255},
		StrokeWidth: 1.5,
	}
	StyleTableCell = DrawStyle{
		StrokeColor: color.RGBA{R: 0, G: 0, B: 255, A: 150},
		FillColor:   color.RGBA{R: 173, G: 216, B: 230, A: 50},
		StrokeWidth: 1.0,
	}
	StyleIntersection = DrawStyle{
		StrokeColor: color.RGBA{R: 255, G: 0, B: 0, A: 255},
		FillColor:   color.RGBA{R: 255, G: 0, B: 0, A: 200},
		StrokeWidth: 1.0,
	}
)
