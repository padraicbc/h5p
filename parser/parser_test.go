package parser

import (
	"strings"
	"testing"

	"github.com/padraicbc/h5p/dom"
	"golang.org/x/net/html"
)

func TestParseAcceptsDifferentInputs(t *testing.T) {
	docFromString, err := Parse("<p>ok</p>")
	if err != nil {
		t.Fatalf("parse string: %v", err)
	}
	if docFromString.Root == nil || len(docFromString.Root.Children) == 0 {
		t.Fatalf("parse string returned empty document")
	}

	docFromBytes, err := Parse([]byte("<p>ok</p>"), WithEncoding("latin-1"))
	if err != nil {
		t.Fatalf("parse bytes: %v", err)
	}
	if docFromBytes.Encoding != "latin-1" {
		t.Fatalf("Parse bytes encoding = %q, want latin-1", docFromBytes.Encoding)
	}

	docFromReader, err := Parse(strings.NewReader("<p>ok</p>"), WithEncoding("latin-1"))
	if err != nil {
		t.Fatalf("parse reader: %v", err)
	}
	if docFromReader.Encoding != "latin-1" {
		t.Fatalf("Parse reader encoding = %q, want latin-1", docFromReader.Encoding)
	}
}

func TestToTextMatchesDomTraversal(t *testing.T) {
	doc, err := Parse("<div> hello <span>world</span> friend </div>")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	text := doc.ToText(" ", true)
	if text != "hello world friend" {
		t.Fatalf("ToText(strip=true) = %q, want %q", text, "hello world friend")
	}

	rawText := doc.ToText("|", false)
	if rawText != " hello |world| friend " {
		t.Fatalf("ToText(strip=false) = %q, want %q", rawText, " hello |world| friend ")
	}
}

func TestQueryHelperUsesSelectorPackage(t *testing.T) {
	doc, err := Parse("<div id='main'><span class='a b'>x</span><p>y</p></div>")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	matches := doc.Root.Query("div > span.a")
	if len(matches) != 1 {
		t.Fatalf("Query returned %d matches, want 1", len(matches))
	}
	if matches[0].Name != "span" || matches[0].Attr("class") != "a b" {
		t.Fatalf("unexpected match %+v", matches[0])
	}
}

func TestParseFragmentContextIsAttached(t *testing.T) {
	ctx := &FragmentContext{TagName: "template", Namespace: ""}
	doc, err := Parse("<span>inner</span>", WithFragmentContext(ctx))
	if err != nil {
		t.Fatalf("parse fragment: %v", err)
	}
	if doc.Root == nil || len(doc.Root.Children) != 1 {
		t.Fatalf("fragment parse missing wrapper elements")
	}
	wrapper := doc.Root.Children[0]
	if wrapper.Name != "template" {
		t.Fatalf("fragment wrapper name = %q, want template", wrapper.Name)
	}
	// fragment context should hang children off the fake context element
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
}

// Ensure the dom helper still works if Root is nil (e.g. failed parse).
func TestHTPHelpersHandleNilRoot(t *testing.T) {
	var doc *HTP
	if got := doc.ToHTML(false, 2); got != "" {
		t.Fatalf("nil doc ToHTML = %q, want empty", got)
	}
	if got := doc.ToText(" ", true); got != "" {
		t.Fatalf("nil doc ToText = %q, want empty", got)
	}
	if got := doc.ToMarkdown(); got != "" {
		t.Fatalf("nil doc ToMarkdown = %q, want empty", got)
	}
	doc = &HTP{}
	if res := doc.Root.Query("div"); res != nil {
		t.Fatalf("nil root Query = %#v, want nil", res)
	}
}

// Sanity check that DOM append helper wires parents when building trees manually.
func TestAppendChildSetsParent(t *testing.T) {
	parent := &dom.Node{Name: "div"}
	child := &dom.Node{Name: "span"}
	parent.AppendChild(child)

	if child.Parent != parent {
		t.Fatalf("expected child parent to be set")
	}
	if len(parent.Children) != 1 || parent.Children[0] != child {
		t.Fatalf("unexpected children slice: %+v", parent.Children)
	}
}

func TestParseCoversNilAndDefaultTypes(t *testing.T) {
	tests := []struct {
		name string
		in   any
	}{
		{name: "nil input", in: nil},
		{name: "numeric default fmt", in: 123},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
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
	doc := &HTP{}
	if doc.ToHTML(false, 2) != "" {
		t.Fatalf("ToHTML on nil root should be empty")
	}
	if doc.ToText(" ", true) != "" {
		t.Fatalf("ToText on nil root should be empty")
	}
	if doc.ToMarkdown() != "" {
		t.Fatalf("ToMarkdown on nil root should be empty")
	}
}

func TestHTPToMarkdown(t *testing.T) {
	doc, err := Parse("<p><strong>bold</strong></p>")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if got := doc.ToMarkdown(); got == "" {
		t.Fatalf("ToMarkdown should return content")
	}
}

func TestConvertHTMLNodeCoversNodeTypes(t *testing.T) {
	src := "<!doctype html><!--note--><template><span>inner</span></template><div id=\"x\">text</div>"
	doc, err := Parse(src)
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
}

func TestConvertHTMLNodeUnknownType(t *testing.T) {
	node := &html.Node{Type: html.ErrorNode}
	if got := convertHTMLNode(node); got != nil {
		t.Fatalf("expected nil for unknown node type, got %#v", got)
	}
}

func TestParseMalformedInput(t *testing.T) {
	doc, err := Parse("<div></span>")
	if err != nil {
		t.Fatalf("expected no error return for malformed input: %v", err)
	}
	if doc == nil {
		t.Fatalf("expected document even with parse issues")
	}
}
