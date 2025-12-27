package dom

import (
	"strconv"
	"strings"
)

// Node represents a DOM node. This mirrors the Python SimpleDomNode structure.
type Node struct {
	Name      string
	Namespace string
	Attrs     map[string]*string
	Children  []*Node
	Parent    *Node
	Data      string
	Template  *Node // for <template> content
}

// NewDocument returns a minimal document root node.
func NewDocument() *Node {
	return &Node{ // document node placeholder
		Name:      "document",
		Namespace: "",
		Attrs:     nil,
		Children:  make([]*Node, 0),
	}
}

// AppendChild attaches a child node.
func (n *Node) AppendChild(child *Node) {
	if n == nil || child == nil {
		return
	}
	n.Children = append(n.Children, child)
	child.Parent = n
}

func (n *Node) ToMarkdown() string {
	if n == nil {
		return ""
	}

	var b strings.Builder
	renderMarkdown(n, &b, "")
	return strings.TrimRight(b.String(), "\n")
}

func renderMarkdown(n *Node, b *strings.Builder, indent string) {
	if n == nil {
		return
	}

	switch strings.ToLower(n.Name) {
	case "#text":
		b.WriteString(n.Data)
		return
	case "#comment":
		return
	case "#document", "#document-fragment", "document":
		for _, c := range n.Children {
			renderMarkdown(c, b, indent)
		}
		return
	case "br":
		b.WriteString("  \n")
		return
	case "p", "div":
		start := b.Len()
		for _, c := range n.Children {
			renderMarkdown(c, b, indent)
		}
		if b.Len() > start {
			b.WriteString("\n\n")
		}
		return
	case "strong", "b":
		b.WriteString("**")
		for _, c := range n.Children {
			renderMarkdown(c, b, indent)
		}
		b.WriteString("**")
		return
	case "em", "i":
		b.WriteString("*")
		for _, c := range n.Children {
			renderMarkdown(c, b, indent)
		}
		b.WriteString("*")
		return
	case "code":
		b.WriteString("`")
		for _, c := range n.Children {
			renderMarkdown(c, b, indent)
		}
		b.WriteString("`")
		return
	case "a":
		text := childTextContent(n)
		href := n.Attr("href")
		if href == "" {
			b.WriteString(text)
			return
		}
		if text == "" {
			text = href
		}
		b.WriteString("[")
		b.WriteString(text)
		b.WriteString("](")
		b.WriteString(href)
		b.WriteString(")")
		return
	case "ul", "ol":
		ordered := strings.ToLower(n.Name) == "ol"
		for _, c := range n.Children {
			if c == nil || strings.ToLower(c.Name) != "li" {
				continue
			}
			renderListItem(c, b, indent, ordered)
		}
		b.WriteString("\n")
		return
	case "li":
		renderListItem(n, b, indent, false)
		return
	case "h1", "h2", "h3", "h4", "h5", "h6":
		level := int(n.Name[1] - '0')
		if level < 1 {
			level = 1
		}
		if level > 6 {
			level = 6
		}
		b.WriteString(strings.Repeat("#", level))
		b.WriteByte(' ')
		for _, c := range n.Children {
			renderMarkdown(c, b, indent)
		}
		b.WriteString("\n\n")
		return
	}

	for _, c := range n.Children {
		renderMarkdown(c, b, indent)
	}
}

func renderListItem(n *Node, b *strings.Builder, indent string, ordered bool) {
	if n == nil {
		return
	}
	b.WriteString(indent)
	if ordered {
		num := listIndex(n)
		if num <= 0 {
			num = 1
		}
		b.WriteString(strconv.Itoa(num))
		b.WriteString(". ")
	} else {
		b.WriteString("- ")
	}
	for _, c := range n.Children {
		renderMarkdown(c, b, indent+"   ")
	}
	b.WriteByte('\n')
}

func listIndex(n *Node) int {
	if n == nil || n.Parent == nil {
		return 0
	}
	count := 0
	for _, c := range n.Parent.Children {
		if c != nil && strings.ToLower(c.Name) == "li" {
			count++
			if c == n {
				return count
			}
		}
	}
	return 0
}

func childTextContent(n *Node) string {
	if n == nil {
		return ""
	}
	var b strings.Builder
	var walk func(*Node)
	walk = func(node *Node) {
		if node == nil {
			return
		}
		if node.Name == "#text" {
			b.WriteString(node.Data)
			return
		}
		for _, c := range node.Children {
			walk(c)
		}
	}
	for _, c := range n.Children {
		walk(c)
	}
	return b.String()
}

// ToText returns concatenated text content with an optional separator and an
// option to trim whitespace.
func (n *Node) ToText(separator string, strip bool) string {
	if n == nil {
		return ""
	}
	var b strings.Builder

	appendText := func(text string) {
		if strip {
			text = strings.TrimSpace(text)
		}
		if text == "" {
			return
		}
		if b.Len() > 0 && separator != "" {
			b.WriteString(separator)
		}
		b.WriteString(text)
	}

	var walk func(*Node)
	walk = func(node *Node) {
		if node == nil {
			return
		}
		if node.Name == "#text" {
			appendText(node.Data)
			return
		}
		for _, child := range node.Children {
			walk(child)
		}
	}
	walk(n)
	return b.String()
}

// ToHTML serializes the node tree rooted at n using the public serializer.
func (n *Node) ToHTML(pretty bool, indentSize int) string {
	return ToHTML(n, pretty, indentSize)
}

// Text returns the concatenated text node data of the node's immediate
// children (but not deeper descendants).
func (n *Node) Text() string {
	if n == nil {
		return ""
	}
	if n.Name == "#text" {
		return n.Data
	}
	if len(n.Children) == 0 {
		return ""
	}
	var b strings.Builder
	for _, c := range n.Children {
		if c != nil && c.Name == "#text" {
			b.WriteString(c.Data)
		}
	}
	return b.String()
}

// AllText returns the concatenated text content for the node and all
// descendants.
func (n *Node) AllText() string {
	if n == nil || len(n.Children) == 0 {
		return ""
	}

	var out strings.Builder
	stack := []*Node{
		n,
	}
	for len(stack) > 0 {
		index := len(stack) - 1
		n = stack[index]
		stack = stack[:index]

		for _, c := range n.Children {
			if c.Name == "#text" {
				out.WriteString(c.Data)
			} else {
				stack = append(stack, c)
			}

		}

	}
	return out.String()
}

// Attr returns the attribute value for the given name or "" if missing.
func (n *Node) Attr(name string) string {
	if n == nil || n.Attrs == nil {
		return ""
	}
	if v, ok := n.Attrs[name]; ok {
		if v != nil {
			return *v
		}
		return ""
	}
	for k, v := range n.Attrs {
		if strings.EqualFold(k, name) && v != nil {
			return *v
		}
	}
	return ""
}
