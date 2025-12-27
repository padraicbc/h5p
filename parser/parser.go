package parser

import (
	"fmt"
	"io"
	"strings"

	"github.com/padraicbc/h5p/dom"
	"github.com/padraicbc/h5p/internal/encoding"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Option configures Parse behavior.
type Option func(*parseConfig)

type parseConfig struct {
	Debug           bool
	Encoding        string
	FragmentContext *FragmentContext
	IframeSrcdoc    bool
}

func WithDebug(debug bool) Option {
	return func(cfg *parseConfig) {
		cfg.Debug = debug
	}
}

func WithEncoding(encoding string) Option {
	return func(cfg *parseConfig) {
		cfg.Encoding = encoding
	}
}

func WithFragmentContext(ctx *FragmentContext) Option {
	return func(cfg *parseConfig) {
		cfg.FragmentContext = ctx
	}
}

func WithIframeSrcdoc(srcdoc bool) Option {
	return func(cfg *parseConfig) {
		cfg.IframeSrcdoc = srcdoc
	}
}

// HTP is the parsed document and helpers.
type HTP struct {
	Debug           bool
	Encoding        string
	Errors          []error
	FragmentContext *FragmentContext
	Root            *dom.Node
}

// FragmentContext represents the element that a fragment parse is rooted in.
type FragmentContext struct {
	TagName   string
	Namespace string
}

// Parse creates a HTP document from the given input. `html` may be a string, []byte, or io.Reader.
func Parse(input any, opts ...Option) (*HTP, error) {
	cfg := &parseConfig{}
	for _, opt := range opts {
		if opt != nil {
			opt(cfg)
		}
	}

	var htmlStr string
	encodingName := cfg.Encoding

	switch v := input.(type) {
	case io.Reader:
		res, err := io.ReadAll(v)
		if err != nil {
			return nil, err
		}
		decoded := encoding.DecodeHTML(res, cfg.Encoding)
		htmlStr = decoded.HTML
		if decoded.Encoding != "" {
			encodingName = decoded.Encoding
		}
	case []byte:
		res := encoding.DecodeHTML(v, cfg.Encoding)
		htmlStr = res.HTML
		if res.Encoding != "" {
			encodingName = res.Encoding
		}
	case string:
		htmlStr = v
	case nil:
		htmlStr = ""
	default:
		htmlStr = fmt.Sprint(v)
	}

	var (
		root *dom.Node
		errs []error
	)

	if cfg.FragmentContext != nil {
		ctxTag := strings.TrimSpace(cfg.FragmentContext.TagName)
		if ctxTag == "" {
			ctxTag = "div"
		}
		// Normalize the fragment context tag to lower-case to match HTML parsing expectations.
		ctxTag = strings.ToLower(ctxTag)
		ctx := &html.Node{
			Type:      html.ElementNode,
			Data:      ctxTag,
			Namespace: cfg.FragmentContext.Namespace,
			// DataAtom must be set for ParseFragment to attach children correctly.
			DataAtom: atom.Lookup([]byte(ctxTag)),
		}
		// Parse the fragment against the supplied context node.
		nodes, err := html.ParseFragment(strings.NewReader(htmlStr), ctx)
		if err != nil {
			errs = append(errs, err)
		}
		// Wrap the fragment nodes with the context element to match existing fragment behavior.
		root = dom.NewDocument()
		wrapper := &dom.Node{Name: ctx.Data, Namespace: ctx.Namespace}
		for _, node := range nodes {
			wrapper.AppendChild(convertHTMLNode(node))
		}
		root.AppendChild(wrapper)
	} else {
		// Full-document parsing uses the upstream html.Parse tree.
		doc, err := html.Parse(strings.NewReader(htmlStr))
		if err != nil {
			errs = append(errs, err)
		}
		root = convertHTMLNode(doc)
	}

	return &HTP{
		Debug:           cfg.Debug,
		Encoding:        encodingName,
		Errors:          errs,
		FragmentContext: cfg.FragmentContext,
		Root:            root,
	}, nil
}

// ToHTML serializes the document.
func (j *HTP) ToHTML(pretty bool, indentSize int) string {
	if j == nil {
		return ""
	}
	return dom.ToHTML(j.Root, pretty, indentSize)
}

// ToText concatenates text nodes.
func (j *HTP) ToText(separator string, strip bool) string {
	if j == nil || j.Root == nil {
		return ""
	}
	return j.Root.ToText(separator, strip)
}

// ToMarkdown returns a Markdown representation.
func (j *HTP) ToMarkdown() string {
	if j == nil || j.Root == nil {
		return ""
	}
	return j.Root.ToMarkdown()
}

// convertHTMLNode walks an upstream html.Node tree and maps it into the
// project dom.Node representation, preserving attributes, namespaces,
// and template content along the way.
func convertHTMLNode(node *html.Node) *dom.Node {
	if node == nil {
		return nil
	}
	switch node.Type {
	case html.DocumentNode:
		// Document nodes become the root "document" placeholder node.
		doc := dom.NewDocument()
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			doc.AppendChild(convertHTMLNode(child))
		}
		return doc
	case html.ElementNode:
		// Element nodes map directly to dom.Node elements with attributes.
		out := &dom.Node{Name: node.Data, Namespace: node.Namespace}
		if len(node.Attr) > 0 {
			out.Attrs = make(map[string]*string, len(node.Attr))
			for _, attr := range node.Attr {
				if attr.Val == "" {
					out.Attrs[attr.Key] = nil
				} else {
					val := attr.Val
					out.Attrs[attr.Key] = &val
				}
			}
		}
		if strings.EqualFold(node.Data, "template") {
			// Preserve template contents as a document fragment to match existing DOM behavior.
			fragment := &dom.Node{Name: "#document-fragment"}
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				fragment.AppendChild(convertHTMLNode(child))
			}
			out.Template = fragment
			return out
		}
		// Regular elements keep their children as-is.
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			out.AppendChild(convertHTMLNode(child))
		}
		return out
	case html.TextNode:
		// Text nodes use the "#text" sentinel name.
		return &dom.Node{Name: "#text", Data: node.Data}
	case html.CommentNode:
		// Comment nodes use the "#comment" sentinel name.
		return &dom.Node{Name: "#comment", Data: node.Data}
	case html.DoctypeNode:
		// Doctype nodes map to the "!doctype" sentinel.
		return &dom.Node{Name: "!doctype", Data: node.Data}
	default:
		// Unknown nodes are ignored.
		return nil
	}
}
