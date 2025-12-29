package parser_test

import (
	"fmt"
	"strings"

	"github.com/padraicbc/h5p/dom"
	"github.com/padraicbc/h5p/parser"
)

// Example demonstrating how to parse HTML and run CSS selectors against the document.
func ExampleParse_selectors() {
	const fixture = `
		<div id="first" class="foo bar">
			<p class="text">Paragraph <em>one</em></p>
			<p class="text highlight">Paragraph <em>two</em></p>
		</div>
		<div id="second" class="foo">
			<p class="text">Paragraph <span>three</span></p>
			<p class="text">Paragraph <span class="highlight">four</span></p>
		</div>
		<ul id="list">
			<li class="item first">One</li>
			<li class="item">Two</li>
			<li class="item last">Three</li>
		</ul>
		<div id="attributes">
			<a data-test="alpha">Link 1</a>
			<a data-test="beta">Link 2</a>
			<a data-test="gamma delta">Link 3</a>
		</div>
		<div id="nested">
			<div class="level1">
				<div class="level2">
					<span class="deep">Deep span</span>
				</div>
			</div>
		</div>
		<div id="siblings">
			<span class="sib first">A</span>
			<span class="sib">B</span>
			<span class="sib last">C</span>
		</div>
		<form id="form">
			<input type="checkbox" name="check" checked />
		</form>
	`

	doc, err := parser.Parse(fixture)
	if err != nil {
		panic(err)
	}

	selectors := []string{
		"#first .text",
		"#second > p",
		"ul#list > li.first",
		"ul#list > li + li",
		"ul#list > li ~ li",
		"#attributes a[data-test]",
		`#attributes a[data-test~="delta"]`,
		"#nested .level1 .level2 > span",
		"#siblings span.sib:last-child",
		`form#form input[type="checkbox"]`,
	}

	for _, sel := range selectors {
		nodes, _ := doc.Root.Query(sel)
		fmt.Printf("%s => %s\n", sel, summarize(nodes))
	}

	// Output:
	// #first .text => p(Paragraph one), p(Paragraph two)
	// #second > p => p(Paragraph three), p(Paragraph four)
	// ul#list > li.first => li(One)
	// ul#list > li + li => li(Two)
	// ul#list > li ~ li => li(Two), li(Three)
	// #attributes a[data-test] => a(Link 1), a(Link 2), a(Link 3)
	// #attributes a[data-test~="delta"] => a(Link 3)
	// #nested .level1 .level2 > span => span(Deep span)
	// #siblings span.sib:last-child => span(C)
	// form#form input[type="checkbox"] => input
}

func summarize(nodes []*dom.Node) string {
	parts := make([]string, 0, len(nodes))
	for _, n := range nodes {
		text := strings.TrimSpace(n.ToText(" ", true))
		if text != "" {
			parts = append(parts, fmt.Sprintf("%s(%s)", n.Name, text))
		} else {
			parts = append(parts, n.Name)
		}
	}
	if len(parts) == 0 {
		return "(none)"
	}
	return strings.Join(parts, ", ")
}
