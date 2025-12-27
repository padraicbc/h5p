package dom

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// CSS selector matcher that supports tag names, #id, .class, attribute operators,
// descendant/child/adjacent/general sibling combinators, selector groups, and a
// subset of pseudo classes used by the test suite.

type attrSelector struct {
	name  string
	op    string
	value string
}

type pseudoSelector struct {
	name string
	arg  string
}

type compoundSelector struct {
	tag     string
	id      string
	classes []string
	attrs   []attrSelector
	pseudos []pseudoSelector
}

type selectorPart struct {
	combinator string // "", " ", ">", "+", "~"
	compound   compoundSelector
}

type selectorChain []selectorPart

// matchChain walks a selector chain from the rightmost part (closest to the
// node) toward the left, following the combinators to ensure each ancestor or
// sibling satisfies its fragment.
func matchChain(node *Node, chain selectorChain) (bool, error) {
	if node == nil {
		return false, nil
	}

	ok, err := matchCompound(node, chain[len(chain)-1].compound)
	if err != nil || !ok {
		return ok, err
	}

	cur := node
	for i := len(chain) - 2; i >= 0; i-- {
		comb := chain[i+1].combinator

		switch comb {
		case ">":
			cur = cur.Parent
			if cur == nil {
				return false, nil
			}
			ok, err = matchCompound(cur, chain[i].compound)
			if err != nil || !ok {
				return ok, err
			}

		case "+":
			cur = previousElementSibling(cur)
			if cur == nil {
				return false, nil
			}
			ok, err = matchCompound(cur, chain[i].compound)
			if err != nil || !ok {
				return ok, err
			}
			// Ensure we only match the immediately preceding sibling and
			// not subsequent elements in a run of matching siblings.
			for prev := previousElementSibling(cur); prev != nil; prev = previousElementSibling(prev) {
				prevOk, prevErr := matchCompound(prev, chain[i].compound)
				if prevErr != nil {
					return false, prevErr
				}
				if prevOk {
					return false, nil
				}
			}

		case "~":
			found := false
			for sib := previousElementSibling(cur); sib != nil; sib = previousElementSibling(sib) {
				ok, err = matchCompound(sib, chain[i].compound)
				if err != nil {
					return false, err
				}
				if ok {
					cur = sib
					found = true
					break
				}
			}
			if !found {
				return false, nil
			}

		case "": // descendant
			found := false
			for p := cur.Parent; p != nil; p = p.Parent {
				ok, err = matchCompound(p, chain[i].compound)
				if err != nil {
					return false, err
				}
				if ok {
					cur = p
					found = true
					break
				}
			}
			if !found {
				return false, nil
			}

		default:
			return false, SelectorError{"invalid combinator: " + comb}
		}
	}

	return true, nil
}

// matchCompound reports whether a node matches a compound selector.
//
// A compound selector represents all simple selectors between combinators,
// for example:
//
//	div#main.container[attr=value]:first-child
//
// Matching rules:
//   - The node must be an element node (not text, comment, etc.).
//   - The type selector (tag name) must match, unless it is '*'.
//   - ID selectors must match exactly (case-sensitive).
//   - Class selectors must match as whitespace-separated tokens.
//   - Attribute selectors are delegated to matchAttr.
//   - Pseudo-classes are delegated to matchPseudo.
//
// A return value of (false, nil) means the node does not match.
// A non-nil error means the selector is semantically invalid and must
// abort matching (e.g. unsupported pseudo-class, invalid nth expression).
func matchCompound(n *Node, comp compoundSelector) (bool, error) {
	// Reject non-element or invalid nodes (e.g. text, comment)
	if n == nil || n.Name == "" || strings.HasPrefix(n.Name, "#") {
		return false, nil
	}

	// --- Type selector (tag name) ---
	// Tag names are case-insensitive in HTML
	tag := strings.ToLower(n.Name)
	if comp.tag != "" && comp.tag != "*" && tag != comp.tag {
		return false, nil
	}

	// --- ID selector (#id) ---
	// IDs are case-sensitive
	if comp.id != "" {
		if n.Attrs == nil {
			return false, nil
		}
		idPtr, ok := n.Attrs["id"]
		if !ok || idPtr == nil || *idPtr != comp.id {
			return false, nil
		}
	}

	// --- Class selectors (.class) ---
	// Classes are matched as whitespace-separated tokens and are case-sensitive
	if len(comp.classes) > 0 {
		classVal := ""
		if n.Attrs != nil {
			for k, v := range n.Attrs {
				if strings.EqualFold(k, "class") && v != nil {
					classVal = *v
					break
				}
			}
		}

		classes := strings.Fields(classVal)
		if len(classes) == 0 {
			return false, nil
		}

		classSet := make(map[string]struct{}, len(classes))
		for _, c := range classes {
			classSet[c] = struct{}{}
		}

		for _, want := range comp.classes {
			parts := strings.Fields(want)
			if len(parts) == 0 {
				return false, nil
			}
			for _, p := range parts {
				if _, ok := classSet[p]; !ok {
					return false, nil
				}
			}
		}
	}

	// --- Attribute selectors ([attr], [attr=value], etc.) ---
	for _, attr := range comp.attrs {
		ok, err := matchAttr(n, attr)
		if err != nil {
			// Semantic selector error (invalid operator, etc.)
			return false, err
		}
		if !ok {
			return false, nil
		}
	}

	// --- Pseudo-classes (:first-child, :not(), :nth-child(), etc.) ---
	for _, pseudo := range comp.pseudos {
		ok, err := matchPseudo(n, pseudo)
		if err != nil {
			// Unsupported or malformed pseudo-class
			return false, err
		}
		if !ok {
			return false, nil
		}
	}

	return true, nil
}

