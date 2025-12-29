package dom

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type xpathAxis int

const (
	xpathAxisChild xpathAxis = iota
	xpathAxisDescendant
	xpathAxisParent
	xpathAxisAncestor
	xpathAxisFollowingSibling
	xpathAxisPrecedingSibling
	xpathAxisFollowing
	xpathAxisPreceding
	xpathAxisAttribute
	xpathAxisSelf
	xpathAxisDescendantOrSelf
	xpathAxisAncestorOrSelf
)

type xpathNodeTest int

const (
	xpathNodeTestName xpathNodeTest = iota
	xpathNodeTestText
	xpathNodeTestNode
	xpathNodeTestComment
	xpathNodeTestProcessingInstruction
)

type xpathStep struct {
	axis       xpathAxis
	nodeTest   xpathNodeTest
	name       string
	predicates []xpathPredicate
}

type xpathPredicate interface {
	filter(nodes []*Node, root *Node) ([]*Node, error)
}

type xpathExprPredicate struct {
	expr xpathExpr
}

func (p xpathExprPredicate) filter(nodes []*Node, root *Node) ([]*Node, error) {
	if len(nodes) == 0 {
		return nil, nil
	}
	out := make([]*Node, 0, len(nodes))
	size := len(nodes)
	for i, n := range nodes {
		ctx := xpathContext{
			root:     root,
			node:     n,
			position: i + 1,
			size:     size,
		}
		val, err := p.expr.eval(ctx)
		if err != nil {
			return nil, err
		}
		switch val.kind {
		case xpathValueNumber:
			if val.num == float64(ctx.position) {
				out = append(out, n)
			}
		default:
			if val.toBool() {
				out = append(out, n)
			}
		}
	}
	return out, nil
}

// XPathError represents an error that occurred during XPath parsing or evaluation
type XPathError struct {
	Message string
}

func (e XPathError) Error() string {
	return fmt.Sprintf("xpath error: %s", e.Message)
}

// QueryXPath returns all nodes that satisfy the provided XPath selector.
// It implements the XPath 1.0 core function set with all axes.
// Returns an error if the selector is invalid.
func (root *Node) QueryXPath(selector string) ([]*Node, error) {
	if root == nil {
		return nil, nil
	}

	// Check for union operator at the top level (not in predicates)
	if hasUnionOperator(selector) {
		return queryXPathUnion(root, selector)
	}

	// Check for attribute selection at path level
	if hasAttributeSelection(selector) {
		return queryXPathWithAttributes(root, selector)
	}

	steps, err := parseXPath(selector)
	if err != nil {
		return nil, err
	}

	nodes := []*Node{root}

	// Handle absolute paths: /root/div means "root element named 'root', then child 'div'"
	// The first step should match against the root itself, not its children
	if len(steps) > 0 && strings.HasPrefix(selector, "/") && !strings.HasPrefix(selector, "//") {
		firstStep := steps[0]
		if matchesNodeTest(root, firstStep.nodeTest, firstStep.name) {
			// Root matches, apply its predicates
			nodes = []*Node{root}
			for _, pred := range firstStep.predicates {
				nodes, err = pred.filter(nodes, root)
				if err != nil {
					return nil, err
				}
			}
			// Skip first step, already processed
			steps = steps[1:]
		} else {
			// Root doesn't match first step
			return nil, nil
		}
	}

	// Process remaining steps
	for _, step := range steps {
		candidates := applyXPathAxis(nodes, step.axis, step.nodeTest, step.name)
		candidates = uniqueNodesInDocOrder(candidates)
		for _, predicate := range step.predicates {
			candidates, err = predicate.filter(candidates, root)
			if err != nil {
				return nil, err
			}
		}
		nodes = candidates
	}
	return nodes, nil
}

func hasUnionOperator(selector string) bool {
	// Check if | appears outside of predicates and strings
	depth := 0
	inString := false
	var stringChar byte

	for i := 0; i < len(selector); i++ {
		ch := selector[i]

		if inString {
			if ch == stringChar {
				inString = false
			}
			continue
		}

		switch ch {
		case '"', '\'':
			inString = true
			stringChar = ch
		case '[':
			depth++
		case ']':
			depth--
		case '|':
			if depth == 0 {
				return true
			}
		}
	}
	return false
}

func queryXPathUnion(root *Node, selector string) ([]*Node, error) {
	// Split the selector by union operator (outside of predicates)
	paths := splitUnionPaths(selector)
	if len(paths) == 0 {
		return nil, XPathError{"invalid union expression"}
	}

	// Execute each path and collect results
	var allNodes []*Node
	for _, path := range paths {
		nodes, err := root.QueryXPath(strings.TrimSpace(path))
		if err != nil {
			return nil, err
		}
		allNodes = append(allNodes, nodes...)
	}

	// Remove duplicates and return
	return uniqueNodesInDocOrder(allNodes), nil
}

func splitUnionPaths(selector string) []string {
	var paths []string
	depth := 0
	inString := false
	var stringChar byte
	start := 0

	for i := 0; i < len(selector); i++ {
		ch := selector[i]

		if inString {
			if ch == stringChar {
				inString = false
			}
			continue
		}

		switch ch {
		case '"', '\'':
			inString = true
			stringChar = ch
		case '[':
			depth++
		case ']':
			depth--
		case '|':
			if depth == 0 {
				paths = append(paths, selector[start:i])
				start = i + 1
			}
		}
	}

	// Add the last path
	if start < len(selector) {
		paths = append(paths, selector[start:])
	}

	return paths
}

func hasAttributeSelection(selector string) bool {
	depth := 0
	inString := false
	var stringChar byte

	for i := 0; i < len(selector); i++ {
		ch := selector[i]

		if inString {
			if ch == stringChar {
				inString = false
			}
			continue
		}

		switch ch {
		case '"', '\'':
			inString = true
			stringChar = ch
		case '[':
			depth++
		case ']':
			depth--
		case '@':
			if depth == 0 {
				return true
			}
		}
	}
	return false
}

func queryXPathWithAttributes(root *Node, selector string) ([]*Node, error) {
	parts := splitAtLastAttribute(selector)
	if len(parts) != 2 {
		return nil, XPathError{"invalid attribute selection"}
	}

	nodePath := parts[0]
	attrName := parts[1]

	// Get nodes
	nodes, err := root.QueryXPath(nodePath)
	if err != nil {
		return nil, err
	}

	// Return nodes that have the requested attribute
	result := make([]*Node, 0)
	for _, n := range nodes {
		if n.Attrs != nil {
			if attrName == "*" {
				if len(n.Attrs) > 0 {
					result = append(result, n)
				}
			} else {
				if _, ok := n.Attrs[attrName]; ok {
					result = append(result, n)
				}
			}
		}
	}

	return result, nil
}

func splitAtLastAttribute(selector string) []string {
	depth := 0
	inString := false
	var stringChar byte
	lastAt := -1

	for i := 0; i < len(selector); i++ {
		ch := selector[i]

		if inString {
			if ch == stringChar {
				inString = false
			}
			continue
		}

		switch ch {
		case '"', '\'':
			inString = true
			stringChar = ch
		case '[':
			depth++
		case ']':
			depth--
		case '@':
			if depth == 0 {
				lastAt = i
			}
		}
	}

	if lastAt == -1 {
		return []string{selector}
	}

	return []string{selector[:lastAt], selector[lastAt+1:]}
}

// QueryXPathFirst returns the first node that satisfies the XPath selector, or
// nil if no match is found.
func (root *Node) QueryXPathFirst(selector string) (*Node, error) {
	matches, err := root.QueryXPath(selector)
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return nil, nil
	}
	return matches[0], nil
}

