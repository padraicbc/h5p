package dom

import (
	"math"
	"testing"
)

func TestQueryXPathBasic(t *testing.T) {
	doc := buildFullSimpleDoc()
	nodes, err := doc.QueryXPath("//p")
	if err != nil {
		t.Fatalf("QueryXPath error: %v", err)
	}
	if len(nodes) != 2 {
		t.Fatalf("QueryXPath(//p) = %d, want 2", len(nodes))
	}
}

func TestQueryXPathAttributeFilters(t *testing.T) {
	doc := buildFullSimpleDoc()
	nodes, err := doc.QueryXPath(`//div[@id="main"]`)
	if err != nil {
		t.Fatalf("QueryXPath error: %v", err)
	}
	if len(nodes) != 1 || nodes[0].Attr("id") != "main" {
		t.Fatalf("QueryXPath(@id) did not return main div")
	}

	nodes, err = doc.QueryXPath(`//*[@data-id]`)
	if err != nil {
		t.Fatalf("QueryXPath error: %v", err)
	}
	if len(nodes) != 1 || nodes[0].Name != "a" {
		t.Fatalf("QueryXPath(@data-id) did not return anchor")
	}

	nodes, err = doc.QueryXPath(`//div[contains(@class,"container")]`)
	if err != nil {
		t.Fatalf("QueryXPath error: %v", err)
	}
	if len(nodes) != 2 {
		t.Fatalf("QueryXPath(contains(@class)) = %d, want 2", len(nodes))
	}
}

func TestQueryXPathPosition(t *testing.T) {
	doc := buildFullSimpleDoc()
	nodes, err := doc.QueryXPath(`//ul/li[2]`)
	if err != nil {
		t.Fatalf("QueryXPath error: %v", err)
	}
	if len(nodes) != 1 || nodes[0].Name != "li" {
		t.Fatalf("QueryXPath position did not return second li")
	}
}

