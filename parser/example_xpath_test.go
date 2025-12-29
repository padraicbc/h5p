package parser_test

import (
	"fmt"

	"github.com/padraicbc/h5p/parser"
)

// Example demonstrating how to parse HTML and run XPath queries against the document.
func ExampleParse_xpath() {
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

	xpaths := []string{
		`//*[@id="first"]//p`,
		`//*[@id="second"]/p`,
		`//ul[@id="list"]/li[@class="item first"]`,
		`//ul[@id="list"]/li[preceding-sibling::li][1]`,
		`//ul[@id="list"]/li[preceding-sibling::li]`,
		`//*[@id="attributes"]/a[@data-test]`,
		`//*[@id="attributes"]/a[contains(@data-test, "delta")]`,
		`//*[@id="nested"]//span`,
		`//*[@id="siblings"]/span[not(following-sibling::span)]`,
		`//form[@id="form"]//input[@type="checkbox"]`,
	}

	for _, xp := range xpaths {
		nodes, _ := doc.Root.QueryXPath(xp)
		fmt.Printf("%s => %s\n", xp, summarize(nodes))
	}

	// Output:
	// //*[@id="first"]//p => p(Paragraph one), p(Paragraph two)
	// //*[@id="second"]/p => p(Paragraph three), p(Paragraph four)
	// //ul[@id="list"]/li[@class="item first"] => li(One)
	// //ul[@id="list"]/li[preceding-sibling::li][1] => li(Two)
	// //ul[@id="list"]/li[preceding-sibling::li] => li(Two), li(Three)
	// //*[@id="attributes"]/a[@data-test] => a(Link 1), a(Link 2), a(Link 3)
	// //*[@id="attributes"]/a[contains(@data-test, "delta")] => a(Link 3)
	// //*[@id="nested"]//span => span(Deep span)
	// //*[@id="siblings"]/span[not(following-sibling::span)] => span(C)
	// //form[@id="form"]//input[@type="checkbox"] => input
}