func parseXPath(selector string) ([]xpathStep, error) {
	if strings.TrimSpace(selector) == "" {
		return nil, XPathError{"empty xpath selector"}
	}

	var steps []xpathStep
	i := 0
	axis := xpathAxisChild
	switch {
	case strings.HasPrefix(selector, "//"):
		axis = xpathAxisDescendantOrSelf
		i = 2
	case strings.HasPrefix(selector, "/"):
		axis = xpathAxisChild
		i = 1
	}

	for i < len(selector) {
		step, next, err := parseXPathStep(selector, i, axis)
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
		i = next
		if i >= len(selector) {
			break
		}
		if strings.HasPrefix(selector[i:], "//") {
			axis = xpathAxisDescendantOrSelf
			i += 2
		} else if selector[i] == '/' {
			axis = xpathAxisChild
			i++
		} else {
			return nil, XPathError{"invalid xpath axis separator"}
		}
	}

	return steps, nil
}

func parseXPathStep(selector string, start int, defaultAxis xpathAxis) (xpathStep, int, error) {
	axis := defaultAxis
	nodeTest := xpathNodeTestName
	i := start

	// Check for explicit axis
	if axisName, axisEnd, hasAxis := tryParseAxis(selector, start); hasAxis {
		var ok bool
		axis, ok = parseAxisName(axisName)
		if !ok {
			return xpathStep{}, 0, XPathError{fmt.Sprintf("unknown axis: %s", axisName)}
		}
		i = axisEnd
	}

	name, nodeTest, i, err := parseNodeTest(selector, i)
	if err != nil {
		return xpathStep{}, 0, err
	}

	step := xpathStep{
		axis:     axis,
		nodeTest: nodeTest,
		name:     name,
	}

	// Parse predicates
	for i < len(selector) && selector[i] == '[' {
		content, next, err := readXPathBracket(selector, i)
		if err != nil {
			return xpathStep{}, 0, err
		}
		predicate, err := parseXPathPredicate(content)
		if err != nil {
			return xpathStep{}, 0, err
		}
		step.predicates = append(step.predicates, predicate)
		i = next
	}

	return step, i, nil
}

func tryParseAxis(selector string, start int) (string, int, bool) {
	colonPos := strings.Index(selector[start:], "::")
	if colonPos == -1 {
		return "", 0, false
	}

	// Check if there's a '[' or '/' before the '::'
	beforeColon := selector[start : start+colonPos]
	if strings.ContainsAny(beforeColon, "[/") {
		return "", 0, false
	}

	axisName := strings.TrimSpace(beforeColon)
	if axisName == "" {
		return "", 0, false
	}

	return axisName, start + colonPos + 2, true
}

func parseAxisName(name string) (xpathAxis, bool) {
	switch strings.ToLower(name) {
	case "child":
		return xpathAxisChild, true
	case "descendant":
		return xpathAxisDescendant, true
	case "parent":
		return xpathAxisParent, true
	case "ancestor":
		return xpathAxisAncestor, true
	case "following-sibling":
		return xpathAxisFollowingSibling, true
	case "preceding-sibling":
		return xpathAxisPrecedingSibling, true
	case "following":
		return xpathAxisFollowing, true
	case "preceding":
		return xpathAxisPreceding, true
	case "attribute":
		return xpathAxisAttribute, true
	case "self":
		return xpathAxisSelf, true
	case "descendant-or-self":
		return xpathAxisDescendantOrSelf, true
	case "ancestor-or-self":
		return xpathAxisAncestorOrSelf, true
	default:
		return 0, false
	}
}

func parseNodeTest(selector string, start int) (string, xpathNodeTest, int, error) {
	if start >= len(selector) {
		return "", 0, 0, XPathError{"missing xpath node test"}
	}

	// Handle . and ..
	if selector[start] == '.' {
		if start+1 < len(selector) && selector[start+1] == '.' {
			return "..", xpathNodeTestName, start + 2, nil
		}
		return ".", xpathNodeTestName, start + 1, nil
	}

	// Handle wildcard
	if selector[start] == '*' {
		return "*", xpathNodeTestName, start + 1, nil
	}

	// Find end of node test
	end := start
	for end < len(selector) {
		ch := selector[end]
		if ch == '/' || ch == '[' {
			break
		}
		end++
	}

	name := strings.TrimSpace(selector[start:end])
	if name == "" {
		return "", 0, 0, XPathError{"missing xpath node test"}
	}

	// Check for node type tests
	if strings.HasSuffix(name, "()") {
		testName := strings.TrimSuffix(name, "()")
		switch strings.ToLower(testName) {
		case "text":
			return "text()", xpathNodeTestText, end, nil
		case "node":
			return "node()", xpathNodeTestNode, end, nil
		case "comment":
			return "comment()", xpathNodeTestComment, end, nil
		case "processing-instruction":
			return "processing-instruction()", xpathNodeTestProcessingInstruction, end, nil
		}
	}

	return name, xpathNodeTestName, end, nil
}

func readXPathBracket(selector string, start int) (string, int, error) {
	depth := 0
	quote := byte(0)

	for i := start; i < len(selector); i++ {
		ch := selector[i]
		switch ch {
		case '"', '\'':
			switch quote {
			case 0:
				quote = ch
			case ch:
				quote = 0
			}
		case '[':
			if quote == 0 {
				depth++
			}
		case ']':
			if quote == 0 {
				depth--
				if depth == 0 {
					return selector[start+1 : i], i + 1, nil
				}
			}
		}
	}

	return "", 0, XPathError{"unclosed xpath predicate"}
}

func parseXPathPredicate(content string) (xpathPredicate, error) {
	parser := newXPathExprParser(content)
	expr, err := parser.parseExpr()
	if err != nil {
		return nil, err
	}
	if parser.peek().typ != xpathTokenEOF {
		return nil, XPathError{"invalid xpath predicate"}
	}
	return xpathExprPredicate{expr: expr}, nil
}

func applyXPathAxis(nodes []*Node, axis xpathAxis, nodeTest xpathNodeTest, name string) []*Node {
	// Handle special cases
	if nodeTest == xpathNodeTestName {
		if name == "." {
			return nodes
		}
		if name == ".." {
			out := make([]*Node, 0, len(nodes))
			for _, n := range nodes {
				if n != nil && n.Parent != nil {
					out = append(out, n.Parent)
				}
			}
			return out
		}
	}

	out := make([]*Node, 0, len(nodes))
	for _, n := range nodes {
		if n == nil {
			continue
		}
		switch axis {
		case xpathAxisChild:
			out = append(out, applyNodeTestToNodes(xpathChildren(n), nodeTest, name)...)
		case xpathAxisDescendant:
			collectDescendants(n, nodeTest, name, &out, false)
		case xpathAxisDescendantOrSelf:
			if matchesNodeTest(n, nodeTest, name) {
				out = append(out, n)
			}
			collectDescendants(n, nodeTest, name, &out, false)
		case xpathAxisParent:
			if n.Parent != nil && matchesNodeTest(n.Parent, nodeTest, name) {
				out = append(out, n.Parent)
			}
		case xpathAxisAncestor:
			collectAncestors(n, nodeTest, name, &out, false)
		case xpathAxisAncestorOrSelf:
			if matchesNodeTest(n, nodeTest, name) {
				out = append(out, n)
			}
			collectAncestors(n, nodeTest, name, &out, false)
		case xpathAxisFollowingSibling:
			out = append(out, applyNodeTestToNodes(getFollowingSiblings(n), nodeTest, name)...)
		case xpathAxisPrecedingSibling:
			out = append(out, applyNodeTestToNodes(getPrecedingSiblings(n), nodeTest, name)...)
		case xpathAxisFollowing:
			collectFollowing(n, nodeTest, name, &out)
		case xpathAxisPreceding:
			collectPreceding(n, nodeTest, name, &out)
		case xpathAxisAttribute:
			// Attribute axis - return nodes that have the requested attribute
			// (Can't return actual attribute nodes with []*Node return type)
			if name == "*" {
				// All attributes
				if len(n.Attrs) > 0 {
					out = append(out, n)
				}
			} else {
				// Specific attribute by name (case-insensitive)
				if n.Attrs != nil {
					for attrName := range n.Attrs {
						if strings.EqualFold(attrName, name) {
							out = append(out, n)
							break
						}
					}
				}
			}
		case xpathAxisSelf:
			if matchesNodeTest(n, nodeTest, name) {
				out = append(out, n)
			}
		}
	}
	return out
}