func TestQueryXPathTextPredicates(t *testing.T) {
	doc := buildNode(el("div", nil,
		&Node{Name: "#text", Data: "hello"},
		el("span", nil, &Node{Name: "#text", Data: "world"}),
	))

	nodes, err := doc.QueryXPath(`//div[text()="helloworld"]`)
	if err != nil {
		t.Fatalf("QueryXPath error: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("QueryXPath text equality = %d, want 1", len(nodes))
	}

	nodes, err = doc.QueryXPath(`//div[contains(text(),"hello")]`)
	if err != nil {
		t.Fatalf("QueryXPath error: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("QueryXPath text contains = %d, want 1", len(nodes))
	}
}

func TestQueryXPathPrecedingSiblingAxis(t *testing.T) {
	doc := buildNode(el("ul", map[string]string{"id": "list"},
		el("li", map[string]string{"class": "item first"}, &Node{Name: "#text", Data: "One"}),
		el("li", map[string]string{"class": "item second"}, &Node{Name: "#text", Data: "Two"}),
		el("li", map[string]string{"class": "item third"}, &Node{Name: "#text", Data: "Three"}),
	))

	nodes, err := doc.QueryXPath(`//ul[@id="list"]/li[preceding-sibling::li]`)
	if err != nil {
		t.Fatalf("QueryXPath error: %v", err)
	}
	if len(nodes) != 2 {
		t.Fatalf("QueryXPath preceding-sibling = %d, want 2", len(nodes))
	}
	if nodes[0].ToText("", false) != "Two" || nodes[1].ToText("", false) != "Three" {
		t.Fatalf("QueryXPath preceding-sibling returned unexpected nodes")
	}

	nodes, err = doc.QueryXPath(`//ul[@id="list"]/li[preceding-sibling::li][1]`)
	if err != nil {
		t.Fatalf("QueryXPath error: %v", err)
	}
	if len(nodes) != 1 || nodes[0].ToText("", false) != "Two" {
		t.Fatalf("QueryXPath preceding-sibling [1] mismatch")
	}
}

func TestQueryXPathFollowingSiblingAxis(t *testing.T) {
	doc := buildNode(el("div", map[string]string{"id": "siblings"},
		el("span", nil, &Node{Name: "#text", Data: "A"}),
		el("span", nil, &Node{Name: "#text", Data: "B"}),
		el("span", nil, &Node{Name: "#text", Data: "C"}),
	))

	nodes, err := doc.QueryXPath(`//*[@id="siblings"]/span[not(following-sibling::span)]`)
	if err != nil {
		t.Fatalf("QueryXPath error: %v", err)
	}
	if len(nodes) != 1 || nodes[0].ToText("", false) != "C" {
		t.Fatalf("QueryXPath following-sibling not() mismatch")
	}
}

func TestQueryXPathFunctions(t *testing.T) {
	doc := buildFullSimpleDoc()

	nodes, err := doc.QueryXPath(`//div[@id="main"][count(p)=2]`)
	if err != nil {
		t.Fatalf("QueryXPath error: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("QueryXPath count() = %d, want 1", len(nodes))
	}

	nodes, err = doc.QueryXPath(`//div[starts-with(@id,"side")]`)
	if err != nil {
		t.Fatalf("QueryXPath error: %v", err)
	}
	if len(nodes) != 1 || nodes[0].Attr("id") != "sidebar" {
		t.Fatalf("QueryXPath starts-with() did not return sidebar")
	}

	nodes, err = doc.QueryXPath(`//div[substring(@id,1,4)="side"]`)
	if err != nil {
		t.Fatalf("QueryXPath error: %v", err)
	}
	if len(nodes) != 1 || nodes[0].Attr("id") != "sidebar" {
		t.Fatalf("QueryXPath substring() did not return sidebar")
	}

	nodes, err = doc.QueryXPath(`//div[normalize-space(@class)="container secondary"]`)
	if err != nil {
		t.Fatalf("QueryXPath error: %v", err)
	}
	if len(nodes) != 1 || nodes[0].Attr("id") != "sidebar" {
		t.Fatalf("QueryXPath normalize-space() did not return sidebar")
	}

	nodes, err = doc.QueryXPath(`//div[translate(@id,"main","MAIN")="MAIN"]`)
	if err != nil {
		t.Fatalf("QueryXPath error: %v", err)
	}
	if len(nodes) != 1 || nodes[0].Attr("id") != "main" {
		t.Fatalf("QueryXPath translate() did not return main")
	}

	nodes, err = doc.QueryXPath(`//div[boolean(@id)]`)
	if err != nil {
		t.Fatalf("QueryXPath error: %v", err)
	}
	if len(nodes) != 2 {
		t.Fatalf("QueryXPath boolean() = %d, want 2", len(nodes))
	}

	nodes, err = doc.QueryXPath(`//div[not(@id)]`)
	if err != nil {
		t.Fatalf("QueryXPath error: %v", err)
	}
	if len(nodes) != 0 {
		t.Fatalf("QueryXPath not() = %d, want 0", len(nodes))
	}

	nodes, err = doc.QueryXPath(`//li[position()=2]`)
	if err != nil {
		t.Fatalf("QueryXPath error: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("QueryXPath position() = %d, want 1", len(nodes))
	}

	nodes, err = doc.QueryXPath(`//li[last()]`)
	if err != nil {
		t.Fatalf("QueryXPath error: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("QueryXPath last() = %d, want 1", len(nodes))
	}
}

func TestXPathQueryFirst(t *testing.T) {
	doc := buildFullSimpleDoc()

	node, err := doc.QueryXPathFirst("//div[@id=\"main\"]")
	if err != nil {
		t.Fatalf("QueryXPathFirst error: %v", err)
	}
	if node == nil || node.Attr("id") != "main" {
		t.Fatalf("QueryXPathFirst did not return main div")
	}

	node, err = doc.QueryXPathFirst("//does-not-exist")
	if err != nil {
		t.Fatalf("QueryXPathFirst error: %v", err)
	}
	if node != nil {
		t.Fatalf("QueryXPathFirst unexpected match")
	}

	_, err = doc.QueryXPathFirst("[")
	if err == nil {
		t.Fatalf("QueryXPathFirst should return error on parse errors")
	}
}

func TestXPathParseErrors(t *testing.T) {
	if _, err := parseXPath(""); err == nil {
		t.Fatalf("parseXPath should fail on empty selector")
	}
	if _, _, _, err := parseNodeTest("div", 3); err == nil {
		t.Fatalf("parseNodeTest should fail on missing node test")
	}
	if _, _, _, err := parseNodeTest("..", 0); err != nil {
		t.Fatalf("parseNodeTest should parse ..")
	}
	if _, _, _, err := parseNodeTest(".", 0); err != nil {
		t.Fatalf("parseNodeTest should parse .")
	}
	if _, _, _, err := parseNodeTest("*", 0); err != nil {
		t.Fatalf("parseNodeTest should parse *")
	}
	if _, err := parseXPath("div["); err == nil {
		t.Fatalf("parseXPath should fail on unclosed predicate")
	}
	if _, _, err := parseXPathStep("div[", 0, xpathAxisChild); err == nil {
		t.Fatalf("parseXPathStep should fail on unclosed predicate")
	}
	if _, err := parseXPathPredicate(""); err == nil {
		t.Fatalf("parseXPathPredicate should fail on empty predicate")
	}
	if _, err := parseXPathPredicate("1,2"); err == nil {
		t.Fatalf("parseXPathPredicate should fail on extra tokens")
	}
	if _, _, err := readXPathBracket("[", 0); err == nil {
		t.Fatalf("readXPathBracket should fail on unclosed bracket")
	}
}

func TestXPathAxisHelpers(t *testing.T) {
	root := &Node{Name: "div"}
	child := &Node{Name: "span"}
	text := &Node{Name: "#text", Data: "hello"}
	child.AppendChild(text)
	root.AppendChild(child)
	root.Template = &Node{Name: "template", Children: []*Node{{Name: "em"}}}

	if got := applyXPathAxis([]*Node{root}, xpathAxisChild, xpathNodeTestName, "."); len(got) != 1 || got[0] != root {
		t.Fatalf("applyXPathAxis . did not return self")
	}
	if got := applyXPathAxis([]*Node{child}, xpathAxisChild, xpathNodeTestName, ".."); len(got) != 1 || got[0] != root {
		t.Fatalf("applyXPathAxis .. did not return parent")
	}
	desc := applyXPathAxis([]*Node{root}, xpathAxisDescendant, xpathNodeTestName, "*")
	if len(desc) == 0 {
		t.Fatalf("applyXPathAxis descendant did not return nodes")
	}
	if !matchesNodeTest(text, xpathNodeTestText, "text()") {
		t.Fatalf("matchesNodeTest should match text()")
	}
	if matchesNodeTest(text, xpathNodeTestName, "div") {
		t.Fatalf("matchesNodeTest should not match text node")
	}
	if !matchesNodeTest(child, xpathNodeTestName, "*") {
		t.Fatalf("matchesNodeTest should match wildcard")
	}

	dupes := uniqueNodesInDocOrder([]*Node{child, child, root})
	if len(dupes) != 2 {
		t.Fatalf("uniqueNodesInDocOrder should remove duplicates")
	}

	if got := applyXPathAxis([]*Node{nil}, xpathAxisChild, xpathNodeTestName, "div"); len(got) != 0 {
		t.Fatalf("applyXPathAxis should ignore nil nodes")
	}
	if got := uniqueNodesInDocOrder([]*Node{nil}); len(got) != 0 {
		t.Fatalf("uniqueNodesInDocOrder should ignore nil entries")
	}
	if matchesNodeTest(&Node{}, xpathNodeTestName, "div") {
		t.Fatalf("matchesNodeTest should reject empty node names")
	}
	if matchesNodeTest(&Node{Name: "#comment"}, xpathNodeTestName, "div") {
		t.Fatalf("matchesNodeTest should reject comment nodes")
	}
	collectDescendants(nil, xpathNodeTestName, "*", &[]*Node{}, false)
	if axis, test, ok := splitXPathAxis("preceding-sibling::li"); !ok || axis != "preceding-sibling" || test != "li" {
		t.Fatalf("splitXPathAxis did not parse preceding-sibling")
	}
	if _, _, ok := splitXPathAxis("badaxis"); ok {
		t.Fatalf("splitXPathAxis should reject missing axis delimiter")
	}
	if _, _, ok := splitXPathAxis("::li"); ok {
		t.Fatalf("splitXPathAxis should reject empty axis")
	}
	if _, _, ok := splitXPathAxis("preceding-sibling::"); ok {
		t.Fatalf("splitXPathAxis should reject empty test")
	}
}

func TestXPathValuesAndConversions(t *testing.T) {
	text := &Node{Name: "#text", Data: "hello"}
	element := &Node{Name: "div", Children: []*Node{text}}

	attrNode := newAttrValue("id", "main")
	if nodeStringValue(attrNode) != "main" {
		t.Fatalf("nodeStringValue attribute mismatch")
	}
	if nodeStringValue(newNodeValue(text)) != "hello" {
		t.Fatalf("nodeStringValue text mismatch")
	}
	if nodeStringValue(newNodeValue(element)) != "hello" {
		t.Fatalf("nodeStringValue element mismatch")
	}
	if nodeStringValue(xpathNode{kind: xpathNodeText}) != "" {
		t.Fatalf("nodeStringValue nil text mismatch")
	}

	if !(xpathValue{kind: xpathValueBoolean, b: true}).toBool() {
		t.Fatalf("toBool boolean true mismatch")
	}
	if (xpathValue{kind: xpathValueNumber, num: 0}).toBool() {
		t.Fatalf("toBool number zero mismatch")
	}
	if (xpathValue{kind: xpathValueString, str: ""}).toBool() {
		t.Fatalf("toBool empty string mismatch")
	}
	if !(xpathValue{kind: xpathValueNodeSet, nodes: []xpathNode{newNodeValue(element)}}).toBool() {
		t.Fatalf("toBool nodeset mismatch")
	}
	if (xpathValue{kind: xpathValueNumber, num: math.NaN()}).toBool() {
		t.Fatalf("toBool NaN mismatch")
	}

	if got := (xpathValue{kind: xpathValueBoolean, b: true}).toNumber(); got != 1 {
		t.Fatalf("toNumber boolean mismatch")
	}
	if got := (xpathValue{kind: xpathValueString, str: "10"}).toNumber(); got != 10 {
		t.Fatalf("toNumber string mismatch")
	}
	if !math.IsNaN((xpathValue{kind: xpathValueString, str: "bad"}).toNumber()) {
		t.Fatalf("toNumber should return NaN for invalid numbers")
	}
	if got := (xpathValue{kind: xpathValueNodeSet, nodes: []xpathNode{newNodeValue(text)}}).toNumber(); !math.IsNaN(got) {
		t.Fatalf("toNumber nodeset mismatch")
	}
	if got := (xpathValue{kind: xpathValueString}).toNumber(); got != 0 {
		t.Fatalf("toNumber empty string mismatch")
	}

	if (xpathValue{kind: xpathValueBoolean, b: true}).toString() != "true" {
		t.Fatalf("toString boolean mismatch")
	}
	if (xpathValue{kind: xpathValueString, str: "x"}).toString() != "x" {
		t.Fatalf("toString string mismatch")
	}
	if (xpathValue{kind: xpathValueNumber, num: math.NaN()}).toString() != "NaN" {
		t.Fatalf("toString NaN mismatch")
	}
	if (xpathValue{kind: xpathValueNumber, num: math.Inf(1)}).toString() != "Infinity" {
		t.Fatalf("toString +Inf mismatch")
	}
	if (xpathValue{kind: xpathValueNumber, num: math.Inf(-1)}).toString() != "-Infinity" {
		t.Fatalf("toString -Inf mismatch")
	}
	if (xpathValue{kind: xpathValueNodeSet}).toString() != "" {
		t.Fatalf("toString empty nodeset mismatch")
	}
	if (xpathValue{kind: xpathValueNodeSet, nodes: []xpathNode{newNodeValue(text)}}).toString() != "hello" {
		t.Fatalf("toString nodeset mismatch")
	}
}

func TestXPathExpressionsAndOperators(t *testing.T) {
	ctx := xpathContext{node: &Node{Name: "div"}, position: 1, size: 3}

	unary := xpathUnaryExpr{op: "-", expr: xpathLiteralExpr{value: xpathValue{kind: xpathValueNumber, num: 3}}}
	if val, _ := unary.eval(ctx); val.num != -3 {
		t.Fatalf("unary expr mismatch")
	}

	expr := xpathBinaryExpr{
		op:    "+",
		left:  xpathLiteralExpr{value: xpathValue{kind: xpathValueNumber, num: 1}},
		right: xpathLiteralExpr{value: xpathValue{kind: xpathValueNumber, num: 2}},
	}
	if val, _ := expr.eval(ctx); val.num != 3 {
		t.Fatalf("binary expr mismatch")
	}

	rel := xpathBinaryExpr{
		op:    "<",
		left:  xpathLiteralExpr{value: xpathValue{kind: xpathValueNumber, num: 1}},
		right: xpathLiteralExpr{value: xpathValue{kind: xpathValueNumber, num: 2}},
	}
	if val, _ := rel.eval(ctx); !val.b {
		t.Fatalf("relational expr mismatch")
	}

	andExpr := xpathBinaryExpr{
		op:    "and",
		left:  xpathLiteralExpr{value: xpathValue{kind: xpathValueBoolean, b: true}},
		right: xpathLiteralExpr{value: xpathValue{kind: xpathValueBoolean, b: false}},
	}
	if val, _ := andExpr.eval(ctx); val.b {
		t.Fatalf("and expr mismatch")
	}

	orExpr := xpathBinaryExpr{
		op:    "or",
		left:  xpathLiteralExpr{value: xpathValue{kind: xpathValueBoolean, b: false}},
		right: xpathLiteralExpr{value: xpathValue{kind: xpathValueBoolean, b: true}},
	}
	if val, _ := orExpr.eval(ctx); !val.b {
		t.Fatalf("or expr mismatch")
	}

	invalid := xpathBinaryExpr{
		op:    "??",
		left:  xpathLiteralExpr{value: xpathValue{kind: xpathValueNumber, num: 1}},
		right: xpathLiteralExpr{value: xpathValue{kind: xpathValueNumber, num: 1}},
	}
	if _, err := invalid.eval(ctx); err == nil {
		t.Fatalf("invalid operator should error")
	}

	if !compareBool(true, true, "=") || compareBool(true, false, "=") {
		t.Fatalf("compareBool mismatch")
	}
	if !compareNumber(1, 2, "<") || compareNumber(2, 1, "<") {
		t.Fatalf("compareNumber mismatch")
	}
	if !compareString("a", "b", "<") || compareString("a", "b", ">=") {
		t.Fatalf("compareString mismatch")
	}
	if compareBool(true, false, "??") {
		t.Fatalf("compareBool default mismatch")
	}
	if compareNumber(1, 2, "??") {
		t.Fatalf("compareNumber default mismatch")
	}
	if compareString("a", "b", "??") {
		t.Fatalf("compareString default mismatch")
	}
}

func TestXPathFunctionsAndErrors(t *testing.T) {
	doc := buildFullSimpleDoc()
	ctx := xpathContext{root: doc, node: &Node{Name: "div", Children: []*Node{{Name: "#text", Data: "text"}}}, position: 2, size: 5}

	for _, name := range []string{"true", "false"} {
		fn := xpathFunctionExpr{name: name}
		if _, err := fn.eval(ctx); err != nil {
			t.Fatalf("%s() should not error", name)
		}
	}

	fn := xpathFunctionExpr{name: "position"}
	if val, _ := fn.eval(ctx); val.num != 2 {
		t.Fatalf("position() mismatch")
	}

	fn = xpathFunctionExpr{name: "last"}
	if val, _ := fn.eval(ctx); val.num != 5 {
		t.Fatalf("last() mismatch")
	}

	fn = xpathFunctionExpr{name: "count", args: []xpathExpr{xpathLiteralExpr{value: xpathValue{kind: xpathValueNodeSet, nodes: []xpathNode{newNodeValue(doc)}}}}}
	if val, _ := fn.eval(ctx); val.num != 1 {
		t.Fatalf("count() mismatch")
	}

	fn = xpathFunctionExpr{name: "id", args: []xpathExpr{xpathLiteralExpr{value: xpathValue{kind: xpathValueString, str: "main sidebar"}}}}
	if val, _ := fn.eval(ctx); val.kind != xpathValueNodeSet || len(val.nodes) != 2 {
		t.Fatalf("id() mismatch")
	}

	fn = xpathFunctionExpr{name: "local-name", args: []xpathExpr{}}
	if val, _ := fn.eval(ctx); val.str == "" {
		t.Fatalf("local-name() mismatch")
	}

	fn = xpathFunctionExpr{name: "namespace-uri"}
	if val, _ := fn.eval(ctx); val.str != "" {
		t.Fatalf("namespace-uri() mismatch")
	}

	fn = xpathFunctionExpr{name: "name"}
	if val, _ := fn.eval(ctx); val.str == "" {
		t.Fatalf("name() mismatch")
	}

	fn = xpathFunctionExpr{name: "string", args: []xpathExpr{}}
	if val, _ := fn.eval(ctx); val.str == "" {
		t.Fatalf("string() mismatch")
	}
	fn = xpathFunctionExpr{name: "string", args: []xpathExpr{
		xpathLiteralExpr{value: xpathValue{kind: xpathValueNumber, num: 42}},
	}}
	if val, _ := fn.eval(ctx); val.str != "42" {
		t.Fatalf("string() arg mismatch")
	}

	fn = xpathFunctionExpr{name: "concat", args: []xpathExpr{
		xpathLiteralExpr{value: xpathValue{kind: xpathValueString, str: "a"}},
		xpathLiteralExpr{value: xpathValue{kind: xpathValueString, str: "b"}},
	}}
	if val, _ := fn.eval(ctx); val.str != "ab" {
		t.Fatalf("concat() mismatch")
	}

	fn = xpathFunctionExpr{name: "starts-with", args: []xpathExpr{
		xpathLiteralExpr{value: xpathValue{kind: xpathValueString, str: "abc"}},
		xpathLiteralExpr{value: xpathValue{kind: xpathValueString, str: "a"}},
	}}
	if val, _ := fn.eval(ctx); !val.b {
		t.Fatalf("starts-with() mismatch")
	}

	fn = xpathFunctionExpr{name: "contains", args: []xpathExpr{
		xpathLiteralExpr{value: xpathValue{kind: xpathValueString, str: "abc"}},
		xpathLiteralExpr{value: xpathValue{kind: xpathValueString, str: "b"}},
	}}
	if val, _ := fn.eval(ctx); !val.b {
		t.Fatalf("contains() mismatch")
	}
	fn = xpathFunctionExpr{name: "contains", args: []xpathExpr{
		errExpr{},
		xpathLiteralExpr{value: xpathValue{kind: xpathValueString, str: "b"}},
	}}
	if _, err := fn.eval(ctx); err == nil {
		t.Fatalf("contains() should propagate errors")
	}

	fn = xpathFunctionExpr{name: "substring-before", args: []xpathExpr{
		xpathLiteralExpr{value: xpathValue{kind: xpathValueString, str: "abc"}},
		xpathLiteralExpr{value: xpathValue{kind: xpathValueString, str: "b"}},
	}}
	if val, _ := fn.eval(ctx); val.str != "a" {
		t.Fatalf("substring-before() mismatch")
	}
	fn = xpathFunctionExpr{name: "substring-before", args: []xpathExpr{
		xpathLiteralExpr{value: xpathValue{kind: xpathValueString, str: "abc"}},
		xpathLiteralExpr{value: xpathValue{kind: xpathValueString, str: ""}},
	}}
	if val, _ := fn.eval(ctx); val.str != "" {
		t.Fatalf("substring-before empty delimiter mismatch")
	}

	fn = xpathFunctionExpr{name: "substring-after", args: []xpathExpr{
		xpathLiteralExpr{value: xpathValue{kind: xpathValueString, str: "abc"}},
		xpathLiteralExpr{value: xpathValue{kind: xpathValueString, str: "b"}},
	}}
	if val, _ := fn.eval(ctx); val.str != "c" {
		t.Fatalf("substring-after() mismatch")
	}
	fn = xpathFunctionExpr{name: "substring-after", args: []xpathExpr{
		xpathLiteralExpr{value: xpathValue{kind: xpathValueString, str: "abc"}},
		xpathLiteralExpr{value: xpathValue{kind: xpathValueString, str: ""}},
	}}
	if val, _ := fn.eval(ctx); val.str != "" {
		t.Fatalf("substring-after empty delimiter mismatch")
	}

	fn = xpathFunctionExpr{name: "substring", args: []xpathExpr{
		xpathLiteralExpr{value: xpathValue{kind: xpathValueString, str: "abcd"}},
		xpathLiteralExpr{value: xpathValue{kind: xpathValueNumber, num: 2}},
		xpathLiteralExpr{value: xpathValue{kind: xpathValueNumber, num: 2}},
	}}
	if val, _ := fn.eval(ctx); val.str != "bc" {
		t.Fatalf("substring() mismatch")
	}

	fn = xpathFunctionExpr{name: "string-length", args: []xpathExpr{
		xpathLiteralExpr{value: xpathValue{kind: xpathValueString, str: "ab"}},
	}}
	if val, _ := fn.eval(ctx); val.num != 2 {
		t.Fatalf("string-length() mismatch")
	}

	fn = xpathFunctionExpr{name: "normalize-space", args: []xpathExpr{
		xpathLiteralExpr{value: xpathValue{kind: xpathValueString, str: " a  b "}},
	}}
	if val, _ := fn.eval(ctx); val.str != "a b" {
		t.Fatalf("normalize-space() mismatch")
	}

	fn = xpathFunctionExpr{name: "translate", args: []xpathExpr{
		xpathLiteralExpr{value: xpathValue{kind: xpathValueString, str: "abc"}},
		xpathLiteralExpr{value: xpathValue{kind: xpathValueString, str: "b"}},
		xpathLiteralExpr{value: xpathValue{kind: xpathValueString, str: "B"}},
	}}
	if val, _ := fn.eval(ctx); val.str != "aBc" {
		t.Fatalf("translate() mismatch")
	}

	fn = xpathFunctionExpr{name: "boolean", args: []xpathExpr{
		xpathLiteralExpr{value: xpathValue{kind: xpathValueString, str: "x"}},
	}}
	if val, _ := fn.eval(ctx); !val.b {
		t.Fatalf("boolean() mismatch")
	}

	fn = xpathFunctionExpr{name: "not", args: []xpathExpr{
		xpathLiteralExpr{value: xpathValue{kind: xpathValueBoolean, b: true}},
	}}
	if val, _ := fn.eval(ctx); val.b {
		t.Fatalf("not() mismatch")
	}

	fn = xpathFunctionExpr{name: "lang", args: []xpathExpr{
		xpathLiteralExpr{value: xpathValue{kind: xpathValueString, str: "en"}},
	}}
	if val, _ := fn.eval(ctx); val.b {
		t.Fatalf("lang() mismatch")
	}

	fn = xpathFunctionExpr{name: "number", args: []xpathExpr{
		xpathLiteralExpr{value: xpathValue{kind: xpathValueString, str: "4"}},
	}}
	if val, _ := fn.eval(ctx); val.num != 4 {
		t.Fatalf("number() mismatch")
	}
	fn = xpathFunctionExpr{name: "number"}
	if val, _ := fn.eval(ctx); !math.IsNaN(val.num) && val.num != 0 {
		t.Fatalf("number() no-arg mismatch")
	}

	fn = xpathFunctionExpr{name: "sum", args: []xpathExpr{
		xpathLiteralExpr{value: xpathValue{kind: xpathValueNodeSet, nodes: []xpathNode{newAttrValue("id", "2"), newAttrValue("id", "3")}}},
	}}
	if val, _ := fn.eval(ctx); val.num != 5 {
		t.Fatalf("sum() mismatch")
	}

	fn = xpathFunctionExpr{name: "floor", args: []xpathExpr{
		xpathLiteralExpr{value: xpathValue{kind: xpathValueNumber, num: 1.9}},
	}}
	if val, _ := fn.eval(ctx); val.num != 1 {
		t.Fatalf("floor() mismatch")
	}

	fn = xpathFunctionExpr{name: "ceiling", args: []xpathExpr{
		xpathLiteralExpr{value: xpathValue{kind: xpathValueNumber, num: 1.1}},
	}}
	if val, _ := fn.eval(ctx); val.num != 2 {
		t.Fatalf("ceiling() mismatch")
	}

	fn = xpathFunctionExpr{name: "round", args: []xpathExpr{
		xpathLiteralExpr{value: xpathValue{kind: xpathValueNumber, num: 1.4}},
	}}
	if val, _ := fn.eval(ctx); val.num != 1 {
		t.Fatalf("round() mismatch")
	}

	fn = xpathFunctionExpr{name: "round", args: []xpathExpr{
		xpathLiteralExpr{value: xpathValue{kind: xpathValueNumber, num: -1.4}},
	}}
	if val, _ := fn.eval(ctx); val.num != -1 {
		t.Fatalf("round() negative mismatch")
	}

	for _, invalid := range []xpathFunctionExpr{
		{name: "count", args: []xpathExpr{}},
		{name: "count", args: []xpathExpr{xpathLiteralExpr{value: xpathValue{kind: xpathValueString, str: "x"}}}},
		{name: "id", args: []xpathExpr{}},
		{name: "starts-with", args: []xpathExpr{}},
		{name: "contains", args: []xpathExpr{}},
		{name: "substring-before", args: []xpathExpr{}},
		{name: "substring-after", args: []xpathExpr{}},
		{name: "substring", args: []xpathExpr{}},
		{name: "translate", args: []xpathExpr{}},
		{name: "boolean", args: []xpathExpr{}},
		{name: "not", args: []xpathExpr{}},
		{name: "lang", args: []xpathExpr{}},
		{name: "number", args: []xpathExpr{xpathLiteralExpr{value: xpathValue{kind: xpathValueNumber, num: 1}}, xpathLiteralExpr{value: xpathValue{kind: xpathValueNumber, num: 2}}}},
		{name: "sum", args: []xpathExpr{}},
		{name: "sum", args: []xpathExpr{xpathLiteralExpr{value: xpathValue{kind: xpathValueString, str: "x"}}}},
		{name: "floor", args: []xpathExpr{}},
		{name: "ceiling", args: []xpathExpr{}},
		{name: "round", args: []xpathExpr{}},
		{name: "text", args: []xpathExpr{xpathLiteralExpr{value: xpathValue{kind: xpathValueString, str: "x"}}}},
		{name: "unknown"},
	} {
		if _, err := invalid.eval(ctx); err == nil {
			t.Fatalf("%s() should error", invalid.name)
		}
	}
}

func TestXPathParserInternals(t *testing.T) {
	parser := newXPathExprParser("1 + 2 * 3")
	expr, err := parser.parseExpr()
	if err != nil {
		t.Fatalf("parseExpr unexpected error: %v", err)
	}
	if parser.peek().typ != xpathTokenEOF {
		t.Fatalf("parser should consume all tokens")
	}

	ctx := xpathContext{}
	if val, _ := expr.eval(ctx); val.num != 7 {
		t.Fatalf("parseExpr eval mismatch")
	}

	parser = newXPathExprParser("@")
	if _, err := parser.parseExpr(); err == nil {
		t.Fatalf("parseExpr should error on missing attribute name")
	}

	parser = newXPathExprParser("func(")
	if _, err := parser.parseExpr(); err == nil {
		t.Fatalf("parseExpr should error on invalid function args")
	}

	parser = newXPathExprParser("1 = 1 or 2 = 2 and 3 = 3")
	if _, err := parser.parseExpr(); err != nil {
		t.Fatalf("parseExpr should parse logical operators: %v", err)
	}

	parser = newXPathExprParser("1 < 2")
	if _, err := parser.parseExpr(); err != nil {
		t.Fatalf("parseExpr should parse relational operators: %v", err)
	}

	parser = newXPathExprParser("(-1 + +2)")
	expr, err = parser.parseExpr()
	if err != nil {
		t.Fatalf("parseExpr should parse unary ops: %v", err)
	}
	if val, _ := expr.eval(xpathContext{}); val.num != 1 {
		t.Fatalf("unary parse eval mismatch")
	}

	parser = newXPathExprParser(".")
	expr, err = parser.parseExpr()
	if err != nil {
		t.Fatalf("parseExpr dot should parse: %v", err)
	}
	if val, _ := expr.eval(xpathContext{node: &Node{Name: "div"}}); val.kind != xpathValueNodeSet {
		t.Fatalf("dot eval mismatch")
	}

	parser = newXPathExprParser("(1")
	if _, err := parser.parseExpr(); err == nil {
		t.Fatalf("parseExpr should error on missing closing paren")
	}
}

func TestXPathLexerCoverage(t *testing.T) {
	lexer := newXPathLexer("  @id != 10 and name() = 'div'")
	tok := lexer.nextToken()
	if tok.typ != xpathTokenAt {
		t.Fatalf("expected @ token")
	}
	_ = lexer.nextToken()
	_ = lexer.nextToken()
	_ = lexer.nextToken()
	_ = lexer.nextToken()

	lexer = newXPathLexer("<= >= = < > + - * !=")
	for tok = lexer.nextToken(); tok.typ != xpathTokenEOF; tok = lexer.nextToken() {
	}

	lexer = newXPathLexer("123.4")
	tok = lexer.nextToken()
	if tok.typ != xpathTokenNumber {
		t.Fatalf("expected number token")
	}

	lexer = newXPathLexer(".")
	tok = lexer.nextToken()
	if tok.typ != xpathTokenDot {
		t.Fatalf("expected dot token")
	}

	lexer = newXPathLexer("node mod")
	tok = lexer.nextToken()
	if tok.typ != xpathTokenIdentifier {
		t.Fatalf("expected identifier token")
	}
	tok = lexer.nextToken()
	if tok.typ != xpathTokenOperator {
		t.Fatalf("expected operator token")
	}

	lexer = newXPathLexer("!")
	tok = lexer.nextToken()
	if tok.typ != xpathTokenOperator {
		t.Fatalf("expected operator token for !")
	}

	lexer = newXPathLexer("@\x80")
	_ = lexer.nextToken()
	tok = lexer.nextToken()
	if tok.typ != xpathTokenEOF {
		t.Fatalf("expected EOF on invalid utf-8")
	}
}

func TestXPathHelperCoverage(t *testing.T) {
	ctx := xpathContext{node: &Node{Name: "div", Attrs: map[string]*string{"lang": strPtr("en-US")}}}

	if compareXPathValues(
		xpathValue{kind: xpathValueBoolean, b: true},
		xpathValue{kind: xpathValueBoolean, b: false},
		"=",
	).b {
		t.Fatalf("compareXPathValues boolean mismatch")
	}

	if !compareXPathValues(
		xpathValue{kind: xpathValueNumber, num: 1},
		xpathValue{kind: xpathValueNumber, num: 2},
		"<",
	).b {
		t.Fatalf("compareXPathValues number mismatch")
	}

	if !compareXPathValues(
		xpathValue{kind: xpathValueString, str: "a"},
		xpathValue{kind: xpathValueString, str: "b"},
		"<",
	).b {
		t.Fatalf("compareXPathValues string mismatch")
	}

	if err := func() error {
		_, err := evalStringArgOptional(ctx, []xpathExpr{
			xpathLiteralExpr{value: xpathValue{kind: xpathValueString, str: "a"}},
			xpathLiteralExpr{value: xpathValue{kind: xpathValueString, str: "b"}},
		})
		return err
	}(); err == nil {
		t.Fatalf("evalStringArgOptional should error on invalid args")
	}
	if got, err := evalStringArgOptional(xpathContext{node: &Node{Name: "div", Children: []*Node{{Name: "#text", Data: "x"}}}}, nil); err != nil || got != "x" {
		t.Fatalf("evalStringArgOptional default mismatch")
	}
	if _, err := evalStringArg(xpathContext{}, errExpr{}); err == nil {
		t.Fatalf("evalStringArg should surface errors")
	}

	if _, err := evalMathUnary(ctx, []xpathExpr{}, math.Floor, "floor"); err == nil {
		t.Fatalf("evalMathUnary should error on missing args")
	}
	if _, err := evalMathUnary(ctx, []xpathExpr{errExpr{}}, math.Floor, "floor"); err == nil {
		t.Fatalf("evalMathUnary should surface errors")
	}

	if got := xpathRound(math.NaN()); !math.IsNaN(got) {
		t.Fatalf("xpathRound should preserve NaN")
	}

	if got := translateString("abc", "b", ""); got != "ac" {
		t.Fatalf("translateString removal mismatch")
	}

	if got := idLookup(nil, nil); got != nil {
		t.Fatalf("idLookup should return nil for nil root")
	}
	if got := idLookup(&Node{Name: "div"}, nil); got != nil {
		t.Fatalf("idLookup should return nil for empty ids")
	}

	if localName(nil) != "" {
		t.Fatalf("localName should return empty string for nil node")
	}

	args := []xpathExpr{
		xpathLiteralExpr{value: xpathValue{kind: xpathValueNodeSet, nodes: []xpathNode{newNodeValue(&Node{Name: "span"})}}},
	}
	if node := nodeForNameFunction(xpathContext{}, args); node == nil || node.Name != "span" {
		t.Fatalf("nodeForNameFunction mismatch")
	}
	if node := nodeForNameFunction(xpathContext{node: &Node{Name: "div"}}, nil); node == nil || node.Name != "div" {
		t.Fatalf("nodeForNameFunction default mismatch")
	}
	if node := nodeForNameFunction(xpathContext{}, []xpathExpr{
		xpathLiteralExpr{value: xpathValue{kind: xpathValueNodeSet}},
	}); node != nil {
		t.Fatalf("nodeForNameFunction empty nodeset mismatch")
	}
	if node := nodeForNameFunction(xpathContext{}, []xpathExpr{errExpr{}}); node != nil {
		t.Fatalf("nodeForNameFunction should return nil on error")
	}

	siblingNodes := precedingSiblingNodes(&Node{Name: "div"}, "span")
	if siblingNodes != nil {
		t.Fatalf("precedingSiblingNodes should return nil without parent")
	}
	siblingNodes = followingSiblingNodes(&Node{Name: "div"}, "span")
	if siblingNodes != nil {
		t.Fatalf("followingSiblingNodes should return nil without parent")
	}

	if !xpathLang(ctx.node, "en") {
		t.Fatalf("xpathLang should match language")
	}
	if xpathLang(ctx.node, "fr") {
		t.Fatalf("xpathLang should not match non-matching language")
	}
}

func TestXPathSubstringEdges(t *testing.T) {
	ctx := xpathContext{}
	val, err := evalSubstring(ctx, []xpathExpr{
		xpathLiteralExpr{value: xpathValue{kind: xpathValueString, str: "abc"}},
		xpathLiteralExpr{value: xpathValue{kind: xpathValueNumber, num: -2}},
	})
	if err != nil || val.str != "abc" {
		t.Fatalf("evalSubstring negative start mismatch")
	}

	val, err = evalSubstring(ctx, []xpathExpr{
		xpathLiteralExpr{value: xpathValue{kind: xpathValueString, str: ""}},
		xpathLiteralExpr{value: xpathValue{kind: xpathValueNumber, num: 1}},
	})
	if err != nil || val.str != "" {
		t.Fatalf("evalSubstring empty string mismatch")
	}

	val, err = evalSubstring(ctx, []xpathExpr{
		xpathLiteralExpr{value: xpathValue{kind: xpathValueString, str: "abc"}},
		xpathLiteralExpr{value: xpathValue{kind: xpathValueNumber, num: 5}},
		xpathLiteralExpr{value: xpathValue{kind: xpathValueNumber, num: 2}},
	})
	if err != nil || val.str != "" {
		t.Fatalf("evalSubstring start beyond length mismatch")
	}

	val, err = evalSubstring(ctx, []xpathExpr{
		xpathLiteralExpr{value: xpathValue{kind: xpathValueString, str: "abc"}},
		xpathLiteralExpr{value: xpathValue{kind: xpathValueNumber, num: 2}},
		xpathLiteralExpr{value: xpathValue{kind: xpathValueNumber, num: -1}},
	})
	if err != nil || val.str != "" {
		t.Fatalf("evalSubstring negative length mismatch")
	}

	val, err = evalSubstring(ctx, []xpathExpr{
		xpathLiteralExpr{value: xpathValue{kind: xpathValueString, str: "abc"}},
		xpathLiteralExpr{value: xpathValue{kind: xpathValueNumber, num: 2}},
		xpathLiteralExpr{value: xpathValue{kind: xpathValueNumber, num: 10}},
	})
	if err != nil || val.str != "bc" {
		t.Fatalf("evalSubstring long length mismatch")
	}
	if _, err := evalSubstring(ctx, []xpathExpr{errExpr{}}); err == nil {
		t.Fatalf("evalSubstring should surface errors")
	}
}

type errExpr struct{}

func (errExpr) eval(ctx xpathContext) (xpathValue, error) {
	return xpathValue{}, XPathError{"boom"}
}

// Helper to create a simple DOM tree for testing
func createTestTree() *Node {
	root := &Node{
		Name: "root",
		Children: []*Node{
			{
				Name: "div",
				Attrs: map[string]*string{
					"id":    strPtr("first"),
					"class": strPtr("container"),
				},
				Children: []*Node{
					{
						Name: "span",
						Children: []*Node{
							{Name: "#text", Data: "Hello"},
						},
					},
					{
						Name: "p",
						Children: []*Node{
							{Name: "#text", Data: "World"},
						},
					},
					{Name: "#text", Data: "Some text"},
				},
			},
			{
				Name: "div",
				Attrs: map[string]*string{
					"id":    strPtr("second"),
					"class": strPtr("item"),
				},
				Children: []*Node{
					{Name: "a", Attrs: map[string]*string{"href": strPtr("http://example.com")}},
					{
						Name: "span",
						Children: []*Node{
							{Name: "#text", Data: "Goodbye"},
						},
					},
				},
			},
			{
				Name: "ul",
				Children: []*Node{
					{
						Name: "li",
						Children: []*Node{
							{Name: "#text", Data: "Item 1"},
						},
					},
					{
						Name: "li",
						Children: []*Node{
							{Name: "#text", Data: "Item 2"},
						},
					},
					{
						Name: "li",
						Children: []*Node{
							{Name: "#text", Data: "Item 3"},
						},
					},
				},
			},
		},
	}

	// Set parent references
	setParents(root)
	return root
}

func setParents(n *Node) {
	for _, child := range n.Children {
		child.Parent = n
		setParents(child)
	}
}

func strPtr(s string) *string {
	return &s
}

// Test all axis support
func TestAllAxes(t *testing.T) {
	root := createTestTree()

	tests := []struct {
		name     string
		xpath    string
		expected int
	}{
		// Child axis (default)
		{"child axis", "/root/div", 2},
		{"explicit child", "/root/child::div", 2},

		// Descendant axis
		{"descendant", "/root/descendant::span", 2},
		{"descendant-or-self", "/root/descendant-or-self::root", 1},

		// Parent axis
		{"parent from span", "//span/parent::div", 2},

		// Ancestor axis
		{"ancestor", "//span/ancestor::root", 1}, // Both spans have root as ancestor, but deduplication returns 1
		{"ancestor-or-self", "//span/ancestor-or-self::span", 2},

		// Following-sibling
		{"following-sibling", "//div[@id='first']/following-sibling::div", 1},

		// Preceding-sibling
		{"preceding-sibling", "//div[@id='second']/preceding-sibling::div", 1},

		// Self axis
		{"self", "//div/self::div", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodes, err := root.QueryXPath(tt.xpath)
			if err != nil {
				t.Fatalf("Error: %v", err)
			}
			if len(nodes) != tt.expected {
				t.Errorf("Expected %d nodes, got %d for xpath: %s", tt.expected, len(nodes), tt.xpath)
			}
		})
	}
}

// Test node type tests
func TestNodeTests(t *testing.T) {
	root := createTestTree()

	tests := []struct {
		name     string
		xpath    string
		expected int
	}{
		{"text nodes", "//text()", 7}, // 7 text nodes in tree
		{"any node", "//node()", 18},  // All descendants: 11 elements + 7 text = 18 (not including root itself)
		{"element by name", "//div", 2},
		{"wildcard", "//div/*", 4}, // * matches only elements: first div has span,p; second div has a,span
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodes, err := root.QueryXPath(tt.xpath)
			if err != nil {
				t.Fatalf("Error: %v", err)
			}
			if len(nodes) != tt.expected {
				t.Errorf("Expected %d nodes, got %d for xpath: %s", tt.expected, len(nodes), tt.xpath)
			}
		})
	}
}

// Test predicates
func TestPredicates(t *testing.T) {
	root := createTestTree()

	tests := []struct {
		name     string
		xpath    string
		expected int
	}{
		{"position predicate", "//li[1]", 1},
		{"last position", "//li[last()]", 1},
		{"position > 1", "//li[position() > 1]", 2},
		{"attribute predicate", "//div[@id='first']", 1},
		{"or predicate", "//div[@id='first' or @id='second']", 2},
		{"and predicate", "//div[@id and @class]", 2},
		{"count function", "//ul[count(li) = 3]", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodes, err := root.QueryXPath(tt.xpath)
			if err != nil {
				t.Fatalf("Error: %v", err)
			}
			if len(nodes) != tt.expected {
				t.Errorf("Expected %d nodes, got %d for xpath: %s", tt.expected, len(nodes), tt.xpath)
			}
		})
	}
}

// Test attribute selection
func TestAttributes(t *testing.T) {
	root := createTestTree()

	tests := []struct {
		name    string
		xpath   string
		hasAttr bool
	}{
		{"select attribute", "//div[@id='first']/@class", true},
		{"attribute wildcard", "//div[@id='first']/@*", true},
		{"attribute axis", "//div/attribute::id", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodes, err := root.QueryXPath(tt.xpath)
			if err != nil {
				t.Fatalf("Error: %v", err)
			}
			if tt.hasAttr && len(nodes) == 0 {
				t.Errorf("Expected attribute nodes for xpath: %s", tt.xpath)
			}
		})
	}
}

