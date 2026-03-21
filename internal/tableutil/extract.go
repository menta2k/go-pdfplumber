package tableutil

import (
	"sort"
	"strings"

	"github.com/menta2k/go-pdfplumber/internal/geometry"
	"github.com/menta2k/go-pdfplumber/pkg/model"
)

// ExtractCellText extracts text content for each cell in the grid.
// It uses midpoint containment: a char belongs to a cell if its midpoint is inside.
func ExtractCellText(chars []model.Char, grid [][]Cell, opts model.TextOptions) [][]string {
	if len(grid) == 0 {
		return nil
	}

	result := make([][]string, len(grid))
	for r, row := range grid {
		result[r] = make([]string, len(row))
		for c, cell := range row {
			cellChars := geometry.FilterCharsByMidpoint(chars, cell.BBox)
			result[r][c] = assembleCharsToText(cellChars, opts)
		}
	}
	return result
}

// assembleCharsToText sorts chars spatially and joins them into text.
func assembleCharsToText(chars []model.Char, opts model.TextOptions) string {
	if len(chars) == 0 {
		return ""
	}

	// Sort top-to-bottom, left-to-right
	sorted := make([]model.Char, len(chars))
	copy(sorted, chars)
	sort.Slice(sorted, func(i, j int) bool {
		if !nearlyEqual(sorted[i].Y, sorted[j].Y, opts.YTolerance) {
			return sorted[i].Y > sorted[j].Y
		}
		return sorted[i].X < sorted[j].X
	})

	// Group into lines by Y, then join chars within each line
	var lines []string
	var currentLine []model.Char
	var lastY float64

	for i, c := range sorted {
		if i == 0 {
			currentLine = append(currentLine, c)
			lastY = c.Y
			continue
		}

		if !nearlyEqual(c.Y, lastY, opts.YTolerance) {
			lines = append(lines, joinLineChars(currentLine, opts.XTolerance))
			currentLine = nil
		}
		currentLine = append(currentLine, c)
		lastY = c.Y
	}
	if len(currentLine) > 0 {
		lines = append(lines, joinLineChars(currentLine, opts.XTolerance))
	}

	return strings.Join(lines, "\n")
}

// joinLineChars joins characters on the same line, inserting spaces where gaps exist.
func joinLineChars(chars []model.Char, xTolerance float64) string {
	if len(chars) == 0 {
		return ""
	}

	var sb strings.Builder
	for i, c := range chars {
		if i > 0 {
			gap := c.X - (chars[i-1].X + chars[i-1].Width)
			if gap > xTolerance {
				sb.WriteByte(' ')
			}
		}
		sb.WriteString(c.Text)
	}
	return strings.TrimSpace(sb.String())
}
