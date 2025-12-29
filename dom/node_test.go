package dom

import (
	"strings"
	"testing"
)

func TestAttrCaseInsensitiveAndNil(t *testing.T) {
	n := &Node{Attrs: map[string]*string{
		"ID":   strPtr("main"),
		"data": nil,
	}}

	if got := n.Attr("id"); got != "main" {
		t.Fatalf("Attr(id) = %q, want main", got)
	}
	if got := n.Attr("ID"); got != "main" {
		t.Fatalf("Attr(ID) = %q, want main", got)
	}
	if got := n.Attr("data"); got != "" {
		t.Fatalf("Attr(data) = %q, want empty string for nil attribute", got)
	}
}

func TestToTextRespectsSeparatorAndStrip(t *testing.T) {
	root := &Node{Name: "div", Children: []*Node{
		{Name: "#text", Data: " hello "},
		{Name: "span", Children: []*Node{{Name: "#text", Data: " world "}}},
		{Name: "#text", Data: "  ! "},
	}}

	gotStrip := root.ToText(" ", true)
	if gotStrip != "hello world !" {
		t.Fatalf("ToText(strip=true) = %q, want %q", gotStrip, "hello world !")
	}

	gotKeep := root.ToText("-", false)
	if gotKeep != " hello - world -  ! " {
		t.Fatalf("ToText(strip=false) = %q, want %q", gotKeep, " hello - world -  ! ")
	}
}

func TestToMarkdownRendersCommonElements(t *testing.T) {
	doc := &Node{Name: "document"}

	h1 := &Node{Name: "h1"}
	h1.AppendChild(&Node{Name: "#text", Data: "Title"})
	doc.AppendChild(h1)

	p := &Node{Name: "p"}
	p.AppendChild(&Node{Name: "#text", Data: "Hello "})
	p.AppendChild(&Node{Name: "strong", Children: []*Node{{Name: "#text", Data: "bold"}}})
	p.AppendChild(&Node{Name: "#text", Data: " and "})
	p.AppendChild(&Node{Name: "em", Children: []*Node{{Name: "#text", Data: "italic"}}})
	p.AppendChild(&Node{Name: "#text", Data: " with "})
	p.AppendChild(&Node{Name: "code", Children: []*Node{{Name: "#text", Data: "code"}}})
	p.AppendChild(&Node{Name: "#text", Data: " and "})
	p.AppendChild(&Node{Name: "a", Attrs: map[string]*string{"href": strPtr("https://example.com")}, Children: []*Node{{Name: "#text", Data: "link"}}})
	doc.AppendChild(p)

	ul := &Node{Name: "ul"}
	li1 := &Node{Name: "li"}
	li1.AppendChild(&Node{Name: "#text", Data: "first"})
	li2 := &Node{Name: "li"}
	li2.AppendChild(&Node{Name: "#text", Data: "second"})
	ul.AppendChild(li1)
	ul.AppendChild(li2)
	doc.AppendChild(ul)

	ol := &Node{Name: "ol"}
	ol1 := &Node{Name: "li"}
	ol1.AppendChild(&Node{Name: "#text", Data: "one"})
	ol2 := &Node{Name: "li"}
	ol2.AppendChild(&Node{Name: "#text", Data: "two"})
	ol.AppendChild(ol1)
	ol.AppendChild(ol2)
	doc.AppendChild(ol)

	md := doc.ToMarkdown()
	if !strings.Contains(md, "# Title") {
		t.Fatalf("markdown missing heading: %q", md)
	}
	if !strings.Contains(md, "Hello **bold** and *italic* with `code` and [link](https://example.com)") {
		t.Fatalf("markdown missing inline formatting: %q", md)
	}
	if !strings.Contains(md, "- first") || !strings.Contains(md, "- second") {
		t.Fatalf("markdown missing unordered list items: %q", md)
	}
	if !strings.Contains(md, "1. one") || !strings.Contains(md, "2. two") {
		t.Fatalf("markdown missing ordered list items: %q", md)
	}
}

func TestToMarkdownHandlesNilReceiver(t *testing.T) {
	var n *Node
	if got := n.ToMarkdown(); got != "" {
		t.Fatalf("nil receiver ToMarkdown = %q, want empty", got)
	}
}

func TestMatchesEmptySelectorPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic on empty selector")
		}
	}()
	Matches(&Node{Name: "div"}, "   ")
}

func TestMatchChainErrorPropagation(t *testing.T) {
	node := &Node{Name: "div", Attrs: map[string]*string{"id": strPtr("x")}}
	chain := selectorChain{
		{compound: compoundSelector{attrs: []attrSelector{{name: "id", op: "!="}}}},
	}
	if ok, err := matchChain(node, chain); err == nil || ok {
		t.Fatalf("matchChain should bubble unsupported operator error, got ok=%v err=%v", ok, err)
	}
}

