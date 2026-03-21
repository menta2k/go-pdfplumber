// Command generate creates test PDF fixtures using go-fpdf
// which produces PDFs with standard Type1 fonts.
package main

import (
	"fmt"
	"os"

	"github.com/go-pdf/fpdf"
)

func main() {
	if err := generateSimple(); err != nil {
		fmt.Fprintf(os.Stderr, "simple: %v\n", err)
		os.Exit(1)
	}
	if err := generateTable(); err != nil {
		fmt.Fprintf(os.Stderr, "table: %v\n", err)
		os.Exit(1)
	}
	if err := generateMultipage(); err != nil {
		fmt.Fprintf(os.Stderr, "multipage: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("All test fixtures generated.")
}

func generateSimple() error {
	pdf := fpdf.New("P", "pt", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Helvetica", "", 14)

	pdf.SetXY(72, 72)
	pdf.Cell(0, 20, "Hello World")

	pdf.SetXY(72, 100)
	pdf.Cell(0, 20, "This is a test PDF document.")

	pdf.SetXY(72, 128)
	pdf.Cell(0, 20, "It has three lines of text.")

	return pdf.OutputFileAndClose("../simple.pdf")
}

func generateTable() error {
	pdf := fpdf.New("P", "pt", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Helvetica", "", 12)

	startX, startY := 72.0, 72.0
	colWidth, rowHeight := 120.0, 30.0

	data := [][]string{
		{"Name", "Age", "City"},
		{"Alice", "30", "NYC"},
		{"Bob", "25", "LA"},
	}

	// Draw table with borders
	for r, row := range data {
		for c, cell := range row {
			x := startX + float64(c)*colWidth
			y := startY + float64(r)*rowHeight
			pdf.Rect(x, y, colWidth, rowHeight, "D")
			pdf.SetXY(x+5, y+8)
			pdf.Cell(colWidth-10, 14, cell)
		}
	}

	return pdf.OutputFileAndClose("../simple_table.pdf")
}

func generateMultipage() error {
	pdf := fpdf.New("P", "pt", "A4", "")
	pdf.SetFont("Helvetica", "", 14)

	for i := 1; i <= 3; i++ {
		pdf.AddPage()
		pdf.SetXY(72, 72)
		pdf.Cell(0, 20, fmt.Sprintf("Page %d of 3", i))
	}

	return pdf.OutputFileAndClose("../multipage.pdf")
}