func applyNodeTestToNodes(nodes []*Node, nodeTest xpathNodeTest, name string) []*Node {
	out := make([]*Node, 0, len(nodes))
	for _, n := range nodes {
		if matchesNodeTest(n, nodeTest, name) {
			out = append(out, n)
		}
	}
	return out
}

func matchesNodeTest(n *Node, nodeTest xpathNodeTest, name string) bool {
	if n == nil {
		return false
	}

	switch nodeTest {
	case xpathNodeTestText:
		return n.Name == "#text"
	case xpathNodeTestComment:
		return n.Name == "#comment"
	case xpathNodeTestProcessingInstruction:
		return strings.HasPrefix(n.Name, "?")
	case xpathNodeTestNode:
		return true
	case xpathNodeTestName:
		if n.Name == "" || strings.HasPrefix(n.Name, "#") {
			return false
		}
		if name == "*" {
			return true
		}
		return strings.EqualFold(n.Name, name)
	}
	return false
}

func xpathChildren(n *Node) []*Node {
	if n.Template != nil {
		return append(n.Children, n.Template)
	}
	return n.Children
}

func collectDescendants(n *Node, nodeTest xpathNodeTest, name string, out *[]*Node, includeSelf bool) {
	if n == nil {
		return
	}
	if includeSelf && matchesNodeTest(n, nodeTest, name) {
		*out = append(*out, n)
	}
	for _, child := range xpathChildren(n) {
		if matchesNodeTest(child, nodeTest, name) {
			*out = append(*out, child)
		}
		collectDescendants(child, nodeTest, name, out, false)
	}
}

func collectAncestors(n *Node, nodeTest xpathNodeTest, name string, out *[]*Node, includeSelf bool) {
	if n == nil {
		return
	}
	current := n.Parent
	if includeSelf {
		current = n
	}
	for current != nil {
		if matchesNodeTest(current, nodeTest, name) {
			*out = append(*out, current)
		}
		current = current.Parent
	}
}

func getFollowingSiblings(n *Node) []*Node {
	if n == nil || n.Parent == nil {
		return nil
	}
	siblings := xpathChildren(n.Parent)
	index := -1
	for i, sib := range siblings {
		if sib == n {
			index = i
			break
		}
	}
	if index == -1 || index+1 >= len(siblings) {
		return nil
	}
	return siblings[index+1:]
}

func getPrecedingSiblings(n *Node) []*Node {
	if n == nil || n.Parent == nil {
		return nil
	}
	siblings := xpathChildren(n.Parent)
	index := -1
	for i, sib := range siblings {
		if sib == n {
			index = i
			break
		}
	}
	if index <= 0 {
		return nil
	}
	// Return in reverse document order (as per XPath spec)
	result := make([]*Node, index)
	for i := 0; i < index; i++ {
		result[index-1-i] = siblings[i]
	}
	return result
}

func collectFollowing(n *Node, nodeTest xpathNodeTest, name string, out *[]*Node) {
	if n == nil {
		return
	}

	// First, get following siblings and their descendants
	for _, sib := range getFollowingSiblings(n) {
		if matchesNodeTest(sib, nodeTest, name) {
			*out = append(*out, sib)
		}
		collectDescendants(sib, nodeTest, name, out, false)
	}

	// Then recursively process parent's following
	if n.Parent != nil {
		collectFollowing(n.Parent, nodeTest, name, out)
	}
}

func collectPreceding(n *Node, nodeTest xpathNodeTest, name string, out *[]*Node) {
	if n == nil {
		return
	}

	// First, recursively process parent's preceding
	if n.Parent != nil {
		collectPreceding(n.Parent, nodeTest, name, out)
	}

	// Then get preceding siblings and their descendants
	for _, sib := range getPrecedingSiblings(n) {
		collectDescendants(sib, nodeTest, name, out, true)
	}
}

func matchesXPathName(n *Node, name string) bool {
	if n == nil {
		return false
	}
	if name == "text()" {
		return n.Name == "#text"
	}
	if n.Name == "" || strings.HasPrefix(n.Name, "#") {
		return false
	}
	if name == "*" {
		return true
	}
	return strings.EqualFold(n.Name, name)
}