func TestMatchPseudoUnsupportedAndErrors(t *testing.T) {
	if _, err := matchPseudo(&Node{Name: "div"}, pseudoSelector{name: "hover"}); err == nil {
		t.Fatalf("unsupported pseudo should return error")
	}
	if _, err := matchPseudo(&Node{Name: "div"}, pseudoSelector{name: "nth-child"}); err == nil {
		t.Fatalf("nth-child without arg should error")
	}
	if _, err := matchPseudo(&Node{Name: "div"}, pseudoSelector{name: "nth-of-type"}); err == nil {
		t.Fatalf("nth-of-type without arg should error")
	}
	if _, err := matchPseudo(&Node{Name: "div"}, pseudoSelector{name: "contains"}); err == nil {
		t.Fatalf(":contains without arg should error")
	}
	if ok, err := matchPseudo(&Node{Name: "div"}, pseudoSelector{name: "not", arg: "["}); !ok || err != nil {
		t.Fatalf(":not with invalid selector should ignore and return true, got %v %v", ok, err)
	}

	// :root must NOT match non-document nodes
	root := &Node{Name: "div"}
	root.AppendChild(&Node{Name: "span"})
	if ok, _ := matchPseudo(root, pseudoSelector{name: "root"}); ok {
		t.Fatalf(":root must not match non-root element")
	}

	// :root MUST match html under document
	doc := &Node{Name: "document"}
	html := &Node{Name: "html"}
	doc.AppendChild(html)

	if ok, err := matchPseudo(html, pseudoSelector{name: "root"}); !ok || err != nil {
		t.Fatalf(":root should match html element under document, got %v %v", ok, err)
	}
}

func TestMatchesNilNodeReturnsFalse(t *testing.T) {
	if Matches(nil, "div") {
		t.Fatalf("Matches(nil) should be false")
	}
}

func TestSelectorErrorImplementsError(t *testing.T) {
	err := SelectorError{Msg: "boom"}
	if err.Error() != "boom" {
		t.Fatalf("SelectorError.Error = %q, want boom", err.Error())
	}
}

func TestAppendChildHandlesNil(t *testing.T) {
	parent := &Node{Name: "div"}
	parent.AppendChild(nil)
	if len(parent.Children) != 0 {
		t.Fatalf("expected no children when appending nil")
	}
	var nilParent *Node
	child := &Node{Name: "span"}
	nilParent.AppendChild(child) // should not panic
}

func TestAttrMissingReturnsEmpty(t *testing.T) {
	n := &Node{Attrs: map[string]*string{"class": strPtr("foo")}}
	if got := n.Attr("id"); got != "" {
		t.Fatalf("missing attr should be empty, got %q", got)
	}
}

func TestPreviousElementSiblingEdges(t *testing.T) {
	if prev := previousElementSibling(nil); prev != nil {
		t.Fatalf("nil node previous sibling should be nil")
	}
	parent := &Node{Name: "div"}
	first := &Node{Name: "p"}
	second := &Node{Name: "span"}
	parent.AppendChild(first)
	parent.AppendChild(second)
	if previousElementSibling(first) != nil {
		t.Fatalf("first child should have nil previous sibling")
	}
	if previousElementSibling(second) != first {
		t.Fatalf("expected first as previous sibling, got %#v", previousElementSibling(second))
	}
}

func TestListIndexAndChildTextContentEdges(t *testing.T) {
	if listIndex(nil) != 0 {
		t.Fatalf("nil listIndex should be 0")
	}
	n := &Node{Name: "li"}
	if listIndex(n) != 0 {
		t.Fatalf("listIndex without parent should be 0")
	}
	parent := &Node{Name: "ul", Children: []*Node{{Name: "li"}}}
	if listIndex(&Node{Name: "li", Parent: parent}) != 0 {
		t.Fatalf("listIndex for node not in parent children should be 0")
	}
	if got := childTextContent(nil); got != "" {
		t.Fatalf("childTextContent(nil) = %q, want empty", got)
	}
	container := &Node{Name: "div", Children: []*Node{
		{Name: "span", Children: []*Node{{Name: "#text", Data: "deep"}}},
	}}
	if got := childTextContent(container); got != "deep" {
		t.Fatalf("childTextContent with nested text = %q, want deep", got)
	}
}

func TestParseSelectorErrors(t *testing.T) {
	if _, err := parseSelector(""); err == nil {
		t.Fatalf("parseSelector should error on empty selector")
	}
	if _, err := parseSelector("div, "); err == nil {
		t.Fatalf("parseSelector should error on empty group")
	}
	if _, err := parseSelector("[]"); err == nil {
		t.Fatalf("parseSelector should error on missing attribute name")
	}
	if _, err := parseSelector("[attr=]"); err == nil {
		t.Fatalf("parseSelector should error on missing attribute value")
	}
	if _, err := parseSelector("div >"); err == nil {
		t.Fatalf("parseSelector should error on trailing combinator")
	}
	if _, err := parseSelector("> div"); err == nil {
		t.Fatalf("parseSelector should error on leading combinator")
	}
	if _, err := parseSelector("div > > p"); err == nil {
		t.Fatalf("parseSelector should error on double combinator")
	}
}

