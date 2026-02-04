package encoding

import (
	"bytes"
	"testing"

	"unicode/utf16"

	"github.com/padraicbc/h5p/internal/common"
)

func TestDecodeHTMLUsesBOMUTF16LE(t *testing.T) {
	cases := []struct {
		name         string
		text         string
		wantEncoding string
		wantHTML     string
	}{
		{
			name:         "decodes utf-16le bom",
			text:         "<p>ok</p>",
			wantEncoding: "utf-16le",
			wantHTML:     "<p>ok</p>",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			u16 := utf16.Encode([]rune(tc.text))
			data := []byte{0xFF, 0xFE} // UTF-16LE BOM
			for _, v := range u16 {
				data = append(data, byte(v), byte(v>>8))
			}

			res := DecodeHTML(data, "")
			if res.Encoding != tc.wantEncoding {
				t.Fatalf("encoding = %q, want %q", res.Encoding, tc.wantEncoding)
			}
			if res.HTML != tc.wantHTML {
				t.Fatalf("decoded HTML = %q, want %q", res.HTML, tc.wantHTML)
			}

		})
	}

}

func TestDecodeHTMLUsesBOMUTF8OverridesTransport(t *testing.T) {
	cases := []struct {
		name         string
		text         string
		transportEnc string
		wantEncoding string
		wantHTML     string
	}{
		{
			name:         "utf-8 bom overrides transport encoding",
			text:         "hi",
			transportEnc: "latin-1",
			wantEncoding: "utf-8",
			wantHTML:     "hi",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			data := append([]byte{0xEF, 0xBB, 0xBF}, []byte(tc.text)...)

			res := DecodeHTML(data, tc.transportEnc)
			if res.Encoding != tc.wantEncoding {
				t.Fatalf("encoding = %q, want %q", res.Encoding, tc.wantEncoding)
			}
			if res.HTML != tc.wantHTML {
				t.Fatalf("decoded HTML = %q, want %q", res.HTML, tc.wantHTML)
			}

		})
	}

}

func TestDecodeHTMLUsesUTF16BE(t *testing.T) {
	cases := []struct {
		name         string
		text         string
		wantEncoding string
		wantHTML     string
	}{
		{
			name:         "decodes utf-16be bom",
			text:         "yo",
			wantEncoding: "utf-16be",
			wantHTML:     "yo",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			u16 := utf16.Encode([]rune(tc.text))
			data := []byte{0xFE, 0xFF}
			for _, v := range u16 {
				data = append(data, byte(v>>8), byte(v))
			}

			res := DecodeHTML(data, "")
			if res.Encoding != tc.wantEncoding {
				t.Fatalf("encoding = %q, want %q", res.Encoding, tc.wantEncoding)
			}
			if res.HTML != tc.wantHTML {
				t.Fatalf("decoded HTML = %q, want %q", res.HTML, tc.wantHTML)
			}

		})
	}

}

func TestDecodeHTMLTransportEncodingNormalized(t *testing.T) {
	cases := []struct {
		name         string
		data         []byte
		transportEnc string
		wantEncoding string
		wantHTML     string
	}{
		{
			name:         "latin-1 maps windows-1252 bytes",
			data:         []byte{0x80}, // maps to Euro sign in windows-1252
			transportEnc: "latin-1",
			wantEncoding: "latin-1",
			wantHTML:     "\u20AC",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			res := DecodeHTML(tc.data, tc.transportEnc)
			if res.Encoding != tc.wantEncoding {
				t.Fatalf("encoding = %q, want %q", res.Encoding, tc.wantEncoding)
			}
			if res.HTML != tc.wantHTML {
				t.Fatalf("decoded HTML = %q, want %q", res.HTML, tc.wantHTML)
			}

		})
	}

}

func TestDecodeHTMLMetaCharset(t *testing.T) {
	cases := []struct {
		name         string
		html         []byte
		wantEncoding string
		wantHTML     string
	}{
		{
			name:         "meta charset takes effect",
			html:         []byte(`<html><head><meta charset="utf-8"></head><body>hi</body></html>`),
			wantEncoding: "utf-8",
			wantHTML:     `<html><head><meta charset="utf-8"></head><body>hi</body></html>`,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			res := DecodeHTML(tc.html, "")
			if res.Encoding != tc.wantEncoding {
				t.Fatalf("encoding = %q, want %q", res.Encoding, tc.wantEncoding)
			}
			if res.HTML != tc.wantHTML {
				t.Fatalf("decoded HTML mismatch; got %q", res.HTML)
			}

		})
	}

}

