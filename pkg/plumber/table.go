package plumber

import (
	"github.com/menta2k/go-pdfplumber/internal/tableutil"
	"github.com/menta2k/go-pdfplumber/pkg/model"
)

// Table represents a detected table on a page.
type Table struct {
	BBox     BBox
	Cells    [][]tableutil.Cell
	RowCount int
	ColCount int

	page *Page
}

// Extract returns the text content of each cell as a 2D string slice (rows x cols).
func (t *Table) Extract(opts ...TextOptions) [][]string {
	o := resolveOpts(opts)
	return tableutil.ExtractCellText(t.page.Chars(), t.Cells, o)
}

// TableFinderOptions controls table detection behavior.
type TableFinderOptions struct {
	// SnapTolerance merges parallel lines within this distance. Default: 3.0
	SnapTolerance float64

	// JoinTolerance merges collinear line segments within this distance. Default: 3.0
	JoinTolerance float64

	// MinEdgeLength discards edges shorter than this. Default: 3.0
	MinEdgeLength float64

	// IntersectionTolerance for finding edge crossings. Default: 3.0
	IntersectionTolerance float64
}

// DefaultTableFinderOptions returns sensible defaults.
func DefaultTableFinderOptions() TableFinderOptions {
	return TableFinderOptions{
		SnapTolerance:         3.0,
		JoinTolerance:         3.0,
		MinEdgeLength:         3.0,
		IntersectionTolerance: 3.0,
	}
}

// FindTables detects tables on the page and returns Table metadata.
func (p *Page) FindTables(opts ...TableFinderOptions) []Table {
	o := resolveTableOpts(opts)
	p.ensureContent()

	edgeOpts := tableutil.EdgeOptions{
		SnapTolerance: o.SnapTolerance,
		MinEdgeLength: o.MinEdgeLength,
	}

	edges := tableutil.ExtractEdges(p.content.lines, p.content.rects, edgeOpts)
	edges = tableutil.MergeEdges(edges, o.SnapTolerance, o.JoinTolerance)

	intersections := tableutil.FindIntersections(edges, o.IntersectionTolerance)
	if len(intersections) < 4 {
		return nil // need at least 4 points to form a cell
	}

	grid := tableutil.BuildCells(intersections, edges, o.IntersectionTolerance)
	if len(grid) == 0 {
		return nil
	}

	// Group cells into distinct tables by finding connected components.
	// For now, treat the entire grid as one table.
	tables := groupIntoTables(grid, p)
	return tables
}

// ExtractTables detects tables and returns their text content.
// Returns [table][row][col] string.
func (p *Page) ExtractTables(tableOpts ...TableFinderOptions) [][][]string {
	tables := p.FindTables(tableOpts...)
	result := make([][][]string, len(tables))
	for i, t := range tables {
		result[i] = t.Extract()
	}
	return result
}

// ExtractTable extracts the largest table on the page.
// Returns [row][col] string, or nil if no tables found.
func (p *Page) ExtractTable(tableOpts ...TableFinderOptions) [][]string {
	tables := p.FindTables(tableOpts...)
	if len(tables) == 0 {
		return nil
	}

	// Find largest by cell count
	largest := 0
	maxCells := 0
	for i, t := range tables {
		cells := t.RowCount * t.ColCount
		if cells > maxCells {
			maxCells = cells
			largest = i
		}
	}
	return tables[largest].Extract()
}

func groupIntoTables(grid [][]Cell, page *Page) []Table {
	if len(grid) == 0 {
		return nil
	}

	bbox := tableutil.GridBBox(grid)
	maxCols := 0
	for _, row := range grid {
		if len(row) > maxCols {
			maxCols = len(row)
		}
	}

	// Normalize grid so each row has the same number of columns
	normalized := normalizeGrid(grid, maxCols)

	return []Table{{
		BBox:     model.BBox(bbox),
		Cells:    normalized,
		RowCount: len(normalized),
		ColCount: maxCols,
		page:     page,
	}}
}

// normalizeGrid ensures all rows have the same column count by padding with empty cells.
func normalizeGrid(grid [][]Cell, cols int) [][]Cell {
	result := make([][]Cell, len(grid))
	for r, row := range grid {
		result[r] = make([]Cell, cols)
		for c, cell := range row {
			if c < cols {
				result[r][c] = cell
			}
		}
	}
	return result
}

// Cell re-exports for public API convenience.
type Cell = tableutil.Cell

func resolveTableOpts(opts []TableFinderOptions) TableFinderOptions {
	if len(opts) > 0 {
		return withTableDefaults(opts[0])
	}
	return DefaultTableFinderOptions()
}

func withTableDefaults(o TableFinderOptions) TableFinderOptions {
	if o.SnapTolerance <= 0 {
		o.SnapTolerance = 3.0
	}
	if o.JoinTolerance <= 0 {
		o.JoinTolerance = 3.0
	}
	if o.MinEdgeLength <= 0 {
		o.MinEdgeLength = 3.0
	}
	if o.IntersectionTolerance <= 0 {
		o.IntersectionTolerance = 3.0
	}
	return o
}
