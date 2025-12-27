package dom

import (
	"fmt"
	"strings"
	"testing"
)

func nodeWithClass(class string) *Node {
	doc := &Node{Name: "document"}

	html := &Node{Name: "html", Parent: doc}
	body := &Node{Name: "body", Parent: html}
	div := &Node{
		Name:   "div",
		Attrs:  map[string]*string{"class": &class},
		Parent: body,
	}

	doc.Children = []*Node{html}
	html.Children = []*Node{body}
	body.Children = []*Node{div}

	return doc
}

// el creates an element node.
func el(name string, attrs map[string]string, children ...*Node) *Node {
	n := &Node{
		Name:  name,
		Attrs: make(map[string]*string),
	}
	for k, v := range attrs {
		val := v
		n.Attrs[k] = &val
	}
	for _, c := range children {
		n.AppendChild(c)
	}
	return n
}

func buildNode(children ...*Node) *Node {
	doc := &Node{Name: "document"}
	html := &Node{Name: "html"}
	body := &Node{Name: "body"}

	doc.AppendChild(html)
	html.AppendChild(body)

	for _, c := range children {
		body.AppendChild(c)
	}
	return doc
}

// helper builders mirroring the original Python test fixtures
func buildFullSimpleDoc() *Node {
	doc := &Node{Name: "document"}
	html := &Node{Name: "html"}
	head := &Node{Name: "head"}
	title := &Node{Name: "title"}
	head.AppendChild(title)

	body := &Node{Name: "body"}
	main := &Node{Name: "div", Attrs: map[string]*string{"id": strPtr("main"), "class": strPtr("container")}}
	h1 := &Node{Name: "h1"}
	pIntro := &Node{Name: "p", Attrs: map[string]*string{"class": strPtr("intro first")}}
	pContent := &Node{Name: "p", Attrs: map[string]*string{"class": strPtr("content")}}
	ul := &Node{Name: "ul"}
	li1 := &Node{Name: "li"}
	li2 := &Node{Name: "li", Attrs: map[string]*string{"class": strPtr("special")}}
	li3 := &Node{Name: "li"}

	ul.AppendChild(li1)
	ul.AppendChild(li2)
	ul.AppendChild(li3)
	main.AppendChild(h1)
	main.AppendChild(pIntro)
	main.AppendChild(pContent)
	main.AppendChild(ul)

	sidebar := &Node{Name: "div", Attrs: map[string]*string{"id": strPtr("sidebar"), "class": strPtr("container secondary")}}
	anchor := &Node{Name: "a", Attrs: map[string]*string{"href": strPtr("http://example.com"), "data-id": strPtr("123")}}
	sidebar.AppendChild(anchor)

	body.AppendChild(main)
	body.AppendChild(sidebar)
	html.AppendChild(head)
	html.AppendChild(body)
	doc.AppendChild(html)
	return doc
}

func buildNestedDoc() *Node {
	doc := &Node{Name: "document"}
	html := &Node{Name: "html"}
	body := &Node{Name: "body"}
	a := &Node{Name: "div", Attrs: map[string]*string{"class": strPtr("a")}}
	b := &Node{Name: "div", Attrs: map[string]*string{"class": strPtr("b")}}
	c := &Node{Name: "div", Attrs: map[string]*string{"class": strPtr("c")}}
	span := &Node{Name: "span", Attrs: map[string]*string{"id": strPtr("deep")}}

	c.AppendChild(span)
	b.AppendChild(c)
	a.AppendChild(b)
	body.AppendChild(a)
	html.AppendChild(body)
	doc.AppendChild(html)
	return doc
}

func buildSiblingDoc() *Node {
	doc := &Node{Name: "document"}
	html := &Node{Name: "html"}
	body := &Node{Name: "body"}
	wrapper := &Node{Name: "div"}

	h1 := &Node{Name: "h1"}
	pFirst := &Node{Name: "p", Attrs: map[string]*string{"class": strPtr("first")}}
	pSecond := &Node{Name: "p", Attrs: map[string]*string{"class": strPtr("second")}}
	pThird := &Node{Name: "p", Attrs: map[string]*string{"class": strPtr("third")}}
	span := &Node{Name: "span"}
	pFourth := &Node{Name: "p", Attrs: map[string]*string{"class": strPtr("fourth")}}

	wrapper.AppendChild(h1)
	wrapper.AppendChild(pFirst)
	wrapper.AppendChild(pSecond)
	wrapper.AppendChild(pThird)
	wrapper.AppendChild(span)
	wrapper.AppendChild(pFourth)

	body.AppendChild(wrapper)
	html.AppendChild(body)
	doc.AppendChild(html)
	return doc
}

func buildEmptyAndRootDoc() *Node {
	doc := &Node{Name: "document"}
	html := &Node{Name: "html"}
	body := &Node{Name: "body"}

	empty := &Node{Name: "div", Attrs: map[string]*string{"class": strPtr("empty")}}
	whitespace := &Node{Name: "div", Attrs: map[string]*string{"class": strPtr("whitespace")}}
	whitespace.AppendChild(&Node{Name: "#text", Data: "   "})
	text := &Node{Name: "div", Attrs: map[string]*string{"class": strPtr("text")}}
	text.AppendChild(&Node{Name: "#text", Data: "content"})
	nested := &Node{Name: "div", Attrs: map[string]*string{"class": strPtr("nested")}}
	nested.AppendChild(&Node{Name: "span"})

	body.AppendChild(empty)
	body.AppendChild(whitespace)
	body.AppendChild(text)
	body.AppendChild(nested)
	html.AppendChild(body)
	doc.AppendChild(html)
	return doc
}

func buildLangDoc(lang string) *Node {
	doc := &Node{Name: "document"}
	html := &Node{Name: "html"}
	body := &Node{Name: "body"}
	p := &Node{Name: "p", Attrs: map[string]*string{"lang": strPtr(lang)}}
	body.AppendChild(p)
	html.AppendChild(body)
	doc.AppendChild(html)
	return doc
}

func buildContainsDoc() *Node {
	doc := &Node{Name: "document"}
	html := &Node{Name: "html"}
	body := &Node{Name: "body"}

	a := &Node{Name: "div", Attrs: map[string]*string{"id": strPtr("a")}}
	btnA := &Node{Name: "button"}
	btnA.AppendChild(&Node{Name: "#text", Data: "click me"})
	a.AppendChild(btnA)

	b := &Node{Name: "div", Attrs: map[string]*string{"id": strPtr("b")}}
	btnB := &Node{Name: "button"}
	btnB.AppendChild(&Node{Name: "#text", Data: "do not click"})
	b.AppendChild(btnB)

	c := &Node{Name: "div", Attrs: map[string]*string{"id": strPtr("c")}}
	span := &Node{Name: "span"}
	span.AppendChild(&Node{Name: "#text", Data: "click"})
	c.AppendChild(span)
	c.AppendChild(&Node{Name: "#text", Data: " me"})

	body.AppendChild(a)
	body.AppendChild(b)
	body.AppendChild(c)
	html.AppendChild(body)
	doc.AppendChild(html)
	return doc
}

func buildTree() *Node {
	root := &Node{Name: "document"}
	div := &Node{Name: "div", Attrs: map[string]*string{"id": strPtr("main"), "class": strPtr("foo bar"), "data-container": strPtr("main")}}
	span := &Node{Name: "span", Attrs: map[string]*string{"class": strPtr("foo bar")}}
	em := &Node{Name: "em", Attrs: map[string]*string{"data-role": strPtr("label")}}
	para := &Node{Name: "p"}
	text := &Node{Name: "#text", Data: "hello"}

	root.AppendChild(div)
	div.AppendChild(span)
	span.AppendChild(em)
	em.AppendChild(text)
	div.AppendChild(para)

	return root
}

func buildSimpleDoc() *Node {
	doc := &Node{Name: "document"}
	html := &Node{Name: "html"}
	head := &Node{Name: "head"}
	title := &Node{Name: "title"}
	head.AppendChild(title)
	body := &Node{Name: "body"}

	main := &Node{Name: "div", Attrs: map[string]*string{"id": strPtr("main"), "class": strPtr("container")}}
	h1 := &Node{Name: "h1"}
	pIntro := &Node{Name: "p", Attrs: map[string]*string{"class": strPtr("intro first")}}
	pContent := &Node{Name: "p", Attrs: map[string]*string{"class": strPtr("content")}}
	ul := &Node{Name: "ul"}
	li1 := &Node{Name: "li"}
	li2 := &Node{Name: "li", Attrs: map[string]*string{"class": strPtr("special")}}
	li3 := &Node{Name: "li"}

	ul.AppendChild(li1)
	ul.AppendChild(li2)
	ul.AppendChild(li3)
	main.AppendChild(h1)
	main.AppendChild(pIntro)
	main.AppendChild(pContent)
	main.AppendChild(ul)

	sidebar := &Node{Name: "div", Attrs: map[string]*string{"id": strPtr("sidebar"), "class": strPtr("container secondary")}}
	anchor := &Node{Name: "a", Attrs: map[string]*string{"href": strPtr("http://example.com"), "data-id": strPtr("123")}}
	sidebar.AppendChild(anchor)

	body.AppendChild(main)
	body.AppendChild(sidebar)
	html.AppendChild(head)
	html.AppendChild(body)
	doc.AppendChild(html)

	return doc
}