// elementChildren returns only the element children, skipping text, comments,
// and nil entries.
func elementChildren(n *Node) []*Node {
	out := make([]*Node, 0, len(n.Children))
	for _, c := range n.Children {
		if c != nil && c.Name != "" && !strings.HasPrefix(c.Name, "#") {
			out = append(out, c)
		}
	}
	return out
}

// previousElementSibling finds the nearest previous sibling that is an
// element node.
func previousElementSibling(n *Node) *Node {
	if n == nil || n.Parent == nil {
		return nil
	}
	children := elementChildren(n.Parent)
	for i, c := range children {
		if c == n {
			if i == 0 {
				return nil
			}
			return children[i-1]
		}
	}
	return nil
}

// matchAttr reports whether a node matches an attribute selector.
//
// It returns:
//   - (true, nil)   → attribute selector matches
//   - (false, nil)  → attribute selector does not match
//   - (false, err)  → selector is semantically invalid
//
// Attribute names are matched case-insensitively (HTML rules).
// Attribute values are matched case-sensitively.
// matchAttr evaluates an attribute selector against a node, honoring the
// operator semantics defined by the CSS selector grammar.
func matchAttr(n *Node, attr attrSelector) (bool, error) {
	if n == nil || n.Attrs == nil {
		return false, nil
	}

	val := ""
	found := false

	// Locate the attribute (case-insensitive name match)
	for k, v := range n.Attrs {
		if strings.EqualFold(k, attr.name) {
			found = true
			if v != nil {
				val = *v
			}
			break
		}
	}

	if !found {
		return false, nil
	}

	switch attr.op {

	// [attr]
	case "exists":
		return true, nil

	// [attr=value]
	case "=":
		return val == attr.value, nil

	// [attr~=value]
	// Value must be one of the whitespace-separated tokens
	case "~=":
		if attr.value == "" {
			return false, nil
		}
		for _, word := range strings.Fields(val) {
			if word == attr.value {
				return true, nil
			}
		}
		return false, nil

	// [attr|=value]
	// Exact match or prefix followed by '-'
	case "|=":
		if attr.value == "" {
			return false, nil
		}
		return val == attr.value || strings.HasPrefix(val, attr.value+"-"), nil

	// [attr^=value]
	// Must start with value (empty value never matches)
	case "^=":
		if attr.value == "" {
			return false, nil
		}
		return strings.HasPrefix(val, attr.value), nil

	// [attr$=value]
	// Must end with value (empty value never matches)
	case "$=":
		if attr.value == "" {
			return false, nil
		}
		return strings.HasSuffix(val, attr.value), nil

	// [attr*=value]
	// Must contain value (empty value never matches)
	case "*=":
		if attr.value == "" {
			return false, nil
		}
		return strings.Contains(val, attr.value), nil
	}

	// Unknown attribute operator → semantic error
	return false, SelectorError{"unsupported attribute operator: " + attr.op}
}

