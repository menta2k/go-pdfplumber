package plumber

import (
	"testing"
)

func TestFindTables(t *testing.T) {
	doc, err := Open(testTablePDF)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer doc.Close()

	page, err := doc.Page(1)
	if err != nil {
		t.Fatalf("Page(1): %v", err)
	}

	t.Logf("Page rects: %d, lines: %d", len(page.Rects()), len(page.Lines()))

	tables := page.FindTables()

	if len(tables) == 0 {
		t.Fatal("expected at least 1 table")
	}

	t.Logf("Found %d table(s)", len(tables))
	for i, tbl := range tables {
		t.Logf("  table[%d]: %dx%d bbox=(%.1f, %.1f, %.1f, %.1f)",
			i, tbl.RowCount, tbl.ColCount,
			tbl.BBox.X0, tbl.BBox.Y0, tbl.BBox.X1, tbl.BBox.Y1)
	}

	// Should find a 3x3 table
	tbl := tables[0]
	if tbl.RowCount != 3 {
		t.Errorf("expected 3 rows, got %d", tbl.RowCount)
	}
	if tbl.ColCount != 3 {
		t.Errorf("expected 3 cols, got %d", tbl.ColCount)
	}
}

func TestExtractTable(t *testing.T) {
	doc, err := Open(testTablePDF)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer doc.Close()

	page, err := doc.Page(1)
	if err != nil {
		t.Fatalf("Page(1): %v", err)
	}

	data := page.ExtractTable()
	if data == nil {
		t.Fatal("expected table data")
	}

	t.Logf("Table data (%d rows):", len(data))
	for r, row := range data {
		t.Logf("  row[%d]: %v", r, row)
	}

	// Verify expected content
	expected := [][]string{
		{"Name", "Age", "City"},
		{"Alice", "30", "NYC"},
		{"Bob", "25", "LA"},
	}

	if len(data) != len(expected) {
		t.Fatalf("expected %d rows, got %d", len(expected), len(data))
	}

	for r, row := range expected {
		if len(data[r]) < len(row) {
			t.Errorf("row %d: expected %d cols, got %d", r, len(row), len(data[r]))
			continue
		}
		for c, want := range row {
			if data[r][c] != want {
				t.Errorf("cell[%d][%d] = %q, want %q", r, c, data[r][c], want)
			}
		}
	}
}

func TestExtractTables(t *testing.T) {
	doc, err := Open(testTablePDF)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer doc.Close()

	page, err := doc.Page(1)
	if err != nil {
		t.Fatalf("Page(1): %v", err)
	}

	tables := page.ExtractTables()
	if len(tables) == 0 {
		t.Fatal("expected at least 1 table")
	}

	// First table should be the 3x3
	if len(tables[0]) < 3 {
		t.Errorf("expected at least 3 rows in first table, got %d", len(tables[0]))
	}
}

func TestExtractTableNoTables(t *testing.T) {
	doc, err := Open(testSimplePDF)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer doc.Close()

	page, err := doc.Page(1)
	if err != nil {
		t.Fatalf("Page(1): %v", err)
	}

	data := page.ExtractTable()
	if data != nil {
		t.Errorf("expected nil for page without tables, got %v", data)
	}
}

func TestFindTablesCustomOptions(t *testing.T) {
	doc, err := Open(testTablePDF)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer doc.Close()

	page, err := doc.Page(1)
	if err != nil {
		t.Fatalf("Page(1): %v", err)
	}

	opts := TableFinderOptions{
		SnapTolerance:         5.0,
		JoinTolerance:         5.0,
		MinEdgeLength:         5.0,
		IntersectionTolerance: 5.0,
	}

	tables := page.FindTables(opts)
	if len(tables) == 0 {
		t.Fatal("expected tables with custom options")
	}
}