func TestAttributeEqualsSelector(t *testing.T) {
	doc := &Node{Name: "document"}
	html := &Node{Name: "html"}
	body := &Node{Name: "body"}
	img := &Node{Name: "img", Attrs: map[string]*string{"alt": strPtr("eISB Site Logo"), "src": strPtr("logo.png")}}

	doc.AppendChild(html)
	html.AppendChild(body)
	body.AppendChild(img)

	imgNodes := doc.Query("img")
	if len(imgNodes) != 1 {
		t.Fatalf("expected 1 img element, got %d", len(imgNodes))
	}

	matches := doc.Query(`img[alt="eISB Site Logo"]`)
	if len(matches) != 1 {
		t.Fatalf("expected 1 match for alt value, got %d (alt=%q)", len(matches), imgNodes[0].Attr("alt"))
	}
}

func TestQueryFirstHonorsCombinators(t *testing.T) {
	root := buildTree()

	if node := root.QueryFirst("div > span em"); node == nil || node.Name != "em" {
		t.Fatalf("QueryFirst descendant chain failed, got %#v", node)
	}
	if node := root.QueryFirst("div > em"); node != nil {
		t.Fatalf("child combinator should not match em without span ancestor, got %#v", node)
	}
}

func TestMatchesChecksAttributesAndClasses(t *testing.T) {
	root := buildTree()
	em := root.QueryFirst("em")
	if em == nil {
		t.Fatalf("expected to find <em> node")
	}

	if !Matches(em, "em[data-role='label']") {
		t.Fatalf("Matches should succeed for data attribute")
	}
	if !Matches(em.Parent, "span.foo.bar") {
		t.Fatalf("Matches should include all classes on span")
	}
	if Matches(em, "em[label]") {
		t.Fatalf("Matches should fail for missing attribute")
	}
}

func TestUniversalSelectorFindsElements(t *testing.T) {
	root := buildTree()
	nodes := root.Query("div *")
	if len(nodes) != 3 { // span, em, p
		t.Fatalf("universal selector matched %d nodes, want 3", len(nodes))
	}
}

func TestTagAndDescendantSelectors(t *testing.T) {
	root := buildTree()

	spanNodes := root.Query("span")
	if len(spanNodes) != 1 {
		t.Fatalf("tag selector matched %d nodes, want 1", len(spanNodes))
	}
	if spanNodes[0].Name != "span" {
		t.Fatalf("expected span node, got %s", spanNodes[0].Name)
	}

	deep := root.Query("div span em")
	if len(deep) != 1 || deep[0].Name != "em" {
		t.Fatalf("descendant selector should find em under div span, got %#v", deep)
	}
}

func TestIDSelector(t *testing.T) {
	root := buildTree()

	nodes := root.Query("#main")
	if len(nodes) != 1 {
		t.Fatalf("id selector matched %d nodes, want 1", len(nodes))
	}
	if nodes[0].Name != "div" {
		t.Fatalf("expected div node with id=main, got %s", nodes[0].Name)
	}
}

func TestClassSelectorMatchesMultiple(t *testing.T) {
	root := buildTree()

	nodes := root.Query(".foo")
	if len(nodes) != 2 { // div and span
		t.Fatalf("class selector matched %d nodes, want 2", len(nodes))
	}
}

func TestAttributePresenceSelector(t *testing.T) {
	root := buildTree()

	emNodes := root.Query("[data-role]")
	if len(emNodes) != 1 || emNodes[0].Name != "em" {
		t.Fatalf("attribute presence selector should find em with data-role, got %#v", emNodes)
	}
}

func TestAttributeIncludesSelector(t *testing.T) {
	root := &Node{Name: "document"}
	tagged := &Node{Name: "div", Attrs: map[string]*string{"data-tags": strPtr("alpha beta gamma")}}
	partial := &Node{Name: "div", Attrs: map[string]*string{"data-tags": strPtr("alphabetagamma")}}

	root.AppendChild(tagged)
	root.AppendChild(partial)

	matches := root.Query(`[data-tags~=beta]`)
	if len(matches) != 1 || matches[0] != tagged {
		t.Fatalf("[data-tags~=beta] should match only the whitespace-separated value, got %#v", matches)
	}
}

func TestAttributeDashPrefixSelector(t *testing.T) {
	root := &Node{Name: "document"}
	englishUS := &Node{Name: "p", Attrs: map[string]*string{"lang": strPtr("en-US")}}
	plainEnglish := &Node{Name: "p", Attrs: map[string]*string{"lang": strPtr("en")}}
	other := &Node{Name: "p", Attrs: map[string]*string{"lang": strPtr("eng-US")}}

	root.AppendChild(englishUS)
	root.AppendChild(plainEnglish)
	root.AppendChild(other)

	matches := root.Query(`[lang|="en"]`)
	if len(matches) != 2 || matches[0] != englishUS || matches[1] != plainEnglish {
		t.Fatalf("[lang|=\"en\"] should match exact or hyphenated values, got %#v", matches)
	}
}

func TestAttributeIncludesRequiresWhitespaceList(t *testing.T) {
	root := &Node{Name: "document"}
	hyphenated := &Node{Name: "p", Attrs: map[string]*string{"lang": strPtr("en-US")}}
	plain := &Node{Name: "p", Attrs: map[string]*string{"lang": strPtr("en")}}
	spaced := &Node{Name: "p", Attrs: map[string]*string{"lang": strPtr("en fr")}}

	root.AppendChild(hyphenated)
	root.AppendChild(plain)
	root.AppendChild(spaced)

	matches := root.Query(`p[lang~="en"]`)

	if len(matches) != 2 ||
		(matches[0] != plain && matches[1] != plain) ||
		(matches[0] != spaced && matches[1] != spaced) {
		t.Fatalf("[lang~=\"en\"] should match plain and spaced tokens, got %#v", matches)
	}
}

func TestAttributeValueWithSpaces(t *testing.T) {
	root := buildTree()

	nodes := root.Query(`div[class="foo bar"]`)
	if len(nodes) != 1 || nodes[0].Name != "div" {
		t.Fatalf("attribute value selector with spaces should find div, got %#v", nodes)
	}
}

func TestCompoundSelectorRequiresAllParts(t *testing.T) {
	root := buildTree()

	nodes := root.Query("div#main.foo.bar[data-container=main]")
	if len(nodes) != 1 || nodes[0].Name != "div" {
		t.Fatalf("compound selector should match div with id, classes, and attribute, got %#v", nodes)
	}
}

func TestDescendantCombinatorTraversesMultipleLevels(t *testing.T) {
	root := buildTree()

	nodes := root.Query("div em")
	if len(nodes) != 1 || nodes[0].Name != "em" {
		t.Fatalf("descendant combinator should find em under div via span, got %#v", nodes)
	}
}

func TestChildCombinatorMatchesDirectChildrenOnly(t *testing.T) {
	root := buildTree()

	children := root.Query("div > span")
	if len(children) != 1 || children[0].Name != "span" {
		t.Fatalf("child combinator should match direct span child of div, got %#v", children)
	}

	if nodes := root.Query("div>em"); len(nodes) != 0 {
		t.Fatalf("child combinator should not match grandchild em, got %#v", nodes)
	}
}

func TestAttributeSelectorsHandleQuotedAndUnquotedValues(t *testing.T) {
	root := buildTree()

	if nodes := root.Query("em[data-role=label]"); len(nodes) != 1 {
		t.Fatalf("unquoted attribute selector should find em, got %#v", nodes)
	}

	if nodes := root.Query(`em[data-role="label"]`); len(nodes) != 1 {
		t.Fatalf("double-quoted attribute selector should find em, got %#v", nodes)
	}

	if nodes := root.Query("em[data-role='label']"); len(nodes) != 1 {
		t.Fatalf("single-quoted attribute selector should find em, got %#v", nodes)
	}
}

func TestAttributePresenceMatchesCaseInsensitiveKeys(t *testing.T) {
	node := &Node{Name: "div", Attrs: map[string]*string{"DATA-Role": strPtr("button")}}
	root := &Node{Name: "document"}
	root.AppendChild(node)

	if matches := root.Query("[data-role]"); len(matches) != 1 || matches[0] != node {
		t.Fatalf("attribute presence should match case-insensitive keys, got %#v", matches)
	}
}

func TestClassSelectorRequiresEachClass(t *testing.T) {
	root := buildTree()

	if nodes := root.Query(".foo.bar"); len(nodes) != 2 {
		t.Fatalf("combined class selector should match div and span, got %d", len(nodes))
	}

	if nodes := root.Query(".foo.baz"); len(nodes) != 0 {
		t.Fatalf("nonexistent class should prevent match, got %#v", nodes)
	}
}

func TestQueryHandlesNilRootAndEmptySelector(t *testing.T) {
	var root *Node
	if nodes := root.Query("div"); nodes != nil {
		t.Fatalf("nil root should return nil result, got %#v", nodes)
	}

	doc := buildTree()
	if nodes := doc.Query("   "); nodes != nil {
		t.Fatalf("empty selector should return nil result, got %#v", nodes)
	}
}

func TestTagSelectorIsCaseInsensitive(t *testing.T) {
	root := buildTree()

	nodes := root.Query("DIV")
	if len(nodes) != 1 || nodes[0].Name != "div" {
		t.Fatalf("expected to match div using uppercase tag, got %#v", nodes)
	}
}