func splitXPathAxis(name string) (string, string, bool) {
	parts := strings.SplitN(name, "::", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	axis := strings.TrimSpace(parts[0])
	test := strings.TrimSpace(parts[1])
	if axis == "" || test == "" {
		return "", "", false
	}
	return axis, test, true
}

func precedingSiblingNodes(node *Node, test string) []xpathNode {
	if node == nil || node.Parent == nil {
		return nil
	}
	siblings := xpathChildren(node.Parent)
	index := -1
	for i, sib := range siblings {
		if sib == node {
			index = i
			break
		}
	}
	if index <= 0 {
		return nil
	}
	out := make([]xpathNode, 0, index)
	for i := index - 1; i >= 0; i-- {
		if matchesXPathName(siblings[i], test) {
			out = append(out, newNodeValue(siblings[i]))
		}
	}
	return out
}

func followingSiblingNodes(node *Node, test string) []xpathNode {
	if node == nil || node.Parent == nil {
		return nil
	}
	siblings := xpathChildren(node.Parent)
	index := -1
	for i, sib := range siblings {
		if sib == node {
			index = i
			break
		}
	}
	if index == -1 || index+1 >= len(siblings) {
		return nil
	}
	out := make([]xpathNode, 0, len(siblings)-index-1)
	for i := index + 1; i < len(siblings); i++ {
		if matchesXPathName(siblings[i], test) {
			out = append(out, newNodeValue(siblings[i]))
		}
	}
	return out
}

// uniqueNodesInDocOrder removes duplicates and ensures document order
func uniqueNodesInDocOrder(nodes []*Node) []*Node {
	if len(nodes) == 0 {
		return nodes
	}

	seen := make(map[*Node]bool, len(nodes))
	unique := make([]*Node, 0, len(nodes))
	for _, n := range nodes {
		if n != nil && !seen[n] {
			seen[n] = true
			unique = append(unique, n)
		}
	}

	return unique
}

type xpathContext struct {
	root     *Node
	node     *Node
	position int
	size     int
}

type xpathValueKind int

const (
	xpathValueNodeSet xpathValueKind = iota
	xpathValueString
	xpathValueNumber
	xpathValueBoolean
)

type xpathValue struct {
	kind  xpathValueKind
	nodes []xpathNode
	str   string
	num   float64
	b     bool
}

type xpathNodeKind int

const (
	xpathNodeElement xpathNodeKind = iota
	xpathNodeText
	xpathNodeAttribute
)

type xpathNode struct {
	node      *Node
	kind      xpathNodeKind
	attrName  string
	attrValue string
}

func newNodeValue(node *Node) xpathNode {
	kind := xpathNodeElement
	if node != nil && node.Name == "#text" {
		kind = xpathNodeText
	}
	return xpathNode{node: node, kind: kind}
}

func newAttrValue(name, value string) xpathNode {
	return xpathNode{kind: xpathNodeAttribute, attrName: name, attrValue: value}
}

func nodeStringValue(n xpathNode) string {
	switch n.kind {
	case xpathNodeAttribute:
		return n.attrValue
	case xpathNodeText:
		if n.node != nil {
			return n.node.Data
		}
		return ""
	default:
		if n.node != nil {
			return n.node.ToText("", false)
		}
		return ""
	}
}

func (v xpathValue) toBool() bool {
	switch v.kind {
	case xpathValueBoolean:
		return v.b
	case xpathValueNumber:
		return v.num != 0 && !math.IsNaN(v.num)
	case xpathValueString:
		return v.str != ""
	case xpathValueNodeSet:
		return len(v.nodes) > 0
	default:
		return false
	}
}

func (v xpathValue) toNumber() float64 {
	switch v.kind {
	case xpathValueNumber:
		return v.num
	case xpathValueBoolean:
		if v.b {
			return 1
		}
		return 0
	case xpathValueString:
		if v.str == "" {
			return 0
		}
		n, err := strconv.ParseFloat(strings.TrimSpace(v.str), 64)
		if err != nil {
			return math.NaN()
		}
		return n
	case xpathValueNodeSet:
		return xpathValue{kind: xpathValueString, str: v.toString()}.toNumber()
	default:
		return math.NaN()
	}
}

func (v xpathValue) toString() string {
	switch v.kind {
	case xpathValueString:
		return v.str
	case xpathValueBoolean:
		if v.b {
			return "true"
		}
		return "false"
	case xpathValueNumber:
		if math.IsNaN(v.num) {
			return "NaN"
		}
		if math.IsInf(v.num, 1) {
			return "Infinity"
		}
		if math.IsInf(v.num, -1) {
			return "-Infinity"
		}
		return strconv.FormatFloat(v.num, 'f', -1, 64)
	case xpathValueNodeSet:
		if len(v.nodes) == 0 {
			return ""
		}
		return nodeStringValue(v.nodes[0])
	default:
		return ""
	}
}

type xpathExpr interface {
	eval(ctx xpathContext) (xpathValue, error)
}

type xpathLiteralExpr struct {
	value xpathValue
}

func (e xpathLiteralExpr) eval(ctx xpathContext) (xpathValue, error) {
	return e.value, nil
}

type xpathUnaryExpr struct {
	op   string
	expr xpathExpr
}

func (e xpathUnaryExpr) eval(ctx xpathContext) (xpathValue, error) {
	val, err := e.expr.eval(ctx)
	if err != nil {
		return xpathValue{}, err
	}
	num := val.toNumber()
	if e.op == "-" {
		num = -num
	}
	return xpathValue{kind: xpathValueNumber, num: num}, nil
}

type xpathBinaryExpr struct {
	op    string
	left  xpathExpr
	right xpathExpr
}

func (e xpathBinaryExpr) eval(ctx xpathContext) (xpathValue, error) {
	left, err := e.left.eval(ctx)
	if err != nil {
		return xpathValue{}, err
	}
	right, err := e.right.eval(ctx)
	if err != nil {
		return xpathValue{}, err
	}

	switch e.op {
	case "or":
		return xpathValue{kind: xpathValueBoolean, b: left.toBool() || right.toBool()}, nil
	case "and":
		return xpathValue{kind: xpathValueBoolean, b: left.toBool() && right.toBool()}, nil
	case "=", "!=", "<", ">", "<=", ">=":
		return compareXPathValues(left, right, e.op), nil
	case "+", "-", "*", "div", "mod":
		ln := left.toNumber()
		rn := right.toNumber()
		var result float64
		switch e.op {
		case "+":
			result = ln + rn
		case "-":
			result = ln - rn
		case "*":
			result = ln * rn
		case "div":
			result = ln / rn
		case "mod":
			result = math.Mod(ln, rn)
		}
		return xpathValue{kind: xpathValueNumber, num: result}, nil
	case "|":
		// Union operator - combine node sets
		if left.kind != xpathValueNodeSet || right.kind != xpathValueNodeSet {
			return xpathValue{}, XPathError{"union operator requires node-sets"}
		}
		combined := append(left.nodes, right.nodes...)
		return xpathValue{kind: xpathValueNodeSet, nodes: combined}, nil
	default:
		return xpathValue{}, XPathError{"unsupported xpath operator: " + e.op}
	}
}

type xpathFunctionExpr struct {
	name string
	args []xpathExpr
}

func (e xpathFunctionExpr) eval(ctx xpathContext) (xpathValue, error) {
	name := strings.ToLower(e.name)
	switch name {
	case "last":
		return xpathValue{kind: xpathValueNumber, num: float64(ctx.size)}, nil
	case "position":
		return xpathValue{kind: xpathValueNumber, num: float64(ctx.position)}, nil
	case "count":
		if len(e.args) != 1 {
			return xpathValue{}, XPathError{"count() expects one argument"}
		}
		val, err := e.args[0].eval(ctx)
		if err != nil {
			return xpathValue{}, err
		}
		if val.kind != xpathValueNodeSet {
			return xpathValue{}, XPathError{"count() expects a node-set"}
		}
		return xpathValue{kind: xpathValueNumber, num: float64(len(val.nodes))}, nil
	case "id":
		if len(e.args) != 1 {
			return xpathValue{}, XPathError{"id() expects one argument"}
		}
		val, err := e.args[0].eval(ctx)
		if err != nil {
			return xpathValue{}, err
		}
		ids := strings.Fields(val.toString())
		return xpathValue{kind: xpathValueNodeSet, nodes: idLookup(ctx.root, ids)}, nil
	case "local-name":
		return xpathValue{kind: xpathValueString, str: localName(nodeForNameFunction(ctx, e.args))}, nil
	case "namespace-uri":
		node := nodeForNameFunction(ctx, e.args)
		if node == nil {
			return xpathValue{kind: xpathValueString, str: ""}, nil
		}
		return xpathValue{kind: xpathValueString, str: node.Namespace}, nil
	case "name":
		node := nodeForNameFunction(ctx, e.args)
		if node == nil {
			return xpathValue{kind: xpathValueString, str: ""}, nil
		}
		return xpathValue{kind: xpathValueString, str: node.Name}, nil
	case "string":
		if len(e.args) == 0 {
			return xpathValue{kind: xpathValueString, str: newNodeValue(ctx.node).node.ToText("", false)}, nil
		}
		val, err := e.args[0].eval(ctx)
		if err != nil {
			return xpathValue{}, err
		}
		return xpathValue{kind: xpathValueString, str: val.toString()}, nil
	case "concat":
		var b strings.Builder
		for _, arg := range e.args {
			val, err := arg.eval(ctx)
			if err != nil {
				return xpathValue{}, err
			}
			b.WriteString(val.toString())
		}
		return xpathValue{kind: xpathValueString, str: b.String()}, nil
	case "starts-with":
		if len(e.args) != 2 {
			return xpathValue{}, XPathError{"starts-with() expects two arguments"}
		}
		first, err := evalStringArg(ctx, e.args[0])
		if err != nil {
			return xpathValue{}, err
		}
		second, err := evalStringArg(ctx, e.args[1])
		if err != nil {
			return xpathValue{}, err
		}
		return xpathValue{kind: xpathValueBoolean, b: strings.HasPrefix(first, second)}, nil
	case "contains":
		if len(e.args) != 2 {
			return xpathValue{}, XPathError{"contains() expects two arguments"}
		}
		first, err := evalStringArg(ctx, e.args[0])
		if err != nil {
			return xpathValue{}, err
		}
		second, err := evalStringArg(ctx, e.args[1])
		if err != nil {
			return xpathValue{}, err
		}
		return xpathValue{kind: xpathValueBoolean, b: strings.Contains(first, second)}, nil
	case "substring-before":
		if len(e.args) != 2 {
			return xpathValue{}, XPathError{"substring-before() expects two arguments"}
		}
		first, err := evalStringArg(ctx, e.args[0])
		if err != nil {
			return xpathValue{}, err
		}
		second, err := evalStringArg(ctx, e.args[1])
		if err != nil {
			return xpathValue{}, err
		}
		if second == "" {
			return xpathValue{kind: xpathValueString, str: ""}, nil
		}
		idx := strings.Index(first, second)
		if idx == -1 {
			return xpathValue{kind: xpathValueString, str: ""}, nil
		}
		return xpathValue{kind: xpathValueString, str: first[:idx]}, nil
	case "substring-after":
		if len(e.args) != 2 {
			return xpathValue{}, XPathError{"substring-after() expects two arguments"}
		}
		first, err := evalStringArg(ctx, e.args[0])
		if err != nil {
			return xpathValue{}, err
		}
		second, err := evalStringArg(ctx, e.args[1])
		if err != nil {
			return xpathValue{}, err
		}
		if second == "" {
			return xpathValue{kind: xpathValueString, str: ""}, nil
		}
		idx := strings.Index(first, second)
		if idx == -1 {
			return xpathValue{kind: xpathValueString, str: ""}, nil
		}
		return xpathValue{kind: xpathValueString, str: first[idx+len(second):]}, nil
	case "substring":
		return evalSubstring(ctx, e.args)
	case "string-length":
		val, err := evalStringArgOptional(ctx, e.args)
		if err != nil {
			return xpathValue{}, err
		}
		return xpathValue{kind: xpathValueNumber, num: float64(len([]rune(val)))}, nil
	case "normalize-space":
		val, err := evalStringArgOptional(ctx, e.args)
		if err != nil {
			return xpathValue{}, err
		}
		return xpathValue{kind: xpathValueString, str: normalizeSpace(val)}, nil
	case "translate":
		if len(e.args) != 3 {
			return xpathValue{}, XPathError{"translate() expects three arguments"}
		}
		first, err := evalStringArg(ctx, e.args[0])
		if err != nil {
			return xpathValue{}, err
		}
		from, err := evalStringArg(ctx, e.args[1])
		if err != nil {
			return xpathValue{}, err
		}
		to, err := evalStringArg(ctx, e.args[2])
		if err != nil {
			return xpathValue{}, err
		}
		return xpathValue{kind: xpathValueString, str: translateString(first, from, to)}, nil
	case "boolean":
		if len(e.args) != 1 {
			return xpathValue{}, XPathError{"boolean() expects one argument"}
		}
		val, err := e.args[0].eval(ctx)
		if err != nil {
			return xpathValue{}, err
		}
		return xpathValue{kind: xpathValueBoolean, b: val.toBool()}, nil
	case "not":
		if len(e.args) != 1 {
			return xpathValue{}, XPathError{"not() expects one argument"}
		}
		val, err := e.args[0].eval(ctx)
		if err != nil {
			return xpathValue{}, err
		}
		return xpathValue{kind: xpathValueBoolean, b: !val.toBool()}, nil
	case "true":
		return xpathValue{kind: xpathValueBoolean, b: true}, nil
	case "false":
		return xpathValue{kind: xpathValueBoolean, b: false}, nil
	case "lang":
		if len(e.args) != 1 {
			return xpathValue{}, XPathError{"lang() expects one argument"}
		}
		tag, err := evalStringArg(ctx, e.args[0])
		if err != nil {
			return xpathValue{}, err
		}
		return xpathValue{kind: xpathValueBoolean, b: xpathLang(ctx.node, tag)}, nil
	case "number":
		if len(e.args) == 0 {
			return xpathValue{kind: xpathValueNumber, num: xpathValue{kind: xpathValueString, str: nodeStringValue(newNodeValue(ctx.node))}.toNumber()}, nil
		}
		if len(e.args) != 1 {
			return xpathValue{}, XPathError{"number() expects one argument"}
		}
		val, err := e.args[0].eval(ctx)
		if err != nil {
			return xpathValue{}, err
		}
		return xpathValue{kind: xpathValueNumber, num: val.toNumber()}, nil
	case "sum":
		if len(e.args) != 1 {
			return xpathValue{}, XPathError{"sum() expects one argument"}
		}
		val, err := e.args[0].eval(ctx)
		if err != nil {
			return xpathValue{}, err
		}
		if val.kind != xpathValueNodeSet {
			return xpathValue{}, XPathError{"sum() expects a node-set"}
		}
		var total float64
		for _, n := range val.nodes {
			num := xpathValue{kind: xpathValueString, str: nodeStringValue(n)}.toNumber()
			if !math.IsNaN(num) {
				total += num
			}
		}
		return xpathValue{kind: xpathValueNumber, num: total}, nil
	case "floor":
		return evalMathUnary(ctx, e.args, math.Floor, "floor")
	case "ceiling":
		return evalMathUnary(ctx, e.args, math.Ceil, "ceiling")
	case "round":
		return evalMathUnary(ctx, e.args, xpathRound, "round")
	case "text":
		if len(e.args) != 0 {
			return xpathValue{}, XPathError{"text() expects no arguments"}
		}
		return xpathValue{kind: xpathValueString, str: nodeStringValue(newNodeValue(ctx.node))}, nil
	default:
		return xpathValue{}, XPathError{"unsupported xpath function: " + name}
	}
}

type xpathAttributeExpr struct {
	name string
}

func (e xpathAttributeExpr) eval(ctx xpathContext) (xpathValue, error) {
	if ctx.node == nil {
		return xpathValue{kind: xpathValueNodeSet}, nil
	}
	if ctx.node.Attrs == nil {
		return xpathValue{kind: xpathValueNodeSet}, nil
	}

	// Handle wildcard
	if e.name == "*" {
		nodes := make([]xpathNode, 0, len(ctx.node.Attrs))
		for k, v := range ctx.node.Attrs {
			value := ""
			if v != nil {
				value = *v
			}
			nodes = append(nodes, newAttrValue(k, value))
		}
		return xpathValue{kind: xpathValueNodeSet, nodes: nodes}, nil
	}

	for k, v := range ctx.node.Attrs {
		if strings.EqualFold(k, e.name) {
			value := ""
			if v != nil {
				value = *v
			}
			return xpathValue{kind: xpathValueNodeSet, nodes: []xpathNode{newAttrValue(k, value)}}, nil
		}
	}
	return xpathValue{kind: xpathValueNodeSet}, nil
}

type xpathNameExpr struct {
	name string
}

func (e xpathNameExpr) eval(ctx xpathContext) (xpathValue, error) {
	if ctx.node == nil {
		return xpathValue{kind: xpathValueNodeSet}, nil
	}
	if axis, test, ok := splitXPathAxis(e.name); ok {
		switch axis {
		case "preceding-sibling":
			return xpathValue{kind: xpathValueNodeSet, nodes: precedingSiblingNodes(ctx.node, test)}, nil
		case "following-sibling":
			return xpathValue{kind: xpathValueNodeSet, nodes: followingSiblingNodes(ctx.node, test)}, nil
		default:
			return xpathValue{kind: xpathValueNodeSet}, nil
		}
	}
	out := make([]xpathNode, 0)
	for _, child := range xpathChildren(ctx.node) {
		if e.name == "*" {
			out = append(out, newNodeValue(child))
			continue
		}
		if child != nil && strings.EqualFold(child.Name, e.name) {
			out = append(out, newNodeValue(child))
		}
	}
	return xpathValue{kind: xpathValueNodeSet, nodes: out}, nil
}

type xpathSelfExpr struct{}

func (e xpathSelfExpr) eval(ctx xpathContext) (xpathValue, error) {
	if ctx.node == nil {
		return xpathValue{kind: xpathValueNodeSet}, nil
	}
	return xpathValue{kind: xpathValueNodeSet, nodes: []xpathNode{newNodeValue(ctx.node)}}, nil
}

// xpathLocationPathExpr represents a location path used as an expression
// For example: count(.//p) or id(//div/@ref)
type xpathLocationPathExpr struct {
	path string // The location path string
}

func (e xpathLocationPathExpr) eval(ctx xpathContext) (xpathValue, error) {
	// Parse and execute the location path from the current context node
	if ctx.node == nil {
		return xpathValue{kind: xpathValueNodeSet}, nil
	}

	// For relative paths, we need to evaluate from the context node
	// Paths like "child::p", "./div", ".//p" are relative
	var nodes []*Node
	var err error

	if strings.HasPrefix(e.path, "/") {
		// Absolute path - find root and query from there
		root := ctx.node
		for root.Parent != nil {
			root = root.Parent
		}
		nodes, err = root.QueryXPath(e.path)
	} else {
		// Relative path - query from context node
		// For paths starting with ".", we can query directly
		// For axis paths like "child::p", they work from the context node
		nodes, err = ctx.node.QueryXPath(e.path)
	}

	if err != nil {
		return xpathValue{}, err
	}

	// Convert []*Node to []xpathNode
	xnodes := make([]xpathNode, len(nodes))
	for i, n := range nodes {
		xnodes[i] = newNodeValue(n)
	}

	return xpathValue{kind: xpathValueNodeSet, nodes: xnodes}, nil
}

type xpathExprParser struct {
	lexer *xpathLexer
	curr  xpathToken
}

func newXPathExprParser(input string) *xpathExprParser {
	lexer := newXPathLexer(input)
	parser := &xpathExprParser{lexer: lexer}
	parser.next()
	return parser
}

func (p *xpathExprParser) next() {
	p.curr = p.lexer.nextToken()
}

func (p *xpathExprParser) peek() xpathToken {
	return p.curr
}

func (p *xpathExprParser) consume(typ xpathTokenType) (xpathToken, error) {
	if p.curr.typ != typ {
		return xpathToken{}, XPathError{"invalid xpath predicate"}
	}
	tok := p.curr
	p.next()
	return tok, nil
}

func (p *xpathExprParser) parseExpr() (xpathExpr, error) {
	return p.parseUnion()
}

func (p *xpathExprParser) parseUnion() (xpathExpr, error) {
	left, err := p.parseOr()
	if err != nil {
		return nil, err
	}
	for p.curr.typ == xpathTokenOperator && p.curr.val == "|" {
		op := p.curr.val
		p.next()
		right, err := p.parseOr()
		if err != nil {
			return nil, err
		}
		left = xpathBinaryExpr{op: op, left: left, right: right}
	}
	return left, nil
}

func (p *xpathExprParser) parseOr() (xpathExpr, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for p.curr.typ == xpathTokenOperator && p.curr.val == "or" {
		op := p.curr.val
		p.next()
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = xpathBinaryExpr{op: op, left: left, right: right}
	}
	return left, nil
}

func (p *xpathExprParser) parseAnd() (xpathExpr, error) {
	left, err := p.parseEquality()
	if err != nil {
		return nil, err
	}
	for p.curr.typ == xpathTokenOperator && p.curr.val == "and" {
		op := p.curr.val
		p.next()
		right, err := p.parseEquality()
		if err != nil {
			return nil, err
		}
		left = xpathBinaryExpr{op: op, left: left, right: right}
	}
	return left, nil
}

func (p *xpathExprParser) parseEquality() (xpathExpr, error) {
	left, err := p.parseRelational()
	if err != nil {
		return nil, err
	}
	for p.curr.typ == xpathTokenOperator && (p.curr.val == "=" || p.curr.val == "!=") {
		op := p.curr.val
		p.next()
		right, err := p.parseRelational()
		if err != nil {
			return nil, err
		}
		left = xpathBinaryExpr{op: op, left: left, right: right}
	}
	return left, nil
}

func (p *xpathExprParser) parseRelational() (xpathExpr, error) {
	left, err := p.parseAdditive()
	if err != nil {
		return nil, err
	}
	for p.curr.typ == xpathTokenOperator && (p.curr.val == "<" || p.curr.val == "<=" || p.curr.val == ">" || p.curr.val == ">=") {
		op := p.curr.val
		p.next()
		right, err := p.parseAdditive()
		if err != nil {
			return nil, err
		}
		left = xpathBinaryExpr{op: op, left: left, right: right}
	}
	return left, nil
}

func (p *xpathExprParser) parseAdditive() (xpathExpr, error) {
	left, err := p.parseMultiplicative()
	if err != nil {
		return nil, err
	}
	for p.curr.typ == xpathTokenOperator && (p.curr.val == "+" || p.curr.val == "-") {
		op := p.curr.val
		p.next()
		right, err := p.parseMultiplicative()
		if err != nil {
			return nil, err
		}
		left = xpathBinaryExpr{op: op, left: left, right: right}
	}
	return left, nil
}

func (p *xpathExprParser) parseMultiplicative() (xpathExpr, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for p.curr.typ == xpathTokenOperator && (p.curr.val == "*" || p.curr.val == "div" || p.curr.val == "mod") {
		op := p.curr.val
		p.next()
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = xpathBinaryExpr{op: op, left: left, right: right}
	}
	return left, nil
}

func (p *xpathExprParser) parseUnary() (xpathExpr, error) {
	if p.curr.typ == xpathTokenOperator && (p.curr.val == "+" || p.curr.val == "-") {
		op := p.curr.val
		p.next()
		expr, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return xpathUnaryExpr{op: op, expr: expr}, nil
	}
	return p.parsePrimary()
}

func (p *xpathExprParser) parsePrimary() (xpathExpr, error) {
	switch p.curr.typ {
	case xpathTokenNumber:
		val, err := strconv.ParseFloat(p.curr.val, 64)
		if err != nil {
			return nil, XPathError{"invalid xpath number"}
		}
		p.next()
		return xpathLiteralExpr{value: xpathValue{kind: xpathValueNumber, num: val}}, nil
	case xpathTokenString:
		val := p.curr.val
		p.next()
		return xpathLiteralExpr{value: xpathValue{kind: xpathValueString, str: val}}, nil
	case xpathTokenAt:
		p.next()
		tok, err := p.consume(xpathTokenIdentifier)
		if err != nil {
			return nil, XPathError{"missing xpath attribute name"}
		}
		return xpathAttributeExpr{name: tok.val}, nil
	case xpathTokenIdentifier:
		name := p.curr.val
		p.next()
		if p.curr.typ == xpathTokenLParen {
			p.next()
			args, err := p.parseArguments()
			if err != nil {
				return nil, err
			}
			return xpathFunctionExpr{name: name, args: args}, nil
		}

		// Check if this is an axis specifier (followed by ::)
		if p.curr.typ == xpathTokenOperator && p.curr.val == "::" {
			// This is a location path like "child::p" or "descendant::div"
			p.next() // consume ::
			return p.parseLocationPathExpr(name + "::")
		}

		return xpathNameExpr{name: name}, nil
	case xpathTokenDot:
		// Check if this is followed by / or // to make it a location path
		p.next()
		if p.curr.typ == xpathTokenOperator && (p.curr.val == "/" || p.curr.val == "//") {
			// This is a location path like ".//p" or "./span"
			// Build the start including the operator
			op := p.curr.val
			p.next()
			return p.parseLocationPathExpr("." + op)
		}
		// Just a self reference
		return xpathSelfExpr{}, nil
	case xpathTokenLParen:
		p.next()
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if _, err := p.consume(xpathTokenRParen); err != nil {
			return nil, err
		}
		return expr, nil
	default:
		return nil, XPathError{"invalid xpath predicate"}
	}
}

// parseLocationPathExpr parses a location path expression when used inside a predicate/function
// startToken is the token we've already consumed (like "." or an axis name)
func (p *xpathExprParser) parseLocationPathExpr(startToken string) (xpathExpr, error) {
	// Since the lexer doesn't tokenize brackets, we need a simpler approach:
	// Just consume tokens until we hit a clear delimiter
	var pathBuilder strings.Builder
	pathBuilder.WriteString(startToken)

	// Keep consuming tokens that are part of the location path
	for {
		switch p.curr.typ {
		case xpathTokenEOF:
			return xpathLocationPathExpr{path: pathBuilder.String()}, nil

		case xpathTokenComma:
			// End of location path
			return xpathLocationPathExpr{path: pathBuilder.String()}, nil

		case xpathTokenRParen:
			// End of location path (closing the function call)
			return xpathLocationPathExpr{path: pathBuilder.String()}, nil

		case xpathTokenOperator:
			op := p.curr.val

			// Path operators - always include
			switch op {
			case "/", "//", "::":
				pathBuilder.WriteString(op)
				p.next()
			case "=", "!=", "<", ">", "<=", ">=", "+", "-", "*", "div", "mod", "and", "or", "|":
				// These operators end the location path
				return xpathLocationPathExpr{path: pathBuilder.String()}, nil
			default:
				// Unknown operator - end path
				return xpathLocationPathExpr{path: pathBuilder.String()}, nil
			}

		case xpathTokenIdentifier:
			pathBuilder.WriteString(p.curr.val)
			p.next()

		case xpathTokenAt:
			pathBuilder.WriteString("@")
			p.next()

		case xpathTokenDot:
			pathBuilder.WriteString(".")
			p.next()

		case xpathTokenNumber:
			pathBuilder.WriteString(p.curr.val)
			p.next()

		case xpathTokenString:
			// Include quotes
			pathBuilder.WriteString("'")
			pathBuilder.WriteString(p.curr.val)
			pathBuilder.WriteString("'")
			p.next()

		case xpathTokenLParen:
			pathBuilder.WriteString("(")
			p.next()
			// Check if this is immediately followed by ) for node tests
			if p.curr.typ == xpathTokenRParen {
				pathBuilder.WriteString(")")
				p.next()
			}

		default:
			// Unknown token - end location path
			return xpathLocationPathExpr{path: pathBuilder.String()}, nil
		}
	}
}

func (p *xpathExprParser) parseArguments() ([]xpathExpr, error) {
	if p.curr.typ == xpathTokenRParen {
		p.next()
		return nil, nil
	}

	args := []xpathExpr{}
	for {
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		args = append(args, expr)
		if p.curr.typ == xpathTokenComma {
			p.next()
			continue
		}
		if p.curr.typ == xpathTokenRParen {
			p.next()
			break
		}
		return nil, XPathError{"invalid xpath function arguments"}
	}
	return args, nil
}

type xpathTokenType int

const (
	xpathTokenEOF xpathTokenType = iota
	xpathTokenIdentifier
	xpathTokenNumber
	xpathTokenString
	xpathTokenOperator
	xpathTokenAt
	xpathTokenDot
	xpathTokenLParen
	xpathTokenRParen
	xpathTokenComma
)

type xpathToken struct {
	typ xpathTokenType
	val string
}

type xpathLexer struct {
	input string
	pos   int
}

func newXPathLexer(input string) *xpathLexer {
	return &xpathLexer{input: input}
}

func (l *xpathLexer) nextToken() xpathToken {
	l.skipWhitespace()
	if l.pos >= len(l.input) {
		return xpathToken{typ: xpathTokenEOF}
	}

	ch := l.input[l.pos]
	switch ch {
	case '@':
		l.pos++
		return xpathToken{typ: xpathTokenAt, val: "@"}
	case '.':
		l.pos++
		return xpathToken{typ: xpathTokenDot, val: "."}
	case '(':
		l.pos++
		return xpathToken{typ: xpathTokenLParen, val: "("}
	case ')':
		l.pos++
		return xpathToken{typ: xpathTokenRParen, val: ")"}
	case ',':
		l.pos++
		return xpathToken{typ: xpathTokenComma, val: ","}
	case '!', '=', '<', '>', '+', '-', '*', '|', '/', ':':
		return l.readOperator()
	case '"', '\'':
		return l.readString()
	default:
		if isDigit(ch) {
			return l.readNumber()
		}
		if isIdentStart(ch) {
			return l.readIdentifierOrOperator()
		}
	}

	l.pos++
	return xpathToken{typ: xpathTokenEOF}
}

func (l *xpathLexer) readOperator() xpathToken {
	start := l.pos
	ch := l.input[l.pos]
	l.pos++

	// Handle // (double slash)
	if ch == '/' && l.pos < len(l.input) && l.input[l.pos] == '/' {
		l.pos++
		return xpathToken{typ: xpathTokenOperator, val: "//"}
	}

	// Handle <= and >=
	if (ch == '<' || ch == '>') && l.pos < len(l.input) && l.input[l.pos] == '=' {
		l.pos++
		return xpathToken{typ: xpathTokenOperator, val: l.input[start:l.pos]}
	}

	// Handle !=
	if ch == '!' && l.pos < len(l.input) && l.input[l.pos] == '=' {
		l.pos++
		return xpathToken{typ: xpathTokenOperator, val: "!="}
	}

	// Handle :: (double colon for axes)
	if ch == ':' && l.pos < len(l.input) && l.input[l.pos] == ':' {
		l.pos++
		return xpathToken{typ: xpathTokenOperator, val: "::"}
	}

	// Single character operators
	if ch == '=' || ch == '<' || ch == '>' || ch == '+' || ch == '-' || ch == '*' || ch == '|' || ch == '/' || ch == ':' {
		return xpathToken{typ: xpathTokenOperator, val: string(ch)}
	}

	return xpathToken{typ: xpathTokenOperator, val: string(ch)}
}

func (l *xpathLexer) readString() xpathToken {
	quote := l.input[l.pos]
	l.pos++
	start := l.pos
	for l.pos < len(l.input) && l.input[l.pos] != quote {
		l.pos++
	}
	val := l.input[start:l.pos]
	if l.pos < len(l.input) {
		l.pos++
	}
	return xpathToken{typ: xpathTokenString, val: val}
}

func (l *xpathLexer) readNumber() xpathToken {
	start := l.pos
	seenDot := false
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == '.' {
			if seenDot {
				break
			}
			seenDot = true
			l.pos++
			continue
		}
		if !isDigit(ch) {
			break
		}
		l.pos++
	}
	return xpathToken{typ: xpathTokenNumber, val: l.input[start:l.pos]}
}

