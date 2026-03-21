package extract

import (
	"github.com/digitorus/pdf"
)

// cidWidthTable holds parsed CIDFont width data.
type cidWidthTable struct {
	defaultWidth float64            // /DW value (default 1000)
	widths       map[int]float64    // CID → width mapping from /W array
}

// lookupWidth returns the width for a given CID.
func (t *cidWidthTable) lookupWidth(cid int) float64 {
	if w, ok := t.widths[cid]; ok {
		return w
	}
	return t.defaultWidth
}

// parseCIDWidths extracts the width table from a CIDFont dictionary.
// fontDict should be the Type0 font dictionary (has /DescendantFonts).
func parseCIDWidths(fontDict pdf.Value) *cidWidthTable {
	table := &cidWidthTable{
		defaultWidth: 1000,
		widths:       make(map[int]float64),
	}

	// Navigate to the CIDFont inside /DescendantFonts
	descendants := fontDict.Key("DescendantFonts")
	if descendants.Len() == 0 {
		return table
	}
	cidFont := descendants.Index(0)

	// Read /DW (default width)
	dw := cidFont.Key("DW")
	if !dw.IsNull() {
		table.defaultWidth = dw.Float64()
	}

	// Parse /W array
	wArray := cidFont.Key("W")
	if wArray.IsNull() {
		return table
	}

	parseWArray(wArray, table)
	return table
}

// parseWArray parses the CIDFont /W array format.
// Format variants (PDF spec 5.11.4):
//
//	[cid [w1 w2 w3 ...]]     — consecutive widths starting at cid
//	[cidFirst cidLast width]  — range of CIDs with same width
func parseWArray(wArray pdf.Value, table *cidWidthTable) {
	i := 0
	length := wArray.Len()

	for i < length {
		cidStart := int(wArray.Index(i).Int64())
		i++
		if i >= length {
			break
		}

		next := wArray.Index(i)

		switch next.Kind() {
		case pdf.Array:
			// [cid [w1 w2 w3 ...]] — consecutive widths
			for j := 0; j < next.Len(); j++ {
				table.widths[cidStart+j] = next.Index(j).Float64()
			}
			i++

		case pdf.Integer, pdf.Real:
			// [cidFirst cidLast width] — range with same width
			if i+1 >= length {
				break
			}
			cidEnd := int(next.Int64())
			i++
			width := wArray.Index(i).Float64()
			i++
			for cid := cidStart; cid <= cidEnd; cid++ {
				table.widths[cid] = width
			}

		default:
			i++
		}
	}
}

// isCIDFont checks if a font dictionary is a Type0 (CIDFont) font.
func isCIDFont(fontDict pdf.Value) bool {
	subtype := fontDict.Key("Subtype").Name()
	return subtype == "Type0"
}

// getCIDFromBytes extracts a CID from raw character bytes.
// For Identity-H encoding, CIDs are 2-byte big-endian values.
func getCIDFromBytes(raw string, pos int, isIdentityH bool) (cid int, advance int) {
	if isIdentityH && pos+1 < len(raw) {
		// 2-byte big-endian CID
		cid = int(raw[pos])<<8 | int(raw[pos+1])
		return cid, 2
	}
	// Single byte
	if pos < len(raw) {
		return int(raw[pos]), 1
	}
	return 0, 1
}

// isIdentityHEncoding checks if the font uses Identity-H encoding.
func isIdentityHEncoding(fontDict pdf.Value) bool {
	enc := fontDict.Key("Encoding")
	if enc.IsNull() {
		return false
	}
	return enc.Name() == "Identity-H" || enc.Name() == "Identity-V"
}
