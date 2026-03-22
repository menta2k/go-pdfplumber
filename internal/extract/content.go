// Package extract provides enhanced PDF content extraction with proper
// per-character positioning, built on top of github.com/digitorus/pdf.
package extract

import (
	"strings"

	"github.com/digitorus/pdf"
)

// Char represents a single character with precise position data.
type Char struct {
	Text     string
	FontName string
	FontSize float64
	X        float64 // left edge
	Y        float64 // baseline (PDF coords: bottom-up)
	W        float64 // advance width in points
}

// Rect represents a rectangle from the PDF content stream.
type Rect struct {
	X, Y, W, H float64
}

// LineOp represents a line drawn via move-to/line-to operators.
type LineOp struct {
	X0, Y0, X1, Y1 float64
}

// Content holds all extracted objects from a page.
type Content struct {
	Chars []Char
	Rects []Rect
	Lines []LineOp
}

type matrix [3][3]float64

var ident = matrix{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}}

func (x matrix) mul(y matrix) matrix {
	var z matrix
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			for k := 0; k < 3; k++ {
				z[i][j] += x[i][k] * y[k][j]
			}
		}
	}
	return z
}

type fontInfo struct {
	pdfFont    pdf.Font
	name       string           // resolved base font name
	cidWidths  *cidWidthTable   // non-nil for CIDFonts
	identityH  bool             // true if Identity-H encoding (2-byte CIDs)
	bfCharMap  map[uint16]rune  // beginbfchar Unicode mappings (supplements CMap)
}

// glyphWidth returns the width of a glyph in 1/1000 units.
func (f *fontInfo) glyphWidth(raw string, pos int) (w0 float64, advance int) {
	if f.cidWidths != nil {
		// CIDFont: look up width by CID
		cid, adv := getCIDFromBytes(raw, pos, f.identityH)
		return f.cidWidths.lookupWidth(cid), adv
	}

	// Simple font: try PDF font's /Widths array first
	if pos < len(raw) {
		w0 = f.pdfFont.Width(int(raw[pos]))
	}
	if w0 == 0 && isStandardFont(f.name) {
		if pos < len(raw) {
			w0 = fontWidth(f.name, raw[pos])
		}
	}
	if w0 == 0 {
		w0 = 600 // fallback
	}
	return w0, 1
}

type graphicsState struct {
	Tc    float64 // character spacing
	Tw    float64 // word spacing
	Th    float64 // horizontal scaling
	Tl    float64 // text leading
	Tf    pdf.Font
	Tfs   float64 // font size
	Trise float64 // text rise
	Tm    matrix  // text matrix
	Tlm   matrix  // text line matrix
	CTM   matrix  // current transformation matrix

	font *fontInfo
}

