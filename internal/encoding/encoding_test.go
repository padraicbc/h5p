package encoding

import (
	"bytes"
	"testing"
	"unicode/utf16"
)

func TestDecodeHTMLUsesBOMUTF16LE(t *testing.T) {
	text := "<p>ok</p>"
	u16 := utf16.Encode([]rune(text))
	data := []byte{0xFF, 0xFE} // UTF-16LE BOM
	for _, v := range u16 {
		data = append(data, byte(v), byte(v>>8))
	}

	res := DecodeHTML(data, "")
	if res.Encoding != "utf-16le" {
		t.Fatalf("encoding = %q, want utf-16le", res.Encoding)
	}
	if res.HTML != text {
		t.Fatalf("decoded HTML = %q, want %q", res.HTML, text)
	}
}

func TestDecodeHTMLUsesBOMUTF8OverridesTransport(t *testing.T) {
	text := "hi"
	data := append([]byte{0xEF, 0xBB, 0xBF}, []byte(text)...)

	res := DecodeHTML(data, "latin-1")
	if res.Encoding != "utf-8" {
		t.Fatalf("encoding = %q, want utf-8", res.Encoding)
	}
	if res.HTML != text {
		t.Fatalf("decoded HTML = %q, want %q", res.HTML, text)
	}
}

func TestDecodeHTMLUsesUTF16BE(t *testing.T) {
	text := "yo"
	u16 := utf16.Encode([]rune(text))
	data := []byte{0xFE, 0xFF}
	for _, v := range u16 {
		data = append(data, byte(v>>8), byte(v))
	}

	res := DecodeHTML(data, "")
	if res.Encoding != "utf-16be" {
		t.Fatalf("encoding = %q, want utf-16be", res.Encoding)
	}
	if res.HTML != text {
		t.Fatalf("decoded HTML = %q, want %q", res.HTML, text)
	}
}

func TestDecodeHTMLTransportEncodingNormalized(t *testing.T) {
	data := []byte{0x80} // maps to Euro sign in windows-1252
	res := DecodeHTML(data, "latin-1")
	if res.Encoding != "latin-1" {
		t.Fatalf("encoding = %q, want latin-1", res.Encoding)
	}
	if res.HTML != "\u20AC" {
		t.Fatalf("decoded HTML = %q, want euro sign", res.HTML)
	}
}

func TestDecodeHTMLMetaCharset(t *testing.T) {
	html := []byte(`<html><head><meta charset="utf-8"></head><body>hi</body></html>`)
	res := DecodeHTML(html, "")
	if res.Encoding != "utf-8" {
		t.Fatalf("encoding = %q, want utf-8", res.Encoding)
	}
	if res.HTML != string(html) {
		t.Fatalf("decoded HTML mismatch; got %q", res.HTML)
	}
}

func TestDecodeHTMLMetaContentCharset(t *testing.T) {
	html := []byte(`<meta http-equiv="Content-Type" content="text/html; charset=iso-8859-1"><p>ok</p>`)
	res := DecodeHTML(html, "")
	if res.Encoding != "iso-8859-1" && res.Encoding != "windows-1252" {
		t.Fatalf("encoding = %q, want iso-8859-1/windows-1252", res.Encoding)
	}
	if res.HTML != string(html) {
		t.Fatalf("decoded HTML mismatch; got %q", res.HTML)
	}
}

func TestDecodeHTMLFallbackWindows1252(t *testing.T) {
	html := []byte("<p>fallback</p>")
	res := DecodeHTML(html, "")
	if res.Encoding != "windows-1252" {
		t.Fatalf("encoding = %q, want windows-1252", res.Encoding)
	}
	if res.HTML != "<p>fallback</p>" {
		t.Fatalf("decoded HTML mismatch; got %q", res.HTML)
	}
}

func TestSniffMetaSearchesSubsequentTags(t *testing.T) {
	html := []byte(`<head><meta><meta charset="utf-8"></head>`)
	if got := sniffMetaEncoding(html); got != "utf-8" {
		t.Fatalf("sniffMetaEncoding = %q, want utf-8", got)
	}
}

