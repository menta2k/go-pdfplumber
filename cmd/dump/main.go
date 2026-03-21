// Command dump prints all extracted words from a PDF for debugging.
package main

import (
	"fmt"
	"os"

	"github.com/menta2k/go-pdfplumber/pkg/plumber"
)

func main() {
	pdfPath := os.Args[1]

	doc, err := plumber.Open(pdfPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer doc.Close()

	page, err := doc.Page(1)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Page: %.0f x %.0f\n\n", page.Width(), page.Height())

	// Dump all chars first
	chars := page.Chars()
	fmt.Printf("=== CHARS (%d total) ===\n", len(chars))
	for i, c := range chars {
		isDot := c.Text == "." || c.Text == "…"
		marker := " "
		if isDot {
			marker = "*"
		}
		fmt.Printf("%s char[%3d]: %q  x=%.1f y=%.1f w=%.1f font=%s size=%.1f\n",
			marker, i, c.Text, c.X, c.Y, c.Width, c.FontName, c.FontSize)
	}

	fmt.Println()

	// Dump words
	opts := plumber.DefaultTextOptions()
	words := page.ExtractWords(opts)
	fmt.Printf("=== WORDS (%d total) ===\n", len(words))
	for i, w := range words {
		hasDots := false
		for _, c := range w.Chars {
			if c.Text == "." || c.Text == "…" {
				hasDots = true
				break
			}
		}
		marker := " "
		if hasDots {
			marker = ">"
		}
		fmt.Printf("%s word[%2d]: %-40q x0=%.1f y0=%.1f x1=%.1f y1=%.1f  nchars=%d\n",
			marker, i, w.Text, w.BBox.X0, w.BBox.Y0, w.BBox.X1, w.BBox.Y1, len(w.Chars))
	}
}
