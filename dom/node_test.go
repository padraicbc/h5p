package dom

import (
	"strings"
	"testing"
)

func TestAttrCaseInsensitiveAndNil(t *testing.T) {
	tests := []struct {
		name  string
		attrs map[string]*string
		key   string
		want  string
	}{
		{
			name:  "lookup is case-insensitive",
			attrs: map[string]*string{"ID": strPtr("main")},
			key:   "id",
			want:  "main",
		},
		{
			name:  "lookup preserves original key",
			attrs: map[string]*string{"ID": strPtr("main")},
			key:   "ID",
			want:  "main",
		},
		{
			name:  "nil attr returns empty",
			attrs: map[string]*string{"data": nil},
			key:   "data",
			want:  "",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			n := &Node{Attrs: tc.attrs}
			if got := n.Attr(tc.key); got != tc.want {
				t.Fatalf("Attr(%s) = %q, want %q", tc.key, got, tc.want)
			}
		})
	}
}

func TestToTextRespectsSeparatorAndStrip(t *testing.T) {
	root := &Node{Name: "div", Children: []*Node{
		{Name: "#text", Data: " hello "},
		{Name: "span", Children: []*Node{{Name: "#text", Data: " world "}}},
		{Name: "#text", Data: "  ! "},
	}}

	tests := []struct {
		name     string
		sep      string
		strip    bool
		wantText string
	}{
		{
			name:     "strip removes extra whitespace",
			sep:      " ",
			strip:    true,
			wantText: "hello world !",
		},
		{
			name:     "no strip keeps separator and spaces",
			sep:      "-",
			strip:    false,
			wantText: " hello - world -  ! ",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := root.ToText(tc.sep, tc.strip); got != tc.wantText {
				t.Fatalf("ToText(strip=%v) = %q, want %q", tc.strip, got, tc.wantText)
			}
		})
	}
}

func TestToMarkdownRendersCommonElements(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "markdown includes headings and lists"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

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
		})
	}
}

func TestToMarkdownHandlesNilReceiver(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "nil receiver returns empty"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var n *Node
			if got := n.ToMarkdown(); got != "" {
				t.Fatalf("nil receiver ToMarkdown = %q, want empty", got)
			}
		})
	}
}

func TestMatchesEmptySelectorPanics(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "empty selector panics"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			defer func() {
				if recover() == nil {
					t.Fatalf("expected panic on empty selector")
				}
			}()
			Matches(&Node{Name: "div"}, "   ")
		})
	}
}

func TestMatchChainErrorPropagation(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "unsupported operator returns error"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			node := &Node{Name: "div", Attrs: map[string]*string{"id": strPtr("x")}}
			chain := selectorChain{
				{compound: compoundSelector{attrs: []attrSelector{{name: "id", op: "!="}}}},
			}
			if ok, err := matchChain(node, chain); err == nil || ok {
				t.Fatalf("matchChain should bubble unsupported operator error, got ok=%v err=%v", ok, err)
			}
		})
	}
}

func TestMatchPseudoUnsupportedAndErrors(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "unsupported pseudo returns error",
			run: func(t *testing.T) {
				if _, err := matchPseudo(&Node{Name: "div"}, pseudoSelector{name: "hover"}); err == nil {
					t.Fatalf("unsupported pseudo should return error")
				}
			},
		},
		{
			name: "nth-child without arg errors",
			run: func(t *testing.T) {
				if _, err := matchPseudo(&Node{Name: "div"}, pseudoSelector{name: "nth-child"}); err == nil {
					t.Fatalf("nth-child without arg should error")
				}
			},
		},
		{
			name: "nth-of-type without arg errors",
			run: func(t *testing.T) {
				if _, err := matchPseudo(&Node{Name: "div"}, pseudoSelector{name: "nth-of-type"}); err == nil {
					t.Fatalf("nth-of-type without arg should error")
				}
			},
		},
		{
			name: "contains without arg errors",
			run: func(t *testing.T) {
				if _, err := matchPseudo(&Node{Name: "div"}, pseudoSelector{name: "contains"}); err == nil {
					t.Fatalf(":contains without arg should error")
				}
			},
		},
		{
			name: "not with invalid selector is ignored",
			run: func(t *testing.T) {
				if ok, err := matchPseudo(&Node{Name: "div"}, pseudoSelector{name: "not", arg: "["}); !ok || err != nil {
					t.Fatalf(":not with invalid selector should ignore and return true, got %v %v", ok, err)
				}
			},
		},
		{
			name: "root does not match non-document nodes",
			run: func(t *testing.T) {
				root := &Node{Name: "div"}
				root.AppendChild(&Node{Name: "span"})
				if ok, _ := matchPseudo(root, pseudoSelector{name: "root"}); ok {
					t.Fatalf(":root must not match non-root element")
				}
			},
		},
		{
			name: "root matches html under document",
			run: func(t *testing.T) {
				doc := &Node{Name: "document"}
				html := &Node{Name: "html"}
				doc.AppendChild(html)
				if ok, err := matchPseudo(html, pseudoSelector{name: "root"}); !ok || err != nil {
					t.Fatalf(":root should match html element under document, got %v %v", ok, err)
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.run(t)
		})
	}
}

func TestMatchesNilNodeReturnsFalse(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "nil node returns false"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if Matches(nil, "div") {
				t.Fatalf("Matches(nil) should be false")
			}
		})
	}
}

func TestSelectorErrorImplementsError(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "Error returns message"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := SelectorError{Msg: "boom"}
			if err.Error() != "boom" {
				t.Fatalf("SelectorError.Error = %q, want boom", err.Error())
			}
		})
	}
}

func TestAppendChildHandlesNil(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "append nil child is ignored",
			run: func(t *testing.T) {
				parent := &Node{Name: "div"}
				parent.AppendChild(nil)
				if len(parent.Children) != 0 {
					t.Fatalf("expected no children when appending nil")
				}
			},
		},
		{
			name: "append child to nil parent does not panic",
			run: func(t *testing.T) {
				var nilParent *Node
				child := &Node{Name: "span"}
				nilParent.AppendChild(child)
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.run(t)
		})
	}
}

func TestAttrMissingReturnsEmpty(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want string
	}{
		{name: "missing attr returns empty", key: "id", want: ""},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			n := &Node{Attrs: map[string]*string{"class": strPtr("foo")}}
			if got := n.Attr(tc.key); got != tc.want {
				t.Fatalf("missing attr should be empty, got %q", got)
			}
		})
	}
}

func TestPreviousElementSiblingEdges(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "nil node has nil previous sibling",
			run: func(t *testing.T) {
				if prev := previousElementSibling(nil); prev != nil {
					t.Fatalf("nil node previous sibling should be nil")
				}
			},
		},
		{
			name: "first child has nil previous sibling",
			run: func(t *testing.T) {
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
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.run(t)
		})
	}
}
