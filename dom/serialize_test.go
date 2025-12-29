package dom

import (
	"testing"
)

func TestToHTMLVoidAndAttrs(t *testing.T) {
	node := &Node{
		Name: "img",
		Attrs: map[string]*string{
			"alt":      strPtr(`foo"bar`),
			"disabled": nil,
		},
	}

	got := ToHTML(node, false, 2)
	want := "<img alt='foo\"bar' disabled>"
	if got != want {
		t.Fatalf("ToHTML(img) = %q, want %q", got, want)
	}
}

func TestToHTMLVoidCaseInsensitive(t *testing.T) {
	node := &Node{
		Name: "IMG",
		Attrs: map[string]*string{
			"alt": strPtr("test"),
		},
	}

	got := ToHTML(node, false, 0)
	want := "<IMG alt=test>"
	if got != want {
		t.Fatalf("ToHTML(IMG) = %q, want %q", got, want)
	}
}

func TestToHTMLPrettyNested(t *testing.T) {
	root := &Node{
		Name: "div",
		Children: []*Node{
			{
				Name: "span",
				Children: []*Node{
					{Name: "#text", Data: " hi "},
				},
			},
			{Name: "#text", Data: "   "}, // trimmed away
		},
	}

	got := ToHTML(root, true, 2)
	want := "<div>\n  <span> hi </span>\n</div>"
	if got != want {
		t.Fatalf("pretty ToHTML(div) = %q, want %q", got, want)
	}
}

func TestToHTMLCompactKeepsText(t *testing.T) {
	root := &Node{
		Name: "div",
		Children: []*Node{
			{Name: "#text", Data: " hi "},
		},
	}

	got := ToHTML(root, false, 2)
	want := "<div> hi </div>"
	if got != want {
		t.Fatalf("compact ToHTML(div) = %q, want %q", got, want)
	}
}

func TestToHTMLDocumentRoot(t *testing.T) {
	doc := NewDocument()
	doc.Children = []*Node{
		{Name: "#comment", Data: "hi"},
		{
			Name: "p",
			Children: []*Node{
				{Name: "#text", Data: "ok"},
			},
		},
	}

	got := ToHTML(doc, true, 2)
	want := "<!--hi-->\n<p>ok</p>"
	if got != want {
		t.Fatalf("ToHTML(document) = %q, want %q", got, want)
	}
}

func TestToHTMLDocumentCompact(t *testing.T) {
	doc := NewDocument()
	doc.Children = []*Node{
		{Name: "p", Children: []*Node{{Name: "#text", Data: "one"}}},
		{Name: "p", Children: []*Node{{Name: "#text", Data: "two"}}},
	}
	if got := ToHTML(doc, false, 0); got != "<p>one</p><p>two</p>" {
		t.Fatalf("compact document ToHTML = %q", got)
	}
}

func TestToHTMLFragment(t *testing.T) {
	frag := &Node{
		Name: "#document-fragment",
		Children: []*Node{
			{Name: "#text", Data: "a"},
			{
				Name: "span",
				Children: []*Node{
					{Name: "#text", Data: " b"},
				},
			},
		},
	}

	got := ToHTML(frag, false, 2)
	want := "a<span> b</span>"
	if got != want {
		t.Fatalf("ToHTML(fragment) = %q, want %q", got, want)
	}
}

func TestToHTMLFragmentPrettyTrimsWhitespace(t *testing.T) {
	frag := &Node{
		Name: "#document-fragment",
		Children: []*Node{
			{Name: "#text", Data: "  spaced "},
			{Name: "#text", Data: "   "},
		},
	}

	got := ToHTML(frag, true, 2)
	want := "spaced"
	if got != want {
		t.Fatalf("pretty ToHTML(fragment) = %q, want %q", got, want)
	}
}

func TestToHTMLEscapesText(t *testing.T) {
	node := &Node{
		Name: "p",
		Children: []*Node{
			{Name: "#text", Data: "<&>"},
		},
	}

	got := ToHTML(node, false, 0)
	want := "<p>&lt;&amp;&gt;</p>"
	if got != want {
		t.Fatalf("ToHTML escapes text = %q, want %q", got, want)
	}
}

func TestEscapeTextEmptyAndNonVoid(t *testing.T) {
	if escapeText("") != "" {
		t.Fatalf("escapeText on empty should return empty")
	}
	if isVoidElement("div") {
		t.Fatalf("div should not be void element")
	}
	if isVoidElement("") {
		t.Fatalf("empty name should not be void element")
	}
}

func TestEscapeAttrValueEmpty(t *testing.T) {
	if got := escapeAttrValue("", '"'); got != "" {
		t.Fatalf("escapeAttrValue on empty should return empty, got %q", got)
	}
}

func TestToHTMLAttributeQuoting(t *testing.T) {
	node := &Node{
		Name: "div",
		Attrs: map[string]*string{
			"single": strPtr("he said \"hi\""),
			"double": strPtr("it's ok"),
		},
	}

	got := ToHTML(node, false, 0)
	want := `<div double="it's ok" single='he said "hi"'></div>`
	if got != want {
		t.Fatalf("ToHTML attribute quoting = %q, want %q", got, want)
	}
}

func TestToHTMLAttributeUnquotedAndBoolean(t *testing.T) {
	node := &Node{
		Name: "input",
		Attrs: map[string]*string{
			"checked": nil,
			"id":      strPtr("x1"),
			"value":   strPtr("fish&chips"),
		},
	}

	got := ToHTML(node, false, 0)
	want := `<input checked id=x1 value=fish&amp;chips>`
	if got != want {
		t.Fatalf("ToHTML attribute variants = %q, want %q", got, want)
	}
}

func TestToHTMLSortedAttributes(t *testing.T) {
	node := &Node{
		Name: "div",
		Attrs: map[string]*string{
			"b": strPtr("2"),
			"a": strPtr("1"),
			"c": strPtr("3"),
		},
	}

	got := ToHTML(node, false, 0)
	want := `<div a=1 b=2 c=3></div>`
	if got != want {
		t.Fatalf("ToHTML sorted attrs = %q, want %q", got, want)
	}
}

func TestToHTMLHandlesEmptyAndNil(t *testing.T) {
	if got := ToHTML(nil, false, 2); got != "" {
		t.Fatalf("ToHTML(nil) = %q, want empty", got)
	}

	doc := NewDocument()
	if got := ToHTML(doc, false, 2); got != "" {
		t.Fatalf("ToHTML(empty doc) = %q, want empty", got)
	}
}

func TestToHTMLDoctypeAndComments(t *testing.T) {
	node := &Node{
		Name: "#document",
		Children: []*Node{
			{Name: "!doctype"},
			{Name: "#comment", Data: "hi"},
			{Name: "html"},
		},
	}

	got := ToHTML(node, true, 0)
	want := "<!DOCTYPE html>\n<!--hi-->\n<html></html>"
	if got != want {
		t.Fatalf("ToHTML doc with doctype = %q, want %q", got, want)
	}
}

func TestToHTMLAllTextPrettyConcats(t *testing.T) {
	node := &Node{
		Name: "p",
		Children: []*Node{
			{Name: "#text", Data: "hello"},
			{Name: "#text", Data: " world"},
		},
	}

	got := ToHTML(node, true, 2)
	want := "<p>hello world</p>"
	if got != want {
		t.Fatalf("pretty ToHTML(text run) = %q, want %q", got, want)
	}
}