func TestTagSelectorNoMatch(t *testing.T) {
	root := buildTree()

	if nodes := root.Query("article"); len(nodes) != 0 {
		t.Fatalf("unknown tag should not match, got %#v", nodes)
	}
}

func TestIDSelectorIsCaseSensitive(t *testing.T) {
	root := buildTree()

	if nodes := root.Query("#MAIN"); len(nodes) != 0 {
		t.Fatalf("id selector should be case-sensitive, got %#v", nodes)
	}
}

func TestClassSelectorNoMatch(t *testing.T) {
	root := buildTree()

	if nodes := root.Query(".missing"); len(nodes) != 0 {
		t.Fatalf("nonexistent class should not match, got %#v", nodes)
	}
}

func TestAttributePresenceMatchesSpecificKeys(t *testing.T) {
	root := buildTree()

	if nodes := root.Query("[data-container]"); len(nodes) != 1 || nodes[0].Name != "div" {
		t.Fatalf("data-container presence should match div, got %#v", nodes)
	}
}

func TestAttributeExactQuotesAndMismatches(t *testing.T) {
	root := buildTree()

	if nodes := root.Query("[id='main']"); len(nodes) != 1 || nodes[0].Name != "div" {
		t.Fatalf("single-quoted attribute selector should match div, got %#v", nodes)
	}
	if nodes := root.Query("[id=wrong]"); len(nodes) != 0 {
		t.Fatalf("non-matching attribute value should not match, got %#v", nodes)
	}
}

func TestDescendantSelectorNoMatch(t *testing.T) {
	root := buildTree()

	if nodes := root.Query("span div"); len(nodes) != 0 {
		t.Fatalf("descendant selector should not match when structure is missing, got %#v", nodes)
	}
}

func TestChildCombinatorRequiresDirectDescendant(t *testing.T) {
	root := buildTree()

	if nodes := root.Query("div > em"); len(nodes) != 0 {
		t.Fatalf("child combinator should not match grandchild em, got %#v", nodes)
	}
	if nodes := root.Query("div > p"); len(nodes) != 1 || nodes[0].Name != "p" {
		t.Fatalf("child combinator should match direct p child, got %#v", nodes)
	}
}

func TestQueryFromSubtreeExcludesSelf(t *testing.T) {
	root := buildTree()
	span := root.QueryFirst("span")
	if span == nil {
		t.Fatalf("expected to find span node")
	}

	if nodes := span.Query("span"); len(nodes) != 0 {
		t.Fatalf("query from subtree should not include self, got %#v", nodes)
	}
}

func TestTagSelectorSupportsHyphenatedNames(t *testing.T) {
	root := &Node{Name: "document"}
	custom := &Node{Name: "my-element"}
	root.AppendChild(custom)

	if nodes := root.Query("my-element"); len(nodes) != 1 || nodes[0] != custom {
		t.Fatalf("hyphenated tag names should match, got %#v", nodes)
	}
}

func TestUniversalSelectorMatchesAllElements(t *testing.T) {
	root := buildSimpleDoc()
	all := root.Query("*")
	if len(all) != 14 {
		t.Fatalf("universal selector should match all elements, got %d", len(all))
	}
}

func TestUniversalSelectorInCompound(t *testing.T) {
	root := buildSimpleDoc()
	matches := root.Query("*.container")
	if len(matches) != 2 {
		t.Fatalf("compound universal selector should match both container divs, got %d", len(matches))
	}
}

func TestIDSelectorWithTagAndMismatchedTag(t *testing.T) {
	root := buildSimpleDoc()

	if nodes := root.Query("div#main"); len(nodes) != 1 {
		t.Fatalf("expected single div#main, got %#v", nodes)
	}
	if nodes := root.Query("span#main"); len(nodes) != 0 {
		t.Fatalf("span#main should not match, got %#v", nodes)
	}
}

func TestClassSelectorsWithTags(t *testing.T) {
	root := buildSimpleDoc()

	if nodes := root.Query(".container"); len(nodes) != 2 {
		t.Fatalf(".container should match both divs, got %d", len(nodes))
	}
	if nodes := root.Query(".intro"); len(nodes) != 1 {
		t.Fatalf(".intro should match exactly one paragraph, got %d", len(nodes))
	}
	if nodes := root.Query("p.intro"); len(nodes) != 1 {
		t.Fatalf("p.intro should match intro paragraph, got %d", len(nodes))
	}
	if nodes := root.Query("div.intro"); len(nodes) != 0 {
		t.Fatalf("div.intro should not match, got %#v", nodes)
	}
	if nodes := root.Query(".container.secondary"); len(nodes) != 1 || nodes[0].Attrs["id"] == nil || *nodes[0].Attrs["id"] != "sidebar" {
		t.Fatalf(".container.secondary should match sidebar div, got %#v", nodes)
	}
	if nodes := root.Query(".special"); len(nodes) != 1 || nodes[0].Name != "li" {
		t.Fatalf(".special should match the li item, got %#v", nodes)
	}
}

func TestAttributePresenceAcrossDocument(t *testing.T) {
	root := buildSimpleDoc()

	if nodes := root.Query("[href]"); len(nodes) != 1 || nodes[0].Name != "a" {
		t.Fatalf("[href] should match anchor, got %#v", nodes)
	}
	if nodes := root.Query("[id]"); len(nodes) != 2 {
		t.Fatalf("[id] should match both divs, got %d", len(nodes))
	}
	if nodes := root.Query("[data-id]"); len(nodes) != 1 || nodes[0].Name != "a" {
		t.Fatalf("[data-id] should match anchor with data-id, got %#v", nodes)
	}
}

func TestAttributeExactQuotedForms(t *testing.T) {
	root := buildSimpleDoc()

	if nodes := root.Query(`[id="main"]`); len(nodes) != 1 {
		t.Fatalf("double-quoted id selector should match main div, got %#v", nodes)
	}
	if nodes := root.Query("[id='main']"); len(nodes) != 1 {
		t.Fatalf("single-quoted id selector should match main div, got %#v", nodes)
	}
	if nodes := root.Query("[id=main]"); len(nodes) != 1 {
		t.Fatalf("unquoted id selector should match main div, got %#v", nodes)
	}
	if nodes := root.Query("[id=wrong]"); len(nodes) != 0 {
		t.Fatalf("mismatched id selector should not match, got %#v", nodes)
	}
}

func TestDescendantAndChildCombinators(t *testing.T) {
	root := buildSimpleDoc()

	if nodes := root.Query("div p"); len(nodes) != 2 {
		t.Fatalf("div p should match two paragraphs, got %d", len(nodes))
	}
	if nodes := root.Query("div > h1"); len(nodes) != 1 || nodes[0].Name != "h1" {
		t.Fatalf("div > h1 should match direct child heading, got %#v", nodes)
	}
	if nodes := root.Query("body > div"); len(nodes) != 2 {
		t.Fatalf("body > div should match both top-level divs, got %d", len(nodes))
	}
	if nodes := root.Query("span div"); len(nodes) != 0 {
		t.Fatalf("span div should not match when hierarchy is missing, got %#v", nodes)
	}
}

func TestMatchesWithCombinators(t *testing.T) {
	root := buildSimpleDoc()
	intro := root.Query("p.intro")
	if len(intro) != 1 {
		t.Fatalf("expected intro paragraph, got %#v", intro)
	}

	if !Matches(intro[0], "div p") {
		t.Fatalf("Matches should return true for descendant chain")
	}
	if Matches(intro[0], "#sidebar p") {
		t.Fatalf("Matches should return false when ancestor chain does not apply")
	}
}

func TestQueryFromSubtreeInComplexDocument(t *testing.T) {
	root := buildSimpleDoc()
	main := root.Query("#main")
	if len(main) != 1 {
		t.Fatalf("expected to find main div, got %#v", main)
	}

	if nodes := main[0].Query("p"); len(nodes) != 2 {
		t.Fatalf("subtree query for paragraphs should return two results, got %d", len(nodes))
	}
	if nodes := main[0].Query("div"); len(nodes) != 0 {
		t.Fatalf("subtree query should not include the starting node, got %#v", nodes)
	}
}

// --- Tag selectors ---
func TestFullTagSelector(t *testing.T) {
	result := buildFullSimpleDoc().Query("p")
	if len(result) != 2 {
		t.Fatalf("expected 2 p elements, got %d", len(result))
	}
	for _, n := range result {
		if n.Name != "p" {
			t.Fatalf("expected only <p> nodes, got %q", n.Name)
		}
	}
}

func TestFullTagSelectorCaseInsensitive(t *testing.T) {
	if got := len(buildFullSimpleDoc().Query("P")); got != 2 {
		t.Fatalf("expected 2 P elements, got %d", got)
	}
}

func TestFullTagSelectorDiv(t *testing.T) {
	if got := len(buildFullSimpleDoc().Query("div")); got != 2 {
		t.Fatalf("expected 2 div elements, got %d", got)
	}
}

func TestFullTagSelectorNoMatch(t *testing.T) {
	if got := len(buildFullSimpleDoc().Query("article")); got != 0 {
		t.Fatalf("expected 0 matches, got %d", got)
	}
}

func TestFullTagSelectorHeadElements(t *testing.T) {
	res := buildFullSimpleDoc().Query("title")
	if len(res) != 1 || res[0].Name != "title" {
		t.Fatalf("expected to find the title element")
	}
}

