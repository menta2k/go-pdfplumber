package textutil

import (
	"testing"

	"github.com/menta2k/go-pdfplumber/pkg/model"
)

func TestSearchWords(t *testing.T) {
	words := []model.Word{
		{Text: "Hello"},
		{Text: "World"},
		{Text: "Hello"},
	}

	matches := SearchWords(words, "Hello")
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}

	matches = SearchWords(words, "xyz")
	if len(matches) != 0 {
		t.Errorf("expected 0 matches, got %d", len(matches))
	}
}

func TestSearchWordsEmpty(t *testing.T) {
	matches := SearchWords(nil, "test")
	if matches != nil {
		t.Error("expected nil for empty words")
	}

	matches = SearchWords([]model.Word{{Text: "test"}}, "")
	if matches != nil {
		t.Error("expected nil for empty query")
	}
}
