package tableutil

import (
	"github.com/menta2k/go-pdfplumber/pkg/model"
)

// Cell represents a single table cell.
type Cell struct {
	BBox model.BBox
	Row  int
	Col  int
}

// BuildCells constructs a grid of cells from intersection points and edges.
// It identifies rectangular cells formed by adjacent intersection coordinates
// where all 4 bounding edges exist.
func BuildCells(points []model.Point, edges []Edge, tolerance float64) [][]Cell {
	xs, ys := UniqueCoords(points, tolerance)

	if len(xs) < 2 || len(ys) < 2 {
		return nil
	}

	rows := len(ys) - 1
	cols := len(xs) - 1

	grid := make([][]Cell, rows)
	for r := 0; r < rows; r++ {
		grid[r] = make([]Cell, 0, cols)
		for c := 0; c < cols; c++ {
			cellBBox := model.BBox{
				X0: xs[c],
				Y0: ys[r],
				X1: xs[c+1],
				Y1: ys[r+1],
			}

			if hasBoundingEdges(cellBBox, edges, tolerance) {
				grid[r] = append(grid[r], Cell{
					BBox: cellBBox,
					Row:  r,
					Col:  c,
				})
			}
		}
	}

	// Remove empty rows and reverse to reading order (top-to-bottom)
	var result [][]Cell
	for i := len(grid) - 1; i >= 0; i-- {
		if len(grid[i]) > 0 {
			result = append(result, grid[i])
		}
	}
	return result
}

// hasBoundingEdges checks whether all 4 sides of the cell bbox are covered by edges.
func hasBoundingEdges(bbox model.BBox, edges []Edge, tolerance float64) bool {
	hasTop := false
	hasBottom := false
	hasLeft := false
	hasRight := false

	for _, e := range edges {
		switch e.Orientation {
		case "horizontal":
			// Bottom edge
			if nearlyEqual(e.Y0, bbox.Y0, tolerance) &&
				e.X0 <= bbox.X0+tolerance && e.X1 >= bbox.X1-tolerance {
				hasBottom = true
			}
			// Top edge
			if nearlyEqual(e.Y0, bbox.Y1, tolerance) &&
				e.X0 <= bbox.X0+tolerance && e.X1 >= bbox.X1-tolerance {
				hasTop = true
			}
		case "vertical":
			// Left edge
			if nearlyEqual(e.X0, bbox.X0, tolerance) &&
				e.Y0 <= bbox.Y0+tolerance && e.Y1 >= bbox.Y1-tolerance {
				hasLeft = true
			}
			// Right edge
			if nearlyEqual(e.X0, bbox.X1, tolerance) &&
				e.Y0 <= bbox.Y0+tolerance && e.Y1 >= bbox.Y1-tolerance {
				hasRight = true
			}
		}
	}

	return hasTop && hasBottom && hasLeft && hasRight
}

// GridBBox returns the bounding box encompassing all cells in the grid.
func GridBBox(grid [][]Cell) model.BBox {
	if len(grid) == 0 {
		return model.BBox{}
	}

	bbox := grid[0][0].BBox
	for _, row := range grid {
		for _, cell := range row {
			if cell.BBox.X0 < bbox.X0 {
				bbox.X0 = cell.BBox.X0
			}
			if cell.BBox.Y0 < bbox.Y0 {
				bbox.Y0 = cell.BBox.Y0
			}
			if cell.BBox.X1 > bbox.X1 {
				bbox.X1 = cell.BBox.X1
			}
			if cell.BBox.Y1 > bbox.Y1 {
				bbox.Y1 = cell.BBox.Y1
			}
		}
	}
	return bbox
}