// --- Universal selectors ---
func TestFullUniversalSelector(t *testing.T) {
	res := buildFullSimpleDoc().Query("*")
	if len(res) <= 10 {
		t.Fatalf("expected many elements matched by *; got %d", len(res))
	}
}

func TestFullUniversalInCompound(t *testing.T) {
	res := buildFullSimpleDoc().Query("*.container")
	if len(res) != 2 {
		t.Fatalf("expected 2 elements with class container, got %d", len(res))
	}
	for _, n := range res {
		if n.Attr("class") == "" {
			t.Fatalf("expected class attribute on result")
		}
	}
}

// --- ID selectors ---
func TestFullIDSelector(t *testing.T) {
	res := buildFullSimpleDoc().Query("#main")
	if len(res) != 1 || res[0].Attr("id") != "main" {
		t.Fatalf("expected to find #main element")
	}
}

func TestFullIDSelectorNoMatch(t *testing.T) {
	if got := len(buildFullSimpleDoc().Query("#nonexistent")); got != 0 {
		t.Fatalf("expected 0 matches, got %d", got)
	}
}

func TestFullIDSelectorCaseSensitive(t *testing.T) {
	if got := len(buildFullSimpleDoc().Query("#MAIN")); got != 0 {
		t.Fatalf("expected case sensitive id lookup to fail, got %d", got)
	}
}

func TestFullIDWithTag(t *testing.T) {
	if got := len(buildFullSimpleDoc().Query("div#main")); got != 1 {
		t.Fatalf("expected 1 div#main, got %d", got)
	}
}

func TestFullIDWithWrongTag(t *testing.T) {
	if got := len(buildFullSimpleDoc().Query("span#main")); got != 0 {
		t.Fatalf("expected 0 span#main, got %d", got)
	}
}

// --- Class selectors ---
func TestFullClassSelector(t *testing.T) {
	if got := len(buildFullSimpleDoc().Query(".container")); got != 2 {
		t.Fatalf("expected 2 .container elements, got %d", got)
	}
}

func TestFullClassSelectorSingle(t *testing.T) {
	if got := len(buildFullSimpleDoc().Query(".intro")); got != 1 {
		t.Fatalf("expected 1 .intro element, got %d", got)
	}
}

func TestFullClassSelectorNoMatch(t *testing.T) {
	if got := len(buildFullSimpleDoc().Query(".nonexistent")); got != 0 {
		t.Fatalf("expected 0 matches, got %d", got)
	}
}

func TestFullClassSelectorCaseSensitive(t *testing.T) {
	if got := len(buildFullSimpleDoc().Query(".Container")); got != 0 {
		t.Fatalf("expected case sensitive class to fail, got %d", got)
	}
}

func TestFullMultipleClasses(t *testing.T) {
	if got := len(buildFullSimpleDoc().Query(".intro.first")); got != 1 {
		t.Fatalf("expected 1 element with both classes, got %d", got)
	}
}

func TestFullClassWithTag(t *testing.T) {
	if got := len(buildFullSimpleDoc().Query("p.intro")); got != 1 {
		t.Fatalf("expected 1 p.intro element, got %d", got)
	}
}

func TestFullClassWithWrongTag(t *testing.T) {
	if got := len(buildFullSimpleDoc().Query("div.intro")); got != 0 {
		t.Fatalf("expected 0 div.intro, got %d", got)
	}
}

// --- Attribute presence selectors ---
func TestFullAttributePresence(t *testing.T) {
	res := buildFullSimpleDoc().Query("[href]")
	if len(res) != 1 || res[0].Name != "a" {
		t.Fatalf("expected a[href] to find anchor")
	}
}

func TestFullAttributePresenceID(t *testing.T) {
	if got := len(buildFullSimpleDoc().Query("[id]")); got != 2 {
		t.Fatalf("expected 2 elements with id, got %d", got)
	}
}

func TestFullAttributePresenceData(t *testing.T) {
	if got := len(buildFullSimpleDoc().Query("[data-id]")); got != 1 {
		t.Fatalf("expected 1 data-id element, got %d", got)
	}
}

// --- Attribute exact match selectors ---
func TestFullAttributeExact(t *testing.T) {
	res := buildFullSimpleDoc().Query(`[id="main"]`)
	if len(res) != 1 || res[0].Attr("id") != "main" {
		t.Fatalf("expected to find element with id=main")
	}
}

func TestFullAttributeExactNoMatch(t *testing.T) {
	if got := len(buildFullSimpleDoc().Query(`[id="wrong"]`)); got != 0 {
		t.Fatalf("expected 0 matches, got %d", got)
	}
}

func TestFullAttributeExactUnquoted(t *testing.T) {
	if got := len(buildFullSimpleDoc().Query(`[id=main]`)); got != 1 {
		t.Fatalf("expected 1 id=main match, got %d", got)
	}
}

func TestFullAttributeExactSingleQuotes(t *testing.T) {
	if got := len(buildFullSimpleDoc().Query("[id='main']")); got != 1 {
		t.Fatalf("expected 1 id=main match, got %d", got)
	}
}

// --- Descendant combinator ---
func TestFullDescendant(t *testing.T) {
	if got := len(buildFullSimpleDoc().Query("div p")); got != 2 {
		t.Fatalf("expected 2 descendant paragraphs, got %d", got)
	}
}

func TestFullDescendantDeep(t *testing.T) {
	if got := len(buildNestedDoc().Query("div span")); got != 1 {
		t.Fatalf("expected 1 deep span, got %d", got)
	}
}

func TestFullDescendantMultipleLevels(t *testing.T) {
	res := buildNestedDoc().Query(".a span")
	if len(res) != 1 || res[0].Attr("id") != "deep" {
		t.Fatalf("expected to find span#deep under .a")
	}
}

func TestFullDescendantNoMatch(t *testing.T) {
	if got := len(buildFullSimpleDoc().Query("span div")); got != 0 {
		t.Fatalf("expected no matches, got %d", got)
	}
}

// --- Child combinator ---
func TestFullChild(t *testing.T) {
	if got := len(buildFullSimpleDoc().Query("div > h1")); got != 1 {
		t.Fatalf("expected 1 child match for div > h1, got %d", got)
	}
}

func TestFullChildDirectOnly(t *testing.T) {
	if got := len(buildNestedDoc().Query("body > span")); got != 0 {
		t.Fatalf("expected no direct span children of body, got %d", got)
	}
}

func TestFullChildWithClass(t *testing.T) {
	if got := len(buildNestedDoc().Query(".a > .b")); got != 1 {
		t.Fatalf("expected one .a > .b match, got %d", got)
	}
}

// --- Node query convenience ---
func TestFullQueryFromDocument(t *testing.T) {
	if got := len(buildFullSimpleDoc().Query("p")); got != 2 {
		t.Fatalf("expected 2 paragraphs from document query, got %d", got)
	}
}

func TestFullQueryFromSubtree(t *testing.T) {
	main := buildFullSimpleDoc().Query("#main")[0]
	if got := len(main.Query("p")); got != 2 {
		t.Fatalf("expected 2 paragraphs inside #main, got %d", got)
	}
}

func TestFullQueryFromSubtreeLimited(t *testing.T) {
	sidebar := buildFullSimpleDoc().Query("#sidebar")[0]
	if got := len(sidebar.Query("p")); got != 0 {
		t.Fatalf("expected 0 paragraphs inside #sidebar, got %d", got)
	}
}

func TestFullQueryFromSubtreeExcludesSelf(t *testing.T) {
	main := buildFullSimpleDoc().Query("#main")[0]
	if got := len(main.Query("div")); got != 0 {
		t.Fatalf("expected query to exclude the starting node itself, got %d", got)
	}
}

// --- Matches helper ---
func TestFullMatchesTrue(t *testing.T) {
	div := buildFullSimpleDoc().Query("#main")[0]
	if !Matches(div, "div") || !Matches(div, "#main") || !Matches(div, ".container") || !Matches(div, "div.container") {
		t.Fatalf("expected div to match multiple selectors")
	}
}

func TestFullMatchesFalse(t *testing.T) {
	div := buildFullSimpleDoc().Query("#main")[0]
	if Matches(div, "span") || Matches(div, "#sidebar") || Matches(div, ".other") {
		t.Fatalf("div should not match unrelated selectors")
	}
}

func TestFullMatchesWithCombinator(t *testing.T) {
	p := buildFullSimpleDoc().Query("p.intro")[0]
	if !Matches(p, "div p") || !Matches(p, "#main p") {
		t.Fatalf("expected paragraph to match ancestor selectors")
	}
	if Matches(p, "#sidebar p") {
		t.Fatalf("paragraph should not match selector under #sidebar")
	}
}

// --- Attribute contains word selectors ---
func TestFullAttributeContainsWord(t *testing.T) {
	if got := len(buildFullSimpleDoc().Query(`[class~="container"]`)); got != 2 {
		t.Fatalf("expected 2 container word matches, got %d", got)
	}
}

func TestFullAttributeContainsWordSingle(t *testing.T) {
	if got := len(buildFullSimpleDoc().Query(`[class~="secondary"]`)); got != 1 {
		t.Fatalf("expected 1 secondary word match, got %d", got)
	}
}

func TestFullAttributeContainsWordNoPartial(t *testing.T) {
	if got := len(buildFullSimpleDoc().Query(`[class~="contain"]`)); got != 0 {
		t.Fatalf("expected no partial word matches, got %d", got)
	}
}