// Test string functions
func TestStringFunctions(t *testing.T) {
	root := createTestTree()

	// Test contains
	nodes, err := root.QueryXPath("//div[contains(@class, 'item')]")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if len(nodes) != 1 {
		t.Errorf("Expected 1 node with contains(), got %d", len(nodes))
	}

	// Test starts-with
	nodes, err = root.QueryXPath("//a[starts-with(@href, 'http://')]")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if len(nodes) != 1 {
		t.Errorf("Expected 1 node with starts-with(), got %d", len(nodes))
	}

	// Test normalize-space (need to test with context)
	nodes, err = root.QueryXPath("//div[normalize-space()]")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if len(nodes) < 1 {
		t.Errorf("Expected nodes with normalize-space()")
	}
}

// Test error handling
func TestErrorHandling(t *testing.T) {
	root := createTestTree()

	tests := []struct {
		name  string
		xpath string
	}{
		{"empty selector", ""},
		{"unclosed bracket", "//div["},
		{"invalid function", "//div[invalid()]"},
		{"unknown axis", "//unknown::div"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := root.QueryXPath(tt.xpath)
			if err == nil {
				t.Errorf("Expected error for xpath: %s", tt.xpath)
			}
			// Check it's an XPathError
			if _, ok := err.(XPathError); !ok && err != nil {
				t.Logf("Got error type: %T", err)
			}
		})
	}
}

