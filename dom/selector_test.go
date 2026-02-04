package dom

import (
	"fmt"
	"strings"
	"testing"

	"github.com/padraicbc/h5p/internal/common"
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
	tests := []common.TestCase{
		{
			Name: "attribute equals selector",
			Run: func(t *testing.T) {
				doc := &Node{Name: "document"}
				html := &Node{Name: "html"}
				body := &Node{Name: "body"}
				img := &Node{Name: "img", Attrs: map[string]*string{"alt": strPtr("eISB Site Logo"), "src": strPtr("logo.png")}}

				doc.AppendChild(html)
				html.AppendChild(body)
				body.AppendChild(img)

				imgNodes, _ := doc.Query("img")
				if len(imgNodes) != 1 {
					t.Fatalf("expected 1 img element, got %d", len(imgNodes))
				}

				matches, _ := doc.Query(`img[alt="eISB Site Logo"]`)
				if len(matches) != 1 {
					t.Fatalf("expected 1 match for alt value, got %d (alt=%q)", len(matches), imgNodes[0].Attr("alt"))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestQueryFirstHonorsCombinators(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "query first honors combinators",
			Run: func(t *testing.T) {
				root := buildTree()

				if node, _ := root.QueryFirst("div > span em"); node == nil || node.Name != "em" {
					t.Fatalf("QueryFirst descendant chain failed, got %#v", node)
				}
				if node, _ := root.QueryFirst("div > em"); node != nil {
					t.Fatalf("child combinator should not match em without span ancestor, got %#v", node)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestMatchesChecksAttributesAndClasses(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "matches checks attributes and classes",
			Run: func(t *testing.T) {
				root := buildTree()
				em, _ := root.QueryFirst("em")
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
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestUniversalSelectorFindsElements(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "universal selector finds elements",
			Run: func(t *testing.T) {
				root := buildTree()
				nodes, _ := root.Query("div *")
				if len(nodes) != 3 { // span, em, p
					t.Fatalf("universal selector matched %d nodes, want 3", len(nodes))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestTagAndDescendantSelectors(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "tag and descendant selectors",
			Run: func(t *testing.T) {
				root := buildTree()

				spanNodes, _ := root.Query("span")
				if len(spanNodes) != 1 {
					t.Fatalf("tag selector matched %d nodes, want 1", len(spanNodes))
				}
				if spanNodes[0].Name != "span" {
					t.Fatalf("expected span node, got %s", spanNodes[0].Name)
				}

				deep, _ := root.Query("div span em")
				if len(deep) != 1 || deep[0].Name != "em" {
					t.Fatalf("descendant selector should find em under div span, got %#v", deep)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestIDSelector(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "id selector",
			Run: func(t *testing.T) {
				root := buildTree()

				nodes, _ := root.Query("#main")
				if len(nodes) != 1 {
					t.Fatalf("id selector matched %d nodes, want 1", len(nodes))
				}
				if nodes[0].Name != "div" {
					t.Fatalf("expected div node with id=main, got %s", nodes[0].Name)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestClassSelectorMatchesMultiple(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "class selector matches multiple",
			Run: func(t *testing.T) {
				root := buildTree()

				nodes, _ := root.Query(".foo")
				if len(nodes) != 2 { // div and span
					t.Fatalf("class selector matched %d nodes, want 2", len(nodes))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestAttributePresenceSelector(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "attribute presence selector",
			Run: func(t *testing.T) {
				root := buildTree()

				emNodes, _ := root.Query("[data-role]")
				if len(emNodes) != 1 || emNodes[0].Name != "em" {
					t.Fatalf("attribute presence selector should find em with data-role, got %#v", emNodes)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestAttributeIncludesSelector(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "attribute includes selector",
			Run: func(t *testing.T) {
				root := &Node{Name: "document"}
				tagged := &Node{Name: "div", Attrs: map[string]*string{"data-tags": strPtr("alpha beta gamma")}}
				partial := &Node{Name: "div", Attrs: map[string]*string{"data-tags": strPtr("alphabetagamma")}}

				root.AppendChild(tagged)
				root.AppendChild(partial)

				matches, _ := root.Query(`[data-tags~=beta]`)
				if len(matches) != 1 || matches[0] != tagged {
					t.Fatalf("[data-tags~=beta] should match only the whitespace-separated value, got %#v", matches)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestAttributeDashPrefixSelector(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "attribute dash prefix selector",
			Run: func(t *testing.T) {
				root := &Node{Name: "document"}
				englishUS := &Node{Name: "p", Attrs: map[string]*string{"lang": strPtr("en-US")}}
				plainEnglish := &Node{Name: "p", Attrs: map[string]*string{"lang": strPtr("en")}}
				other := &Node{Name: "p", Attrs: map[string]*string{"lang": strPtr("eng-US")}}

				root.AppendChild(englishUS)
				root.AppendChild(plainEnglish)
				root.AppendChild(other)

				matches, _ := root.Query(`[lang|="en"]`)
				if len(matches) != 2 || matches[0] != englishUS || matches[1] != plainEnglish {
					t.Fatalf("[lang|=\"en\"] should match exact or hyphenated values, got %#v", matches)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestAttributeIncludesRequiresWhitespaceList(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "attribute includes requires whitespace list",
			Run: func(t *testing.T) {
				root := &Node{Name: "document"}
				hyphenated := &Node{Name: "p", Attrs: map[string]*string{"lang": strPtr("en-US")}}
				plain := &Node{Name: "p", Attrs: map[string]*string{"lang": strPtr("en")}}
				spaced := &Node{Name: "p", Attrs: map[string]*string{"lang": strPtr("en fr")}}

				root.AppendChild(hyphenated)
				root.AppendChild(plain)
				root.AppendChild(spaced)

				matches, _ := root.Query(`p[lang~="en"]`)

				if len(matches) != 2 ||
					(matches[0] != plain && matches[1] != plain) ||
					(matches[0] != spaced && matches[1] != spaced) {
					t.Fatalf("[lang~=\"en\"] should match plain and spaced tokens, got %#v", matches)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestAttributeValueWithSpaces(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "attribute value with spaces",
			Run: func(t *testing.T) {
				root := buildTree()

				nodes, _ := root.Query(`div[class="foo bar"]`)
				if len(nodes) != 1 || nodes[0].Name != "div" {
					t.Fatalf("attribute value selector with spaces should find div, got %#v", nodes)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestCompoundSelectorRequiresAllParts(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "compound selector requires all parts",
			Run: func(t *testing.T) {
				root := buildTree()

				nodes, _ := root.Query("div#main.foo.bar[data-container=main]")
				if len(nodes) != 1 || nodes[0].Name != "div" {
					t.Fatalf("compound selector should match div with id, classes, and attribute, got %#v", nodes)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestDescendantCombinatorTraversesMultipleLevels(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "descendant combinator traverses multiple levels",
			Run: func(t *testing.T) {
				root := buildTree()

				nodes, _ := root.Query("div em")
				if len(nodes) != 1 || nodes[0].Name != "em" {
					t.Fatalf("descendant combinator should find em under div via span, got %#v", nodes)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestChildCombinatorMatchesDirectChildrenOnly(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "child combinator matches direct children only",
			Run: func(t *testing.T) {
				root := buildTree()

				children, _ := root.Query("div > span")
				if len(children) != 1 || children[0].Name != "span" {
					t.Fatalf("child combinator should match direct span child of div, got %#v", children)
				}

				if nodes, _ := root.Query("div>em"); len(nodes) != 0 {
					t.Fatalf("child combinator should not match grandchild em, got %#v", nodes)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestAttributeSelectorsHandleQuotedAndUnquotedValues(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "attribute selectors handle quoted and unquoted values",
			Run: func(t *testing.T) {
				root := buildTree()

				if nodes, _ := root.Query("em[data-role=label]"); len(nodes) != 1 {
					t.Fatalf("unquoted attribute selector should find em, got %#v", nodes)
				}

				if nodes, _ := root.Query(`em[data-role="label"]`); len(nodes) != 1 {
					t.Fatalf("double-quoted attribute selector should find em, got %#v", nodes)
				}

				if nodes, _ := root.Query("em[data-role='label']"); len(nodes) != 1 {
					t.Fatalf("single-quoted attribute selector should find em, got %#v", nodes)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestAttributePresenceMatchesCaseInsensitiveKeys(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "attribute presence matches case insensitive keys",
			Run: func(t *testing.T) {
				node := &Node{Name: "div", Attrs: map[string]*string{"DATA-Role": strPtr("button")}}
				root := &Node{Name: "document"}
				root.AppendChild(node)

				if matches, _ := root.Query("[data-role]"); len(matches) != 1 || matches[0] != node {
					t.Fatalf("attribute presence should match case-insensitive keys, got %#v", matches)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestClassSelectorRequiresEachClass(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "class selector requires each class",
			Run: func(t *testing.T) {
				root := buildTree()

				if nodes, _ := root.Query(".foo.bar"); len(nodes) != 2 {
					t.Fatalf("combined class selector should match div and span, got %d", len(nodes))
				}

				if nodes, _ := root.Query(".foo.baz"); len(nodes) != 0 {
					t.Fatalf("nonexistent class should prevent match, got %#v", nodes)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestQueryHandlesNilRootAndEmptySelector(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "query handles nil root and empty selector",
			Run: func(t *testing.T) {
				var root *Node
				if nodes, _ := root.Query("div"); nodes != nil {
					t.Fatalf("nil root should return nil result, got %#v", nodes)
				}

				doc := buildTree()
				if nodes, _ := doc.Query("   "); nodes != nil {
					t.Fatalf("empty selector should return nil result, got %#v", nodes)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestTagSelectorIsCaseInsensitive(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "tag selector is case insensitive",
			Run: func(t *testing.T) {
				root := buildTree()

				nodes, _ := root.Query("DIV")
				if len(nodes) != 1 || nodes[0].Name != "div" {
					t.Fatalf("expected to match div using uppercase tag, got %#v", nodes)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestTagSelectorNoMatch(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "tag selector no match",
			Run: func(t *testing.T) {
				root := buildTree()

				if nodes, _ := root.Query("article"); len(nodes) != 0 {
					t.Fatalf("unknown tag should not match, got %#v", nodes)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestIDSelectorIsCaseSensitive(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "id selector is case sensitive",
			Run: func(t *testing.T) {
				root := buildTree()

				if nodes, _ := root.Query("#MAIN"); len(nodes) != 0 {
					t.Fatalf("id selector should be case-sensitive, got %#v", nodes)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestClassSelectorNoMatch(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "class selector no match",
			Run: func(t *testing.T) {
				root := buildTree()

				if nodes, _ := root.Query(".missing"); len(nodes) != 0 {
					t.Fatalf("nonexistent class should not match, got %#v", nodes)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestAttributePresenceMatchesSpecificKeys(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "attribute presence matches specific keys",
			Run: func(t *testing.T) {
				root := buildTree()

				if nodes, _ := root.Query("[data-container]"); len(nodes) != 1 || nodes[0].Name != "div" {
					t.Fatalf("data-container presence should match div, got %#v", nodes)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestAttributeExactQuotesAndMismatches(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "attribute exact quotes and mismatches",
			Run: func(t *testing.T) {
				root := buildTree()

				if nodes, _ := root.Query("[id='main']"); len(nodes) != 1 || nodes[0].Name != "div" {
					t.Fatalf("single-quoted attribute selector should match div, got %#v", nodes)
				}
				if nodes, _ := root.Query("[id=wrong]"); len(nodes) != 0 {
					t.Fatalf("non-matching attribute value should not match, got %#v", nodes)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestDescendantSelectorNoMatch(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "descendant selector no match",
			Run: func(t *testing.T) {
				root := buildTree()

				if nodes, _ := root.Query("span div"); len(nodes) != 0 {
					t.Fatalf("descendant selector should not match when structure is missing, got %#v", nodes)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestChildCombinatorRequiresDirectDescendant(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "child combinator requires direct descendant",
			Run: func(t *testing.T) {
				root := buildTree()

				if nodes, _ := root.Query("div > em"); len(nodes) != 0 {
					t.Fatalf("child combinator should not match grandchild em, got %#v", nodes)
				}
				if nodes, _ := root.Query("div > p"); len(nodes) != 1 || nodes[0].Name != "p" {
					t.Fatalf("child combinator should match direct p child, got %#v", nodes)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestQueryFromSubtreeExcludesSelf(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "query from subtree excludes self",
			Run: func(t *testing.T) {
				root := buildTree()
				span, _ := root.QueryFirst("span")
				if span == nil {
					t.Fatalf("expected to find span node")
				}

				if nodes, _ := span.Query("span"); len(nodes) != 0 {
					t.Fatalf("query from subtree should not include self, got %#v", nodes)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestTagSelectorSupportsHyphenatedNames(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "tag selector supports hyphenated names",
			Run: func(t *testing.T) {
				root := &Node{Name: "document"}
				custom := &Node{Name: "my-element"}
				root.AppendChild(custom)

				if nodes, _ := root.Query("my-element"); len(nodes) != 1 || nodes[0] != custom {
					t.Fatalf("hyphenated tag names should match, got %#v", nodes)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestUniversalSelectorMatchesAllElements(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "universal selector matches all elements",
			Run: func(t *testing.T) {
				root := buildSimpleDoc()
				all, _ := root.Query("*")
				if len(all) != 14 {
					t.Fatalf("universal selector should match all elements, got %d", len(all))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestUniversalSelectorInCompound(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "universal selector in compound",
			Run: func(t *testing.T) {
				root := buildSimpleDoc()
				matches, _ := root.Query("*.container")
				if len(matches) != 2 {
					t.Fatalf("compound universal selector should match both container divs, got %d", len(matches))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestIDSelectorWithTagAndMismatchedTag(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "id selector with tag and mismatched tag",
			Run: func(t *testing.T) {
				root := buildSimpleDoc()

				if nodes, _ := root.Query("div#main"); len(nodes) != 1 {
					t.Fatalf("expected single div#main, got %#v", nodes)
				}
				if nodes, _ := root.Query("span#main"); len(nodes) != 0 {
					t.Fatalf("span#main should not match, got %#v", nodes)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestClassSelectorsWithTags(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "class selectors with tags",
			Run: func(t *testing.T) {
				root := buildSimpleDoc()

				if nodes, _ := root.Query(".container"); len(nodes) != 2 {
					t.Fatalf(".container should match both divs, got %d", len(nodes))
				}
				if nodes, _ := root.Query(".intro"); len(nodes) != 1 {
					t.Fatalf(".intro should match exactly one paragraph, got %d", len(nodes))
				}
				if nodes, _ := root.Query("p.intro"); len(nodes) != 1 {
					t.Fatalf("p.intro should match intro paragraph, got %d", len(nodes))
				}
				if nodes, _ := root.Query("div.intro"); len(nodes) != 0 {
					t.Fatalf("div.intro should not match, got %#v", nodes)
				}
				if nodes, _ := root.Query(".container.secondary"); len(nodes) != 1 || nodes[0].Attrs["id"] == nil || *nodes[0].Attrs["id"] != "sidebar" {
					t.Fatalf(".container.secondary should match sidebar div, got %#v", nodes)
				}
				if nodes, _ := root.Query(".special"); len(nodes) != 1 || nodes[0].Name != "li" {
					t.Fatalf(".special should match the li item, got %#v", nodes)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestAttributePresenceAcrossDocument(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "attribute presence across document",
			Run: func(t *testing.T) {
				root := buildSimpleDoc()

				if nodes, _ := root.Query("[href]"); len(nodes) != 1 || nodes[0].Name != "a" {
					t.Fatalf("[href] should match anchor, got %#v", nodes)
				}
				if nodes, _ := root.Query("[id]"); len(nodes) != 2 {
					t.Fatalf("[id] should match both divs, got %d", len(nodes))
				}
				if nodes, _ := root.Query("[data-id]"); len(nodes) != 1 || nodes[0].Name != "a" {
					t.Fatalf("[data-id] should match anchor with data-id, got %#v", nodes)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestAttributeExactQuotedForms(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "attribute exact quoted forms",
			Run: func(t *testing.T) {
				root := buildSimpleDoc()

				if nodes, _ := root.Query(`[id="main"]`); len(nodes) != 1 {
					t.Fatalf("double-quoted id selector should match main div, got %#v", nodes)
				}
				if nodes, _ := root.Query("[id='main']"); len(nodes) != 1 {
					t.Fatalf("single-quoted id selector should match main div, got %#v", nodes)
				}
				if nodes, _ := root.Query("[id=main]"); len(nodes) != 1 {
					t.Fatalf("unquoted id selector should match main div, got %#v", nodes)
				}
				if nodes, _ := root.Query("[id=wrong]"); len(nodes) != 0 {
					t.Fatalf("mismatched id selector should not match, got %#v", nodes)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestDescendantAndChildCombinators(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "descendant and child combinators",
			Run: func(t *testing.T) {
				root := buildSimpleDoc()

				if nodes, _ := root.Query("div p"); len(nodes) != 2 {
					t.Fatalf("div p should match two paragraphs, got %d", len(nodes))
				}
				if nodes, _ := root.Query("div > h1"); len(nodes) != 1 || nodes[0].Name != "h1" {
					t.Fatalf("div > h1 should match direct child heading, got %#v", nodes)
				}
				if nodes, _ := root.Query("body > div"); len(nodes) != 2 {
					t.Fatalf("body > div should match both top-level divs, got %d", len(nodes))
				}
				if nodes, _ := root.Query("span div"); len(nodes) != 0 {
					t.Fatalf("span div should not match when hierarchy is missing, got %#v", nodes)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestMatchesWithCombinators(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "matches with combinators",
			Run: func(t *testing.T) {
				root := buildSimpleDoc()
				intro, _ := root.Query("p.intro")
				if len(intro) != 1 {
					t.Fatalf("expected intro paragraph, got %#v", intro)
				}

				if !Matches(intro[0], "div p") {
					t.Fatalf("Matches should return true for descendant chain")
				}
				if Matches(intro[0], "#sidebar p") {
					t.Fatalf("Matches should return false when ancestor chain does not apply")
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestQueryFromSubtreeInComplexDocument(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "query from subtree in complex document",
			Run: func(t *testing.T) {
				root := buildSimpleDoc()
				main, _ := root.Query("#main")
				if len(main) != 1 {
					t.Fatalf("expected to find main div, got %#v", main)
				}

				nodes, _ := main[0].Query("p")
				if len(nodes) != 2 {
					t.Fatalf("subtree query for paragraphs should return two results, got %d", len(nodes))
				}
				nodes, _ = main[0].Query("div")
				if len(nodes) != 0 {
					t.Fatalf("subtree query should not include the starting node, got %#v", nodes)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

// --- Tag selectors ---
func TestFullTagSelector(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full tag selector",
			Run: func(t *testing.T) {
				result, _ := buildFullSimpleDoc().Query("p")
				if len(result) != 2 {
					t.Fatalf("expected 2 p elements, got %d", len(result))
				}
				for _, n := range result {
					if n.Name != "p" {
						t.Fatalf("expected only <p> nodes, got %q", n.Name)
					}
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullTagSelectorCaseInsensitive(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full tag selector case insensitive",
			Run: func(t *testing.T) {
				nodes, _ := buildFullSimpleDoc().Query("P")
				if got := len(nodes); got != 2 {
					t.Fatalf("expected 2 P elements, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullTagSelectorDiv(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full tag selector div",
			Run: func(t *testing.T) {
				nodes, _ := buildFullSimpleDoc().Query("div")
				if got := len(nodes); got != 2 {
					t.Fatalf("expected 2 div elements, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullTagSelectorNoMatch(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full tag selector no match",
			Run: func(t *testing.T) {
				nodes, _ := buildFullSimpleDoc().Query("article")
				if got := len(nodes); got != 0 {
					t.Fatalf("expected 0 matches, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullTagSelectorHeadElements(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full tag selector head elements",
			Run: func(t *testing.T) {
				res, _ := buildFullSimpleDoc().Query("title")
				if len(res) != 1 || res[0].Name != "title" {
					t.Fatalf("expected to find the title element")
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

// --- Universal selectors ---
func TestFullUniversalSelector(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full universal selector",
			Run: func(t *testing.T) {
				res, _ := buildFullSimpleDoc().Query("*")
				if len(res) <= 10 {
					t.Fatalf("expected many elements matched by *; got %d", len(res))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullUniversalInCompound(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full universal in compound",
			Run: func(t *testing.T) {
				res, _ := buildFullSimpleDoc().Query("*.container")
				if len(res) != 2 {
					t.Fatalf("expected 2 elements with class container, got %d", len(res))
				}
				for _, n := range res {
					if n.Attr("class") == "" {
						t.Fatalf("expected class attribute on result")
					}
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

// --- ID selectors ---
func TestFullIDSelector(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full id selector",
			Run: func(t *testing.T) {
				res, _ := buildFullSimpleDoc().Query("#main")
				if len(res) != 1 || res[0].Attr("id") != "main" {
					t.Fatalf("expected to find #main element")
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullIDSelectorNoMatch(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full id selector no match",
			Run: func(t *testing.T) {
				nodes, _ := buildFullSimpleDoc().Query("#nonexistent")
				if got := len(nodes); got != 0 {
					t.Fatalf("expected 0 matches, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullIDSelectorCaseSensitive(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full id selector case sensitive",
			Run: func(t *testing.T) {
				nodes, _ := buildFullSimpleDoc().Query("#MAIN")
				if got := len(nodes); got != 0 {
					t.Fatalf("expected case sensitive id lookup to fail, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullIDWithTag(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full id with tag",
			Run: func(t *testing.T) {
				nodes, _ := buildFullSimpleDoc().Query("div#main")
				if got := len(nodes); got != 1 {
					t.Fatalf("expected 1 div#main, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullIDWithWrongTag(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full id with wrong tag",
			Run: func(t *testing.T) {
				nodes, _ := buildFullSimpleDoc().Query("span#main")
				if got := len(nodes); got != 0 {
					t.Fatalf("expected 0 span#main, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

// --- Class selectors ---
func TestFullClassSelector(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full class selector",
			Run: func(t *testing.T) {
				nodes, _ := buildFullSimpleDoc().Query(".container")
				if got := len(nodes); got != 2 {
					t.Fatalf("expected 2 .container elements, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullClassSelectorSingle(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full class selector single",
			Run: func(t *testing.T) {
				nodes, _ := buildFullSimpleDoc().Query(".intro")
				if got := len(nodes); got != 1 {
					t.Fatalf("expected 1 .intro element, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullClassSelectorNoMatch(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full class selector no match",
			Run: func(t *testing.T) {
				result, _ := buildFullSimpleDoc().Query(".nonexistent")
				if got := len(result); got != 0 {
					t.Fatalf("expected 0 matches, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullClassSelectorCaseSensitive(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full class selector case sensitive",
			Run: func(t *testing.T) {
				result, _ := buildFullSimpleDoc().Query(".Container")
				if got := len(result); got != 0 {
					t.Fatalf("expected case sensitive class to fail, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullMultipleClasses(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full multiple classes",
			Run: func(t *testing.T) {
				result, _ := buildFullSimpleDoc().Query(".intro.first")
				if got := len(result); got != 1 {
					t.Fatalf("expected 1 element with both classes, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullClassWithTag(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full class with tag",
			Run: func(t *testing.T) {
				result, _ := buildFullSimpleDoc().Query("p.intro")
				if got := len(result); got != 1 {
					t.Fatalf("expected 1 p.intro element, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullClassWithWrongTag(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full class with wrong tag",
			Run: func(t *testing.T) {
				result, _ := buildFullSimpleDoc().Query("div.intro")
				if got := len(result); got != 0 {
					t.Fatalf("expected 0 div.intro, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

// --- Attribute presence selectors ---
func TestFullAttributePresence(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full attribute presence",
			Run: func(t *testing.T) {
				res, _ := buildFullSimpleDoc().Query("[href]")
				if len(res) != 1 || res[0].Name != "a" {
					t.Fatalf("expected a[href] to find anchor")
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullAttributePresenceID(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full attribute presence id",
			Run: func(t *testing.T) {
				result, _ := buildFullSimpleDoc().Query("[id]")
				if got := len(result); got != 2 {
					t.Fatalf("expected 2 elements with id, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullAttributePresenceData(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full attribute presence data",
			Run: func(t *testing.T) {
				result, _ := buildFullSimpleDoc().Query("[data-id]")
				if got := len(result); got != 1 {
					t.Fatalf("expected 1 data-id element, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

// --- Attribute exact match selectors ---
func TestFullAttributeExact(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full attribute exact",
			Run: func(t *testing.T) {
				res, _ := buildFullSimpleDoc().Query(`[id="main"]`)
				if len(res) != 1 || res[0].Attr("id") != "main" {
					t.Fatalf("expected to find element with id=main")
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullAttributeExactNoMatch(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full attribute exact no match",
			Run: func(t *testing.T) {
				result, _ := buildFullSimpleDoc().Query(`[id="wrong"]`)
				if got := len(result); got != 0 {
					t.Fatalf("expected 0 matches, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullAttributeExactUnquoted(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full attribute exact unquoted",
			Run: func(t *testing.T) {
				result, _ := buildFullSimpleDoc().Query(`[id=main]`)
				if got := len(result); got != 1 {
					t.Fatalf("expected 1 id=main match, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullAttributeExactSingleQuotes(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full attribute exact single quotes",
			Run: func(t *testing.T) {
				result, _ := buildFullSimpleDoc().Query("[id='main']")
				if got := len(result); got != 1 {
					t.Fatalf("expected 1 id=main match, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

// --- Descendant combinator ---
func TestFullDescendant(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full descendant",
			Run: func(t *testing.T) {
				result, _ := buildFullSimpleDoc().Query("div p")
				if got := len(result); got != 2 {
					t.Fatalf("expected 2 descendant paragraphs, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullDescendantDeep(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full descendant deep",
			Run: func(t *testing.T) {
				result, _ := buildNestedDoc().Query("div span")
				if got := len(result); got != 1 {
					t.Fatalf("expected 1 deep span, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullDescendantMultipleLevels(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full descendant multiple levels",
			Run: func(t *testing.T) {
				res, _ := buildNestedDoc().Query(".a span")
				if len(res) != 1 || res[0].Attr("id") != "deep" {
					t.Fatalf("expected to find span#deep under .a")
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullDescendantNoMatch(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full descendant no match",
			Run: func(t *testing.T) {
				result, _ := buildFullSimpleDoc().Query("span div")
				if got := len(result); got != 0 {
					t.Fatalf("expected no matches, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

// --- Child combinator ---
func TestFullChild(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full child",
			Run: func(t *testing.T) {
				result, _ := buildFullSimpleDoc().Query("div > h1")
				if got := len(result); got != 1 {
					t.Fatalf("expected 1 child match for div > h1, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullChildDirectOnly(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full child direct only",
			Run: func(t *testing.T) {
				result, _ := buildNestedDoc().Query("body > span")
				if got := len(result); got != 0 {
					t.Fatalf("expected no direct span children of body, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullChildWithClass(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full child with class",
			Run: func(t *testing.T) {
				result, _ := buildNestedDoc().Query(".a > .b")
				if got := len(result); got != 1 {
					t.Fatalf("expected one .a > .b match, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

// --- Node query convenience ---
func TestFullQueryFromDocument(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full query from document",
			Run: func(t *testing.T) {
				result, _ := buildFullSimpleDoc().Query("p")
				if got := len(result); got != 2 {
					t.Fatalf("expected 2 paragraphs from document query, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullQueryFromSubtree(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full query from subtree",
			Run: func(t *testing.T) {
				main, _ := buildFullSimpleDoc().Query("#main")
				result, _ := main[0].Query("p")
				if got := len(result); got != 2 {
					t.Fatalf("expected 2 paragraphs inside #main, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullQueryFromSubtreeLimited(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full query from subtree limited",
			Run: func(t *testing.T) {
				sidebar, _ := buildFullSimpleDoc().Query("#sidebar")
				result, _ := sidebar[0].Query("p")
				if got := len(result); got != 0 {
					t.Fatalf("expected 0 paragraphs inside #sidebar, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullQueryFromSubtreeExcludesSelf(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full query from subtree excludes self",
			Run: func(t *testing.T) {
				main, _ := buildFullSimpleDoc().Query("#main")
				result, _ := main[0].Query("div")
				if got := len(result); got != 0 {
					t.Fatalf("expected query to exclude the starting node itself, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

// --- Matches helper ---
func TestFullMatchesTrue(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full matches true",
			Run: func(t *testing.T) {
				div, _ := buildFullSimpleDoc().Query("#main")
				if !Matches(div[0], "div") || !Matches(div[0], "#main") || !Matches(div[0], ".container") || !Matches(div[0], "div.container") {
					t.Fatalf("expected div to match multiple selectors")
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullMatchesFalse(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full matches false",
			Run: func(t *testing.T) {
				div, _ := buildFullSimpleDoc().Query("#main")
				if Matches(div[0], "span") || Matches(div[0], "#sidebar") || Matches(div[0], ".other") {
					t.Fatalf("div should not match unrelated selectors")
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullMatchesWithCombinator(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full matches with combinator",
			Run: func(t *testing.T) {
				p, _ := buildFullSimpleDoc().Query("p.intro")
				if !Matches(p[0], "div p") || !Matches(p[0], "#main p") {
					t.Fatalf("expected paragraph to match ancestor selectors")
				}
				if Matches(p[0], "#sidebar p") {
					t.Fatalf("paragraph should not match selector under #sidebar")
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

// --- Attribute contains word selectors ---
func TestFullAttributeContainsWord(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full attribute contains word",
			Run: func(t *testing.T) {
				result, _ := buildFullSimpleDoc().Query(`[class~="container"]`)
				if got := len(result); got != 2 {
					t.Fatalf("expected 2 container word matches, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullAttributeContainsWordSingle(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full attribute contains word single",
			Run: func(t *testing.T) {
				result, _ := buildFullSimpleDoc().Query(`[class~="secondary"]`)
				if got := len(result); got != 1 {
					t.Fatalf("expected 1 secondary word match, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullAttributeContainsWordNoPartial(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full attribute contains word no partial",
			Run: func(t *testing.T) {
				result, _ := buildFullSimpleDoc().Query(`[class~="contain"]`)
				if got := len(result); got != 0 {
					t.Fatalf("expected no partial word matches, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

// --- Attribute hyphen prefix selectors ---
func TestFullAttributeHyphenExact(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full attribute hyphen exact",
			Run: func(t *testing.T) {
				result, _ := buildLangDoc("en").Query(`[lang|="en"]`)
				if got := len(result); got != 1 {
					t.Fatalf("expected 1 lang hyphen exact match, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullAttributeHyphenPrefix(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full attribute hyphen prefix",
			Run: func(t *testing.T) {
				result, _ := buildLangDoc("en-US").Query(`[lang|="en"]`)
				if got := len(result); got != 1 {
					t.Fatalf("expected 1 lang hyphen prefix match, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullAttributeHyphenNoMatch(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full attribute hyphen no match",
			Run: func(t *testing.T) {
				result, _ := buildLangDoc("eng").Query(`[lang|="en"]`)
				if got := len(result); got != 0 {
					t.Fatalf("expected no hyphen matches, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

// --- Attribute starts/ends/contains selectors ---
func TestFullAttributeStartsWith(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full attribute starts with",
			Run: func(t *testing.T) {
				result, _ := buildFullSimpleDoc().Query(`[href^="http"]`)
				if got := len(result); got != 1 {
					t.Fatalf("expected href starting with http, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullAttributeEndsWith(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full attribute ends with",
			Run: func(t *testing.T) {
				result, _ := buildFullSimpleDoc().Query(`[href$=".com"]`)
				if got := len(result); got != 1 {
					t.Fatalf("expected href ending with .com, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullAttributeContains(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full attribute contains",
			Run: func(t *testing.T) {
				result, _ := buildFullSimpleDoc().Query(`[href*="example"]`)
				if got := len(result); got != 1 {
					t.Fatalf("expected href containing example, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

// --- Sibling combinators ---
func TestFullAdjacentSibling(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full adjacent sibling",
			Run: func(t *testing.T) {
				res, _ := buildSiblingDoc().Query("h1 + p")
				if len(res) != 1 || !strings.Contains(res[0].Attr("class"), "first") {
					t.Fatalf("expected first paragraph adjacent to h1")
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullAdjacentSiblingChain(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full adjacent sibling chain",
			Run: func(t *testing.T) {
				res, _ := buildSiblingDoc().Query(".first + p")
				if len(res) != 1 || !strings.Contains(res[0].Attr("class"), "second") {
					t.Fatalf("expected second paragraph after .first")
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullAdjacentSiblingNoMatch(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full adjacent sibling no match",
			Run: func(t *testing.T) {
				result, _ := buildSiblingDoc().Query(".first + span")
				if got := len(result); got != 0 {
					t.Fatalf("expected no match for .first + span, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullGeneralSibling(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full general sibling",
			Run: func(t *testing.T) {
				result, _ := buildSiblingDoc().Query("h1 ~ p")
				if got := len(result); got != 4 {
					t.Fatalf("expected 4 p siblings after h1, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestAdjacentSiblingDoesNotCascade(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "adjacent sibling does not cascade",
			Run: func(t *testing.T) {
				doc := buildFullSimpleDoc()
				res, _ := doc.Query("li + li")
				if len(res) != 1 {
					t.Fatalf("expected only the first adjacent li, got %d", len(res))
				}
				if res[0].Attr("class") != "special" {
					t.Fatalf("expected the second li with class special, got class %q", res[0].Attr("class"))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullGeneralSiblingWithClass(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full general sibling with class",
			Run: func(t *testing.T) {
				result, _ := buildSiblingDoc().Query(".first ~ p")
				if got := len(result); got != 3 {
					t.Fatalf("expected 3 p siblings after .first, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

// --- Child position pseudo classes ---
func TestFullFirstChild(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full first child",
			Run: func(t *testing.T) {
				result, _ := buildFullSimpleDoc().Query("li:first-child")
				if got := len(result); got != 1 {
					t.Fatalf("expected first li child, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullFirstChildWithTag(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full first child with tag",
			Run: func(t *testing.T) {
				res, _ := buildSiblingDoc().Query("div > :first-child")
				if len(res) != 1 || res[0].Name != "h1" {
					t.Fatalf("expected h1 as first child of div")
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullLastChild(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full last child",
			Run: func(t *testing.T) {
				result, _ := buildFullSimpleDoc().Query("li:last-child")
				if got := len(result); got != 1 {
					t.Fatalf("expected last li child, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullNthChildNumber(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full nth child number",
			Run: func(t *testing.T) {
				res, _ := buildFullSimpleDoc().Query("li:nth-child(2)")
				if len(res) != 1 || !strings.Contains(res[0].Attr("class"), "special") {
					t.Fatalf("expected second li with class special")
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestNthChildWhitespaceOnlyArgument(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "nth child whitespace only argument",
			Run: func(t *testing.T) {
				doc := buildNode(
					el("ul", nil,
						el("li", nil),
						el("li", nil),
					),
				)

				result, _ := doc.Query("li:nth-child(   )")
				if got := len(result); got != 0 {
					t.Fatalf("nth-child with whitespace-only argument should match nothing, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullNthChildOddEven(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full nth child odd even",
			Run: func(t *testing.T) {
				result, _ := buildFullSimpleDoc().Query("li:nth-child(odd)")
				if got := len(result); got != 2 {
					t.Fatalf("expected two odd li children, got %d", got)
				}
				result2, _ := buildFullSimpleDoc().Query("li:nth-child(even)")
				if got := len(result2); got != 1 {
					t.Fatalf("expected one even li child, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullNthChildFormula(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full nth child formula",
			Run: func(t *testing.T) {
				result, _ := buildFullSimpleDoc().Query("li:nth-child(2n+1)")
				if got := len(result); got != 2 {
					t.Fatalf("expected nth-child(2n+1) to match 2 nodes, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullNotSelector(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full not selector",
			Run: func(t *testing.T) {
				res, _ := buildFullSimpleDoc().Query("div:not(#sidebar)")
				if len(res) != 1 || res[0].Attr("id") != "main" {
					t.Fatalf("expected to exclude #sidebar div")
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullOnlyChild(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full only child",
			Run: func(t *testing.T) {
				doc := buildLangDoc("")
				wrapper := &Node{Name: "div"}
				span := &Node{Name: "span"}
				span.AppendChild(&Node{Name: "#text", Data: "Only"})
				wrapper.AppendChild(span)
				doc.Children[0].Children[0].AppendChild(wrapper)

				result, _ := doc.Query("span:only-child")
				if got := len(result); got != 1 {
					t.Fatalf("expected only-child span, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

// --- Empty and root pseudo classes ---
func TestFullEmpty(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full empty",
			Run: func(t *testing.T) {
				result, _ := buildEmptyAndRootDoc().Query(".empty:empty")
				if got := len(result); got != 1 {
					t.Fatalf("expected empty div to match :empty, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullEmptyWhitespace(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full empty whitespace",
			Run: func(t *testing.T) {
				result, _ := buildEmptyAndRootDoc().Query(".whitespace:empty")
				if got := len(result); got != 1 {
					t.Fatalf("expected whitespace div treated as empty, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullRoot(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full root",
			Run: func(t *testing.T) {
				res, _ := buildFullSimpleDoc().Query(":root")
				if len(res) != 1 || res[0].Name != "html" {
					t.Fatalf("expected html element as :root")
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

// --- Type-based pseudo classes ---
func TestFullFirstOfType(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full first of type",
			Run: func(t *testing.T) {
				res, _ := buildSiblingDoc().Query("p:first-of-type")
				if len(res) != 1 || !strings.Contains(res[0].Attr("class"), "first") {
					t.Fatalf("expected first-of-type p")
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullLastOfType(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full last of type",
			Run: func(t *testing.T) {
				res, _ := buildSiblingDoc().Query("p:last-of-type")
				if len(res) != 1 || !strings.Contains(res[0].Attr("class"), "fourth") {
					t.Fatalf("expected last-of-type p")
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullNthOfType(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full nth of type",
			Run: func(t *testing.T) {
				res, _ := buildSiblingDoc().Query("p:nth-of-type(2)")
				if len(res) != 1 || !strings.Contains(res[0].Attr("class"), "second") {
					t.Fatalf("expected second p as nth-of-type(2)")
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullOnlyOfType(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full only of type",
			Run: func(t *testing.T) {
				res, _ := buildSiblingDoc().Query("h1:only-of-type")
				if len(res) != 1 {
					t.Fatalf("expected single h1 to match only-of-type")
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

// --- Selector groups and complex selectors ---
func TestFullSelectorGroups(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full selector groups",
			Run: func(t *testing.T) {
				result, _ := buildFullSimpleDoc().Query("h1, p, li")
				if got := len(result); got != 6 {
					t.Fatalf("expected 6 elements across selector group, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullComplexSelector(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full complex selector",
			Run: func(t *testing.T) {
				result, _ := buildFullSimpleDoc().Query("div.container > ul li.special")
				if got := len(result); got != 1 {
					t.Fatalf("expected complex selector to match special li, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

// --- Template content ---
func TestFullTemplateQuery(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full template query",
			Run: func(t *testing.T) {
				tplContent := &Node{Name: "div", Attrs: map[string]*string{"class": strPtr("inside")}}
				template := &Node{Name: "template", Template: tplContent}
				doc := &Node{Name: "document"}
				html := &Node{Name: "html"}
				body := &Node{Name: "body"}
				body.AppendChild(template)
				html.AppendChild(body)
				doc.AppendChild(html)

				result, _ := doc.Query(".inside")
				if got := len(result); got != 1 {
					t.Fatalf("expected to find element inside template, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

// --- Pseudo :contains ---
func TestFullContainsPseudo(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full contains pseudo",
			Run: func(t *testing.T) {
				doc := buildContainsDoc()
				res, _ := doc.Query(`button:contains("click me")`)
				if len(res) != 1 || res[0].Name != "button" {
					t.Fatalf("expected button containing 'click me'")
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestFullContainsDescendants(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "full contains descendants",
			Run: func(t *testing.T) {
				doc := buildContainsDoc()
				result, _ := doc.Query(`div:contains("click me")`)
				ids := make(map[string]struct{})
				for _, n := range result {
					ids[n.Attr("id")] = struct{}{}
				}
				if len(ids) != 2 {
					t.Fatalf("expected two divs containing text, got %d", len(ids))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestAttributeStartsWithEmptyValue(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "attribute starts with empty value",
			Run: func(t *testing.T) {
				doc := buildNode(
					el("div", map[string]string{"data-x": "abc"}),
				)

				result, _ := doc.Query(`[data-x^=""]`)
				if got := len(result); got != 0 {
					t.Fatalf("[data-x^=\"\"] should match nothing, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestAttributeEndsWithEmptyValue(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "attribute ends with empty value",
			Run: func(t *testing.T) {
				doc := buildNode(
					el("div", map[string]string{"data-x": "abc"}),
				)

				result, _ := doc.Query(`[data-x$=""]`)
				if got := len(result); got != 0 {
					t.Fatalf("[data-x$=\"\"] should match nothing, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestAttributeContainsEmptyValue(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "attribute contains empty value",
			Run: func(t *testing.T) {
				doc := buildNode(
					el("div", map[string]string{"data-x": "abc"}),
				)

				result, _ := doc.Query(`[data-x*=""]`)
				if got := len(result); got != 0 {
					t.Fatalf("[data-x*=\"\"] should match nothing, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestAttributeExactMatchEmptyValue(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "attribute exact match empty value",
			Run: func(t *testing.T) {
				doc := buildNode(
					el("input", map[string]string{"type": ""}),
				)

				result, _ := doc.Query(`[type=""]`)
				if got := len(result); got != 1 {
					t.Fatalf(`[type=""] should match element with empty attribute, got %d`, got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestAttributeHyphenPrefixEmptyValue(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "attribute hyphen prefix empty value",
			Run: func(t *testing.T) {
				doc := buildNode(
					el("p", map[string]string{"lang": "en"}),
				)

				result, _ := doc.Query(`[lang|=""]`)
				if got := len(result); got != 0 {
					t.Fatalf(`[lang|=""] should match nothing, got %d`, got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestAttributeContainsWordEmptyValue(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "attribute contains word empty value",
			Run: func(t *testing.T) {
				doc := buildNode(
					el("div", map[string]string{"class": "a b c"}),
				)

				result, _ := doc.Query(`[class~=""]`)
				if got := len(result); got != 0 {
					t.Fatalf(`[class~=""] should match nothing, got %d`, got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestNotEmptyArgumentMatchesAll(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "not empty argument matches all",
			Run: func(t *testing.T) {
				doc := buildNode(
					el("div", nil),
					el("span", nil),
				)

				result, _ := doc.Query(`div:not()`)
				if got := len(result); got != 1 {
					t.Fatalf("div:not() should match all divs, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestNotRejectsComplexSelector(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "not rejects complex selector",
			Run: func(t *testing.T) {
				doc := buildNode(
					el("div", nil,
						el("p", nil),
					),
				)

				// Python ignores the :not() and matches div
				result, _ := doc.Query(`div:not(div > p)`)
				if got := len(result); got != 1 {
					t.Fatalf("complex :not() should be ignored")
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestUnicodeClassSelectors(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "unicode class selectors",
			Run: func(t *testing.T) {
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
					result, _ := doc.Query(sel)
					if got := len(result); got != 1 {
						t.Fatalf("unicode selector %q failed", sel)
					}
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestEscapedClassSelector(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "escaped class selector",
			Run: func(t *testing.T) {
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

				matches, _ := doc.Query(`.foo\:bar`)
				if len(matches) != 1 {
					t.Fatalf("escaped class selector should match, got %d", len(matches))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestContainsMissingArgumentErrors(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "contains missing argument errors",
			Run: func(t *testing.T) {
				_, err := parseSelector(`button:contains()`)
				if err == nil {
					t.Fatalf(":contains() without argument should error")
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestNumericEscapeInClassSelector(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "numeric escape in class selector",
			Run: func(t *testing.T) {
				doc := nodeWithClass("foo:bar")
				result, _ := doc.Query(`.foo\3A bar`)
				if len(result) != 1 {
					t.Fatalf("numeric escape failed")
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestEscapedNewlineInSelector(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "escaped newline in selector",
			Run: func(t *testing.T) {
				doc := buildNode(el("div", map[string]string{"class": "foobar"}))

				result, _ := doc.Query(".foo\\\nbar")
				if got := len(result); got != 1 {
					t.Fatalf("escaped newline should collapse selector")
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestAttributeStartsWithEmptyValue_Diagnostic(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "attribute starts with empty value_diagnostic",
			Run: func(t *testing.T) {
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
			},
		},
	}

	common.RunTestCases(t, tests)
}

func docWithNode(n *Node) *Node {
	doc := &Node{Name: "document"}
	doc.AppendChild(n)
	return doc
}

func TestEscapesInsideAttributeSelectorAreLiteral(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "escapes inside attribute selector are literal",
			Run: func(t *testing.T) {
				val := `foo\ bar`
				div := &Node{
					Name: "div",
					Attrs: map[string]*string{
						"data-x": strPtr(val),
					},
				}

				doc := docWithNode(div)

				result, _ := doc.Query(`[data-x="foo\ bar"]`)
				if got := len(result); got != 1 {
					t.Fatalf("attribute escape should be literal, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestEscapedClassInsideNot(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "escaped class inside not",
			Run: func(t *testing.T) {
				div := &Node{
					Name: "div",
					Attrs: map[string]*string{
						"class": strPtr("foo:bar"),
					},
				}

				doc := docWithNode(div)

				result, _ := doc.Query(`:not(.foo\:bar)`)
				if got := len(result); got != 0 {
					t.Fatalf(":not() with escaped class should exclude node")
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestEscapedCombinatorIsLiteral(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "escaped combinator is literal",
			Run: func(t *testing.T) {
				div := &Node{
					Name: "div",
					Attrs: map[string]*string{
						"class": strPtr("foo>bar"),
					},
				}

				doc := docWithNode(div)

				result, _ := doc.Query(`.foo\>bar`)
				if got := len(result); got != 1 {
					t.Fatalf("escaped combinator should be literal, got %d", got)
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestMultipleNumericEscapesInClass(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "multiple numeric escapes in class",
			Run: func(t *testing.T) {
				div := &Node{
					Name: "div",
					Attrs: map[string]*string{
						"class": strPtr("123"),
					},
				}

				doc := docWithNode(div)

				result, _ := doc.Query(`.\31\32\33`)
				if got := len(result); got != 1 {
					t.Fatalf("multiple numeric escapes should decode to '123'")
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestInvalidEscapeSequenceIsLiteral(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "invalid escape sequence is literal",
			Run: func(t *testing.T) {
				div := &Node{
					Name: "div",
					Attrs: map[string]*string{
						"class": strPtr("foozbar"),
					},
				}

				doc := docWithNode(div)

				result, _ := doc.Query(`.foo\zbar`)
				if got := len(result); got != 1 {
					t.Fatalf("invalid escape should be treated as literal")
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestEscapedNewlineCollapsesSelector(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "escaped newline collapses selector",
			Run: func(t *testing.T) {
				div := &Node{
					Name: "div",
					Attrs: map[string]*string{
						"class": strPtr("foobar"),
					},
				}

				doc := docWithNode(div)

				result, _ := doc.Query(".foo\\\nbar")
				if got := len(result); got != 1 {
					t.Fatalf("escaped newline should collapse selector")
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestUnicodeNumericEscapeInClass(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "unicode numeric escape in class",
			Run: func(t *testing.T) {
				div := &Node{
					Name: "div",
					Attrs: map[string]*string{
						"class": strPtr("über"),
					},
				}

				doc := docWithNode(div)

				result, _ := doc.Query(`.\FC ber`)
				if got := len(result); got != 1 {
					t.Fatalf("unicode numeric escape should match 'über'")
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestEscapedSpaceInClassSelector(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "escaped space in class selector",
			Run: func(t *testing.T) {
				div := &Node{
					Name: "div",
					Attrs: map[string]*string{
						"class": strPtr("foo bar"),
					},
				}

				doc := docWithNode(div)

				result, _ := doc.Query(`.foo\ bar`)
				if got := len(result); got != 1 {
					t.Fatalf("escaped space should be part of class name")
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestRootDoesNotMatchDocumentNode(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "root does not match document node",
			Run: func(t *testing.T) {
				doc := &Node{Name: "document"}

				if Matches(doc, ":root") {
					t.Fatalf("document node must not match :root")
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

// Error Handling Tests
func TestEmptySelector(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "empty selector",
			Run: func(t *testing.T) {
				doc := buildSimpleDoc()
				result, _ := doc.Query("")
				if result != nil {
					t.Fatalf("empty selector should return nil")
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestWhitespaceOnlySelector(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "whitespace only selector",
			Run: func(t *testing.T) {
				doc := buildSimpleDoc()
				result, _ := doc.Query("   ")
				if result != nil {
					t.Fatalf("whitespace-only selector should return nil")
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestUnclosedAttributeBracket(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "unclosed attribute bracket",
			Run: func(t *testing.T) {
				_, err := parseSelector("[attr")
				if err == nil {
					t.Fatalf("expected error for unclosed attribute bracket")
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestMissingAttributeName(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "missing attribute name",
			Run: func(t *testing.T) {
				_, err := parseSelector("[]")
				if err == nil {
					t.Fatalf("expected error for missing attribute name")
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestUnclosedString(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "unclosed string",
			Run: func(t *testing.T) {
				_, err := parseSelector(`[attr="unclosed]`)
				if err == nil {
					t.Fatalf("expected error for unclosed string")
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestMissingPseudoName(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "missing pseudo name",
			Run: func(t *testing.T) {
				_, err := parseSelector("div:")
				if err == nil {
					t.Fatalf("expected error for missing pseudo name")
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestDanglingCombinator(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "dangling combinator",
			Run: func(t *testing.T) {
				_, err := parseSelector("div >")
				if err == nil {
					t.Fatalf("expected error for dangling combinator")
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestDoubleCombinator(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "double combinator",
			Run: func(t *testing.T) {
				_, err := parseSelector("div > > p")
				if err == nil {
					t.Fatalf("expected error for double combinator")
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestMissingIDName(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "missing id name",
			Run: func(t *testing.T) {
				_, err := parseSelector("#")
				if err == nil {
					t.Fatalf("expected error for missing ID name")
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestMissingClassName(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "missing class name",
			Run: func(t *testing.T) {
				_, err := parseSelector(".")
				if err == nil {
					t.Fatalf("expected error for missing class name")
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

// Edge Cases Tests
func TestDeeplyNestedElements(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "deeply nested elements",
			Run: func(t *testing.T) {
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

				result, _ := doc.Query("span")
				if len(result) != 1 {
					t.Fatalf("expected to find deeply nested span, got %d", len(result))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestManySiblings(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "many siblings",
			Run: func(t *testing.T) {
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

				result, _ := doc.Query("li:nth-child(50)")
				if len(result) != 1 {
					t.Fatalf("expected to find 50th child, got %d", len(result))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestEmptyDocument(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "empty document",
			Run: func(t *testing.T) {
				doc := &Node{Name: "document"}
				result, _ := doc.Query("div")
				if len(result) != 0 {
					t.Fatalf("empty document should have no divs, got %d", len(result))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestSpecialAttributeValuesWithSpaces(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "special attribute values with spaces",
			Run: func(t *testing.T) {
				doc := &Node{Name: "document"}
				a := &Node{
					Name:  "a",
					Attrs: map[string]*string{"href": strPtr("has spaces")},
				}
				doc.AppendChild(a)

				result, _ := doc.Query(`[href="has spaces"]`)
				if len(result) != 1 {
					t.Fatalf("expected to find element with spaces in attribute, got %d", len(result))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestQueryOnTextNode(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "query on text node",
			Run: func(t *testing.T) {
				doc := buildSimpleDoc()
				body, _ := doc.Query("body")

				// Text nodes shouldn't match element selectors
				for _, child := range body[0].Children {
					if child.Name == "#text" && Matches(child, "div") {
						t.Fatalf("text node should not match element selector")
					}
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestNthChildZero(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "nth child zero",
			Run: func(t *testing.T) {
				doc := buildNode(
					el("ul", nil,
						el("li", nil),
						el("li", nil),
					),
				)

				result, _ := doc.Query("li:nth-child(0)")
				if len(result) != 0 {
					t.Fatalf("nth-child(0) should match nothing, got %d", len(result))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestNthChildNegative(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "nth child negative",
			Run: func(t *testing.T) {
				doc := buildNode(
					el("ul", nil,
						el("li", nil),
						el("li", nil),
					),
				)

				result, _ := doc.Query("li:nth-child(-1)")
				if len(result) != 0 {
					t.Fatalf("nth-child(-1) should match nothing, got %d", len(result))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestNthChildLargeNumber(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "nth child large number",
			Run: func(t *testing.T) {
				doc := buildNode(
					el("ul", nil,
						el("li", nil),
						el("li", nil),
					),
				)

				result, _ := doc.Query("li:nth-child(100)")
				if len(result) != 0 {
					t.Fatalf("nth-child(100) should match nothing, got %d", len(result))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestClassWithHyphen(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "class with hyphen",
			Run: func(t *testing.T) {
				doc := buildNode(
					el("div", map[string]string{"class": "my-class"}),
				)

				result, _ := doc.Query(".my-class")
				if len(result) != 1 {
					t.Fatalf("hyphenated class should match, got %d", len(result))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestIDWithHyphen(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "id with hyphen",
			Run: func(t *testing.T) {
				doc := buildNode(
					el("div", map[string]string{"id": "my-id"}),
				)

				result, _ := doc.Query("#my-id")
				if len(result) != 1 {
					t.Fatalf("hyphenated ID should match, got %d", len(result))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

// Advanced nth-child Tests
func TestNthChildN(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "nth child n",
			Run: func(t *testing.T) {
				doc := buildNode(
					el("ul", nil,
						el("li", nil),
						el("li", nil),
						el("li", nil),
						el("li", nil),
						el("li", nil),
					),
				)

				result, _ := doc.Query("li:nth-child(n)")
				if len(result) != 5 {
					t.Fatalf("nth-child(n) should match all 5 elements, got %d", len(result))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestNthChild2N(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "nth child2n",
			Run: func(t *testing.T) {
				doc := buildNode(
					el("ul", nil,
						el("li", nil),
						el("li", nil),
						el("li", nil),
						el("li", nil),
						el("li", nil),
					),
				)

				result, _ := doc.Query("li:nth-child(2n)")
				if len(result) != 2 {
					t.Fatalf("nth-child(2n) should match 2 elements (2nd and 4th), got %d", len(result))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestNthChildNegativeOffset(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "nth child negative offset",
			Run: func(t *testing.T) {
				doc := buildNode(
					el("ul", nil,
						el("li", nil),
						el("li", nil),
						el("li", nil),
						el("li", nil),
						el("li", nil),
					),
				)

				result, _ := doc.Query("li:nth-child(-n+3)")
				if len(result) != 3 {
					t.Fatalf("nth-child(-n+3) should match first 3 elements, got %d", len(result))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

// Tokenizer Edge Cases
func TestAttributeWithSpacesAroundOperator(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "attribute with spaces around operator",
			Run: func(t *testing.T) {
				doc := buildNode(
					el("div", map[string]string{"id": "test"}),
				)

				result, _ := doc.Query("[ id = test ]")
				if len(result) != 1 {
					t.Fatalf("attribute selector with spaces should work, got %d", len(result))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestCombinatorWithExtraSpaces(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "combinator with extra spaces",
			Run: func(t *testing.T) {
				doc := buildNode(
					el("div", nil,
						el("p", nil),
					),
				)

				result, _ := doc.Query("div   >   p")
				if len(result) != 1 {
					t.Fatalf("combinator with extra spaces should work, got %d", len(result))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestMultiplePseudoClasses(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "multiple pseudo classes",
			Run: func(t *testing.T) {
				doc := buildNode(
					el("ul", nil,
						el("li", nil),
					),
				)

				result, _ := doc.Query("li:first-child:last-child")
				if len(result) != 1 {
					t.Fatalf("multiple pseudo-classes should work, got %d", len(result))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestPseudoWithArgContainingSpaces(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "pseudo with arg containing spaces",
			Run: func(t *testing.T) {
				doc := buildNode(
					el("ul", nil,
						el("li", nil),
						el("li", nil),
						el("li", nil),
					),
				)

				result, _ := doc.Query("li:nth-child( 2n + 1 )")
				if len(result) != 2 {
					t.Fatalf("nth-child with spaces in arg should work, got %d", len(result))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

// Matcher Edge Cases
func TestEmptyPseudoWithComments(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "empty pseudo with comments",
			Run: func(t *testing.T) {
				doc := buildNode(
					el("div", nil,
						&Node{Name: "#comment", Data: "comment"},
					),
				)

				result, _ := doc.Query("div:empty")
				// Comments should be ignored per CSS spec
				if len(result) != 1 {
					t.Fatalf("div with only comment should match :empty, got %d", len(result))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestNthChildWithInvalidExpression(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "nth child with invalid expression",
			Run: func(t *testing.T) {
				doc := buildNode(
					el("ul", nil,
						el("li", nil),
						el("li", nil),
					),
				)

				result, _ := doc.Query("li:nth-child(invalid)")
				if len(result) != 0 {
					t.Fatalf("invalid nth-child expression should match nothing, got %d", len(result))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestNthOfTypeWithInvalidExpression(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "nth of type with invalid expression",
			Run: func(t *testing.T) {
				doc := buildNode(
					el("div", nil),
					el("div", nil),
				)

				result, _ := doc.Query("div:nth-of-type(invalid)")
				if len(result) != 0 {
					t.Fatalf("invalid nth-of-type expression should match nothing, got %d", len(result))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestOnlyChildNoMatch(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "only child no match",
			Run: func(t *testing.T) {
				doc := buildSimpleDoc()
				result, _ := doc.Query("li:only-child")
				if len(result) != 0 {
					t.Fatalf("only-child should not match when there are siblings, got %d", len(result))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestOnlyOfTypeNoMatch(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "only of type no match",
			Run: func(t *testing.T) {
				doc := buildSiblingDoc()
				result, _ := doc.Query("p:only-of-type")
				if len(result) != 0 {
					t.Fatalf("only-of-type should not match when there are multiple of same type, got %d", len(result))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestLastChildNotLast(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "last child not last",
			Run: func(t *testing.T) {
				doc := buildSiblingDoc()
				result, _ := doc.Query("h1:last-child")
				if len(result) != 0 {
					t.Fatalf("h1 is not last child, should not match, got %d", len(result))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestEmptyWithText(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "empty with text",
			Run: func(t *testing.T) {
				doc := buildEmptyAndRootDoc()
				result, _ := doc.Query(".text:empty")
				if len(result) != 0 {
					t.Fatalf("element with text should not match :empty, got %d", len(result))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestRootWithTag(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "root with tag",
			Run: func(t *testing.T) {
				doc := buildSimpleDoc()
				result, _ := doc.Query("html:root")
				if len(result) != 1 {
					t.Fatalf("html:root should match, got %d", len(result))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestNthOfTypeOdd(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "nth of type odd",
			Run: func(t *testing.T) {
				doc := buildSiblingDoc()
				result, _ := doc.Query("p:nth-of-type(odd)")
				if len(result) != 2 {
					t.Fatalf("nth-of-type(odd) should match 2 paragraphs, got %d", len(result))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

// Selector Groups
func TestTwoSelectors(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "two selectors",
			Run: func(t *testing.T) {
				doc := buildSimpleDoc()
				result, _ := doc.Query("h1, h2")
				if len(result) != 1 || result[0].Name != "h1" {
					t.Fatalf("selector group should find h1, got %d results", len(result))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestComplexSelectorGroups(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "complex selector groups",
			Run: func(t *testing.T) {
				doc := buildSimpleDoc()
				result, _ := doc.Query("#main p, #sidebar a")
				if len(result) != 3 {
					t.Fatalf("complex selector group should find 3 elements, got %d", len(result))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

// Complex Compound Selectors
func TestCompoundSelectorWithAttribute(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "compound selector with attribute",
			Run: func(t *testing.T) {
				doc := buildSimpleDoc()
				result, _ := doc.Query("a[href][data-id]")
				if len(result) != 1 {
					t.Fatalf("compound selector with multiple attributes should match, got %d", len(result))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestMultipleCombinators(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "multiple combinators",
			Run: func(t *testing.T) {
				doc := buildNestedDoc()
				result, _ := doc.Query(".a > .b > .c span")
				if len(result) != 1 {
					t.Fatalf("selector with multiple combinators should match, got %d", len(result))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestPseudoWithCombinator(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "pseudo with combinator",
			Run: func(t *testing.T) {
				doc := buildSiblingDoc()
				result, _ := doc.Query("div > p:first-child")
				if len(result) != 0 {
					t.Fatalf("div > p:first-child should not match (h1 is first), got %d", len(result))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

// Invalid Character Test
func TestInvalidCharacterInSelector(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "invalid character in selector",
			Run: func(t *testing.T) {
				_, err := parseSelector("div@foo")
				if err == nil {
					t.Fatalf("expected error for invalid character @")
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

// Case Sensitivity Tests
func TestClassSelectorCaseSensitive(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "class selector case sensitive",
			Run: func(t *testing.T) {
				doc := buildSimpleDoc()
				result, _ := doc.Query(".Container")
				if len(result) != 0 {
					t.Fatalf("class selector should be case-sensitive, got %d", len(result))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

// Additional Attribute Tests
func TestAttributeHyphenPrefixNoMatchWithoutHyphen(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "attribute hyphen prefix no match without hyphen",
			Run: func(t *testing.T) {
				doc := buildLangDoc("english")
				result, _ := doc.Query(`[lang|="en"]`)
				if len(result) != 0 {
					t.Fatalf("lang|=en should not match 'english', got %d", len(result))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestAttributeContainsWordEmptyClass(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "attribute contains word empty class",
			Run: func(t *testing.T) {
				doc := buildNode(
					el("div", map[string]string{"class": ""}),
				)

				result, _ := doc.Query(`[class~="foo"]`)
				if len(result) != 0 {
					t.Fatalf("~= should not match empty class, got %d", len(result))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

// First/Last of Type Edge Cases
func TestFirstOfTypeMultipleTypes(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "first of type multiple types",
			Run: func(t *testing.T) {
				doc := buildNode(
					el("div", nil),
					el("span", nil),
					el("div", nil),
				)

				result, _ := doc.Query("div:first-of-type")
				if len(result) != 1 {
					t.Fatalf("first-of-type should find first div, got %d", len(result))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestLastOfTypeMultipleTypes(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "last of type multiple types",
			Run: func(t *testing.T) {
				doc := buildNode(
					el("div", nil),
					el("span", nil),
					el("div", nil),
				)

				result, _ := doc.Query("div:last-of-type")
				if len(result) != 1 {
					t.Fatalf("last-of-type should find last div, got %d", len(result))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

// Not pseudo-class with empty argument
func TestNotWithClassSelector(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "not with class selector",
			Run: func(t *testing.T) {
				doc := buildSimpleDoc()
				result, _ := doc.Query("p:not(.intro)")
				if len(result) != 1 {
					t.Fatalf("p:not(.intro) should match one paragraph, got %d", len(result))
				}
				if strings.Contains(result[0].Attr("class"), "intro") {
					t.Fatalf("matched paragraph should not have intro class")
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestNotWithCombinator(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "not with combinator",
			Run: func(t *testing.T) {
				doc := buildSimpleDoc()
				result, _ := doc.Query("div > p:not(.content)")
				if len(result) != 1 {
					t.Fatalf("div > p:not(.content) should match one paragraph, got %d", len(result))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

// Sibling combinator edge cases
func TestGeneralSiblingNoMatch(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "general sibling no match",
			Run: func(t *testing.T) {
				doc := buildNode(
					el("div", nil,
						el("span", nil),
						el("p", nil),
					),
				)

				result, _ := doc.Query("div ~ p")
				if len(result) != 0 {
					t.Fatalf("div ~ p should not match when no div before p, got %d", len(result))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

func TestAdjacentSiblingNoMatch(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "adjacent sibling no match",
			Run: func(t *testing.T) {
				doc := buildNode(
					el("div", nil,
						el("h1", nil),
						el("span", nil),
						el("p", nil),
					),
				)

				result, _ := doc.Query("h1 + p")
				if len(result) != 0 {
					t.Fatalf("h1 + p should not match when span is between, got %d", len(result))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

// Universal selector variations
func TestUniversalSelectorSkipsTextNodes(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "universal selector skips text nodes",
			Run: func(t *testing.T) {
				root := buildTree()

				all, _ := root.Query("*")
				if len(all) != 4 { // div, span, em, p (exclude #text)
					t.Fatalf("universal selector matched %d nodes, want 4", len(all))
				}
				doc := buildNode(
					el("div", nil,
						&Node{Name: "#text", Data: "text"},
						el("span", nil),
					),
				)

				result, _ := doc.Query("*")
				// Should only match div and span, not text node
				for _, n := range result {
					if strings.HasPrefix(n.Name, "#") {
						t.Fatalf("universal selector should skip text nodes")
					}
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

// Minimal reproduction case
func TestDebugUnicodeClass(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "debug unicode class",
			Run: func(t *testing.T) {
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
				result, _ := doc.Query("div")

				if len(result) != 1 {
					t.Fatalf("Basic div query failed")
				}

				// Test 3: Parse the selector
				chains, err := parseSelector(`.über`)
				if err != nil {
					t.Fatalf("Failed to parse selector: %v", err)
				}

				if len(chains) > 0 && len(chains[0]) > 0 {
					className := chains[0][0].compound.classes[0]
					fmt.Printf("Parsed class Name: %q (bytes: %v)\n", className, []byte(className))
				}

				// Test 4: Try to match
				result, _ = doc.Query(`.über`)

				if len(result) != 1 {
					t.Fatalf("Unicode class query failed - expected 1, got %d", len(result))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

// Test if the problem is specific to class selectors
func TestDebugUnicodeAttribute(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "debug unicode attribute",
			Run: func(t *testing.T) {
				strPtr := func(s string) *string { return &s }

				div := &Node{
					Name:  "div",
					Attrs: map[string]*string{"data-name": strPtr("über")},
				}

				doc := &Node{Name: "document"}
				doc.AppendChild(div)

				// Try attribute selector
				result, _ := doc.Query(`[data-name="über"]`)

				if len(result) != 1 {
					t.Fatalf("Unicode attribute query failed - expected 1, got %d", len(result))
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}

// Test the readIdent function directly
func TestDebugReadIdent(t *testing.T) {
	tests := []common.TestCase{
		{
			Name: "debug read ident",
			Run: func(t *testing.T) {
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

					if gotIdent != tc.wantIdent {
						t.Errorf("readIdent(%q, %d): got ident %q, want %q",
							tc.input, tc.start, gotIdent, tc.wantIdent)
					}
					if gotNext != tc.wantNext {
						t.Errorf("readIdent(%q, %d): got next %d, want %d",
							tc.input, tc.start, gotNext, tc.wantNext)
					}
				}
			},
		},
	}

	common.RunTestCases(t, tests)
}