// --- Attribute hyphen prefix selectors ---
func TestFullAttributeHyphenExact(t *testing.T) {
	if got := len(buildLangDoc("en").Query(`[lang|="en"]`)); got != 1 {
		t.Fatalf("expected 1 lang hyphen exact match, got %d", got)
	}
}

func TestFullAttributeHyphenPrefix(t *testing.T) {
	if got := len(buildLangDoc("en-US").Query(`[lang|="en"]`)); got != 1 {
		t.Fatalf("expected 1 lang hyphen prefix match, got %d", got)
	}
}

func TestFullAttributeHyphenNoMatch(t *testing.T) {
	if got := len(buildLangDoc("eng").Query(`[lang|="en"]`)); got != 0 {
		t.Fatalf("expected no hyphen matches, got %d", got)
	}
}

// --- Attribute starts/ends/contains selectors ---
func TestFullAttributeStartsWith(t *testing.T) {
	if got := len(buildFullSimpleDoc().Query(`[href^="http"]`)); got != 1 {
		t.Fatalf("expected href starting with http, got %d", got)
	}
}

func TestFullAttributeEndsWith(t *testing.T) {
	if got := len(buildFullSimpleDoc().Query(`[href$=".com"]`)); got != 1 {
		t.Fatalf("expected href ending with .com, got %d", got)
	}
}

func TestFullAttributeContains(t *testing.T) {
	if got := len(buildFullSimpleDoc().Query(`[href*="example"]`)); got != 1 {
		t.Fatalf("expected href containing example, got %d", got)
	}
}

// --- Sibling combinators ---
func TestFullAdjacentSibling(t *testing.T) {
	res := buildSiblingDoc().Query("h1 + p")
	if len(res) != 1 || !strings.Contains(res[0].Attr("class"), "first") {
		t.Fatalf("expected first paragraph adjacent to h1")
	}
}

func TestFullAdjacentSiblingChain(t *testing.T) {
	res := buildSiblingDoc().Query(".first + p")
	if len(res) != 1 || !strings.Contains(res[0].Attr("class"), "second") {
		t.Fatalf("expected second paragraph after .first")
	}
}

func TestFullAdjacentSiblingNoMatch(t *testing.T) {
	if got := len(buildSiblingDoc().Query(".first + span")); got != 0 {
		t.Fatalf("expected no match for .first + span, got %d", got)
	}
}

func TestFullGeneralSibling(t *testing.T) {
	if got := len(buildSiblingDoc().Query("h1 ~ p")); got != 4 {
		t.Fatalf("expected 4 p siblings after h1, got %d", got)
	}
}

func TestAdjacentSiblingDoesNotCascade(t *testing.T) {
	doc := buildFullSimpleDoc()
	res := doc.Query("li + li")
	if len(res) != 1 {
		t.Fatalf("expected only the first adjacent li, got %d", len(res))
	}
	if res[0].Attr("class") != "special" {
		t.Fatalf("expected the second li with class special, got class %q", res[0].Attr("class"))
	}
}

func TestFullGeneralSiblingWithClass(t *testing.T) {
	if got := len(buildSiblingDoc().Query(".first ~ p")); got != 3 {
		t.Fatalf("expected 3 p siblings after .first, got %d", got)
	}
}

// --- Child position pseudo classes ---
func TestFullFirstChild(t *testing.T) {
	if got := len(buildFullSimpleDoc().Query("li:first-child")); got != 1 {
		t.Fatalf("expected first li child, got %d", got)
	}
}

func TestFullFirstChildWithTag(t *testing.T) {
	res := buildSiblingDoc().Query("div > :first-child")
	if len(res) != 1 || res[0].Name != "h1" {
		t.Fatalf("expected h1 as first child of div")
	}
}

func TestFullLastChild(t *testing.T) {
	if got := len(buildFullSimpleDoc().Query("li:last-child")); got != 1 {
		t.Fatalf("expected last li child, got %d", got)
	}
}

func TestFullNthChildNumber(t *testing.T) {
	res := buildFullSimpleDoc().Query("li:nth-child(2)")
	if len(res) != 1 || !strings.Contains(res[0].Attr("class"), "special") {
		t.Fatalf("expected second li with class special")
	}
}
func TestNthChildWhitespaceOnlyArgument(t *testing.T) {
	doc := buildNode(
		el("ul", nil,
			el("li", nil),
			el("li", nil),
		),
	)

	if got := len(doc.Query("li:nth-child(   )")); got != 0 {
		t.Fatalf("nth-child with whitespace-only argument should match nothing, got %d", got)
	}
}

func TestFullNthChildOddEven(t *testing.T) {
	if got := len(buildFullSimpleDoc().Query("li:nth-child(odd)")); got != 2 {
		t.Fatalf("expected two odd li children, got %d", got)
	}
	if got := len(buildFullSimpleDoc().Query("li:nth-child(even)")); got != 1 {
		t.Fatalf("expected one even li child, got %d", got)
	}
}

func TestFullNthChildFormula(t *testing.T) {
	if got := len(buildFullSimpleDoc().Query("li:nth-child(2n+1)")); got != 2 {
		t.Fatalf("expected nth-child(2n+1) to match 2 nodes, got %d", got)
	}
}

func TestFullNotSelector(t *testing.T) {
	res := buildFullSimpleDoc().Query("div:not(#sidebar)")
	if len(res) != 1 || res[0].Attr("id") != "main" {
		t.Fatalf("expected to exclude #sidebar div")
	}
}

func TestFullOnlyChild(t *testing.T) {
	doc := buildLangDoc("")
	wrapper := &Node{Name: "div"}
	span := &Node{Name: "span"}
	span.AppendChild(&Node{Name: "#text", Data: "Only"})
	wrapper.AppendChild(span)
	doc.Children[0].Children[0].AppendChild(wrapper)

	if got := len(doc.Query("span:only-child")); got != 1 {
		t.Fatalf("expected only-child span, got %d", got)
	}
}

// --- Empty and root pseudo classes ---
func TestFullEmpty(t *testing.T) {
	if got := len(buildEmptyAndRootDoc().Query(".empty:empty")); got != 1 {
		t.Fatalf("expected empty div to match :empty, got %d", got)
	}
}

func TestFullEmptyWhitespace(t *testing.T) {
	if got := len(buildEmptyAndRootDoc().Query(".whitespace:empty")); got != 1 {
		t.Fatalf("expected whitespace div treated as empty, got %d", got)
	}
}

func TestFullRoot(t *testing.T) {
	res := buildFullSimpleDoc().Query(":root")
	if len(res) != 1 || res[0].Name != "html" {
		t.Fatalf("expected html element as :root")
	}
}

// --- Type-based pseudo classes ---
func TestFullFirstOfType(t *testing.T) {
	res := buildSiblingDoc().Query("p:first-of-type")
	if len(res) != 1 || !strings.Contains(res[0].Attr("class"), "first") {
		t.Fatalf("expected first-of-type p")
	}
}

func TestFullLastOfType(t *testing.T) {
	res := buildSiblingDoc().Query("p:last-of-type")
	if len(res) != 1 || !strings.Contains(res[0].Attr("class"), "fourth") {
		t.Fatalf("expected last-of-type p")
	}
}

func TestFullNthOfType(t *testing.T) {
	res := buildSiblingDoc().Query("p:nth-of-type(2)")
	if len(res) != 1 || !strings.Contains(res[0].Attr("class"), "second") {
		t.Fatalf("expected second p as nth-of-type(2)")
	}
}

func TestFullOnlyOfType(t *testing.T) {
	res := buildSiblingDoc().Query("h1:only-of-type")
	if len(res) != 1 {
		t.Fatalf("expected single h1 to match only-of-type")
	}
}

// --- Selector groups and complex selectors ---
func TestFullSelectorGroups(t *testing.T) {
	if got := len(buildFullSimpleDoc().Query("h1, p, li")); got != 6 {
		t.Fatalf("expected 6 elements across selector group, got %d", got)
	}
}

func TestFullComplexSelector(t *testing.T) {
	if got := len(buildFullSimpleDoc().Query("div.container > ul li.special")); got != 1 {
		t.Fatalf("expected complex selector to match special li, got %d", got)
	}
}

// --- Template content ---
func TestFullTemplateQuery(t *testing.T) {
	tplContent := &Node{Name: "div", Attrs: map[string]*string{"class": strPtr("inside")}}
	template := &Node{Name: "template", Template: tplContent}
	doc := &Node{Name: "document"}
	html := &Node{Name: "html"}
	body := &Node{Name: "body"}
	body.AppendChild(template)
	html.AppendChild(body)
	doc.AppendChild(html)

	if got := len(doc.Query(".inside")); got != 1 {
		t.Fatalf("expected to find element inside template, got %d", got)
	}
}

// --- Pseudo :contains ---
func TestFullContainsPseudo(t *testing.T) {
	doc := buildContainsDoc()
	res := doc.Query(`button:contains("click me")`)
	if len(res) != 1 || res[0].Name != "button" {
		t.Fatalf("expected button containing 'click me'")
	}
}

