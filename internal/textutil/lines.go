package textutil

import (
	"strings"

	"github.com/menta2k/go-pdfplumber/internal/geometry"
	"github.com/menta2k/go-pdfplumber/pkg/model"
)

// GroupIntoLines groups words into text lines. Words on the same baseline
// (within yTolerance) are combined into a single TextLine, ordered left-to-right.
func GroupIntoLines(words []model.Word, yTolerance float64) []model.TextLine {
	if len(words) == 0 {
		return nil
	}

	type lineGroup struct {
		baselineY float64
		words     []model.Word
	}

	var groups []lineGroup

	for _, w := range words {
		baseY := w.BBox.Y0
		found := false
		for i := range groups {
			if nearlyEqual(groups[i].baselineY, baseY, yTolerance) {
				groups[i].words = append(groups[i].words, w)
				found = true
				break
			}
		}
		if !found {
			groups = append(groups, lineGroup{
				baselineY: baseY,
				words:     []model.Word{w},
			})
		}
	}

	// Sort groups top-to-bottom (higher Y first in PDF coords)
	for i := 0; i < len(groups)-1; i++ {
		for j := i + 1; j < len(groups); j++ {
			if groups[j].baselineY > groups[i].baselineY {
				groups[i], groups[j] = groups[j], groups[i]
			}
		}
	}

	lines := make([]model.TextLine, 0, len(groups))
	for _, g := range groups {
		sortWordsLTR(g.words)
		lines = append(lines, buildTextLine(g.words))
	}

	return lines
}

func sortWordsLTR(words []model.Word) {
	for i := 0; i < len(words)-1; i++ {
		for j := i + 1; j < len(words); j++ {
			if words[j].BBox.X0 < words[i].BBox.X0 {
				words[i], words[j] = words[j], words[i]
			}
		}
	}
}

func buildTextLine(words []model.Word) model.TextLine {
	var sb strings.Builder
	boxes := make([]model.BBox, len(words))
	for i, w := range words {
		if i > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(w.Text)
		boxes[i] = w.BBox
	}

	lineWords := make([]model.Word, len(words))
	copy(lineWords, words)

	return model.TextLine{
		Text:  sb.String(),
		BBox:  geometry.UnionAll(boxes),
		Words: lineWords,
	}
}