func (l *xpathLexer) readIdentifierOrOperator() xpathToken {
	start := l.pos
	for l.pos < len(l.input) {
		r, size := utf8.DecodeRuneInString(l.input[l.pos:])
		if r == utf8.RuneError && size == 1 {
			break
		}
		// Don't include ':' in identifiers anymore since we tokenize :: as operator
		if !isXPathIdentRune(r) {
			break
		}
		l.pos += size
	}
	val := l.input[start:l.pos]
	switch val {
	case "and", "or", "div", "mod":
		return xpathToken{typ: xpathTokenOperator, val: val}
	default:
		return xpathToken{typ: xpathTokenIdentifier, val: val}
	}
}

func (l *xpathLexer) skipWhitespace() {
	for l.pos < len(l.input) {
		if !isSpace(l.input[l.pos]) {
			return
		}
		l.pos++
	}
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isIdentStart(ch byte) bool {
	return unicode.IsLetter(rune(ch)) || ch == '_' || ch == ':'
}

func isSpace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

func compareXPathValues(left, right xpathValue, op string) xpathValue {
	if left.kind == xpathValueNodeSet || right.kind == xpathValueNodeSet {
		leftStr := left.toString()
		rightStr := right.toString()
		return compareXPathValues(xpathValue{kind: xpathValueString, str: leftStr}, xpathValue{kind: xpathValueString, str: rightStr}, op)
	}

	if left.kind == xpathValueBoolean || right.kind == xpathValueBoolean {
		lb := left.toBool()
		rb := right.toBool()
		return xpathValue{kind: xpathValueBoolean, b: compareBool(lb, rb, op)}
	}

	if left.kind == xpathValueNumber || right.kind == xpathValueNumber {
		ln := left.toNumber()
		rn := right.toNumber()
		return xpathValue{kind: xpathValueBoolean, b: compareNumber(ln, rn, op)}
	}

	return xpathValue{kind: xpathValueBoolean, b: compareString(left.toString(), right.toString(), op)}
}

func compareBool(left, right bool, op string) bool {
	switch op {
	case "=":
		return left == right
	case "!=":
		return left != right
	default:
		return false
	}
}

func compareNumber(left, right float64, op string) bool {
	switch op {
	case "=":
		return left == right
	case "!=":
		return left != right
	case "<":
		return left < right
	case "<=":
		return left <= right
	case ">":
		return left > right
	case ">=":
		return left >= right
	default:
		return false
	}
}

func compareString(left, right, op string) bool {
	switch op {
	case "=":
		return left == right
	case "!=":
		return left != right
	case "<":
		return left < right
	case "<=":
		return left <= right
	case ">":
		return left > right
	case ">=":
		return left >= right
	default:
		return false
	}
}

func evalStringArg(ctx xpathContext, expr xpathExpr) (string, error) {
	val, err := expr.eval(ctx)
	if err != nil {
		return "", err
	}
	return val.toString(), nil
}

func evalStringArgOptional(ctx xpathContext, args []xpathExpr) (string, error) {
	if len(args) == 0 {
		return nodeStringValue(newNodeValue(ctx.node)), nil
	}
	if len(args) != 1 {
		return "", XPathError{"invalid xpath function arguments"}
	}
	return evalStringArg(ctx, args[0])
}

func evalSubstring(ctx xpathContext, args []xpathExpr) (xpathValue, error) {
	if len(args) < 2 || len(args) > 3 {
		return xpathValue{}, XPathError{"substring() expects two or three arguments"}
	}
	str, err := evalStringArg(ctx, args[0])
	if err != nil {
		return xpathValue{}, err
	}
	startVal, err := args[1].eval(ctx)
	if err != nil {
		return xpathValue{}, err
	}

	// XPath 1.0 uses 1-based indexing with proper rounding
	startNum := startVal.toNumber()
	start := int(math.Round(startNum)) - 1

	chars := []rune(str)
	if len(chars) == 0 {
		return xpathValue{kind: xpathValueString, str: ""}, nil
	}
	if start < 0 {
		start = 0
	}
	end := len(chars)
	if len(args) == 3 {
		lengthVal, err := args[2].eval(ctx)
		if err != nil {
			return xpathValue{}, err
		}
		length := int(math.Round(lengthVal.toNumber()))
		if length < 0 {
			length = 0
		}
		end = start + length
		if end > len(chars) {
			end = len(chars)
		}
	}
	if start > len(chars) {
		return xpathValue{kind: xpathValueString, str: ""}, nil
	}
	return xpathValue{kind: xpathValueString, str: string(chars[start:end])}, nil
}

func evalMathUnary(ctx xpathContext, args []xpathExpr, op func(float64) float64, name string) (xpathValue, error) {
	if len(args) != 1 {
		return xpathValue{}, XPathError{name + "() expects one argument"}
	}
	val, err := args[0].eval(ctx)
	if err != nil {
		return xpathValue{}, err
	}
	return xpathValue{kind: xpathValueNumber, num: op(val.toNumber())}, nil
}

func xpathRound(val float64) float64 {
	if math.IsNaN(val) || math.IsInf(val, 0) {
		return val
	}
	if val >= 0 {
		return math.Floor(val + 0.5)
	}
	return math.Ceil(val - 0.5)
}

func normalizeSpace(input string) string {
	fields := strings.Fields(input)
	return strings.Join(fields, " ")
}

func translateString(input, from, to string) string {
	fromRunes := []rune(from)
	toRunes := []rune(to)
	mapping := make(map[rune]*rune, len(fromRunes))
	for i, r := range fromRunes {
		if i < len(toRunes) {
			mapping[r] = &toRunes[i]
		} else {
			mapping[r] = nil
		}
	}

	var out strings.Builder
	for _, r := range input {
		replacement, ok := mapping[r]
		if !ok {
			out.WriteRune(r)
			continue
		}
		if replacement != nil {
			out.WriteRune(*replacement)
		}
	}
	return out.String()
}

func idLookup(root *Node, ids []string) []xpathNode {
	if root == nil || len(ids) == 0 {
		return nil
	}
	idSet := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		idSet[id] = struct{}{}
	}
	var out []xpathNode
	var walk func(n *Node)
	walk = func(n *Node) {
		if n == nil {
			return
		}
		if n.Attrs != nil {
			if idPtr, ok := n.Attrs["id"]; ok && idPtr != nil {
				if _, ok := idSet[*idPtr]; ok {
					out = append(out, newNodeValue(n))
				}
			}
		}
		for _, child := range xpathChildren(n) {
			walk(child)
		}
	}
	walk(root)
	return out
}

func localName(node *Node) string {
	if node == nil {
		return ""
	}
	parts := strings.Split(node.Name, ":")
	return parts[len(parts)-1]
}

func nodeForNameFunction(ctx xpathContext, args []xpathExpr) *Node {
	if len(args) == 0 {
		return ctx.node
	}
	val, err := args[0].eval(ctx)
	if err != nil {
		return nil
	}
	if val.kind != xpathValueNodeSet || len(val.nodes) == 0 {
		return nil
	}
	return val.nodes[0].node
}

func xpathLang(node *Node, lang string) bool {
	if node == nil {
		return false
	}
	lang = strings.ToLower(lang)
	for n := node; n != nil; n = n.Parent {
		val := strings.ToLower(n.Attr("lang"))
		if val == "" {
			continue
		}
		// First lang attribute found in ancestor chain determines the result
		return val == lang || strings.HasPrefix(val, lang+"-")
	}
	return false
}

func isXPathIdentRune(r rune) bool {
	return r == '_' || r == '-' || r == '*' ||
		unicode.IsLetter(r) ||
		unicode.IsDigit(r)
}