func TestFullContainsDescendants(t *testing.T) {
	doc := buildContainsDoc()
	ids := make(map[string]struct{})
	for _, n := range doc.Query(`div:contains("click me")`) {
		ids[n.Attr("id")] = struct{}{}
	}
	if len(ids) != 2 {
		t.Fatalf("expected two divs containing text, got %d", len(ids))
	}
}
func TestAttributeStartsWithEmptyValue(t *testing.T) {
	doc := buildNode(
		el("div", map[string]string{"data-x": "abc"}),
	)

	if got := len(doc.Query(`[data-x^=""]`)); got != 0 {
		t.Fatalf("[data-x^=\"\"] should match nothing, got %d", got)
	}
}
func TestAttributeEndsWithEmptyValue(t *testing.T) {
	doc := buildNode(
		el("div", map[string]string{"data-x": "abc"}),
	)

	if got := len(doc.Query(`[data-x$=""]`)); got != 0 {
		t.Fatalf("[data-x$=\"\"] should match nothing, got %d", got)
	}
}
func TestAttributeContainsEmptyValue(t *testing.T) {
	doc := buildNode(
		el("div", map[string]string{"data-x": "abc"}),
	)

	if got := len(doc.Query(`[data-x*=""]`)); got != 0 {
		t.Fatalf("[data-x*=\"\"] should match nothing, got %d", got)
	}
}
func TestAttributeExactMatchEmptyValue(t *testing.T) {
	doc := buildNode(
		el("input", map[string]string{"type": ""}),
	)

	if got := len(doc.Query(`[type=""]`)); got != 1 {
		t.Fatalf(`[type=""] should match element with empty attribute, got %d`, got)
	}
}
func TestAttributeHyphenPrefixEmptyValue(t *testing.T) {
	doc := buildNode(
		el("p", map[string]string{"lang": "en"}),
	)

	if got := len(doc.Query(`[lang|=""]`)); got != 0 {
		t.Fatalf(`[lang|=""] should match nothing, got %d`, got)
	}
}
func TestAttributeContainsWordEmptyValue(t *testing.T) {
	doc := buildNode(
		el("div", map[string]string{"class": "a b c"}),
	)

	if got := len(doc.Query(`[class~=""]`)); got != 0 {
		t.Fatalf(`[class~=""] should match nothing, got %d`, got)
	}
}

func TestNotEmptyArgumentMatchesAll(t *testing.T) {
	doc := buildNode(
		el("div", nil),
		el("span", nil),
	)

	if got := len(doc.Query(`div:not()`)); got != 1 {
		t.Fatalf("div:not() should match all divs, got %d", got)
	}
}
func TestNotRejectsComplexSelector(t *testing.T) {
	doc := buildNode(
		el("div", nil,
			el("p", nil),
		),
	)

	// Python ignores the :not() and matches div
	if got := len(doc.Query(`div:not(div > p)`)); got != 1 {
		t.Fatalf("complex :not() should be ignored")
	}
}
func TestUnicodeClassSelectors(t *testing.T) {
	doc := buildNode(
		el("div", map[string]string{
			"class": "über 日本語 normal",
		}),
	)

	tests := []string{
		".über",
		".日本語",
		".normal",
	}

	for _, sel := range tests {
		if got := len(doc.Query(sel)); got != 1 {
			t.Fatalf("unicode selector %q failed", sel)
		}
	}
}
func TestEscapedClassSelector(t *testing.T) {
	doc := &Node{Name: "document"}
	html := &Node{Name: "html", Parent: doc}
	body := &Node{Name: "body", Parent: html}
	div := &Node{
		Name:   "div",
		Attrs:  map[string]*string{"class": strPtr("foo:bar")},
		Parent: body,
	}

	doc.Children = []*Node{html}
	html.Children = []*Node{body}
	body.Children = []*Node{div}

	matches := doc.Query(`.foo\:bar`)
	if len(matches) != 1 {
		t.Fatalf("escaped class selector should match, got %d", len(matches))
	}
}
func TestContainsMissingArgumentErrors(t *testing.T) {
	_, err := parseSelector(`button:contains()`)
	if err == nil {
		t.Fatalf(":contains() without argument should error")
	}
}
func TestNumericEscapeInClassSelector(t *testing.T) {
	doc := nodeWithClass("foo:bar")
	if len(doc.Query(`.foo\3A bar`)) != 1 {
		t.Fatalf("numeric escape failed")
	}
}
func TestEscapedNewlineInSelector(t *testing.T) {
	doc := buildNode(el("div", map[string]string{"class": "foobar"}))

	if got := len(doc.Query(".foo\\\nbar")); got != 1 {
		t.Fatalf("escaped newline should collapse selector")
	}
}

func TestAttributeStartsWithEmptyValue_Diagnostic(t *testing.T) {
	chains, err := parseSelector(`[data-x^=""]`)
	if err != nil {
		t.Fatalf("parseSelector errored: %v", err)
	}

	if len(chains) != 1 || len(chains[0]) != 1 {
		t.Fatalf("unexpected selector chain: %#v", chains)
	}

	comp := chains[0][0].compound
	if len(comp.attrs) != 1 {
		t.Fatalf("expected 1 attribute selector, got %d", len(comp.attrs))
	}

	attr := comp.attrs[0]

	if attr.op != "^=" {
		t.Fatalf("expected ^= operator, got %q", attr.op)
	}
	if attr.value != "" {
		t.Fatalf("expected empty value, got %q", attr.value)
	}
}
func docWithNode(n *Node) *Node {
	doc := &Node{Name: "document"}
	doc.AppendChild(n)
	return doc
}
func TestEscapesInsideAttributeSelectorAreLiteral(t *testing.T) {
	val := `foo\ bar`
	div := &Node{
		Name: "div",
		Attrs: map[string]*string{
			"data-x": strPtr(val),
		},
	}

	doc := docWithNode(div)

	if got := len(doc.Query(`[data-x="foo\ bar"]`)); got != 1 {
		t.Fatalf("attribute escape should be literal, got %d", got)
	}
}
func TestEscapedClassInsideNot(t *testing.T) {
	div := &Node{
		Name: "div",
		Attrs: map[string]*string{
			"class": strPtr("foo:bar"),
		},
	}

	doc := docWithNode(div)

	if got := len(doc.Query(`:not(.foo\:bar)`)); got != 0 {
		t.Fatalf(":not() with escaped class should exclude node")
	}
}
func TestEscapedCombinatorIsLiteral(t *testing.T) {
	div := &Node{
		Name: "div",
		Attrs: map[string]*string{
			"class": strPtr("foo>bar"),
		},
	}

	doc := docWithNode(div)

	if got := len(doc.Query(`.foo\>bar`)); got != 1 {
		t.Fatalf("escaped combinator should be literal, got %d", got)
	}
}

func TestMultipleNumericEscapesInClass(t *testing.T) {
	div := &Node{
		Name: "div",
		Attrs: map[string]*string{
			"class": strPtr("123"),
		},
	}

	doc := docWithNode(div)

	if got := len(doc.Query(`.\31\32\33`)); got != 1 {
		t.Fatalf("multiple numeric escapes should decode to '123'")
	}
}
func TestInvalidEscapeSequenceIsLiteral(t *testing.T) {
	div := &Node{
		Name: "div",
		Attrs: map[string]*string{
			"class": strPtr("foozbar"),
		},
	}

	doc := docWithNode(div)

	if got := len(doc.Query(`.foo\zbar`)); got != 1 {
		t.Fatalf("invalid escape should be treated as literal")
	}
}

func TestEscapedNewlineCollapsesSelector(t *testing.T) {
	div := &Node{
		Name: "div",
		Attrs: map[string]*string{
			"class": strPtr("foobar"),
		},
	}

	doc := docWithNode(div)

	if got := len(doc.Query(".foo\\\nbar")); got != 1 {
		t.Fatalf("escaped newline should collapse selector")
	}
}

func TestUnicodeNumericEscapeInClass(t *testing.T) {
	div := &Node{
		Name: "div",
		Attrs: map[string]*string{
			"class": strPtr("über"),
		},
	}

	doc := docWithNode(div)

	if got := len(doc.Query(`.\FC ber`)); got != 1 {
		t.Fatalf("unicode numeric escape should match 'über'")
	}
}
func TestEscapedSpaceInClassSelector(t *testing.T) {
	div := &Node{
		Name: "div",
		Attrs: map[string]*string{
			"class": strPtr("foo bar"),
		},
	}

	doc := docWithNode(div)

	if got := len(doc.Query(`.foo\ bar`)); got != 1 {
		t.Fatalf("escaped space should be part of class name")
	}
}
func TestRootDoesNotMatchDocumentNode(t *testing.T) {
	doc := &Node{Name: "document"}

	if Matches(doc, ":root") {
		t.Fatalf("document node must not match :root")
	}
}

// cl

// Error Handling Tests
func TestEmptySelector(t *testing.T) {
	doc := buildSimpleDoc()
	result := doc.Query("")
	if result != nil {
		t.Fatalf("empty selector should return nil")
	}
}

func TestWhitespaceOnlySelector(t *testing.T) {
	doc := buildSimpleDoc()
	result := doc.Query("   ")
	if result != nil {
		t.Fatalf("whitespace-only selector should return nil")
	}
}

func TestUnclosedAttributeBracket(t *testing.T) {
	_, err := parseSelector("[attr")
	if err == nil {
		t.Fatalf("expected error for unclosed attribute bracket")
	}
}

func TestMissingAttributeName(t *testing.T) {
	_, err := parseSelector("[]")
	if err == nil {
		t.Fatalf("expected error for missing attribute name")
	}
}

// func TestInvalidAttributeOperator(t *testing.T) {
// 	_, err := parseSelector("[attr!=value]")
// 	if err == nil {
// 		t.Fatalf("expected error for invalid attribute operator")
// 	}
// }