// matchPseudo reports whether a node matches a pseudo-class selector.
//
// It returns (false, nil) when the node simply does not match,
// and (false, error) when the pseudo-class is invalid or unsupported.
//
// Semantic validation (e.g. nth-child expressions) is performed here,
// not during parsing, to match CSS and Python behavior.
// matchPseudo evaluates supported pseudo-classes on a node, covering
// positional pseudos like :first-child and :nth-child as well as content-based
// pseudos such as :contains.
func matchPseudo(n *Node, pseudo pseudoSelector) (bool, error) {
	if n == nil {
		return false, nil
	}

	switch pseudo.name {

	// :first-child
	case "first-child":
		siblings := elementChildren(n.Parent)
		return len(siblings) > 0 && siblings[0] == n, nil

	// :last-child
	case "last-child":
		siblings := elementChildren(n.Parent)
		return len(siblings) > 0 && siblings[len(siblings)-1] == n, nil

	// :only-child
	case "only-child":
		siblings := elementChildren(n.Parent)
		return len(siblings) == 1 && siblings[0] == n, nil

	// :first-of-type
	case "first-of-type":
		siblings := elementChildren(n.Parent)
		for _, s := range siblings {
			if s.Name == n.Name {
				return s == n, nil
			}
		}
		return false, nil

	// :last-of-type
	case "last-of-type":
		siblings := elementChildren(n.Parent)
		for i := len(siblings) - 1; i >= 0; i-- {
			if siblings[i].Name == n.Name {
				return siblings[i] == n, nil
			}
		}
		return false, nil

	// :only-of-type
	case "only-of-type":
		siblings := elementChildren(n.Parent)
		count := 0
		for _, s := range siblings {
			if s.Name == n.Name {
				count++
			}
		}
		return count == 1, nil

	// :nth-child(...)
	case "nth-child":
		if pseudo.arg == "" {
			return false, SelectorError{"nth-child requires an argument"}
		}
		siblings := elementChildren(n.Parent)
		index := -1
		for i, s := range siblings {
			if s == n {
				index = i + 1 // 1-indexed
				break
			}
		}
		if index == -1 {
			return false, nil
		}
		ok, err := matchNthExpression(index, pseudo.arg)
		if err != nil {
			return false, err
		}
		return ok, nil

	// :nth-of-type(...)
	case "nth-of-type":
		if pseudo.arg == "" {
			return false, SelectorError{"nth-of-type requires an argument"}
		}
		siblings := elementChildren(n.Parent)
		pos := 0
		for _, s := range siblings {
			if s.Name == n.Name {
				pos++
				if s == n {
					break
				}
			}
		}
		ok, err := matchNthExpression(pos, pseudo.arg)
		if err != nil {
			return false, err
		}
		return ok, nil

	// :root
	case "root":
		if n.Name == "document" {
			return false, nil
		}
		return n.Parent != nil && n.Parent.Name == "document", nil

	// :empty
	case "empty":
		for _, c := range n.Children {
			if c == nil {
				continue
			}
			// Ignore comments
			if strings.HasPrefix(c.Name, "#") {
				if c.Name == "#text" && strings.TrimSpace(c.Data) != "" {
					return false, nil
				}
				continue
			}
			return false, nil
		}
		return true, nil

	// :not(...)
	case "not":
		// Empty :not() matches everything
		if strings.TrimSpace(pseudo.arg) == "" {
			return true, nil
		}

		chains, err := parseSelector(pseudo.arg)
		if err != nil {
			// Invalid selector inside :not() → ignore :not()
			return true, nil
		}

		// CSS Level 3 :not() allows only a single compound selector
		if len(chains) != 1 || len(chains[0]) != 1 {
			// Complex selector inside :not() → ignore :not()
			return true, nil
		}

		ok, err := matchCompound(n, chains[0][0].compound)
		if err != nil {
			// Defensive: ignore :not() on error
			return true, nil
		}

		return !ok, nil

		// :contains(...) — non-standard, but supported
	case "contains":
		if strings.TrimSpace(pseudo.arg) == "" {
			return false, SelectorError{"contains() requires an argument"}
		}

		text := n.ToText("", false)
		arg := strings.Trim(pseudo.arg, "\"'")
		return strings.Contains(text, arg), nil

	}

	// Unsupported pseudo-class
	return false, SelectorError{"unsupported pseudo-class: :" + pseudo.name}
}

