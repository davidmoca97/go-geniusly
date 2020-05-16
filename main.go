package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"golang.org/x/net/html"
)

func main() {

	var URL, output, help string
	// Flags provided by the user through the command line
	flag.StringVar(&URL, "url", "", "URL of the lyrics of the song on genius.com")
	flag.StringVar(&URL, "u", "", "URL of the lyrics of the song on genius.com")
	flag.StringVar(&output, "output", "", "Path of the file where the lyrics will be saved")
	flag.StringVar(&output, "o", "", "Path of the file where the lyrics will be saved")
	flag.StringVar(&help, "help", "", "Help! 😰")

	flag.Parse()

	if URL == "" {
		handleErr("The URL of the song was not provided")
	}
	if _, err := url.ParseRequestURI(URL); err != nil {
		handleErr("The parameter provided is not a URL")
	}

	// // To compute the time that the program took to do all
	// startTime := time.Now()

	// Make the GET request
	request, err := http.Get(URL)
	if err != nil {
		handleErr("Error in the request: \n%v", err)
	}
	defer request.Body.Close()

	// Read and save the body of the response after being executed
	body, _ := ioutil.ReadAll(request.Body)
	if request.StatusCode != http.StatusOK {
		handleErr("The page does not exist")
	}

	// Parse the HTML of the response to an HTML node
	HTML, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		handleErr("Error parsing the HTML of the genius page: \n%v", err)
	}

	// These are the nodes we're looking for
	// << This node has the lyrics inside a <p> tag >>
	NodesWeLookFor := []html.Node{
		{ // Node that contains the lyrics in V1
			Type: html.ElementNode,
			Data: "div",
			Attr: []html.Attribute{
				{
					Key: "class",
					Val: "lyrics",
				},
			},
		},
		{ // Node that contains the lyrics in V2
			Type: html.ElementNode,
			Data: "div",
			Attr: []html.Attribute{
				{
					Key: "class",
					Val: "SongPage__Section-sc-19xhmoi-3 cXvCRB",
				},
			},
		},
	}

	var lyricsDiv *html.Node
	for idx, nodeWeWant := range NodesWeLookFor {
		// Get the node that has the lyrics
		lyricsDiv, err = LookForANode(HTML, &nodeWeWant)
		if err != nil && idx == len(NodesWeLookFor) {
			handleErr("Could not get the Lyrics node: \n%v", err)
		} else if err == nil {
			break
		}
	}

	// Get the lyrics
	lyricsString, err := getLyrics(lyricsDiv)
	if err != nil {
		handleErr("Could not retrieve the lyrics from the HTML: \n%v", err)
	}

	if output != "" {
		// Write the lyrics on a txt file
		err = ioutil.WriteFile(path.Clean(output), []byte(lyricsString), 0777)
		if err != nil {
			handleErr("Could not write in file: \n%v", err)
		}
	} else {
		fmt.Println(lyricsString)
	}

	// // Finish
	// duration := time.Now().Sub(startTime)
	// // How long did the program took?
	// fmt.Fprintf(os.Stdout, "Program finished after %v", duration)
}

// LookForANode : Looks for a node within a node (recursive)
func LookForANode(n, o *html.Node) (*html.Node, error) {
	// Node to return
	var b *html.Node
	var seekFunc func(*html.Node, bool) func()

	// Func that looks for the node
	seekFunc = func(n *html.Node, stopAtFirstMatch bool) func() {
		var _seekFunc func()
		var stop bool
		_seekFunc = func() {
			// If the current node equals with both data and type
			if n.Data == o.Data && n.Type == o.Type {
				var numAttrsInNode int
				// Loop over the attr and find if they are equal with the wanted attrs
				for _, attr := range o.Attr {
					if hasAttr(n, attr.Key, attr.Val, true) {
						numAttrsInNode++
					}
				}
				if numAttrsInNode == len(o.Attr) {
					b = n
					if stopAtFirstMatch {
						stop = true
						return
					}
				}
			}

			// Loop over the siblings
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				n = c
				_seekFunc()
				if stop {
					return
				}
			}
		}
		return _seekFunc
	}
	seekFunc(n, true)()
	if b != nil {
		return b, nil
	}
	return nil, fmt.Errorf("Missing %v node in the node tree", o.Data)
}

// Get all the text within the provided node
func getLyrics(doc *html.Node) (string, error) {
	var actual *html.Node
	var paragraph *html.Node

	actual = doc.FirstChild

	// Look for the <p> tag inside the div
	for actual != nil {
		if (actual.Type == html.ElementNode && actual.Data == "p") ||
			(actual.Data == "div" &&
				hasAttr(actual, "class", "SongPageGrid-sc-1vi6xda-0 DGVcp Lyrics__Root-sc-1ynbvzw-0 jvlKWy", true)) {
			paragraph = actual
			break
		}
		actual = actual.NextSibling
	}
	if actual == nil {
		handleErr("Lyrics node was not found")
	}

	// Actual is now the first child of the paragraph
	actual = paragraph.FirstChild

	// Getting the Lyrics
	lyrics := getTextOfNodes(actual)

	return (lyrics), nil
}

// Get all the text within a Node
func getTextOfNodes(n *html.Node) string {
	var str string

	// Do the loop while the node is different from nil
	for ; n != nil; n = n.NextSibling {

		// If the node is not a text or an html element
		if n.Type != html.ElementNode && n.Type != html.TextNode {
			continue
		}

		// If the node is pure text
		if n.Type == html.TextNode {
			str = str + strings.Replace(n.Data, "\n", "", -1)
			continue
		}

		// If the node is an html tag, different from <br>
		if n.Type == html.ElementNode && n.Data != "br" {
			str = str + getTextOfNodes(n.FirstChild)
		}

		// Add line break
		if n.Data == "br" {
			str = str + "\n"
		}
	}

	return str
}

// Renders the html node and returns a buffer with the data
func renderNode(n *html.Node) bytes.Buffer {
	var buf bytes.Buffer
	w := io.Writer(&buf)
	html.Render(w, n)
	return buf
}

// Looks for an attribute of a node
// Returns true if the node has the attr
// CheckForVal is a flag to wether or not check the value of an attr
func hasAttr(n *html.Node, key, val string, checkForVal bool) bool {
	for _, a := range n.Attr {
		if a.Key == key {
			if (checkForVal && a.Val == val) || !checkForVal {
				return true
			}
		}
	}
	return false
}

// Search for an attribute and a value of a node
// Returns true if the node has same attr and value provided in the params
func hasAttrWithVal(n *html.Node, attr, val string) bool {
	for _, a := range n.Attr {
		if a.Key == attr && a.Val == val {
			return true
		}
	}
	return false
}

func handleErr(text string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, text+"\n", args...)
	os.Exit(1)
}

// Check if the flag was provided through the command line
func wasFlagProvided(flagName string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == flagName {
			found = true
		}
	})
	return found
}