func TestDecodeHTMLMetaContentCharset(t *testing.T) {
	cases := []struct {
		name         string
		html         []byte
		wantEncoding []string
		wantHTML     string
	}{
		{
			name:         "meta content-type charset",
			html:         []byte(`<meta http-equiv="Content-Type" content="text/html; charset=iso-8859-1"><p>ok</p>`),
			wantEncoding: []string{"iso-8859-1", "windows-1252"},
			wantHTML:     `<meta http-equiv="Content-Type" content="text/html; charset=iso-8859-1"><p>ok</p>`,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			res := DecodeHTML(tc.html, "")
			matched := false
			for _, enc := range tc.wantEncoding {
				if res.Encoding == enc {
					matched = true
					break
				}
			}
			if !matched {
				t.Fatalf("encoding = %q, want one of %v", res.Encoding, tc.wantEncoding)
			}
			if res.HTML != tc.wantHTML {
				t.Fatalf("decoded HTML mismatch; got %q", res.HTML)
			}

		})
	}

}

func TestDecodeHTMLFallbackWindows1252(t *testing.T) {
	cases := []struct {
		name         string
		html         []byte
		wantEncoding string
		wantHTML     string
	}{
		{
			name:         "fallback encoding is windows-1252",
			html:         []byte("<p>fallback</p>"),
			wantEncoding: "windows-1252",
			wantHTML:     "<p>fallback</p>",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			res := DecodeHTML(tc.html, "")
			if res.Encoding != tc.wantEncoding {
				t.Fatalf("encoding = %q, want %q", res.Encoding, tc.wantEncoding)
			}
			if res.HTML != tc.wantHTML {
				t.Fatalf("decoded HTML mismatch; got %q", res.HTML)
			}

		})
	}

}

func TestSniffMetaSearchesSubsequentTags(t *testing.T) {
	cases := []struct {
		name string
		html []byte
		want string
	}{
		{
			name: "finds charset after empty meta",
			html: []byte(`<head><meta><meta charset="utf-8"></head>`),
			want: "utf-8",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			if got := sniffMetaEncoding(tc.html); got != tc.want {
				t.Fatalf("sniffMetaEncoding = %q, want %q", got, tc.want)
			}

		})
	}

}

func TestSniffMetaEmptyAndMalformedTags(t *testing.T) {
	cases := []struct {
		name string
		html []byte
		want string
	}{
		{
			name: "empty input returns empty",
			html: nil,
			want: "",
		},
		{
			name: "missing closing bracket returns empty",
			html: []byte("<meta charset=\"utf-8\""),
			want: "",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			if got := sniffMetaEncoding(tc.html); got != tc.want {
				t.Fatalf("sniffMetaEncoding = %q, want %q", got, tc.want)
			}

		})
	}

}

func TestSniffMetaOffsetReturn(t *testing.T) {
	cases := []struct {
		name string
		html []byte
		want string
	}{
		{
			name: "lone meta returns empty",
			html: []byte("<meta>"),
			want: "",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			if got := sniffMetaEncoding(tc.html); got != tc.want {
				t.Fatalf("sniffMetaEncoding = %q, want %q", got, tc.want)
			}

		})
	}

}

func TestExtractCharsetContentBranch(t *testing.T) {
	cases := []struct {
		name string
		tag  []byte
		want string
	}{
		{
			name: "extracts charset from content",
			tag:  []byte(`<meta charset= content="text/html; charset=utf-8"`),
			want: "utf-8",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			tagLower := bytes.ToLower(tc.tag)
			if got := extractCharset(tc.tag, tagLower); got != tc.want {
				t.Fatalf("extractCharset = %q, want %q", got, tc.want)
			}

		})
	}

}

func TestExtractCharsetContentValueEmpty(t *testing.T) {
	cases := []struct {
		name string
		tag  []byte
		want string
	}{
		{
			name: "empty content value returns empty",
			tag:  []byte(`<meta content=>`),
			want: "",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			tagLower := bytes.ToLower(tc.tag)
			if got := extractCharset(tc.tag, tagLower); got != tc.want {
				t.Fatalf("extractCharset empty content value = %q, want %q", got, tc.want)
			}

		})
	}

}

