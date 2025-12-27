package encoding

import (
	"bytes"
	"strings"
	"unicode/utf16"
)

// Package encoding will handle HTML transport decoding and charset sniffing.

type Result struct {
	HTML     string
	Encoding string
}

// DecodeHTML takes raw bytes and an optional
// transport-supplied encoding and return decoded HTML plus the chosen encoding.
func DecodeHTML(input []byte, transportEncoding string) Result {
	data := input

	encName, bomLen := sniffBOM(data)
	if encName != "" && bomLen > 0 {
		data = data[bomLen:]
		transportEncoding = encName
	}

	if encName == "" {
		if resolved := normalizeEncodingLabel(transportEncoding); resolved != "" {
			encName = resolved
			transportEncoding = strings.TrimSpace(strings.ToLower(transportEncoding))
		}
	}

	if encName == "" {
		if resolved := sniffMetaEncoding(data); resolved != "" {
			encName = resolved
			transportEncoding = resolved
		}
	}

	if encName == "" {
		encName = "windows-1252"
		transportEncoding = encName
	}

	html := decodeBytes(data, encName)
	return Result{HTML: html, Encoding: transportEncoding}
}

func sniffBOM(data []byte) (string, int) {
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return "utf-8", 3
	}
	if len(data) >= 2 {
		if data[0] == 0xFE && data[1] == 0xFF {
			return "utf-16be", 2
		}
		if data[0] == 0xFF && data[1] == 0xFE {
			return "utf-16le", 2
		}
	}
	return "", 0
}

func sniffMetaEncoding(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	limit := min(len(data), 1024)

	sample := data[:limit]
	lower := bytes.ToLower(sample)
	offset := 0
	for {
		idx := bytes.Index(lower[offset:], []byte("<meta"))
		if idx == -1 {
			return ""
		}
		idx += offset

		rest := lower[idx+len("<meta"):]
		gt := bytes.IndexByte(rest, '>')
		if gt == -1 {
			return ""
		}
		end := idx + len("<meta") + gt

		tagOrig := sample[idx:end]
		tagLower := lower[idx:end]

		if charset := extractCharset(tagOrig, tagLower); charset != "" {
			if normalized := normalizeEncodingLabel(charset); normalized != "" {
				return normalized
			}
		}

		offset = end + 1
		if offset >= len(sample) {
			return ""
		}
	}
}

func extractCharset(tagOrig, tagLower []byte) string {
	if idx := bytes.Index(tagLower, []byte("charset=")); idx != -1 {
		if val := parseAttrValue(tagOrig, idx+len("charset=")); val != "" {
			return val
		}
	}

	if contentIdx := bytes.Index(tagLower, []byte("content=")); contentIdx != -1 {
		contentVal := parseAttrValue(tagOrig, contentIdx+len("content="))
		if contentVal == "" {
			return ""
		}
		contentLower := strings.ToLower(contentVal)
		if ci := strings.Index(contentLower, "charset="); ci != -1 {
			value := contentLower[ci+len("charset="):]
			for i, r := range value {
				if r == '"' || r == '\'' || r == ';' || r == ' ' || r == '\t' || r == '\n' || r == '\r' {
					value = value[:i]
					break
				}
			}
			value = strings.TrimSpace(value)
			if value != "" {
				return value
			}
		}
	}

	return ""
}

func parseAttrValue(tag []byte, start int) string {
	if start >= len(tag) {
		return ""
	}

	if tag[start] == '"' || tag[start] == '\'' {
		quote := tag[start]
		start++
		if start >= len(tag) {
			return ""
		}
		if end := bytes.IndexByte(tag[start:], quote); end != -1 {
			return strings.TrimSpace(string(tag[start : start+end]))
		}
		return strings.TrimSpace(string(tag[start:]))
	}

	end := start
	for end < len(tag) && tag[end] > ' ' && tag[end] != '>' {
		end++
	}

	return strings.TrimSpace(string(tag[start:end]))
}

func normalizeEncodingLabel(label string) string {
	label = strings.TrimSpace(strings.ToLower(label))
	switch label {
	case "":
		return ""
	case "utf-8", "utf8", "unicode-1-1-utf-8":
		return "utf-8"
	case "utf-16", "utf-16le", "utf16le", "utf-16-le", "utf-16le-bom":
		return "utf-16le"
	case "utf-16be", "utf16be", "utf-16-be", "utf-16be-bom":
		return "utf-16be"
	case "iso-8859-1", "iso_8859-1", "latin-1", "latin1", "latin 1", "l1", "windows-1252", "cp1252", "us-ascii", "ascii":
		return "windows-1252"
	default:
		return ""
	}
}

func decodeBytes(data []byte, encoding string) string {
	switch encoding {
	case "utf-8":
		return string(data)
	case "utf-16le":
		return decodeUTF16(data, true)
	case "utf-16be":
		return decodeUTF16(data, false)
	default: // fallback to windows-1252
		return decodeWindows1252(data)
	}
}

func decodeUTF16(data []byte, littleEndian bool) string {
	if len(data) < 2 {
		return ""
	}

	u16 := make([]uint16, 0, len(data)/2)
	for i := 0; i+1 < len(data); i += 2 {
		if littleEndian {
			u16 = append(u16, uint16(data[i])|uint16(data[i+1])<<8)
		} else {
			u16 = append(u16, uint16(data[i])<<8|uint16(data[i+1]))
		}
	}

	return string(utf16.Decode(u16))
}

func decodeWindows1252(data []byte) string {
	runes := make([]rune, len(data))
	for i, b := range data {
		runes[i] = windows1252Rune(b)
	}
	return string(runes)
}

func windows1252Rune(b byte) rune {
	switch {
	case b < 0x80:
		return rune(b)
	case b >= 0xA0:
		return rune(b)
	default:
		return windows1252Table[b-0x80]
	}
}

var windows1252Table = [...]rune{
	0x20AC, 0x0081, 0x201A, 0x0192, 0x201E, 0x2026, 0x2020, 0x2021,
	0x02C6, 0x2030, 0x0160, 0x2039, 0x0152, 0x008D, 0x017D, 0x008F,
	0x0090, 0x2018, 0x2019, 0x201C, 0x201D, 0x2022, 0x2013, 0x2014,
	0x02DC, 0x2122, 0x0161, 0x203A, 0x0153, 0x009D, 0x017E, 0x0178,
}
