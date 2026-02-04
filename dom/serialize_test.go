package dom

import (
	"testing"

	"github.com/padraicbc/h5p/internal/common"
)

func TestToHTMLVoidAndAttrs(t *testing.T) {
	cases := []struct {
		name   string
		node   *Node
		pretty bool
		indent int
		want   string
	}{
		{
			name: "void element renders attrs and boolean",
			node: &Node{
				Name: "img",
				Attrs: map[string]*string{
					"alt":      strPtr(`foo"bar`),
					"disabled": nil,
				},
			},
			pretty: false,
			indent: 2,
			want:   "<img alt='foo\"bar' disabled>",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			if got := ToHTML(tc.node, tc.pretty, tc.indent); got != tc.want {
				t.Fatalf("ToHTML(img) = %q, want %q", got, tc.want)
			}

		})
	}

}

func TestToHTMLVoidCaseInsensitive(t *testing.T) {
	cases := []struct {
		name string
		node *Node
		want string
	}{
		{
			name: "void tag is case-insensitive",
			node: &Node{
				Name: "IMG",
				Attrs: map[string]*string{
					"alt": strPtr("test"),
				},
			},
			want: "<IMG alt=test>",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			if got := ToHTML(tc.node, false, 0); got != tc.want {
				t.Fatalf("ToHTML(IMG) = %q, want %q", got, tc.want)
			}

		})
	}

}

func TestToHTMLPrettyNested(t *testing.T) {
	cases := []struct {
		name string
		node *Node
		want string
	}{
		{
			name: "pretty printing trims whitespace text nodes",
			node: &Node{
				Name: "div",
				Children: []*Node{
					{
						Name:     "span",
						Children: []*Node{{Name: "#text", Data: " hi "}},
					},
					{Name: "#text", Data: "   "},
				},
			},
			want: "<div>\n  <span> hi </span>\n</div>",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			if got := ToHTML(tc.node, true, 2); got != tc.want {
				t.Fatalf("pretty ToHTML(div) = %q, want %q", got, tc.want)
			}

		})
	}

}

func TestToHTMLCompactKeepsText(t *testing.T) {
	cases := []struct {
		name string
		node *Node
		want string
	}{
		{
			name: "compact keeps text",
			node: &Node{
				Name:     "div",
				Children: []*Node{{Name: "#text", Data: " hi "}},
			},
			want: "<div> hi </div>",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			if got := ToHTML(tc.node, false, 2); got != tc.want {
				t.Fatalf("compact ToHTML(div) = %q, want %q", got, tc.want)
			}

		})
	}

}

func TestToHTMLDocumentRoot(t *testing.T) {
	cases := []struct {
		name string
		doc  *Node
		want string
	}{
		{
			name: "document root renders children",
			doc: func() *Node {
				doc := NewDocument()
				doc.Children = []*Node{
					{Name: "#comment", Data: "hi"},
					{Name: "p", Children: []*Node{{Name: "#text", Data: "ok"}}},
				}
				return doc
			}(),
			want: "<!--hi-->\n<p>ok</p>",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			if got := ToHTML(tc.doc, true, 2); got != tc.want {
				t.Fatalf("ToHTML(document) = %q, want %q", got, tc.want)
			}

		})
	}

}

func TestToHTMLDocumentCompact(t *testing.T) {
	cases := []struct {
		name string
		doc  *Node
		want string
	}{
		{
			name: "compact document is concatenated",
			doc: func() *Node {
				doc := NewDocument()
				doc.Children = []*Node{
					{Name: "p", Children: []*Node{{Name: "#text", Data: "one"}}},
					{Name: "p", Children: []*Node{{Name: "#text", Data: "two"}}},
				}
				return doc
			}(),
			want: "<p>one</p><p>two</p>",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			if got := ToHTML(tc.doc, false, 0); got != tc.want {
				t.Fatalf("compact document ToHTML = %q", got)
			}

		})
	}

}

func TestToHTMLFragment(t *testing.T) {
	cases := []struct {
		name string
		node *Node
		want string
	}{
		{
			name: "document fragment renders children",
			node: &Node{
				Name: "#document-fragment",
				Children: []*Node{
					{Name: "#text", Data: "a"},
					{Name: "span", Children: []*Node{{Name: "#text", Data: " b"}}},
				},
			},
			want: "a<span> b</span>",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			if got := ToHTML(tc.node, false, 2); got != tc.want {
				t.Fatalf("ToHTML(fragment) = %q, want %q", got, tc.want)
			}

		})
	}

}

func TestToHTMLFragmentPrettyTrimsWhitespace(t *testing.T) {
	cases := []struct {
		name string
		node *Node
		want string
	}{
		{
			name: "pretty fragment trims whitespace",
			node: &Node{
				Name: "#document-fragment",
				Children: []*Node{
					{Name: "#text", Data: "  spaced "},
					{Name: "#text", Data: "   "},
				},
			},
			want: "spaced",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			if got := ToHTML(tc.node, true, 2); got != tc.want {
				t.Fatalf("pretty ToHTML(fragment) = %q, want %q", got, tc.want)
			}

		})
	}

}

