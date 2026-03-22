package extract

import (
	"encoding/binary"

	"github.com/digitorus/pdf"
)

// parseBfCharMap parses the ToUnicode CMap's beginbfchar and beginbfrange
// entries directly from the CMap stream data. Returns a code→Unicode mapping.
// digitorus/pdf only handles beginbfrange, not beginbfchar, and may fail
// on TrueType fonts with custom single-byte encodings.
func parseBfCharMap(fontDict pdf.Value) map[uint16]rune {
	toUnicode := fontDict.Key("ToUnicode")
	if toUnicode.IsNull() {
		return nil
	}

	data := toUnicode.Data()
	if len(data) == 0 {
		return nil
	}

	return parseCMapEntries(data)
}

// parseCMapEntries extracts both beginbfchar and beginbfrange mappings from raw CMap data.
func parseCMapEntries(data []byte) map[uint16]rune {
	result := make(map[uint16]rune)
	s := string(data)

	// Parse all beginbfchar...endbfchar sections
	remaining := s
	for {
		start := indexOf(remaining, "beginbfchar")
		if start < 0 {
			break
		}
		remaining = remaining[start+len("beginbfchar"):]

		end := indexOf(remaining, "endbfchar")
		if end < 0 {
			break
		}
		section := remaining[:end]
		remaining = remaining[end+len("endbfchar"):]

		parseBfCharSection(section, result)
	}

	// Parse all beginbfrange...endbfrange sections
	remaining = s
	for {
		start := indexOf(remaining, "beginbfrange")
		if start < 0 {
			break
		}
		remaining = remaining[start+len("beginbfrange"):]

		end := indexOf(remaining, "endbfrange")
		if end < 0 {
			break
		}
		section := remaining[:end]
		remaining = remaining[end+len("endbfrange"):]

		parseBfRangeSection(section, result)
	}

	if len(result) == 0 {
		return nil
	}
	return result
}

func parseBfCharSection(section string, result map[uint16]rune) {
	i := 0
	for i < len(section) {
		// Find source hex: <xxxx>
		srcStart := indexOf(section[i:], "<")
		if srcStart < 0 {
			break
		}
		srcStart += i
		srcEnd := indexOf(section[srcStart+1:], ">")
		if srcEnd < 0 {
			break
		}
		srcEnd += srcStart + 1
		srcHex := section[srcStart+1 : srcEnd]

		// Find destination hex: <xxxx>
		i = srcEnd + 1
		dstStart := indexOf(section[i:], "<")
		if dstStart < 0 {
			break
		}
		dstStart += i
		dstEnd := indexOf(section[dstStart+1:], ">")
		if dstEnd < 0 {
			break
		}
		dstEnd += dstStart + 1
		dstHex := section[dstStart+1 : dstEnd]

		i = dstEnd + 1

		src := decodeHex(srcHex)
		dst := decodeHex(dstHex)

		code := hexToCode(src)
		uni := hexToRune(dst)
		if code >= 0 && uni > 0 {
			result[uint16(code)] = uni
		}
	}
}

// parseBfRangeSection parses beginbfrange entries: <start> <end> <dstStart>
func parseBfRangeSection(section string, result map[uint16]rune) {
	i := 0
	for i < len(section) {
		// Parse 3 hex values: <start> <end> <dstStart>
		var hexVals [3][]byte
		for h := 0; h < 3; h++ {
			start := indexOf(section[i:], "<")
			if start < 0 {
				return
			}
			start += i
			end := indexOf(section[start+1:], ">")
			if end < 0 {
				return
			}
			end += start + 1
			hexVals[h] = decodeHex(section[start+1 : end])
			i = end + 1
		}

		rangeStart := hexToCode(hexVals[0])
		rangeEnd := hexToCode(hexVals[1])
		dstStart := hexToRune(hexVals[2])

		if rangeStart < 0 || rangeEnd < 0 || dstStart <= 0 {
			continue
		}

		for code := rangeStart; code <= rangeEnd; code++ {
			result[uint16(code)] = dstStart + rune(code-rangeStart)
		}
	}
}

// hexToCode converts decoded hex bytes to a code value (1 or 2 bytes).
func hexToCode(b []byte) int {
	switch len(b) {
	case 1:
		return int(b[0])
	case 2:
		return int(binary.BigEndian.Uint16(b))
	default:
		return -1
	}
}

// hexToRune converts decoded hex bytes to a Unicode rune (2 or 4 bytes).
func hexToRune(b []byte) rune {
	switch len(b) {
	case 2:
		return rune(binary.BigEndian.Uint16(b))
	case 4:
		return rune(binary.BigEndian.Uint32(b))
	default:
		return 0
	}
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func decodeHex(hex string) []byte {
	var result []byte
	for i := 0; i+1 < len(hex); i += 2 {
		b := hexByte(hex[i])<<4 | hexByte(hex[i+1])
		result = append(result, b)
	}
	return result
}

func hexByte(c byte) byte {
	switch {
	case c >= '0' && c <= '9':
		return c - '0'
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10
	}
	return 0
}
