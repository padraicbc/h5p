package parser

import (
	"strings"
	"testing"

	"github.com/padraicbc/h5p/internal/common"

	"github.com/padraicbc/h5p/dom"
	"golang.org/x/net/html"
)

func TestParseAcceptsDifferentInputs(t *testing.T) {
	cases := []struct {
		name         string
		input        any
		opts         []Option
		wantEncoding string
		wantRoot     bool
	}{
		{
			name:     "string input produces root",
			input:    "<p>ok</p>",
			wantRoot: true,
		},
		{
			name:         "byte input honors encoding option",
			input:        []byte("<p>ok</p>"),
			opts:         []Option{WithEncoding("latin-1")},
			wantEncoding: "latin-1",
			wantRoot:     true,
		},
		{
			name:         "reader input honors encoding option",
			input:        strings.NewReader("<p>ok</p>"),
			opts:         []Option{WithEncoding("latin-1")},
			wantEncoding: "latin-1",
			wantRoot:     true,
		},
	}

	for _, tc := range cases {

		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			doc, err := Parse(tc.input, tc.opts...)
			if err != nil {
				t.Fatalf("Parse returned error: %v", err)
			}
			if tc.wantRoot {
				if doc.Root == nil || len(doc.Root.Children) == 0 {
					t.Fatalf("Parse returned empty document")
				}
			}
			if tc.wantEncoding != "" && doc.Encoding != tc.wantEncoding {
				t.Fatalf("Parse encoding = %q, want %q", doc.Encoding, tc.wantEncoding)
			}

		})
	}

}

func TestToTextMatchesDomTraversal(t *testing.T) {
	cases := []struct {
		name     string
		html     string
		sep      string
		strip    bool
		wantText string
	}{
		{
			name:     "striped text collapses whitespace",
			html:     "<div> hello <span>world</span> friend </div>",
			sep:      " ",
			strip:    true,
			wantText: "hello world friend",
		},
		{
			name:     "unstripped preserves separators",
			html:     "<div> hello <span>world</span> friend </div>",
			sep:      "|",
			strip:    false,
			wantText: " hello |world| friend ",
		},
	}

	for _, tc := range cases {

		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			doc, err := Parse(tc.html)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			if text := doc.ToText(tc.sep, tc.strip); text != tc.wantText {
				t.Fatalf("ToText(strip=%v) = %q, want %q", tc.strip, text, tc.wantText)
			}

		})
	}

}

func TestQueryHelperUsesSelectorPackage(t *testing.T) {
	cases := []struct {
		name string
		html string
		sel  string
	}{
		{
			name: "selector query matches span",
			html: "<div id='main'><span class='a b'>x</span><p>y</p></div>",
			sel:  "div > span.a",
		},
	}

	for _, tc := range cases {

		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			doc, err := Parse(tc.html)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			matches, _ := doc.Root.Query(tc.sel)
			if len(matches) != 1 {
				t.Fatalf("Query returned %d matches, want 1", len(matches))
			}
			if matches[0].Name != "span" || matches[0].Attr("class") != "a b" {
				t.Fatalf("unexpected match %+v", matches[0])
			}

		})
	}

}

func TestParseFragmentContextIsAttached(t *testing.T) {
	cases := []struct {
		name    string
		ctx     *FragmentContext
		html    string
		wrapper string
	}{
		{
			name:    "fragment context attaches wrapper",
			ctx:     &FragmentContext{TagName: "template", Namespace: ""},
			html:    "<span>inner</span>",
			wrapper: "template",
		},
	}

	for _, tc := range cases {

		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			doc, err := Parse(tc.html, WithFragmentContext(tc.ctx))
			if err != nil {
				t.Fatalf("Parse fragment error: %v", err)
			}
			if doc.Root == nil || len(doc.Root.Children) != 1 {
				t.Fatalf("fragment parse missing wrapper elements")
			}
			wrapper := doc.Root.Children[0]
			if wrapper.Name != tc.wrapper {
				t.Fatalf("fragment wrapper name = %q, want %q", wrapper.Name, tc.wrapper)
			}
			if len(wrapper.Children) == 0 {
				t.Fatalf("fragment wrapper missing children")
			}

			foundSpan := false
			for _, child := range wrapper.Children {
				if child.Name == "span" && child.ToText("", false) == "inner" {
					foundSpan = true
				}
			}
			if !foundSpan {
				t.Fatalf("fragment children did not include parsed span: %+v", wrapper.Children)
			}

		})
	}

}

func TestHTPHelpersHandleNilRoot(t *testing.T) {
	cases := []struct {
		name string
		doc  *HTP
		want string
	}{
		{name: "nil doc ToHTML", doc: nil, want: ""},
		{name: "nil doc ToText", doc: nil, want: ""},
		{name: "nil doc ToMarkdown", doc: nil, want: ""},
		{name: "empty doc ToHTML", doc: &HTP{}, want: ""},
		{name: "empty doc ToText", doc: &HTP{}, want: ""},
		{name: "empty doc ToMarkdown", doc: &HTP{}, want: ""},
	}

	for _, tc := range cases {

		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			switch tc.name {
			case "nil doc ToHTML", "empty doc ToHTML":
				if got := tc.doc.ToHTML(false, 2); got != tc.want {
					t.Fatalf("ToHTML = %q, want empty", got)
				}
			case "nil doc ToText", "empty doc ToText":
				if got := tc.doc.ToText(" ", true); got != tc.want {
					t.Fatalf("ToText = %q, want empty", got)
				}
			case "nil doc ToMarkdown", "empty doc ToMarkdown":
				if got := tc.doc.ToMarkdown(); got != tc.want {
					t.Fatalf("ToMarkdown = %q, want empty", got)
				}
			default:
				t.Fatalf("unhandled case %q", tc.name)
			}

		})
	}

	queryCases := []struct {
		name string
		doc  *HTP
	}{
		{name: "nil root Query returns nil", doc: &HTP{}},
	}

	for _, tc := range queryCases {

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if res, _ := tc.doc.Root.Query("div"); res != nil {
				t.Fatalf("nil root Query = %#v, want nil", res)
			}
		})
	}
}

