package plumber

import "errors"

var (
	// ErrClosed is returned when operating on a closed document.
	ErrClosed = errors.New("plumber: document is closed")

	// ErrPageOutOfRange is returned when a page number is outside [1, NumPages].
	ErrPageOutOfRange = errors.New("plumber: page number out of range")

	// ErrInvalidBBox is returned when a bounding box has zero or negative area.
	ErrInvalidBBox = errors.New("plumber: invalid bounding box")

	// ErrEmptyPage is returned when a page has no content.
	ErrEmptyPage = errors.New("plumber: page has no content")
)