func TestNthExpressionParsers(t *testing.T) {
	if ok, err := matchNthExpression(0, "1"); ok || err != nil {
		t.Fatalf("pos<=0 should be false nil, got %v %v", ok, err)
	}
	if _, err := matchNthExpression(1, ""); err == nil {
		t.Fatalf("empty nth expression should error")
	}
	if _, err := matchNthExpression(1, "abc"); err == nil {
		t.Fatalf("invalid nth expression should error")
	}
	if ok, err := matchNthExpression(3, "0n+3"); !ok || err != nil {
		t.Fatalf("0n+3 should match pos 3, got %v %v", ok, err)
	}
	if ok, err := matchNthExpression(2, "2n+3"); ok || err != nil {
		t.Fatalf("2n+3 should not match pos 2, got %v %v", ok, err)
	}
	if ok, err := matchNthExpression(4, "2n"); !ok || err != nil {
		t.Fatalf("2n should match pos 4, got %v %v", ok, err)
	}
	if ok, err := matchNthExpression(1, "even"); ok || err != nil {
		t.Fatalf("even should not match pos 1, got %v %v", ok, err)
	}
	if ok, err := matchNthExpression(4, "2n"); !ok || err != nil {
		t.Fatalf("2n should match pos 4, got %v %v", ok, err)
	}
	if _, _, ok := parseAnPlusB("abc"); ok {
		t.Fatalf("parseAnPlusB without n should fail")
	}
	if _, err := parseInt(" "); err == nil {
		t.Fatalf("parseInt on empty should error")
	}
	if a, b, ok := parseAnPlusB("xn+1"); ok || a != 0 || b != 0 {
		t.Fatalf("parseAnPlusB with invalid coefficient should fail, got %d %d %v", a, b, ok)
	}
	if a, b, ok := parseAnPlusB("n"); !ok || a != 1 || b != 0 {
		t.Fatalf("parseAnPlusB(\"n\") = %d,%d,%v, want 1,0,true", a, b, ok)
	}
	if a, b, ok := parseAnPlusB("-n+2"); !ok || a != -1 || b != 2 {
		t.Fatalf("parseAnPlusB(\"-n+2\") = %d,%d,%v, want -1,2,true", a, b, ok)
	}
}

func TestTextAndAllTextHelpers(t *testing.T) {
	root := &Node{Name: "div"}
	textOnly := &Node{Name: "#text", Data: "hello"}
	innerText := &Node{Name: "#text", Data: "world"}
	child := &Node{Name: "span", Children: []*Node{innerText}}
	root.AppendChild(textOnly)
	root.AppendChild(child)

	if got := root.Text(); got != "hello" {
		t.Fatalf("Text() = %q, want hello", got)
	}
	if got := (&Node{Name: "div"}).Text(); got != "" {
		t.Fatalf("Text() with no children = %q, want empty", got)
	}
	if got := textOnly.Text(); got != "hello" {
		t.Fatalf("Text on text node = %q, want hello", got)
	}
	if got := root.AllText(); got != "helloworld" {
		t.Fatalf("AllText() = %q, want helloworld", got)
	}
	if got := (&Node{Name: "div"}).AllText(); got != "" {
		t.Fatalf("AllText() with no children = %q, want empty", got)
	}
}

func TestQueryFirstSearchesTemplate(t *testing.T) {
	tmplChild := &Node{Name: "span", Attrs: map[string]*string{"id": strPtr("inside")}}
	host := &Node{Name: "template"}
	host.Template = &Node{Name: "#document-fragment", Children: []*Node{tmplChild}}
	root := &Node{Name: "div", Children: []*Node{host}}

	found, _ := root.QueryFirst("#inside")
	if found != tmplChild {
		t.Fatalf("expected to find template child, got %#v", found)
	}
}

func TestNodeToHTMLDelegates(t *testing.T) {
	node := &Node{Name: "p", Children: []*Node{{Name: "#text", Data: "ok"}}}
	if got := node.ToHTML(false, 2); got != "<p>ok</p>" {
		t.Fatalf("ToHTML wrapper = %q, want <p>ok</p>", got)
	}
}

func TestRenderMarkdownSpecialCases(t *testing.T) {
	doc := &Node{Name: "document"}
	doc.AppendChild(&Node{Name: "br"})
	doc.AppendChild(&Node{Name: "#comment", Data: "ignore"})
	doc.AppendChild(&Node{Name: "custom", Children: []*Node{
		{Name: "#text", Data: "inside"},
	}})
	md := doc.ToMarkdown()
	if !strings.Contains(md, "  \ninside") {
		t.Fatalf("expected markdown to include br and nested text, got %q", md)
	}

	// ordered list item with missing parent linkage should still render (defaults to 1)
	li := &Node{Name: "li", Children: []*Node{{Name: "#text", Data: "orphan"}}}
	renderListItem(li, &strings.Builder{}, "", true)
}
