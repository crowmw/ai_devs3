package html

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

// Processor handles HTML document processing
type Processor struct {
	baseURL string
}

// NewProcessor creates a new HTML processor
func NewProcessor(baseURL string) *Processor {
	return &Processor{
		baseURL: baseURL,
	}
}

// ParseHTML parses HTML content into a document
func (p *Processor) ParseHTML(content []byte) (*html.Node, error) {
	doc, err := htmlquery.Parse(bytes.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("error parsing HTML: %w", err)
	}
	return doc, nil
}

// ConvertToPlainText converts HTML to plain text by removing all HTML tags
func (p *Processor) ConvertToPlainText(doc *html.Node) string {
	htmlStr := htmlquery.OutputHTML(doc, true)
	re := regexp.MustCompile("<[^>]*>")
	plainText := re.ReplaceAllString(htmlStr, "")
	return strings.TrimSpace(plainText)
}

// GetElementsByTag finds all elements with the given tag name
func (p *Processor) GetElementsByTag(doc *html.Node, tag string) []*html.Node {
	return htmlquery.Find(doc, "//"+tag)
}

// GetElementAttribute gets an attribute value from a node
func (p *Processor) GetElementAttribute(node *html.Node, attr string) string {
	return htmlquery.SelectAttr(node, attr)
}

// ReplaceNode replaces a node with a new text node
func (p *Processor) ReplaceNode(node *html.Node, text string) {
	descriptionNode := &html.Node{
		Type: html.TextNode,
		Data: text,
	}

	parent := node.Parent
	if parent != nil {
		parent.InsertBefore(descriptionNode, node)
		parent.RemoveChild(node)
	}
}

// GetFullURL returns the full URL for a relative path
func (p *Processor) GetFullURL(path string) string {
	if strings.HasPrefix(path, "http") {
		return path
	}
	return p.baseURL + "/dane/" + path
}
