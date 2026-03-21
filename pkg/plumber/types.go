// Package plumber provides a high-level API for extracting text, tables,
// and objects from PDF documents, similar to Python's pdfplumber.
package plumber

import "github.com/menta2k/go-pdfplumber/pkg/model"

// Re-export core types from model package for convenience.
type (
	Point       = model.Point
	BBox        = model.BBox
	BBoxer      = model.BBoxer
	Char        = model.Char
	LineSegment = model.LineSegment
	RectObject  = model.RectObject
	Curve       = model.Curve
	Word        = model.Word
	TextLine    = model.TextLine
	TextOptions = model.TextOptions
)

// DefaultTextOptions returns TextOptions with sensible defaults.
var DefaultTextOptions = model.DefaultTextOptions