// Test QueryXPathFirst
func TestQueryXPathFirst(t *testing.T) {
	root := createTestTree()

	// Should return first match
	node, err := root.QueryXPathFirst("//div")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if node == nil {
		t.Error("Expected a node")
	}
	if node.Attr("id") != "first" {
		t.Errorf("Expected first div, got id=%s", node.Attr("id"))
	}

	// Should return nil for no match
	node, err = root.QueryXPathFirst("//nonexistent")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if node != nil {
		t.Error("Expected nil for no match")
	}
}

// Test complex expressions
func TestComplexExpressions(t *testing.T) {
	root := createTestTree()

	tests := []struct {
		name     string
		xpath    string
		expected int
	}{
		// Multiple predicates: first [position() > 1] filters to items 2,3
		// Then [position() < 3] on that new set gets items at positions 1,2 of filtered set
		// So we get original items 2 and 3
		{"multiple predicates", "//li[position() > 1][position() < 3]", 2},
		{"nested path", "//div[@id='first']/span/parent::div", 1},
		{"arithmetic", "//ul[count(li) + 0 = 3]", 1},
		{"boolean logic", "//div[@id='first' and @class='container']", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodes, err := root.QueryXPath(tt.xpath)
			if err != nil {
				t.Fatalf("Error: %v", err)
			}
			if len(nodes) != tt.expected {
				t.Errorf("Expected %d nodes, got %d for xpath: %s", tt.expected, len(nodes), tt.xpath)
			}
		})
	}
}

