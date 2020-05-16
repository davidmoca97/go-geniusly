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

func TestRenderNode(t *testing.T) {
	node := html.Node{
		Type: html.ElementNode,
		Data: "div",
		Attr: []html.Attribute{
			{
				Key: "class",
				Val: "parent",
			},
		},
	}
	expected := `<div class="parent"></div>`
	buffer := renderNode(&node)
	htmlNode := buffer.String()

	if htmlNode != expected {
		t.Errorf("The result does not match with the expected result. Got:\n %v\n and expected: %v", htmlNode, expected)
	}
}

func TestHasAttr(t *testing.T) {
	node := html.Node{
		Type: html.ElementNode,
		Data: "div",
		Attr: []html.Attribute{
			{
				Key: "class",
				Val: "parent",
			},
			{
				Key: "data-testid",
			},
		},
	}
	expected := []bool{true, true, false, true, true}
	parameters := []struct {
		key         string
		val         string
		checkForVal bool
	}{
		{key: "class", val: "parent", checkForVal: true},
		{key: "class", val: "", checkForVal: false},
		{key: "style", val: "border", checkForVal: true},
		{key: "data-testid", val: "", checkForVal: true},
		{key: "data-testid", val: "", checkForVal: false},
	}
	for i := 0; i < len(parameters); i++ {
		got := hasAttr(&node, parameters[i].key, parameters[i].val, parameters[i].checkForVal)
		if got != expected[i] {
			t.Errorf("Case #%v failed. The result does not match with the expected result. Got: %v and expected: %v", i, got, expected[i])
		}
	}
}
