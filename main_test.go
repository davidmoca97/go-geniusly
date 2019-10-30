package main

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

const (
	test1 = `
		<div>
			<h5>Test</h5>
			<div class="im_here">
				<p>That was the <b>test</b></p>
			</div>
		</div>
	`
)

const (
	expected1 = `<div class="im_here"><p>That was the <b>test</b></p></div>`
)

func TestLookForANode(t *testing.T) {
	n, _ := html.Parse(strings.NewReader(test1))
	NodeWeLookFor := html.Node{
		Type: html.ElementNode,
		Data: "div",
		Attr: []html.Attribute{
			{
				Key: "class",
				Val: "im_here",
			},
		},
	}
	result, err := LookForANode(n, &NodeWeLookFor)
	if err != nil {
		t.Fatalf("Something wrong happened during the test: <<%v>>", err)
	}
	gotNode := renderNode(result)
	got := strings.Replace(gotNode.String(), "\n", "", -1)
	got = strings.Replace(got, "\t", "", -1)
	got = strings.TrimSpace(got)

	if got != expected1 {
		t.Errorf("The result does not match with the expected result. Got:\n %v\n and expected: %v", got, expected1)
	}
}