// Test math functions
func TestMathFunctions(t *testing.T) {
	root := createTestTree()

	// Test with position calculations
	nodes, err := root.QueryXPath("//li[position() mod 2 = 1]")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if len(nodes) != 2 { // positions 1 and 3
		t.Errorf("Expected 2 odd-positioned nodes, got %d", len(nodes))
	}
}

// Benchmark basic operations
func BenchmarkQueryXPath(b *testing.B) {
	root := createTestTree()

	benchmarks := []struct {
		name  string
		xpath string
	}{
		{"simple child", "/root/div"},
		{"descendant", "//span"},
		{"attribute", "//div[@id='first']"},
		{"predicate", "//li[position() > 1]"},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				root.QueryXPath(bm.xpath)
			}
		})
	}
}

func TestUnionOperator(t *testing.T) {
	// Create a test tree
	root := &Node{
		Name: "root",
		Children: []*Node{
			{Name: "div", Attrs: map[string]*string{"id": strPtr("first")}},
			{Name: "span", Attrs: map[string]*string{"id": strPtr("second")}},
			{Name: "p", Attrs: map[string]*string{"id": strPtr("third")}},
			{
				Name: "section",
				Children: []*Node{
					{Name: "div", Attrs: map[string]*string{"id": strPtr("nested")}},
					{Name: "span", Attrs: map[string]*string{"id": strPtr("nested2")}},
				},
			},
		},
	}
	setParents(root)

	tests := []struct {
		name     string
		xpath    string
		expected int
		checkIDs []string
	}{
		{
			name:     "union two element names",
			xpath:    "//div | //span",
			expected: 4, // 2 divs + 2 spans
			checkIDs: []string{"first", "second", "nested", "nested2"},
		},
		{
			name:     "union with predicates",
			xpath:    "//div[@id='first'] | //span[@id='second']",
			expected: 2,
			checkIDs: []string{"first", "second"},
		},
		{
			name:     "union three paths",
			xpath:    "//div | //span | //p",
			expected: 5, // 2 divs + 2 spans + 1 p
		},
		{
			name:     "union with different paths",
			xpath:    "/root/div | //section/span",
			expected: 2, // direct child div + nested span
		},
		{
			name:     "union with overlapping results",
			xpath:    "//div | //*[@id='first']",
			expected: 2, // 2 divs (one is duplicate)
		},
		{
			name:     "union with descendant axes",
			xpath:    "/root/div | /root//span",
			expected: 3, // 1 direct div + 2 spans (direct and nested)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodes, err := root.QueryXPath(tt.xpath)
			if err != nil {
				t.Fatalf("QueryXPath error: %v", err)
			}
			if len(nodes) != tt.expected {
				t.Errorf("Expected %d nodes, got %d for xpath: %s", tt.expected, len(nodes), tt.xpath)
				for i, n := range nodes {
					t.Logf("  Node %d: %s (id=%s)", i, n.Name, n.Attr("id"))
				}
			}

			// Check specific IDs if provided
			if tt.checkIDs != nil {
				found := make(map[string]bool)
				for _, n := range nodes {
					if id := n.Attr("id"); id != "" {
						found[id] = true
					}
				}
				for _, expectedID := range tt.checkIDs {
					if !found[expectedID] {
						t.Errorf("Expected to find node with id='%s'", expectedID)
					}
				}
			}
		})
	}
}