func TestSniffMetaEmptyAndMalformedTags(t *testing.T) {
	if got := sniffMetaEncoding(nil); got != "" {
		t.Fatalf("sniff empty = %q, want empty", got)
	}
	if got := sniffMetaEncoding([]byte("<meta charset=\"utf-8\"")); got != "" {
		t.Fatalf("sniff missing > = %q, want empty", got)
	}
}

func TestSniffMetaOffsetReturn(t *testing.T) {
	if got := sniffMetaEncoding([]byte("<meta>")); got != "" {
		t.Fatalf("sniff lone meta = %q, want empty", got)
	}
}

func TestExtractCharsetContentBranch(t *testing.T) {
	tag := []byte(`<meta charset= content="text/html; charset=utf-8"`)
	tagLower := bytes.ToLower(tag)
	if got := extractCharset(tag, tagLower); got != "utf-8" {
		t.Fatalf("extractCharset content branch = %q, want utf-8", got)
	}
}

func TestExtractCharsetContentValueEmpty(t *testing.T) {
	tag := []byte(`<meta content=>`)
	tagLower := bytes.ToLower(tag)
	if got := extractCharset(tag, tagLower); got != "" {
		t.Fatalf("extractCharset empty content value = %q, want empty", got)
	}
}

func TestExtractCharsetDelimiterTrimming(t *testing.T) {
	tag := []byte(`<meta charset= content="text/html; charset=utf-8; other=1"`)
	tagLower := bytes.ToLower(tag)
	if got := extractCharset(tag, tagLower); got != "utf-8" {
		t.Fatalf("extractCharset with delimiter = %q, want utf-8", got)
	}
}

func TestSniffMetaRespects1kLimit(t *testing.T) {
	padding := bytes.Repeat([]byte("a"), 1040)
	html := append(padding, []byte(`<meta charset="utf-8">`)...)
	if got := sniffMetaEncoding(html); got != "" {
		t.Fatalf("sniffMetaEncoding with late meta = %q, want empty", got)
	}
}

func TestNormalizeEncodingLabelVariants(t *testing.T) {
	cases := map[string]string{
		"":           "",
		"UTF8":       "utf-8",
		"utf-16":     "utf-16le",
		"utf-16be":   "utf-16be",
		"ascii":      "windows-1252",
		"unknown-xx": "",
	}
	for in, want := range cases {
		if got := normalizeEncodingLabel(in); got != want {
			t.Fatalf("normalizeEncodingLabel(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestParseAttrValuePaths(t *testing.T) {
	if v := parseAttrValue([]byte(`charset=utf-8>`), len("charset=")); v != "utf-8" {
		t.Fatalf("unquoted attr value = %q, want utf-8", v)
	}
	if v := parseAttrValue([]byte(`charset="utf-8`), len("charset=")); v != "utf-8" {
		t.Fatalf("unterminated quoted attr = %q, want utf-8", v)
	}
	if v := parseAttrValue([]byte(`charset='utf-8'>`), len("charset=")); v != "utf-8" {
		t.Fatalf("single-quoted attr = %q, want utf-8", v)
	}
	if v := parseAttrValue([]byte(`charset=`), len("charset=")); v != "" {
		t.Fatalf("empty attr = %q, want empty", v)
	}
}

func TestDecodeUTF16AndWindows1252Helpers(t *testing.T) {
	if got := decodeUTF16([]byte{0x00}, true); got != "" {
		t.Fatalf("decodeUTF16 short = %q, want empty", got)
	}
	if got := decodeUTF16([]byte{0x00, 0x48, 0x00, 0x69}, false); got != "Hi" {
		t.Fatalf("decodeUTF16 big endian = %q, want Hi", got)
	}
	win := decodeWindows1252([]byte{0x80, 0x82, 0xA0})
	if win != "\u20AC\u201A\u00A0" {
		t.Fatalf("decodeWindows1252 = %q, want Euro+201A+NBSP", win)
	}
	if r := windows1252Rune(0x7F); r != rune(0x7F) {
		t.Fatalf("windows1252Rune ASCII = %U, want 0x7F", r)
	}
	if v := parseAttrValue([]byte{'"'}, 0); v != "" {
		t.Fatalf("parseAttrValue lone quote = %q, want empty", v)
	}
}