func TestExtractCharsetDelimiterTrimming(t *testing.T) {
	cases := []struct {
		name string
		tag  []byte
		want string
	}{
		{
			name: "charset is trimmed before delimiter",
			tag:  []byte(`<meta charset= content="text/html; charset=utf-8; other=1"`),
			want: "utf-8",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			tagLower := bytes.ToLower(tc.tag)
			if got := extractCharset(tc.tag, tagLower); got != tc.want {
				t.Fatalf("extractCharset with delimiter = %q, want %q", got, tc.want)
			}

		})
	}

}

func TestSniffMetaRespects1kLimit(t *testing.T) {
	cases := []struct {
		name string
		pad  int
		want string
	}{
		{
			name: "meta after 1k is ignored",
			pad:  1040,
			want: "",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			padding := bytes.Repeat([]byte("a"), tc.pad)
			html := append(padding, []byte(`<meta charset="utf-8">`)...)
			if got := sniffMetaEncoding(html); got != tc.want {
				t.Fatalf("sniffMetaEncoding with late meta = %q, want %q", got, tc.want)
			}

		})
	}

}

func TestNormalizeEncodingLabelVariants(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty label", in: "", want: ""},
		{name: "utf8 normalizes", in: "UTF8", want: "utf-8"},
		{name: "utf-16 normalizes to le", in: "utf-16", want: "utf-16le"},
		{name: "utf-16be preserved", in: "utf-16be", want: "utf-16be"},
		{name: "ascii maps to windows-1252", in: "ascii", want: "windows-1252"},
		{name: "unknown label empty", in: "unknown-xx", want: ""},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			if got := normalizeEncodingLabel(tc.in); got != tc.want {
				t.Fatalf("normalizeEncodingLabel(%q) = %q, want %q", tc.in, got, tc.want)
			}

		})
	}

}

func TestParseAttrValuePaths(t *testing.T) {
	cases := []struct {
		name   string
		data   []byte
		offset int
		want   string
	}{
		{
			name:   "unquoted value",
			data:   []byte(`charset=utf-8>`),
			offset: len("charset="),
			want:   "utf-8",
		},
		{
			name:   "unterminated double quote",
			data:   []byte(`charset="utf-8`),
			offset: len("charset="),
			want:   "utf-8",
		},
		{
			name:   "single quoted value",
			data:   []byte(`charset='utf-8'>`),
			offset: len("charset="),
			want:   "utf-8",
		},
		{
			name:   "empty value",
			data:   []byte(`charset=`),
			offset: len("charset="),
			want:   "",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			if v := parseAttrValue(tc.data, tc.offset); v != tc.want {
				t.Fatalf("parseAttrValue = %q, want %q", v, tc.want)
			}

		})
	}

}

func TestDecodeUTF16AndWindows1252Helpers(t *testing.T) {
	cases := []common.TestCase{
		{
			Name: "decodeUTF16 short returns empty",
			Run: func(t *testing.T) {
				if got := decodeUTF16([]byte{0x00}, true); got != "" {
					t.Fatalf("decodeUTF16 short = %q, want empty", got)
				}
			},
		},
		{
			Name: "decodeUTF16 big endian",
			Run: func(t *testing.T) {
				if got := decodeUTF16([]byte{0x00, 0x48, 0x00, 0x69}, false); got != "Hi" {
					t.Fatalf("decodeUTF16 big endian = %q, want Hi", got)
				}
			},
		},
		{
			Name: "decodeWindows1252 maps bytes",
			Run: func(t *testing.T) {
				win := decodeWindows1252([]byte{0x80, 0x82, 0xA0})
				if win != "\u20AC\u201A\u00A0" {
					t.Fatalf("decodeWindows1252 = %q, want Euro+201A+NBSP", win)
				}
			},
		},
		{
			Name: "windows1252Rune preserves ASCII",
			Run: func(t *testing.T) {
				if r := windows1252Rune(0x7F); r != rune(0x7F) {
					t.Fatalf("windows1252Rune ASCII = %U, want 0x7F", r)
				}
			},
		},
		{
			Name: "parseAttrValue lone quote returns empty",
			Run: func(t *testing.T) {
				if v := parseAttrValue([]byte{'"'}, 0); v != "" {
					t.Fatalf("parseAttrValue lone quote = %q, want empty", v)
				}
			},
		},
	}
	common.RunTestCases(t, cases)

}