func TestUnionWithComplexPredicates(t *testing.T) {
	root := &Node{
		Name: "root",
		Children: []*Node{
			{
				Name:  "ul",
				Attrs: map[string]*string{"id": strPtr("list1")},
				Children: []*Node{
					{Name: "li", Attrs: map[string]*string{"class": strPtr("item")}},
					{Name: "li", Attrs: map[string]*string{"class": strPtr("item active")}},
					{Name: "li", Attrs: map[string]*string{"class": strPtr("item")}},
				},
			},
			{
				Name:  "ul",
				Attrs: map[string]*string{"id": strPtr("list2")},
				Children: []*Node{
					{Name: "li", Attrs: map[string]*string{"class": strPtr("special")}},
					{Name: "li", Attrs: map[string]*string{"class": strPtr("item")}},
				},
			},
		},
	}
	setParents(root)

	tests := []struct {
		name     string
		xpath    string
		expected int
	}{
		{
			name:     "union with position predicates",
			xpath:    "//ul[@id='list1']/li[1] | //ul[@id='list2']/li[1]",
			expected: 2,
		},
		{
			name:     "union with contains",
			xpath:    "//li[contains(@class, 'active')] | //li[contains(@class, 'special')]",
			expected: 2,
		},
		{
			name:     "union with multiple predicates",
			xpath:    "//li[@class='item'][1] | //li[@class='special']",
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodes, err := root.QueryXPath(tt.xpath)
			if err != nil {
				t.Fatalf("QueryXPath error: %v", err)
			}
			if len(nodes) != tt.expected {
				t.Errorf("Expected %d nodes, got %d for xpath: %s", tt.expected, len(nodes), tt.xpath)
			}
		})
	}
}

