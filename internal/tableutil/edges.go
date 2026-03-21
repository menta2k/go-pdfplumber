// Package tableutil provides table detection and extraction algorithms
// for PDF pages, following pdfplumber's line-intersection approach.
package tableutil

import (
	"math"
	"sort"

	"github.com/menta2k/go-pdfplumber/pkg/model"
)

// Edge represents a horizontal or vertical line segment used for table detection.
type Edge struct {
	X0, Y0, X1, Y1 float64
	Orientation     string // "horizontal" or "vertical"
}

// EdgeOptions controls edge extraction behavior.
type EdgeOptions struct {
	SnapTolerance float64 // snap nearly-H/V edges to exact H/V (default 3.0)
	MinEdgeLength float64 // discard edges shorter than this (default 3.0)
}

// DefaultEdgeOptions returns sensible defaults.
func DefaultEdgeOptions() EdgeOptions {
	return EdgeOptions{
		SnapTolerance: 3.0,
		MinEdgeLength: 3.0,
	}
}

// ExtractEdges collects edges from page lines and rects.
// Rects are decomposed into 4 edges each.
func ExtractEdges(lines []model.LineSegment, rects []model.RectObject, opts EdgeOptions) []Edge {
	var edges []Edge

	// From explicit lines
	for _, l := range lines {
		if l.Orientation == "diagonal" {
			continue
		}
		edges = append(edges, Edge{
			X0: l.X0, Y0: l.Y0, X1: l.X1, Y1: l.Y1,
			Orientation: l.Orientation,
		})
	}

	// Decompose rects into 4 edges
	for _, r := range rects {
		b := r.BBox
		// Bottom
		edges = append(edges, Edge{X0: b.X0, Y0: b.Y0, X1: b.X1, Y1: b.Y0, Orientation: "horizontal"})
		// Top
		edges = append(edges, Edge{X0: b.X0, Y0: b.Y1, X1: b.X1, Y1: b.Y1, Orientation: "horizontal"})
		// Left
		edges = append(edges, Edge{X0: b.X0, Y0: b.Y0, X1: b.X0, Y1: b.Y1, Orientation: "vertical"})
		// Right
		edges = append(edges, Edge{X0: b.X1, Y0: b.Y0, X1: b.X1, Y1: b.Y1, Orientation: "vertical"})
	}

	// Snap and classify
	edges = snapEdges(edges, opts.SnapTolerance)

	// Filter short edges
	edges = filterByLength(edges, opts.MinEdgeLength)

	return edges
}

// snapEdges snaps nearly horizontal/vertical edges to exact H/V.
func snapEdges(edges []Edge, tolerance float64) []Edge {
	result := make([]Edge, 0, len(edges))
	for _, e := range edges {
		if math.Abs(e.Y0-e.Y1) <= tolerance && e.Orientation != "vertical" {
			// Snap to horizontal
			midY := (e.Y0 + e.Y1) / 2
			result = append(result, Edge{
				X0: math.Min(e.X0, e.X1), Y0: midY,
				X1: math.Max(e.X0, e.X1), Y1: midY,
				Orientation: "horizontal",
			})
		} else if math.Abs(e.X0-e.X1) <= tolerance && e.Orientation != "horizontal" {
			// Snap to vertical
			midX := (e.X0 + e.X1) / 2
			result = append(result, Edge{
				X0: midX, Y0: math.Min(e.Y0, e.Y1),
				X1: midX, Y1: math.Max(e.Y0, e.Y1),
				Orientation: "vertical",
			})
		} else {
			result = append(result, e)
		}
	}
	return result
}

func filterByLength(edges []Edge, minLength float64) []Edge {
	var result []Edge
	for _, e := range edges {
		length := edgeLength(e)
		if length >= minLength {
			result = append(result, e)
		}
	}
	return result
}

func edgeLength(e Edge) float64 {
	dx := e.X1 - e.X0
	dy := e.Y1 - e.Y0
	return math.Sqrt(dx*dx + dy*dy)
}

// MergeEdges merges collinear edges that are close together (within joinTolerance).
func MergeEdges(edges []Edge, snapTolerance, joinTolerance float64) []Edge {
	var horizontal, vertical, other []Edge
	for _, e := range edges {
		switch e.Orientation {
		case "horizontal":
			horizontal = append(horizontal, e)
		case "vertical":
			vertical = append(vertical, e)
		default:
			other = append(other, e)
		}
	}

	horizontal = mergeCollinear(horizontal, snapTolerance, joinTolerance, true)
	vertical = mergeCollinear(vertical, snapTolerance, joinTolerance, false)

	result := make([]Edge, 0, len(horizontal)+len(vertical)+len(other))
	result = append(result, horizontal...)
	result = append(result, vertical...)
	result = append(result, other...)
	return result
}

func mergeCollinear(edges []Edge, snapTol, joinTol float64, isHorizontal bool) []Edge {
	if len(edges) == 0 {
		return nil
	}

	// Group by the constant coordinate (Y for horizontal, X for vertical)
	type group struct {
		coord float64
		edges []Edge
	}
	var groups []group

	for _, e := range edges {
		coord := e.Y0
		if !isHorizontal {
			coord = e.X0
		}

		found := false
		for i := range groups {
			if math.Abs(groups[i].coord-coord) <= snapTol {
				groups[i].edges = append(groups[i].edges, e)
				found = true
				break
			}
		}
		if !found {
			groups = append(groups, group{coord: coord, edges: []Edge{e}})
		}
	}

	var result []Edge
	for _, g := range groups {
		// Sort by start position
		sort.Slice(g.edges, func(i, j int) bool {
			if isHorizontal {
				return g.edges[i].X0 < g.edges[j].X0
			}
			return g.edges[i].Y0 < g.edges[j].Y0
		})

		// Merge overlapping/adjacent
		merged := []Edge{g.edges[0]}
		for _, e := range g.edges[1:] {
			last := &merged[len(merged)-1]
			var lastEnd, eStart float64
			if isHorizontal {
				lastEnd = last.X1
				eStart = e.X0
			} else {
				lastEnd = last.Y1
				eStart = e.Y0
			}

			if eStart <= lastEnd+joinTol {
				// Extend
				if isHorizontal {
					last.X1 = math.Max(last.X1, e.X1)
				} else {
					last.Y1 = math.Max(last.Y1, e.Y1)
				}
			} else {
				merged = append(merged, e)
			}
		}
		result = append(result, merged...)
	}
	return result
}
