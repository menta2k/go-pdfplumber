package plumber

import (
	"github.com/menta2k/go-pdfplumber/internal/textutil"
)

// ExtractWords groups characters into words using spatial proximity.
// Uses the provided options or DefaultTextOptions if none given.
func (p *Page) ExtractWords(opts ...TextOptions) []Word {
	o := resolveOpts(opts)
	return textutil.GroupIntoWords(p.Chars(), o)
}

// ExtractTextLines groups characters into lines of text.
// Returns lines sorted top-to-bottom with words left-to-right.
func (p *Page) ExtractTextLines(opts ...TextOptions) []TextLine {
	o := resolveOpts(opts)
	words := textutil.GroupIntoWords(p.Chars(), o)
	return textutil.GroupIntoLines(words, o.YTolerance)
}

// ExtractText extracts all text from the page as a single string.
// Lines are separated by newlines, with paragraph breaks as double newlines.
func (p *Page) ExtractText(opts ...TextOptions) string {
	o := resolveOpts(opts)
	words := textutil.GroupIntoWords(p.Chars(), o)
	lines := textutil.GroupIntoLines(words, o.YTolerance)
	return textutil.AssembleText(lines)
}

// Search finds all occurrences of a text pattern on the page.
// Returns the matching words. For simple substring matching, not regex.
func (p *Page) Search(query string, opts ...TextOptions) []Word {
	o := resolveOpts(opts)
	words := textutil.GroupIntoWords(p.Chars(), o)
	return textutil.SearchWords(words, query)
}

func resolveOpts(opts []TextOptions) TextOptions {
	if len(opts) > 0 {
		return withDefaults(opts[0])
	}
	return DefaultTextOptions()
}

func withDefaults(o TextOptions) TextOptions {
	if o.XTolerance <= 0 {
		o.XTolerance = 3.0
	}
	if o.YTolerance <= 0 {
		o.YTolerance = 3.0
	}
	if o.XDensity <= 0 {
		o.XDensity = 1.0
	}
	if o.YDensity <= 0 {
		o.YDensity = 1.0
	}
	return o
}