// matchNthExpression evaluates whether a 1-based position matches an
// :nth-child() / :nth-of-type() expression.
//
// Supported forms:
//   - odd
//   - even
//   - <integer>           (e.g. 3)
//   - an+b                (e.g. 2n+1, n, -n+3)
//
// It returns:
//   - (true, nil)   → position matches the expression
//   - (false, nil)  → position does not match
//   - (false, err)  → expression is syntactically invalid
//
// Semantic validation happens here (not during parsing), matching CSS
// and Python selector behavior.
// matchNthExpression checks whether a position satisfies an "an+b" style nth
// expression (e.g. "2n+1" for odd positions).
func matchNthExpression(pos int, expr string) (bool, error) {
	if pos <= 0 {
		return false, nil
	}

	expr = strings.TrimSpace(expr)
	if expr == "" {
		return false, SelectorError{"empty nth-child expression"}
	}

	switch strings.ToLower(expr) {

	// odd → 1, 3, 5, ...
	case "odd":
		return pos%2 == 1, nil

	// even → 2, 4, 6, ...
	case "even":
		return pos%2 == 0, nil
	}

	// Simple integer form: :nth-child(3)
	if !strings.ContainsAny(expr, "nN") {
		val, err := parseInt(expr)
		if err != nil {
			return false, SelectorError{"invalid nth-child expression: " + expr}
		}
		return pos == val, nil
	}

	// an+b form
	a, b, ok := parseAnPlusB(expr)
	if !ok {
		return false, SelectorError{"invalid nth-child expression: " + expr}
	}

	// a == 0 → equivalent to b
	if a == 0 {
		return pos == b, nil
	}

	// Check if pos satisfies pos = a*n + b for n >= 0
	diff := pos - b
	if diff%a != 0 {
		return false, nil
	}

	n := diff / a
	return n >= 0, nil
}

func parseInt(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, strconv.ErrSyntax
	}
	return strconv.Atoi(s)
}

func parseAnPlusB(expr string) (a int, b int, ok bool) {
	expr = strings.ToLower(strings.ReplaceAll(expr, " ", ""))
	nIndex := strings.IndexRune(expr, 'n')
	if nIndex == -1 {
		return 0, 0, false
	}

	aPart := expr[:nIndex]
	switch aPart {
	case "", "+":
		a = 1
	case "-":
		a = -1
	default:
		var err error
		a, err = parseInt(aPart)
		if err != nil {
			return 0, 0, false
		}
	}

	bPart := expr[nIndex+1:]
	if bPart == "" {
		b = 0
		return a, b, true
	}
	var err error
	b, err = parseInt(bPart)
	if err != nil {
		return 0, 0, false
	}
	return a, b, true
}

func readIdent(s string, i int) (string, int) {
	var out strings.Builder

	for i < len(s) {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size == 1 {
			break
		}

		// CSS escape
		if r == '\\' {
			i += size
			if i >= len(s) {
				break
			}

			r2, size2 := utf8.DecodeRuneInString(s[i:])

			// escaped newline → discard both
			if r2 == '\n' || r2 == '\r' || r2 == '\f' {
				i += size2
				continue
			}

			// numeric escape (up to 6 hex digits)
			if isHexDigit(r2) {
				hex := string(r2)
				i += size2

				for count := 0; count < 5 && i < len(s); count++ { // allow up to 6 hex digits total
					rh, sh := utf8.DecodeRuneInString(s[i:])
					if !isHexDigit(rh) {
						break
					}
					hex += string(rh)
					i += sh
				}

				val, err := strconv.ParseInt(hex, 16, 32)
				if err == nil {
					out.WriteRune(rune(val))
				}

				// optional whitespace after numeric escape
				if i < len(s) {
					rw, sw := utf8.DecodeRuneInString(s[i:])
					if isCSSSpace(rw) {
						i += sw
					}
				}
				continue
			}

			// escaped single character
			out.WriteRune(r2)
			i += size2
			continue
		}

		// normal identifier character
		if !isIdentRune(r) {
			break
		}

		out.WriteRune(r)
		i += size
	}

	return out.String(), i
}
func isCSSSpace(r rune) bool {
	return r == ' ' || r == '\n' || r == '\r' || r == '\t' || r == '\f'
}

func isIdentRune(r rune) bool {
	return r == '_' || r == '-' || r == '*' ||
		unicode.IsLetter(r) ||
		unicode.IsDigit(r)
}