func TestUnclosedString(t *testing.T) {
	_, err := parseSelector(`[attr="unclosed]`)
	if err == nil {
		t.Fatalf("expected error for unclosed string")
	}
}

func TestMissingPseudoName(t *testing.T) {
	_, err := parseSelector("div:")
	if err == nil {
		t.Fatalf("expected error for missing pseudo name")
	}
}

// func TestUnsupportedPseudoClass(t *testing.T) {
// 	_, err := parseSelector("div:hover")
// 	if err == nil {
// 		t.Fatalf("expected error for unsupported pseudo-class :hover")
// 	}
// }

func TestDanglingCombinator(t *testing.T) {
	_, err := parseSelector("div >")
	if err == nil {
		t.Fatalf("expected error for dangling combinator")
	}
}

func TestDoubleCombinator(t *testing.T) {
	_, err := parseSelector("div > > p")
	if err == nil {
		t.Fatalf("expected error for double combinator")
	}
}

func TestMissingIDName(t *testing.T) {
	_, err := parseSelector("#")
	if err == nil {
		t.Fatalf("expected error for missing ID name")
	}
}

func TestMissingClassName(t *testing.T) {
	_, err := parseSelector(".")
	if err == nil {
		t.Fatalf("expected error for missing class name")
	}
}

// Edge Cases Tests
func TestDeeplyNestedElements(t *testing.T) {
	// Build 100 levels of nesting
	doc := &Node{Name: "document"}
	current := doc
	for i := 0; i < 100; i++ {
		div := &Node{Name: "div"}
		current.AppendChild(div)
		current = div
	}
	span := &Node{Name: "span"}
	current.AppendChild(span)

	result := doc.Query("span")
	if len(result) != 1 {
		t.Fatalf("expected to find deeply nested span, got %d", len(result))
	}
}

func TestManySiblings(t *testing.T) {
	doc := &Node{Name: "document"}
	html := &Node{Name: "html"}
	body := &Node{Name: "body"}
	ul := &Node{Name: "ul"}

	for i := 0; i < 100; i++ {
		li := &Node{Name: "li"}
		ul.AppendChild(li)
	}

	body.AppendChild(ul)
	html.AppendChild(body)
	doc.AppendChild(html)

	result := doc.Query("li:nth-child(50)")
	if len(result) != 1 {
		t.Fatalf("expected to find 50th child, got %d", len(result))
	}
}

func TestEmptyDocument(t *testing.T) {
	doc := &Node{Name: "document"}
	result := doc.Query("div")
	if len(result) != 0 {
		t.Fatalf("empty document should have no divs, got %d", len(result))
	}
}

func TestSpecialAttributeValuesWithSpaces(t *testing.T) {
	doc := &Node{Name: "document"}
	a := &Node{
		Name:  "a",
		Attrs: map[string]*string{"href": strPtr("has spaces")},
	}
	doc.AppendChild(a)

	result := doc.Query(`[href="has spaces"]`)
	if len(result) != 1 {
		t.Fatalf("expected to find element with spaces in attribute, got %d", len(result))
	}
}

func TestQueryOnTextNode(t *testing.T) {
	doc := buildSimpleDoc()
	body := doc.Query("body")[0]

	// Text nodes shouldn't match element selectors
	for _, child := range body.Children {
		if child.Name == "#text" && Matches(child, "div") {
			t.Fatalf("text node should not match element selector")
		}
	}
}

func TestNthChildZero(t *testing.T) {
	doc := buildNode(
		el("ul", nil,
			el("li", nil),
			el("li", nil),
		),
	)

	result := doc.Query("li:nth-child(0)")
	if len(result) != 0 {
		t.Fatalf("nth-child(0) should match nothing, got %d", len(result))
	}
}

func TestNthChildNegative(t *testing.T) {
	doc := buildNode(
		el("ul", nil,
			el("li", nil),
			el("li", nil),
		),
	)

	result := doc.Query("li:nth-child(-1)")
	if len(result) != 0 {
		t.Fatalf("nth-child(-1) should match nothing, got %d", len(result))
	}
}

func TestNthChildLargeNumber(t *testing.T) {
	doc := buildNode(
		el("ul", nil,
			el("li", nil),
			el("li", nil),
		),
	)

	result := doc.Query("li:nth-child(100)")
	if len(result) != 0 {
		t.Fatalf("nth-child(100) should match nothing, got %d", len(result))
	}
}

func TestClassWithHyphen(t *testing.T) {
	doc := buildNode(
		el("div", map[string]string{"class": "my-class"}),
	)

	result := doc.Query(".my-class")
	if len(result) != 1 {
		t.Fatalf("hyphenated class should match, got %d", len(result))
	}
}

func TestIDWithHyphen(t *testing.T) {
	doc := buildNode(
		el("div", map[string]string{"id": "my-id"}),
	)

	result := doc.Query("#my-id")
	if len(result) != 1 {
		t.Fatalf("hyphenated ID should match, got %d", len(result))
	}
}

// Advanced nth-child Tests
func TestNthChildN(t *testing.T) {
	doc := buildNode(
		el("ul", nil,
			el("li", nil),
			el("li", nil),
			el("li", nil),
			el("li", nil),
			el("li", nil),
		),
	)

	result := doc.Query("li:nth-child(n)")
	if len(result) != 5 {
		t.Fatalf("nth-child(n) should match all 5 elements, got %d", len(result))
	}
}

func TestNthChild2N(t *testing.T) {
	doc := buildNode(
		el("ul", nil,
			el("li", nil),
			el("li", nil),
			el("li", nil),
			el("li", nil),
			el("li", nil),
		),
	)

	result := doc.Query("li:nth-child(2n)")
	if len(result) != 2 {
		t.Fatalf("nth-child(2n) should match 2 elements (2nd and 4th), got %d", len(result))
	}
}

func TestNthChildNegativeOffset(t *testing.T) {
	doc := buildNode(
		el("ul", nil,
			el("li", nil),
			el("li", nil),
			el("li", nil),
			el("li", nil),
			el("li", nil),
		),
	)

	result := doc.Query("li:nth-child(-n+3)")
	if len(result) != 3 {
		t.Fatalf("nth-child(-n+3) should match first 3 elements, got %d", len(result))
	}
}

// Tokenizer Edge Cases
func TestAttributeWithSpacesAroundOperator(t *testing.T) {
	doc := buildNode(
		el("div", map[string]string{"id": "test"}),
	)

	result := doc.Query("[ id = test ]")
	if len(result) != 1 {
		t.Fatalf("attribute selector with spaces should work, got %d", len(result))
	}
}

func TestCombinatorWithExtraSpaces(t *testing.T) {
	doc := buildNode(
		el("div", nil,
			el("p", nil),
		),
	)

	result := doc.Query("div   >   p")
	if len(result) != 1 {
		t.Fatalf("combinator with extra spaces should work, got %d", len(result))
	}
}

func TestMultiplePseudoClasses(t *testing.T) {
	doc := buildNode(
		el("ul", nil,
			el("li", nil),
		),
	)

	result := doc.Query("li:first-child:last-child")
	if len(result) != 1 {
		t.Fatalf("multiple pseudo-classes should work, got %d", len(result))
	}
}

func TestPseudoWithArgContainingSpaces(t *testing.T) {
	doc := buildNode(
		el("ul", nil,
			el("li", nil),
			el("li", nil),
			el("li", nil),
		),
	)

	result := doc.Query("li:nth-child( 2n + 1 )")
	if len(result) != 2 {
		t.Fatalf("nth-child with spaces in arg should work, got %d", len(result))
	}
}

// Matcher Edge Cases
func TestEmptyPseudoWithComments(t *testing.T) {
	doc := buildNode(
		el("div", nil,
			&Node{Name: "#comment", Data: "comment"},
		),
	)

	result := doc.Query("div:empty")
	// Comments should be ignored per CSS spec
	if len(result) != 1 {
		t.Fatalf("div with only comment should match :empty, got %d", len(result))
	}
}

func TestNthChildWithInvalidExpression(t *testing.T) {
	doc := buildNode(
		el("ul", nil,
			el("li", nil),
			el("li", nil),
		),
	)

	result := doc.Query("li:nth-child(invalid)")
	if len(result) != 0 {
		t.Fatalf("invalid nth-child expression should match nothing, got %d", len(result))
	}
}

func TestNthOfTypeWithInvalidExpression(t *testing.T) {
	doc := buildNode(
		el("div", nil),
		el("div", nil),
	)

	result := doc.Query("div:nth-of-type(invalid)")
	if len(result) != 0 {
		t.Fatalf("invalid nth-of-type expression should match nothing, got %d", len(result))
	}
}

func TestOnlyChildNoMatch(t *testing.T) {
	doc := buildSimpleDoc()
	result := doc.Query("li:only-child")
	if len(result) != 0 {
		t.Fatalf("only-child should not match when there are siblings, got %d", len(result))
	}
}

func TestOnlyOfTypeNoMatch(t *testing.T) {
	doc := buildSiblingDoc()
	result := doc.Query("p:only-of-type")
	if len(result) != 0 {
		t.Fatalf("only-of-type should not match when there are multiple of same type, got %d", len(result))
	}
}

func TestLastChildNotLast(t *testing.T) {
	doc := buildSiblingDoc()
	result := doc.Query("h1:last-child")
	if len(result) != 0 {
		t.Fatalf("h1 is not last child, should not match, got %d", len(result))
	}
}

