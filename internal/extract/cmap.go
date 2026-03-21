package extract

import (
	"encoding/binary"

	"github.com/digitorus/pdf"
)

// parseBfCharMap parses the ToUnicode CMap's beginbfchar entries directly
// from the CMap stream data. Returns a CID→Unicode mapping.
// digitorus/pdf only handles beginbfrange, not beginbfchar.
func parseBfCharMap(fontDict pdf.Value) map[uint16]rune {
	toUnicode := fontDict.Key("ToUnicode")
	if toUnicode.IsNull() {
		return nil
	}

	data := toUnicode.Data()
	if len(data) == 0 {
		return nil
	}

	return parseCMapBfChar(data)
}

// parseCMapBfChar extracts beginbfchar mappings from raw CMap data.
// Format: <srcCode> <dstUnicode>
func parseCMapBfChar(data []byte) map[uint16]rune {
	result := make(map[uint16]rune)
	s := string(data)

	// Find all beginbfchar...endbfchar sections
	for {
		start := indexOf(s, "beginbfchar")
		if start < 0 {
			break
		}
		s = s[start+len("beginbfchar"):]

		end := indexOf(s, "endbfchar")
		if end < 0 {
			break
		}
		section := s[:end]
		s = s[end+len("endbfchar"):]

		// Parse pairs: <hex1> <hex2>
		parseBfCharSection(section, result)
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

		if len(src) >= 2 && len(dst) >= 2 {
			cid := binary.BigEndian.Uint16(src)
			uni := rune(binary.BigEndian.Uint16(dst))
			result[cid] = uni
		}
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