//	func isSpace(b byte) bool {
//		return b == ' ' || b == '\n' || b == '\t' || b == '\r' || b == '\f'
//	}
func isHexDigit(r rune) bool {
	return r >= '0' && r <= '9' ||
		r >= 'a' && r <= 'f' ||
		r >= 'A' && r <= 'F'
}

func parseSelector(sel string) ([]selectorChain, error) {
	if strings.TrimSpace(sel) == "" {
		return nil, SelectorError{"empty selector"}
	}

	groups := splitTopLevel(sel, ',')
	chains := make([]selectorChain, 0, len(groups))

	for _, group := range groups {
		if strings.TrimSpace(group) == "" {
			return nil, SelectorError{"empty selector group"}
		}
		parts, err := parseChain(group)
		if err != nil {
			return nil, err
		}

		if len(parts) == 0 {
			return nil, SelectorError{"invalid selector: " + group}
		}

		chains = append(chains, parts)
	}

	return chains, nil
}

func splitTopLevel(s string, sep rune) []string {
	depth := 0
	quote := byte(0)
	var parts []string
	start := 0

	for i := 0; i < len(s); i++ {
		ch := s[i]
		switch ch {
		case '"', '\'':
			switch quote {
			case 0:
				quote = ch
			case ch:
				quote = 0
			}
		case '[', '(':
			if quote == 0 {
				depth++
			}
		case ']', ')':
			if quote == 0 && depth > 0 {
				depth--
			}
		default:
			if quote == 0 && depth == 0 && rune(ch) == sep {
				parts = append(parts, s[start:i])
				start = i + 1
			}
		}
	}

	parts = append(parts, s[start:])
	return parts
}

// parseChain parses a single CSS selector chain (no commas).
//
// A selector chain is a sequence of compound selectors connected by combinators,
// for example:
//
//	div > ul li.special
//
// This function:
//   - Splits the selector into ordered selectorParts (left → right)
//   - Correctly handles combinators: descendant (whitespace), >, +, ~
//   - Tracks bracket and quote depth so combinators inside [], () or strings
//     are not treated as separators
//   - Enforces CSS validity rules:
//   - no leading combinator        ("> div")
//   - no trailing combinator       ("div >")
//   - no consecutive combinators   ("div > > p")
//   - no empty compound selectors
//   - balanced brackets and quotes
//
// The combinator is stored on the *right-hand* selectorPart, matching CSS
// semantics (e.g. in "div > p", the '>' belongs to 'p').
//
// On any structural or syntax error, a SelectorError is returned.
func parseChain(sel string) (selectorChain, error) {
	if strings.TrimSpace(sel) == "" {
		return nil, SelectorError{"empty selector"}
	}

	var chain selectorChain
	current := ""
	combinator := ""
	depth := 0
	quote := byte(0)

	flush := func() error {
		if len(current) == 0 {
			return SelectorError{"empty compound selector"}
		}
		compound, err := parseCompound(current)
		if err != nil {
			return err
		}
		chain = append(chain, selectorPart{
			combinator: combinator,
			compound:   compound,
		})
		current = ""
		combinator = ""
		return nil
	}

	nextNonSpace := func(start int) byte {
		for j := start; j < len(sel); j++ {
			if !isSpace(sel[j]) {
				return sel[j]
			}
		}
		return 0
	}

	for i := 0; i < len(sel); {
		r, size := utf8.DecodeRuneInString(sel[i:])
		if r == utf8.RuneError && size == 1 {
			return nil, SelectorError{"invalid utf-8 in selector"}
		}

		if r == '\\' {
			i += size

			if i < len(sel) {
				r2, size2 := utf8.DecodeRuneInString(sel[i:])

				// Escaped newline => discard both (line continuation)
				if r2 == '\n' || r2 == '\r' || r2 == '\f' {
					i += size2
					if r2 == '\r' && i < len(sel) && sel[i] == '\n' {
						i++
					}
					continue
				}

				current += string('\\')
				current += string(r2)
				i += size2

				// Numeric escape: consume up to 6 hex digits
				for count := 0; count < 5 && i < len(sel); count++ {
					rh, sh := utf8.DecodeRuneInString(sel[i:])
					if !isHexDigit(rh) {
						break
					}
					current += string(rh)
					i += sh
				}

				// Optional whitespace after numeric escape
				if i < len(sel) {
					rw, sw := utf8.DecodeRuneInString(sel[i:])
					if isCSSSpace(rw) {
						current += string(rw)
						i += sw
					}
				}
			} else {
				current += string('\\')
			}
			continue
		}

		if quote == 0 {
			switch r {
			case '"', '\'':
				quote = byte(r)

			case '[', '(':
				depth++

			case ']', ')':
				if depth > 0 {
					depth--
				}

			default:
				if depth == 0 {
					if r == '>' || r == '+' || r == '~' {
						if len(current) == 0 {
							return nil, SelectorError{"invalid combinator placement"}
						}
						if err := flush(); err != nil {
							return nil, err
						}
						combinator = string(r)
						i += size
						continue
					}

					if isCSSSpace(r) {
						// If whitespace is followed by an explicit combinator, ignore it
						if next := nextNonSpace(i + size); next == '>' || next == '+' || next == '~' {
							i += size
							continue
						}

						// Otherwise, whitespace acts as descendant combinator
						if len(current) != 0 {
							if err := flush(); err != nil {
								return nil, err
							}
							combinator = ""
						}
						i += size
						continue
					}
				}
			}
		} else if r == rune(quote) {
			quote = 0
		}

		current += string(r)
		i += size
	}

	if quote != 0 || depth != 0 {
		return nil, SelectorError{"unclosed bracket or quote"}
	}

	if len(current) != 0 {
		if err := flush(); err != nil {
			return nil, err
		}
	} else if combinator != "" {
		return nil, SelectorError{"dangling combinator"}
	}

	return chain, nil
}

