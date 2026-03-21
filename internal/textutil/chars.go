// Package textutil provides algorithms for grouping PDF characters
// into words, lines, and full-page text.
package textutil

import (
	"sort"
	"strings"
	"unicode"

	"github.com/menta2k/go-pdfplumber/pkg/model"
)

// SortChars returns a new slice of chars sorted top-to-bottom (descending Y),
// then left-to-right (ascending X). Does not mutate the input.
func SortChars(chars []model.Char) []model.Char {
	sorted := make([]model.Char, len(chars))
	copy(sorted, chars)
	sort.Slice(sorted, func(i, j int) bool {
		if !nearlyEqual(sorted[i].Y, sorted[j].Y, 0.5) {
			return sorted[i].Y > sorted[j].Y // higher Y = higher on page
		}
		return sorted[i].X < sorted[j].X
	})
	return sorted
}

// DedupeChars removes duplicate overlapping characters.
func DedupeChars(chars []model.Char, tolerance float64) []model.Char {
	if len(chars) == 0 {
		return nil
	}

	result := make([]model.Char, 0, len(chars))
	result = append(result, chars[0])

	for i := 1; i < len(chars); i++ {
		isDupe := false
		for j := len(result) - 1; j >= 0 && j >= len(result)-5; j-- {
			if chars[i].Text == result[j].Text &&
				nearlyEqual(chars[i].X, result[j].X, tolerance) &&
				nearlyEqual(chars[i].Y, result[j].Y, tolerance) &&
				chars[i].FontName == result[j].FontName &&
				nearlyEqual(chars[i].FontSize, result[j].FontSize, tolerance) {
				isDupe = true
				break
			}
		}
		if !isDupe {
			result = append(result, chars[i])
		}
	}
	return result
}

// FilterBlankChars removes whitespace-only characters from the slice.
func FilterBlankChars(chars []model.Char) []model.Char {
	result := make([]model.Char, 0, len(chars))
	for _, c := range chars {
		if !isBlank(c.Text) {
			result = append(result, c)
		}
	}
	return result
}

func isBlank(s string) bool {
	for _, r := range s {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

func isPunctuation(s string) bool {
	for _, r := range s {
		if !unicode.IsPunct(r) && !unicode.IsSymbol(r) {
			return false
		}
	}
	return len(s) > 0
}

func nearlyEqual(a, b, tolerance float64) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d <= tolerance
}

func charsMatchAttrs(a, b model.Char, attrs []string) bool {
	for _, attr := range attrs {
		switch strings.ToLower(attr) {
		case "fontname":
			if a.FontName != b.FontName {
				return false
			}
		case "fontsize":
			if !nearlyEqual(a.FontSize, b.FontSize, 0.1) {
				return false
			}
		}
	}
	return true
}