func TestAppendChildSetsParent(t *testing.T) {
	cases := []struct {
		name string
	}{
		{name: "append child sets parent and slice"},
	}

	for _, tc := range cases {

		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			parent := &dom.Node{Name: "div"}
			child := &dom.Node{Name: "span"}
			parent.AppendChild(child)

			if child.Parent != parent {
				t.Fatalf("expected child parent to be set")
			}
			if len(parent.Children) != 1 || parent.Children[0] != child {
				t.Fatalf("unexpected children slice: %+v", parent.Children)
			}

		})
	}

}

func TestParseCoversNilAndDefaultTypes(t *testing.T) {
	cases := []struct {
		name string
		in   any
	}{
		{name: "nil input", in: nil},
		{name: "numeric default fmt", in: 123},
	}

	for _, tc := range cases {

		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			doc, err := Parse(tc.in)
			if err != nil {
				t.Fatalf("Parse(%v) returned error: %v", tc.in, err)
			}
			if doc == nil || doc.Root == nil {
				t.Fatalf("Parse(%v) returned nil document", tc.in)
			}

		})
	}

}

func TestHTPNilRootHelpers(t *testing.T) {
	cases := []common.TestCase{
		{
			Name: "ToHTML on nil root returns empty",
			Run: func(t *testing.T) {
				doc := &HTP{}
				if doc.ToHTML(false, 2) != "" {
					t.Fatalf("ToHTML on nil root should be empty")
				}
			},
		},
		{
			Name: "ToText on nil root returns empty",
			Run: func(t *testing.T) {
				doc := &HTP{}
				if doc.ToText(" ", true) != "" {
					t.Fatalf("ToText on nil root should be empty")
				}
			},
		},
		{
			Name: "ToMarkdown on nil root returns empty",
			Run: func(t *testing.T) {
				doc := &HTP{}
				if doc.ToMarkdown() != "" {
					t.Fatalf("ToMarkdown on nil root should be empty")
				}
			},
		},
	}
	common.RunTestCases(t, cases)

}

func TestHTPToMarkdown(t *testing.T) {
	cases := []struct {
		name string
		html string
	}{
		{name: "markdown returns content", html: "<p><strong>bold</strong></p>"},
	}

	for _, tc := range cases {

		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			doc, err := Parse(tc.html)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}
			if got := doc.ToMarkdown(); got == "" {
				t.Fatalf("ToMarkdown should return content")
			}

		})
	}

}

func TestConvertHTMLNodeCoversNodeTypes(t *testing.T) {
	cases := []struct {
		name string
		html string
	}{
		{
			name: "converts doctype comment template div",
			html: "<!doctype html><!--note--><template><span>inner</span></template><div id=\"x\">text</div>",
		},
	}

	for _, tc := range cases {

		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			doc, err := Parse(tc.html)
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}
			if doc.Root == nil || len(doc.Root.Children) == 0 {
				t.Fatalf("expected parsed document root")
			}

			var (
				foundDoctype  bool
				foundComment  bool
				foundTemplate bool
				foundDiv      bool
			)

			var walk func(*dom.Node)
			walk = func(n *dom.Node) {
				if n == nil {
					return
				}
				switch n.Name {
				case "!doctype":
					foundDoctype = true
				case "#comment":
					foundComment = true
				case "template":
					foundTemplate = true
					if n.Template == nil || len(n.Template.Children) == 0 {
						t.Fatalf("template should preserve fragment children")
					}
				case "div":
					foundDiv = true
					if n.Attr("id") != "x" {
						t.Fatalf("expected div id attribute to be preserved")
					}
				}
				for _, child := range n.Children {
					walk(child)
				}
				if n.Template != nil {
					walk(n.Template)
				}
			}

			walk(doc.Root)

			if !foundDoctype || !foundComment || !foundTemplate || !foundDiv {
				t.Fatalf("expected doctype=%v comment=%v template=%v div=%v", foundDoctype, foundComment, foundTemplate, foundDiv)
			}

		})
	}

}

func TestConvertHTMLNodeUnknownType(t *testing.T) {
	cases := []struct {
		name string
		node *html.Node
	}{
		{name: "unknown node type returns nil", node: &html.Node{Type: html.ErrorNode}},
	}

	for _, tc := range cases {

		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			if got := convertHTMLNode(tc.node); got != nil {
				t.Fatalf("expected nil for unknown node type, got %#v", got)
			}

		})
	}

}

func TestParseMalformedInput(t *testing.T) {
	cases := []struct {
		name string
		html string
	}{
		{name: "malformed input still returns document", html: "<div></span>"},
	}

	for _, tc := range cases {

		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			doc, err := Parse(tc.html)
			if err != nil {
				t.Fatalf("expected no error return for malformed input: %v", err)
			}
			if doc == nil {
				t.Fatalf("expected document even with parse issues")
			}

		})
	}

}