func TestUnionEdgeCases(t *testing.T) {
	root := &Node{
		Name: "root",
		Children: []*Node{
			{Name: "div"},
			{Name: "span"},
		},
	}
	setParents(root)

	tests := []struct {
		name      string
		xpath     string
		shouldErr bool
		expected  int
	}{
		{
			name:     "single path (no union)",
			xpath:    "//div",
			expected: 1,
		},
		{
			name:     "empty result on left",
			xpath:    "//nonexistent | //span",
			expected: 1,
		},
		{
			name:     "empty result on right",
			xpath:    "//div | //nonexistent",
			expected: 1,
		},
		{
			name:     "both empty",
			xpath:    "//foo | //bar",
			expected: 0,
		},
		{
			name:     "union with spaces",
			xpath:    "//div  |  //span",
			expected: 2,
		},
		{
			name:     "union in predicate (uses | in expression, not path union)",
			xpath:    "//div[@id='a' or @id='b']",
			expected: 0, // No divs with those IDs
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodes, err := root.QueryXPath(tt.xpath)
			if tt.shouldErr {
				if err == nil {
					t.Fatalf("Expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("QueryXPath error: %v", err)
			}
			if len(nodes) != tt.expected {
				t.Errorf("Expected %d nodes, got %d for xpath: %s", tt.expected, len(nodes), tt.xpath)
			}
		})
	}
}

func TestUnionDoesNotAffectPredicates(t *testing.T) {
	root := &Node{
		Name: "root",
		Children: []*Node{
			{
				Name:  "div",
				Attrs: map[string]*string{"data-items": strPtr("a|b|c")},
			},
			{
				Name:  "span",
				Attrs: map[string]*string{"data-value": strPtr("x|y")},
			},
		},
	}
	setParents(root)

	// The | inside the attribute value should not be treated as union
	nodes, err := root.QueryXPath("//div[@data-items='a|b|c']")
	if err != nil {
		t.Fatalf("QueryXPath error: %v", err)
	}
	if len(nodes) != 1 {
		t.Errorf("Expected 1 node, got %d - | in string should not be treated as union", len(nodes))
	}
}

func TestXPathLocationPathInFunction(t *testing.T) {
	// Create test tree
	root := &Node{
		Name: "root",
		Children: []*Node{
			{
				Name:  "article",
				Attrs: map[string]*string{"id": strPtr("a1")},
				Children: []*Node{
					{Name: "p", Children: []*Node{{Name: "#text", Data: "Para 1"}}},
					{Name: "p", Children: []*Node{{Name: "#text", Data: "Para 2"}}},
				},
			},
			{
				Name:  "article",
				Attrs: map[string]*string{"id": strPtr("a2")},
				Children: []*Node{
					{Name: "p", Children: []*Node{{Name: "#text", Data: "Para 3"}}},
				},
			},
			{
				Name:  "article",
				Attrs: map[string]*string{"id": strPtr("a3")},
				Children: []*Node{
					{Name: "div", Children: []*Node{
						{Name: "p", Children: []*Node{{Name: "#text", Data: "Nested"}}},
					}},
				},
			},
		},
	}
	setParents(root)

	// Test count(.//p) - count all descendant p elements
	nodes, err := root.QueryXPath("//article[count(.//p) > 1]")
	if err != nil {
		t.Fatalf("QueryXPath error: %v", err)
	}
	if len(nodes) != 1 {
		t.Errorf("Expected 1 article with > 1 descendant p, got %d", len(nodes))
	}
	if len(nodes) > 0 && nodes[0].Attr("id") != "a1" {
		t.Errorf("Expected article a1, got %s", nodes[0].Attr("id"))
	}

	// Debug: Test if p works directly (without count)
	// Article a1 should have 2 direct p children
	a1, err := root.QueryXPathFirst("//article[@id='a1']")
	if err != nil || a1 == nil {
		t.Fatalf("Could not find article a1")
	}
	directChildren, err := a1.QueryXPath("p")
	if err != nil {
		t.Fatalf("p query error: %v", err)
	}
	t.Logf("Direct p children of a1: %d", len(directChildren))

	// Test count(p) - count direct children (p is equivalent to child::p)
	nodes, err = root.QueryXPath("//article[count(p) = 2]")
	if err != nil {
		t.Fatalf("QueryXPath error: %v", err)
	}
	t.Logf("Articles with count(p)=2: %d", len(nodes))
	if len(nodes) != 1 {
		t.Errorf("Expected 1 article with 2 child p elements, got %d", len(nodes))
	}
}

func TestXPathLocationChildPathInFunction(t *testing.T) {
	// Create test tree
	root := &Node{
		Name: "root",
		Children: []*Node{
			{
				Name:  "article",
				Attrs: map[string]*string{"id": strPtr("a1")},
				Children: []*Node{
					{Name: "p", Children: []*Node{{Name: "#text", Data: "Para 1"}}},
					{Name: "p", Children: []*Node{{Name: "#text", Data: "Para 2"}}},
				},
			},
			{
				Name:  "article",
				Attrs: map[string]*string{"id": strPtr("a2")},
				Children: []*Node{
					{Name: "p", Children: []*Node{{Name: "#text", Data: "Para 3"}}},
				},
			},
			{
				Name:  "article",
				Attrs: map[string]*string{"id": strPtr("a3")},
				Children: []*Node{
					{Name: "div", Children: []*Node{
						{Name: "p", Children: []*Node{{Name: "#text", Data: "Nested"}}},
					}},
				},
			},
		},
	}
	setParents(root)

	// Test count(.//p) - count all descendant p elements
	nodes, err := root.QueryXPath("//article[count(.//p) > 1]")
	if err != nil {
		t.Fatalf("QueryXPath error: %v", err)
	}
	if len(nodes) != 1 {
		t.Errorf("Expected 1 article with > 1 descendant p, got %d", len(nodes))
	}
	if len(nodes) > 0 && nodes[0].Attr("id") != "a1" {
		t.Errorf("Expected article a1, got %s", nodes[0].Attr("id"))
	}

	// Debug: Test if child::p works in main parser
	directChildAxis, err := root.QueryXPath("//article[@id='a1']/child::p")
	if err != nil {
		t.Fatalf("child::p in main parser error: %v", err)
	}
	t.Logf("Main parser child::p results: %d", len(directChildAxis))

	// Debug: Test if p works directly (without count)
	// Article a1 should have 2 direct p children
	a1, err := root.QueryXPathFirst("//article[@id='a1']")
	if err != nil || a1 == nil {
		t.Fatalf("Could not find article a1")
	}
	directChildren, err := a1.QueryXPath("p")
	if err != nil {
		t.Fatalf("p query error: %v", err)
	}
	t.Logf("Direct p children of a1: %d", len(directChildren))

	// Debug: Test if child::p works when called directly on a node
	childAxisDirect, err := a1.QueryXPath("child::p")
	if err != nil {
		t.Fatalf("a1.QueryXPath('child::p') error: %v", err)
	}
	t.Logf("a1.QueryXPath('child::p') results: %d", len(childAxisDirect))

	// Debug: Manually test what count(child::p) returns for a1
	// Create a predicate parser and evaluate it with a1 as context
	testPred := "[count(child::p)]"
	content := testPred[1 : len(testPred)-1]
	pred, err := parseXPathPredicate(content)
	if err != nil {
		t.Fatalf("Failed to parse test predicate: %v", err)
	}

	// Evaluate with a1 as one of the nodes
	testNodes := []*Node{a1}
	resultNodes, err := pred.filter(testNodes, root)
	if err != nil {
		t.Fatalf("Failed to filter: %v", err)
	}
	t.Logf("Nodes after filter with [count(child::p)]: %d (should be 1 if count returns 2)", len(resultNodes))

	// Test count(p) - count direct children (p is equivalent to child::p)
	nodes, err = root.QueryXPath("//article[count(p) = 2]")
	if err != nil {
		t.Fatalf("QueryXPath error: %v", err)
	}
	t.Logf("Articles with count(p)=2: %d", len(nodes))
	if len(nodes) != 1 {
		t.Errorf("Expected 1 article with 2 child p elements, got %d", len(nodes))
	}

	// Test count(child::p) - explicit axis syntax
	nodes, err = root.QueryXPath("//article[count(child::p) = 2]")
	if err != nil {
		t.Fatalf("count(child::p) error: %v", err)
	}
	t.Logf("Articles with count(child::p)=2: %d", len(nodes))
	if len(nodes) != 1 {
		t.Errorf("Expected 1 article with count(child::p)=2, got %d", len(nodes))
	}
}