// ExtractContent extracts characters, rectangles, and lines from a PDF page
// with proper per-character positioning, including CIDFont support.
func ExtractContent(page pdf.Page) (result Content) {
	defer func() {
		if r := recover(); r != nil {
			result = Content{}
		}
	}()

	strm := page.V.Key("Contents")
	var enc pdf.TextEncoding = &nopEncoder{}

	g := graphicsState{
		Th:   1,
		CTM:  ident,
		font: &fontInfo{},
	}

	// Cache CID width tables per font name to avoid re-parsing
	cidCache := make(map[string]*cidWidthTable)

	var chars []Char
	var rects []Rect
	var lines []LineOp
	var curX, curY float64

	showText := func(s string) {
		// For CIDFonts with bfchar map: decode CIDs directly from raw bytes
		if g.font.bfCharMap != nil && g.font.identityH {
			showTextCID(s, g.font, &g, &chars)
			return
		}

		// For non-CID fonts with bfchar map (e.g. TrueType with custom encoding):
		// decode each byte using the ToUnicode map directly
		if g.font.bfCharMap != nil && !g.font.identityH {
			showTextBfChar(s, g.font, &g, &chars)
			return
		}

		// Standard path: use digitorus/pdf encoder
		decoded := enc.Decode(s)
		rawPos := 0

		for _, ch := range decoded {
			trm := matrix{
				{g.Tfs * g.Th, 0, 0},
				{0, g.Tfs, 0},
				{0, g.Trise, 1},
			}.mul(g.Tm).mul(g.CTM)

			w0, advance := g.font.glyphWidth(s, rawPos)
			rawPos += advance

			charWidth := w0 / 1000 * trm[0][0]

			if ch != ' ' && ch != 0 {
				chars = append(chars, Char{
					Text:     string(ch),
					FontName: g.font.name,
					FontSize: trm[0][0],
					X:        trm[2][0],
					Y:        trm[2][1],
					W:        charWidth,
				})
			}

			tx := w0/1000*g.Tfs + g.Tc
			if ch == ' ' {
				tx += g.Tw
			}
			tx *= g.Th
			g.Tm = matrix{{1, 0, 0}, {0, 1, 0}, {tx, 0, 1}}.mul(g.Tm)
		}
	}

	var gstack []graphicsState

	pdf.Interpret(strm, func(stk *pdf.Stack, op string) {
		n := stk.Len()
		args := make([]pdf.Value, n)
		for i := n - 1; i >= 0; i-- {
			args[i] = stk.Pop()
		}

		switch op {
		case "cm":
			if len(args) != 6 {
				return
			}
			var m matrix
			for i := 0; i < 6; i++ {
				m[i/2][i%2] = args[i].Float64()
			}
			m[2][2] = 1
			g.CTM = m.mul(g.CTM)

		case "q":
			gstack = append(gstack, g)

		case "Q":
			if len(gstack) > 0 {
				g = gstack[len(gstack)-1]
				gstack = gstack[:len(gstack)-1]
			}

		case "re":
			if len(args) != 4 {
				return
			}
			rects = append(rects, Rect{
				X: args[0].Float64(),
				Y: args[1].Float64(),
				W: args[2].Float64(),
				H: args[3].Float64(),
			})

		case "m":
			if len(args) >= 2 {
				curX = args[0].Float64()
				curY = args[1].Float64()
			}

		case "l":
			if len(args) >= 2 {
				x1 := args[0].Float64()
				y1 := args[1].Float64()
				lines = append(lines, LineOp{
					X0: curX, Y0: curY,
					X1: x1, Y1: y1,
				})
				curX = x1
				curY = y1
			}

		case "S", "s":
		case "f", "F", "f*", "B", "B*", "b", "b*":

		case "BT":
			g.Tm = ident
			g.Tlm = g.Tm

		case "ET":

		case "T*":
			x := matrix{{1, 0, 0}, {0, 1, 0}, {0, -g.Tl, 1}}
			g.Tlm = x.mul(g.Tlm)
			g.Tm = g.Tlm

		case "Tc":
			if len(args) == 1 {
				g.Tc = args[0].Float64()
			}

		case "TD":
			if len(args) == 2 {
				g.Tl = -args[1].Float64()
			}
			fallthrough
		case "Td":
			if len(args) == 2 {
				tx := args[0].Float64()
				ty := args[1].Float64()
				x := matrix{{1, 0, 0}, {0, 1, 0}, {tx, ty, 1}}
				g.Tlm = x.mul(g.Tlm)
				g.Tm = g.Tlm
			}

		case "Tf":
			if len(args) == 2 {
				fname := args[0].Name()
				g.Tf = page.Font(fname)
				enc = g.Tf.Encoder()
				if enc == nil {
					enc = &nopEncoder{}
				}
				g.Tfs = args[1].Float64()

				// Resolve base font name
				baseName := g.Tf.BaseFont()
				if i := strings.Index(baseName, "+"); i >= 0 {
					baseName = baseName[i+1:]
				}

				// Build font info with CIDFont support
				fontDict := g.Tf.V
				fi := &fontInfo{
					pdfFont: g.Tf,
					name:    baseName,
				}

				if isCIDFont(fontDict) {
					fi.identityH = isIdentityHEncoding(fontDict)
					cacheKey := fname + ":" + baseName
					if cached, ok := cidCache[cacheKey]; ok {
						fi.cidWidths = cached
					} else {
						fi.cidWidths = parseCIDWidths(fontDict)
						cidCache[cacheKey] = fi.cidWidths
					}
				}

				// Parse ToUnicode bfchar/bfrange mappings for all fonts
				// (TrueType fonts with custom encoding also need this)
				fi.bfCharMap = parseBfCharMap(fontDict)

				g.font = fi
			}

		case "\"":
			if len(args) == 3 {
				g.Tw = args[0].Float64()
				g.Tc = args[1].Float64()
				x := matrix{{1, 0, 0}, {0, 1, 0}, {0, -g.Tl, 1}}
				g.Tlm = x.mul(g.Tlm)
				g.Tm = g.Tlm
				showText(args[2].RawString())
			}

		case "'":
			if len(args) == 1 {
				x := matrix{{1, 0, 0}, {0, 1, 0}, {0, -g.Tl, 1}}
				g.Tlm = x.mul(g.Tlm)
				g.Tm = g.Tlm
				showText(args[0].RawString())
			}

		case "Tj":
			if len(args) == 1 {
				showText(args[0].RawString())
			}

		case "TJ":
			if len(args) == 1 {
				v := args[0]
				for i := 0; i < v.Len(); i++ {
					x := v.Index(i)
					if x.Kind() == pdf.String {
						showText(x.RawString())
					} else {
						tx := -x.Float64() / 1000 * g.Tfs * g.Th
						g.Tm = matrix{{1, 0, 0}, {0, 1, 0}, {tx, 0, 1}}.mul(g.Tm)
					}
				}
			}

		case "TL":
			if len(args) == 1 {
				g.Tl = args[0].Float64()
			}

		case "Tm":
			if len(args) == 6 {
				var m matrix
				for i := 0; i < 6; i++ {
					m[i/2][i%2] = args[i].Float64()
				}
				m[2][2] = 1
				g.Tm = m
				g.Tlm = m
			}

		case "Tr":
		case "Ts":
			if len(args) == 1 {
				g.Trise = args[0].Float64()
			}

		case "Tw":
			if len(args) == 1 {
				g.Tw = args[0].Float64()
			}

		case "Tz":
			if len(args) == 1 {
				g.Th = args[0].Float64() / 100
			}

		case "gs":
		}
	})

	return Content{Chars: chars, Rects: rects, Lines: lines}
}