func TestToHTMLEscapesText(t *testing.T) {
	cases := []struct {
		name string
		node *Node
		want string
	}{
		{
			name: "text nodes are escaped",
			node: &Node{
				Name:     "p",
				Children: []*Node{{Name: "#text", Data: "<&>"}},
			},
			want: "<p>&lt;&amp;&gt;</p>",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			if got := ToHTML(tc.node, false, 0); got != tc.want {
				t.Fatalf("ToHTML escapes text = %q, want %q", got, tc.want)
			}

		})
	}

}

func TestEscapeTextEmptyAndNonVoid(t *testing.T) {
	cases := []common.TestCase{
		{
			Name: "escapeText empty returns empty",
			Run: func(t *testing.T) {
				if escapeText("") != "" {
					t.Fatalf("escapeText on empty should return empty")
				}
			},
		},
		{
			Name: "div is not void element",
			Run: func(t *testing.T) {
				if isVoidElement("div") {
					t.Fatalf("div should not be void element")
				}
			},
		},
		{
			Name: "empty name is not void element",
			Run: func(t *testing.T) {
				if isVoidElement("") {
					t.Fatalf("empty name should not be void element")
				}
			},
		},
	}
	common.RunTestCases(t, cases)

}

func TestEscapeAttrValueEmpty(t *testing.T) {
	cases := []struct {
		name string
	}{
		{name: "escapeAttrValue empty returns empty"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			if got := escapeAttrValue("", '"'); got != "" {
				t.Fatalf("escapeAttrValue on empty should return empty, got %q", got)
			}

		})
	}

}

func TestToHTMLAttributeQuoting(t *testing.T) {
	cases := []struct {
		name string
		node *Node
		want string
	}{
		{
			name: "attribute quoting chooses best quotes",
			node: &Node{
				Name: "div",
				Attrs: map[string]*string{
					"single": strPtr("he said \"hi\""),
					"double": strPtr("it's ok"),
				},
			},
			want: `<div double="it's ok" single='he said "hi"'></div>`,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			if got := ToHTML(tc.node, false, 0); got != tc.want {
				t.Fatalf("ToHTML attribute quoting = %q, want %q", got, tc.want)
			}

		})
	}

}

func TestToHTMLAttributeUnquotedAndBoolean(t *testing.T) {
	cases := []struct {
		name string
		node *Node
		want string
	}{
		{
			name: "unquoted and boolean attributes",
			node: &Node{
				Name: "input",
				Attrs: map[string]*string{
					"checked": nil,
					"id":      strPtr("x1"),
					"value":   strPtr("fish&chips"),
				},
			},
			want: `<input checked id=x1 value=fish&amp;chips>`,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			if got := ToHTML(tc.node, false, 0); got != tc.want {
				t.Fatalf("ToHTML attribute variants = %q, want %q", got, tc.want)
			}

		})
	}

}

func TestToHTMLSortedAttributes(t *testing.T) {
	cases := []struct {
		name string
		node *Node
		want string
	}{
		{
			name: "attributes are sorted",
			node: &Node{
				Name: "div",
				Attrs: map[string]*string{
					"b": strPtr("2"),
					"a": strPtr("1"),
					"c": strPtr("3"),
				},
			},
			want: `<div a=1 b=2 c=3></div>`,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			if got := ToHTML(tc.node, false, 0); got != tc.want {
				t.Fatalf("ToHTML sorted attrs = %q, want %q", got, tc.want)
			}

		})
	}

}

func TestToHTMLHandlesEmptyAndNil(t *testing.T) {
	cases := []struct {
		name string
		node *Node
		want string
	}{
		{name: "nil node returns empty", node: nil, want: ""},
		{name: "empty document returns empty", node: NewDocument(), want: ""},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			if got := ToHTML(tc.node, false, 2); got != tc.want {
				t.Fatalf("ToHTML(%s) = %q, want %q", tc.name, got, tc.want)
			}

		})
	}

}

func TestToHTMLDoctypeAndComments(t *testing.T) {
	cases := []struct {
		name string
		node *Node
		want string
	}{
		{
			name: "doctype and comments render",
			node: &Node{
				Name: "#document",
				Children: []*Node{
					{Name: "!doctype"},
					{Name: "#comment", Data: "hi"},
					{Name: "html"},
				},
			},
			want: "<!DOCTYPE html>\n<!--hi-->\n<html></html>",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			if got := ToHTML(tc.node, true, 0); got != tc.want {
				t.Fatalf("ToHTML doc with doctype = %q, want %q", got, tc.want)
			}

		})
	}

}

func TestToHTMLAllTextPrettyConcats(t *testing.T) {
	cases := []struct {
		name string
		node *Node
		want string
	}{
		{
			name: "pretty concatenates adjacent text",
			node: &Node{
				Name: "p",
				Children: []*Node{
					{Name: "#text", Data: "hello"},
					{Name: "#text", Data: " world"},
				},
			},
			want: "<p>hello world</p>",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			if got := ToHTML(tc.node, true, 2); got != tc.want {
				t.Fatalf("pretty ToHTML(text run) = %q, want %q", got, tc.want)
			}

		})
	}

}