// parseCompound parses a single compound selector.
//
// A compound selector is the portion of a CSS selector between combinators,
// for example:
//
//	div#main.container[attr=value]:first-child
//
// This function parses (in any order):
//   - an optional type selector (tag name or '*')
//   - zero or more ID selectors (#id)
//   - zero or more class selectors (.class)
//   - zero or more attribute selectors ([attr], [attr=value], etc.)
//   - zero or more pseudo-classes (:first-child, :not(...), etc.)
//
// It enforces strict CSS validity rules and returns a SelectorError on:
//   - missing identifiers (e.g. '#', '.', ':')
//   - multiple type selectors in the same compound ('div span')
//   - unclosed attribute selectors or pseudo arguments
//   - invalid characters in the selector
//
// The returned compoundSelector contains normalized (lowercased) tag and
// pseudo names. Semantic validation of pseudo-classes (e.g. nth-child syntax)
// is performed later during matching, not here.
func parseCompound(part string) (compoundSelector, error) {
	comp := compoundSelector{}
	i := 0

	for i < len(part) {
		ch := part[i]

		switch ch {
		case '#':
			name, next := readIdent(part, i+1)
			if name == "" {
				return comp, SelectorError{"missing id name"}
			}
			comp.id = name
			i = next

		case '.':
			name, next := readIdent(part, i+1)
			if name == "" {
				return comp, SelectorError{"missing class name"}
			}
			comp.classes = append(comp.classes, name)
			i = next

		case '[':
			end := i + 1
			depth := 1
			quote := byte(0)

			for end < len(part) && depth > 0 {
				if quote == 0 {
					switch part[end] {
					case '"', '\'':
						quote = part[end]
					case '[':
						depth++
					case ']':
						depth--
					}
				} else if part[end] == quote {
					quote = 0
				}
				end++
			}

			if depth != 0 || quote != 0 {
				return comp, SelectorError{"unclosed attribute selector"}
			}

			attrContent := strings.TrimSpace(part[i+1 : end-1])
			attr, err := parseAttr(attrContent)
			if err != nil {
				return comp, err
			}
			comp.attrs = append(comp.attrs, attr)
			i = end

		case ':':
			name, next := readIdent(part, i+1)
			if name == "" {
				return comp, SelectorError{"missing pseudo name"}
			}
			i = next

			arg := ""
			if i < len(part) && part[i] == '(' {
				depth := 1
				j := i + 1
				quote := byte(0)

				for j < len(part) && depth > 0 {
					if quote == 0 {
						switch part[j] {
						case '"', '\'':
							quote = part[j]
						case '(':
							depth++
						case ')':
							depth--
						}
					} else if part[j] == quote {
						quote = 0
					}
					j++
				}

				if depth != 0 || quote != 0 {
					return comp, SelectorError{"unclosed pseudo argument"}
				}

				arg = strings.TrimSpace(part[i+1 : j-1])
				i = j
			}

			lname := strings.ToLower(name)
			if lname == "contains" && strings.TrimSpace(arg) == "" {
				return comp, SelectorError{"contains() requires an argument"}
			}

			comp.pseudos = append(comp.pseudos, pseudoSelector{
				name: lname,
				arg:  arg,
			})

		default:
			if comp.tag != "" {
				return comp, SelectorError{"multiple type selectors"}
			}

			tag, next := readIdent(part, i)
			if tag == "" {
				return comp, SelectorError{"invalid character in selector"}
			}

			comp.tag = strings.ToLower(tag)
			i = next
		}
	}

	if comp.tag == "" {
		comp.tag = "*"
	}

	return comp, nil
}

