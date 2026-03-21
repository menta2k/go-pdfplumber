package textutil

import (
	"strings"

	"github.com/menta2k/go-pdfplumber/pkg/model"
)

const paragraphGapMultiplier = 1.5

// AssembleText joins text lines into a single string.
func AssembleText(lines []model.TextLine) string {
	if len(lines) == 0 {
		return ""
	}

	if len(lines) == 1 {
		return lines[0].Text
	}

	avgHeight := averageLineHeight(lines)

	var sb strings.Builder
	for i, line := range lines {
		if i > 0 {
			prevBottom := lines[i-1].BBox.Y0
			currTop := line.BBox.Y1
			gap := prevBottom - currTop

			if avgHeight > 0 && gap > avgHeight*paragraphGapMultiplier {
				sb.WriteByte('\n')
			}
			sb.WriteByte('\n')
		}
		sb.WriteString(line.Text)
	}

	return sb.String()
}

func averageLineHeight(lines []model.TextLine) float64 {
	if len(lines) == 0 {
		return 0
	}
	var total float64
	for _, l := range lines {
		total += l.BBox.Height()
	}
	return total / float64(len(lines))
}
