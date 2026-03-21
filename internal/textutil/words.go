package textutil

import (
	"strings"

	"github.com/menta2k/go-pdfplumber/internal/geometry"
	"github.com/menta2k/go-pdfplumber/pkg/model"
)

// GroupIntoWords groups sorted characters into words based on spatial proximity.
func GroupIntoWords(chars []model.Char, opts model.TextOptions) []model.Word {
	if len(chars) == 0 {
		return nil
	}

	sorted := SortChars(chars)

	if !opts.KeepBlankChars {
		sorted = FilterBlankChars(sorted)
	}

	if len(sorted) == 0 {
		return nil
	}

	var words []model.Word
	currentChars := []model.Char{sorted[0]}

	for i := 1; i < len(sorted); i++ {
		prev := currentChars[len(currentChars)-1]
		curr := sorted[i]

		sameLine := nearlyEqual(prev.Y, curr.Y, opts.YTolerance)
		gap := curr.X - (prev.X + prev.Width)
		closeEnough := gap <= opts.XTolerance

		attrsMatch := charsMatchAttrs(prev, curr, opts.ExtraAttrs)

		splitPunct := opts.SplitAtPunctuation &&
			(isPunctuation(prev.Text) != isPunctuation(curr.Text))

		if sameLine && closeEnough && attrsMatch && !splitPunct {
			currentChars = append(currentChars, curr)
		} else {
			words = append(words, buildWord(currentChars))
			currentChars = []model.Char{curr}
		}
	}

	if len(currentChars) > 0 {
		words = append(words, buildWord(currentChars))
	}

	return words
}

func buildWord(chars []model.Char) model.Word {
	var sb strings.Builder
	boxes := make([]model.BBox, len(chars))
	for i, c := range chars {
		sb.WriteString(c.Text)
		boxes[i] = c.BBox
	}

	wordChars := make([]model.Char, len(chars))
	copy(wordChars, chars)

	return model.Word{
		Text:  sb.String(),
		BBox:  geometry.UnionAll(boxes),
		Chars: wordChars,
	}
}