// parseAttr parses the contents of an attribute selector.
//
// Examples of valid inputs:
//
//	"attr"
//	"attr=value"
//	"attr~=value"
//	"attr|=value"
//	"attr^=value"
//	"attr$=value"
//	"attr*=value"
//
// The returned attrSelector contains:
//   - name  → attribute name (lowercased)
//   - op    → operator ("exists", "=", "~=", "|=", "^=", "$=", "*=")
//   - value → unquoted attribute value (if applicable)
//
// This function performs syntactic validation only.
// Semantic validation (matching behavior) is handled by matchAttr.
func parseAttr(content string) (attrSelector, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return attrSelector{}, SelectorError{"empty attribute selector"}
	}

	ops := []string{"~=", "|=", "^=", "$=", "*=", "="}

	for _, op := range ops {
		if idx := strings.Index(content, op); idx != -1 {
			name := strings.TrimSpace(content[:idx])
			val := strings.TrimSpace(content[idx+len(op):])

			if name == "" {
				return attrSelector{}, SelectorError{"missing attribute name"}
			}
			if val == "" {
				return attrSelector{}, SelectorError{"missing attribute value"}
			}

			// Strip surrounding quotes from the value, if present
			val = strings.Trim(val, "\"'")

			return attrSelector{
				name:  strings.ToLower(name),
				op:    op,
				value: val,
			}, nil
		}
	}

	// [attr] → existence selector
	name := strings.TrimSpace(content)
	if name == "" {
		return attrSelector{}, SelectorError{"missing attribute name"}
	}

	return attrSelector{
		name: strings.ToLower(name),
		op:   "exists",
	}, nil
}

// Query returns all descendant nodes that satisfy the provided CSS selector.
func (root *Node) Query(selector string) []*Node {
	if root == nil {
		return nil
	}

	chains, err := parseSelector(selector)
	if err != nil {
		return nil
	}

	var out []*Node

	var walk func(n *Node)
	walk = func(n *Node) {
		for _, child := range n.Children {
			for _, chain := range chains {
				ok, err := matchChain(child, chain)
				if err == nil && ok {
					out = append(out, child)
					break
				}
			}
			walk(child)
		}

		if n.Template != nil {
			for _, chain := range chains {
				ok, err := matchChain(n.Template, chain)
				if err == nil && ok {
					out = append(out, n.Template)
					break
				}
			}
			walk(n.Template)
		}
	}

	walk(root)
	return out
}

// QueryFirst returns the first descendant node that satisfies the selector.
func (root *Node) QueryFirst(selector string) *Node {
	if root == nil {
		return nil
	}

	chains, err := parseSelector(selector)
	if err != nil {
		return nil
	}

	var match *Node
	var walk func(n *Node) bool
	walk = func(n *Node) bool {
		for _, child := range n.Children {
			for _, chain := range chains {
				ok, err := matchChain(child, chain)
				if err == nil && ok {
					match = child
					return true
				}
			}
			if walk(child) {
				return true
			}
		}

		if n.Template != nil {
			for _, chain := range chains {
				ok, err := matchChain(n.Template, chain)
				if err == nil && ok {
					match = n.Template
					return true
				}
			}
			if walk(n.Template) {
				return true
			}
		}
		return false
	}

	walk(root)
	return match
}

// Matches reports whether the node matches the selector.
func Matches(node *Node, selector string) bool {
	if node == nil {
		return false
	}

	if strings.TrimSpace(selector) == "" {
		panic(SelectorError{"empty selector"})
	}

	chains, err := parseSelector(selector)
	if err != nil {
		panic(err)
	}

	for _, chain := range chains {
		ok, err := matchChain(node, chain)
		if err != nil {
			panic(err)
		}
		if ok {
			return true
		}
	}
	return false
}