// showTextCID handles text rendering for CIDFonts with bfchar Unicode maps.
// It iterates raw bytes in 2-byte chunks (Identity-H), looks up each CID
// in the bfchar map for proper Unicode, and uses the CID width table for positioning.
func showTextCID(s string, fi *fontInfo, g *graphicsState, chars *[]Char) {
	pos := 0
	for pos+1 < len(s) {
		cid := uint16(s[pos])<<8 | uint16(s[pos+1])
		pos += 2

		trm := matrix{
			{g.Tfs * g.Th, 0, 0},
			{0, g.Tfs, 0},
			{0, g.Trise, 1},
		}.mul(g.Tm).mul(g.CTM)

		w0 := fi.cidWidths.lookupWidth(int(cid))
		charWidth := w0 / 1000 * trm[0][0]

		// Look up Unicode from bfchar map
		ch, ok := fi.bfCharMap[cid]
		if !ok {
			ch = rune(cid) // fallback to raw CID value
		}

		if ch != ' ' && ch != 0 {
			*chars = append(*chars, Char{
				Text:     string(ch),
				FontName: fi.name,
				FontSize: trm[0][0],
				X:        trm[2][0],
				Y:        trm[2][1],
				W:        charWidth,
			})
		}

		tx := w0/1000*g.Tfs + g.Tc
		if ch == ' ' {
			tx += g.Tw
		}
		tx *= g.Th
		g.Tm = matrix{{1, 0, 0}, {0, 1, 0}, {tx, 0, 1}}.mul(g.Tm)
	}
}

// showTextBfChar handles text rendering for non-CID fonts (e.g. TrueType)
// that have a ToUnicode bfchar map with single-byte codes.
func showTextBfChar(s string, fi *fontInfo, g *graphicsState, chars *[]Char) {
	for pos := 0; pos < len(s); pos++ {
		code := uint16(s[pos])

		trm := matrix{
			{g.Tfs * g.Th, 0, 0},
			{0, g.Tfs, 0},
			{0, g.Trise, 1},
		}.mul(g.Tm).mul(g.CTM)

		w0, _ := fi.glyphWidth(s, pos)
		charWidth := w0 / 1000 * trm[0][0]

		ch, ok := fi.bfCharMap[code]
		if !ok {
			ch = rune(code) // fallback
		}

		if ch != ' ' && ch != 0 {
			*chars = append(*chars, Char{
				Text:     string(ch),
				FontName: fi.name,
				FontSize: trm[0][0],
				X:        trm[2][0],
				Y:        trm[2][1],
				W:        charWidth,
			})
		}

		tx := w0/1000*g.Tfs + g.Tc
		if ch == ' ' {
			tx += g.Tw
		}
		tx *= g.Th
		g.Tm = matrix{{1, 0, 0}, {0, 1, 0}, {tx, 0, 1}}.mul(g.Tm)
	}
}

type nopEncoder struct{}

func (e *nopEncoder) Decode(raw string) string { return raw }
