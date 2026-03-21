package textutil

import (
	"strings"

	"github.com/menta2k/go-pdfplumber/pkg/model"
)

// SearchWords finds words that contain the query string (case-sensitive).
func SearchWords(words []model.Word, query string) []model.Word {
	if query == "" {
		return nil
	}

	var matches []model.Word
	for _, w := range words {
		if strings.Contains(w.Text, query) {
			matches = append(matches, w)
		}
	}
	return matches
}