func TestEmptyWithText(t *testing.T) {
	doc := buildEmptyAndRootDoc()
	result := doc.Query(".text:empty")
	if len(result) != 0 {
		t.Fatalf("element with text should not match :empty, got %d", len(result))
	}
}

func TestRootWithTag(t *testing.T) {
	doc := buildSimpleDoc()
	result := doc.Query("html:root")
	if len(result) != 1 {
		t.Fatalf("html:root should match, got %d", len(result))
	}
}

func TestNthOfTypeOdd(t *testing.T) {
	doc := buildSiblingDoc()
	result := doc.Query("p:nth-of-type(odd)")
	if len(result) != 2 {
		t.Fatalf("nth-of-type(odd) should match 2 paragraphs, got %d", len(result))
	}
}

// Selector Groups
func TestTwoSelectors(t *testing.T) {
	doc := buildSimpleDoc()
	result := doc.Query("h1, h2")
	if len(result) != 1 || result[0].Name != "h1" {
		t.Fatalf("selector group should find h1, got %d results", len(result))
	}
}

func TestComplexSelectorGroups(t *testing.T) {
	doc := buildSimpleDoc()
	result := doc.Query("#main p, #sidebar a")
	if len(result) != 3 {
		t.Fatalf("complex selector group should find 3 elements, got %d", len(result))
	}
}

// Complex Compound Selectors
func TestCompoundSelectorWithAttribute(t *testing.T) {
	doc := buildSimpleDoc()
	result := doc.Query("a[href][data-id]")
	if len(result) != 1 {
		t.Fatalf("compound selector with multiple attributes should match, got %d", len(result))
	}
}

func TestMultipleCombinators(t *testing.T) {
	doc := buildNestedDoc()
	result := doc.Query(".a > .b > .c span")
	if len(result) != 1 {
		t.Fatalf("selector with multiple combinators should match, got %d", len(result))
	}
}

func TestPseudoWithCombinator(t *testing.T) {
	doc := buildSiblingDoc()
	result := doc.Query("div > p:first-child")
	if len(result) != 0 {
		t.Fatalf("div > p:first-child should not match (h1 is first), got %d", len(result))
	}
}

// Invalid Character Test
func TestInvalidCharacterInSelector(t *testing.T) {
	_, err := parseSelector("div@foo")
	if err == nil {
		t.Fatalf("expected error for invalid character @")
	}
}

// Case Sensitivity Tests
func TestClassSelectorCaseSensitive(t *testing.T) {
	doc := buildSimpleDoc()
	result := doc.Query(".Container")
	if len(result) != 0 {
		t.Fatalf("class selector should be case-sensitive, got %d", len(result))
	}
}

// Additional Attribute Tests
func TestAttributeHyphenPrefixNoMatchWithoutHyphen(t *testing.T) {
	doc := buildLangDoc("english")
	result := doc.Query(`[lang|="en"]`)
	if len(result) != 0 {
		t.Fatalf("lang|=en should not match 'english', got %d", len(result))
	}
}

func TestAttributeContainsWordEmptyClass(t *testing.T) {
	doc := buildNode(
		el("div", map[string]string{"class": ""}),
	)

	result := doc.Query(`[class~="foo"]`)
	if len(result) != 0 {
		t.Fatalf("~= should not match empty class, got %d", len(result))
	}
}

// First/Last of Type Edge Cases
func TestFirstOfTypeMultipleTypes(t *testing.T) {
	doc := buildNode(
		el("div", nil),
		el("span", nil),
		el("div", nil),
	)

	result := doc.Query("div:first-of-type")
	if len(result) != 1 {
		t.Fatalf("first-of-type should find first div, got %d", len(result))
	}
}

func TestLastOfTypeMultipleTypes(t *testing.T) {
	doc := buildNode(
		el("div", nil),
		el("span", nil),
		el("div", nil),
	)

	result := doc.Query("div:last-of-type")
	if len(result) != 1 {
		t.Fatalf("last-of-type should find last div, got %d", len(result))
	}
}

// Not pseudo-class with empty argument
func TestNotWithClassSelector(t *testing.T) {
	doc := buildSimpleDoc()
	result := doc.Query("p:not(.intro)")
	if len(result) != 1 {
		t.Fatalf("p:not(.intro) should match one paragraph, got %d", len(result))
	}
	if strings.Contains(result[0].Attr("class"), "intro") {
		t.Fatalf("matched paragraph should not have intro class")
	}
}

func TestNotWithCombinator(t *testing.T) {
	doc := buildSimpleDoc()
	result := doc.Query("div > p:not(.content)")
	if len(result) != 1 {
		t.Fatalf("div > p:not(.content) should match one paragraph, got %d", len(result))
	}
}

// Sibling combinator edge cases
func TestGeneralSiblingNoMatch(t *testing.T) {
	doc := buildNode(
		el("div", nil,
			el("span", nil),
			el("p", nil),
		),
	)

	result := doc.Query("div ~ p")
	if len(result) != 0 {
		t.Fatalf("div ~ p should not match when no div before p, got %d", len(result))
	}
}

func TestAdjacentSiblingNoMatch(t *testing.T) {
	doc := buildNode(
		el("div", nil,
			el("h1", nil),
			el("span", nil),
			el("p", nil),
		),
	)

	result := doc.Query("h1 + p")
	if len(result) != 0 {
		t.Fatalf("h1 + p should not match when span is between, got %d", len(result))
	}
}

// Universal selector variations
func TestUniversalSelectorSkipsTextNodes(t *testing.T) {
	root := buildTree()

	all := root.Query("*")
	if len(all) != 4 { // div, span, em, p (exclude #text)
		t.Fatalf("universal selector matched %d nodes, want 4", len(all))
	}
	doc := buildNode(
		el("div", nil,
			&Node{Name: "#text", Data: "text"},
			el("span", nil),
		),
	)

	result := doc.Query("*")
	// Should only match div and span, not text node
	for _, n := range result {
		if strings.HasPrefix(n.Name, "#") {
			t.Fatalf("universal selector should skip text nodes")
		}
	}
}

// Minimal reproduction case
func TestDebugUnicodeClass(t *testing.T) {
	// Helper
	strPtr := func(s string) *string { return &s }

	// Create node with unicode class
	div := &Node{
		Name:  "div",
		Attrs: map[string]*string{"class": strPtr("über")},
	}

	doc := &Node{Name: "document"}
	doc.AppendChild(div)

	// Test 1: Can we find a div at all?
	result := doc.Query("div")
	fmt.Printf("Query 'div': found %d nodes\n", len(result))
	if len(result) != 1 {
		t.Fatalf("Basic div query failed")
	}

	// Test 2: What does the class attribute actually contain?
	if result[0].Attrs != nil {
		if classPtr, ok := result[0].Attrs["class"]; ok && classPtr != nil {
			fmt.Printf("Actual class value: %q (bytes: %v)\n", *classPtr, []byte(*classPtr))
		}
	}

	// Test 3: Parse the selector
	chains, err := parseSelector(`.über`)
	if err != nil {
		t.Fatalf("Failed to parse selector: %v", err)
	}
	fmt.Printf("Parsed selector has %d chains\n", len(chains))
	if len(chains) > 0 && len(chains[0]) > 0 {
		className := chains[0][0].compound.classes[0]
		fmt.Printf("Parsed class name: %q (bytes: %v)\n", className, []byte(className))
	}

	// Test 4: Try to match
	result = doc.Query(`.über`)
	fmt.Printf("Query '.über': found %d nodes\n", len(result))

	if len(result) != 1 {
		t.Fatalf("Unicode class query failed - expected 1, got %d", len(result))
	}
}

// Test if the problem is specific to class selectors
func TestDebugUnicodeAttribute(t *testing.T) {
	strPtr := func(s string) *string { return &s }

	div := &Node{
		Name:  "div",
		Attrs: map[string]*string{"data-name": strPtr("über")},
	}

	doc := &Node{Name: "document"}
	doc.AppendChild(div)

	// Try attribute selector
	result := doc.Query(`[data-name="über"]`)
	fmt.Printf("Query '[data-name=\"über\"]': found %d nodes\n", len(result))

	if len(result) != 1 {
		t.Fatalf("Unicode attribute query failed - expected 1, got %d", len(result))
	}
}

// Test the readIdent function directly
func TestDebugReadIdent(t *testing.T) {
	testCases := []struct {
		input     string
		start     int
		wantIdent string
		wantNext  int
	}{
		{"über", 0, "über", 5}, // ü is 2 bytes in UTF-8
		{"test", 0, "test", 4},
		{"foo-bar", 0, "foo-bar", 7},
		{"123abc", 0, "123abc", 6},
	}

	for _, tc := range testCases {
		gotIdent, gotNext := readIdent(tc.input, tc.start)
		fmt.Printf("readIdent(%q, %d) = (%q, %d)\n", tc.input, tc.start, gotIdent, gotNext)

		if gotIdent != tc.wantIdent {
			t.Errorf("readIdent(%q, %d): got ident %q, want %q",
				tc.input, tc.start, gotIdent, tc.wantIdent)
		}
		if gotNext != tc.wantNext {
			t.Errorf("readIdent(%q, %d): got next %d, want %d",
				tc.input, tc.start, gotNext, tc.wantNext)
		}
	}
}
