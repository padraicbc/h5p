package dom

import (
	"sort"
	"strings"
)

var (
	textReplacer       = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
	attrDoubleReplacer = strings.NewReplacer("&", "&amp;", "\"", "&quot;")
	attrSingleReplacer = strings.NewReplacer("&", "&amp;", "'", "&#39;")
	voidElements       = map[string]struct{}{
		"area":   {},
		"base":   {},
		"br":     {},
		"col":    {},
		"embed":  {},
		"hr":     {},
		"img":    {},
		"input":  {},
		"link":   {},
		"meta":   {},
		"param":  {},
		"source": {},
		"track":  {},
		"wbr":    {},
	}
)

// ToHTML serializes a node to HTML.
func ToHTML(node *Node, pretty bool, indentSize int) string {
	if node == nil {
		return ""
	}

	if isDocument(node.Name) {
		if len(node.Children) == 0 {
			return ""
		}

		if pretty {
			parts := make([]string, 0, len(node.Children))
			for _, child := range node.Children {
				if child == nil {
					continue
				}
				html := nodeToHTML(child, 0, indentSize, pretty)
				if html != "" {
					parts = append(parts, html)
				}
			}
			return strings.Join(parts, "\n")
		}

		var b strings.Builder
		for _, child := range node.Children {
			if child == nil {
				continue
			}
			b.WriteString(nodeToHTML(child, 0, indentSize, pretty))
		}
		return b.String()
	}

	return nodeToHTML(node, 0, indentSize, pretty)
}

// nodeToHTML recursively serializes a node and its children, applying
// indentation when pretty-printing is requested.
func nodeToHTML(node *Node, indent, indentSize int, pretty bool) string {
	if node == nil {
		return ""
	}

	prefix := ""
	newline := ""
	if pretty {
		if indent > 0 && indentSize > 0 {
			prefix = strings.Repeat(" ", indent*indentSize)
		}
		newline = "\n"
	}

	name := node.Name

	switch name {
	case "#text":
		text := node.Data
		if pretty {
			text = strings.TrimSpace(text)
			if text == "" {
				return ""
			}
			return prefix + escapeText(text)
		}
		if text == "" {
			return ""
		}
		return escapeText(text)

	case "#comment":
		return prefix + "<!--" + node.Data + "-->"

	case "!doctype":
		return prefix + "<!DOCTYPE html>"

	case "#document-fragment":
		if len(node.Children) == 0 {
			return ""
		}
		if pretty {
			parts := make([]string, 0, len(node.Children))
			for _, child := range node.Children {
				if child == nil {
					continue
				}
				html := nodeToHTML(child, indent, indentSize, pretty)
				if html != "" {
					parts = append(parts, html)
				}
			}
			return strings.Join(parts, newline)
		}
		var b strings.Builder
		for _, child := range node.Children {
			if child == nil {
				continue
			}
			b.WriteString(nodeToHTML(child, indent, indentSize, pretty))
		}
		return b.String()
	}

	openTag := serializeStartTag(name, node.Attrs)

	if isVoidElement(name) {
		return prefix + openTag
	}

	children := node.Children
	if len(children) == 0 {
		return prefix + openTag + serializeEndTag(name)
	}

	allText := true
	for _, child := range children {
		if child == nil {
			continue
		}
		if child.Name != "#text" {
			allText = false
			break
		}
	}

	if allText && pretty {
		return prefix + openTag + escapeText(concatText(children)) + serializeEndTag(name)
	}

	parts := make([]string, 0, len(children)+2)
	parts = append(parts, prefix+openTag)

	for _, child := range children {
		if child == nil {
			continue
		}
		html := nodeToHTML(child, indent+1, indentSize, pretty)
		if html != "" {
			parts = append(parts, html)
		}
	}
	parts = append(parts, prefix+serializeEndTag(name))

	if pretty {
		return strings.Join(parts, newline)
	}

	var b strings.Builder
	total := 0
	for _, p := range parts {
		total += len(p)
	}
	b.Grow(total)
	for _, p := range parts {
		b.WriteString(p)
	}
	return b.String()

}

// escapeText applies HTML escaping to text node contents.
func escapeText(s string) string {
	if s == "" {
		return ""
	}
	return textReplacer.Replace(s)
}

// chooseAttrQuote selects which quote character to use around an attribute
// value, preferring a single quote when it avoids escaping.
func chooseAttrQuote(value string) byte {
	if strings.Contains(value, "\"") && !strings.Contains(value, "'") {
		return '\''
	}
	return '"'
}

// escapeAttrValue escapes an attribute value according to the surrounding
// quote character.
func escapeAttrValue(value string, quote byte) string {
	if value == "" {
		return ""
	}
	if quote == '"' {
		return attrDoubleReplacer.Replace(value)
	}
	return attrSingleReplacer.Replace(value)
}

// canUnquoteAttrValue reports whether an attribute value can be emitted without
// quotes under HTML serialization rules.
func canUnquoteAttrValue(value string) bool {
	if value == "" {
		return false
	}
	for i := 0; i < len(value); i++ {
		ch := value[i]
		switch ch {
		case '>', '"', '\'', '=':
			return false
		case ' ', '\t', '\n', '\f', '\r':
			return false
		}
	}
	return true
}

// serializeStartTag builds an opening tag string, sorting attributes for
// determinism and escaping as needed.
func serializeStartTag(name string, attrs map[string]*string) string {
	var b strings.Builder
	b.WriteByte('<')
	b.WriteString(name)

	if len(attrs) > 0 {
		keys := make([]string, 0, len(attrs))
		for k := range attrs {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			b.WriteByte(' ')
			valPtr := attrs[key]
			if valPtr == nil || *valPtr == "" {
				b.WriteString(key)
				continue
			}
			value := *valPtr
			if canUnquoteAttrValue(value) {
				b.WriteString(key)
				b.WriteByte('=')
				b.WriteString(strings.ReplaceAll(value, "&", "&amp;"))
				continue
			}
			quote := chooseAttrQuote(value)
			b.WriteString(key)
			b.WriteByte('=')
			b.WriteByte(quote)
			b.WriteString(escapeAttrValue(value, quote))
			b.WriteByte(quote)
		}
	}

	b.WriteByte('>')
	return b.String()
}

// serializeEndTag returns the closing tag for an element name.
func serializeEndTag(name string) string {
	return "</" + name + ">"
}

// concatText joins the data fields of text children without additional
// separators.
func concatText(children []*Node) string {
	var b strings.Builder
	for _, child := range children {
		if child == nil {
			continue
		}
		b.WriteString(child.Data)
	}
	return b.String()
}

// isVoidElement reports whether an element name is an HTML void element.
func isVoidElement(name string) bool {
	if name == "" {
		return false
	}
	_, ok := voidElements[strings.ToLower(name)]
	return ok
}

// isDocument reports whether a name represents a document root node.
func isDocument(name string) bool {
	return name == "#document" || name == "document"
}
