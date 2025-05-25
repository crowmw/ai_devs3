package processor

import (
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
)

// GetHTMLString converts HTML byte data to string
func GetHTMLString(data []byte) string {
	return string(data)
}

// HTMLToMarkdown converts HTML content to Markdown
func HTMLToMarkdown(htmlContent string) string {
	// First sanitize HTML to remove potentially harmful content
	p := bluemonday.StrictPolicy()
	sanitizedHTML := p.Sanitize(htmlContent)

	// Convert sanitized HTML to Markdown
	md := blackfriday.Run([]byte(sanitizedHTML), blackfriday.WithExtensions(blackfriday.CommonExtensions))
	return string(md)
}
