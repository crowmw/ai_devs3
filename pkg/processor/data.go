package processor

import (
	"fmt"
	"os"
	"strings"

	"github.com/antchfx/htmlquery"
)

func ReadMarkdownFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error reading markdown file: %w", err)
	}
	return string(content), nil
}

// ReadLinesFromTextFile processes raw data into filtered lines
func ReadLinesFromTextFile(data []byte) []string {
	lines := strings.Split(string(data), "\n")
	filteredLines := []string{}
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			filteredLines = append(filteredLines, line)
		}
	}
	return filteredLines
}

func ReadLinesFromTextFileAsString(data []byte) string {
	return strings.Join(ReadLinesFromTextFile(data), "\n")
}

// GetHTMLString converts HTML byte data to string
func GetHTMLString(data []byte) string {
	return string(data)
}

// ExtractTextFromHTML extracts text content from HTML using XPath
// xpathQuery - XPath query to find the element (e.g. "//p[@id='human-question']")
func ExtractTextFromHTML(htmlString string, xpathQuery string) string {
	doc, err := htmlquery.Parse(strings.NewReader(htmlString))
	if err != nil {
		return ""
	}

	// Find the element using XPath
	node := htmlquery.FindOne(doc, xpathQuery)
	if node == nil {
		return ""
	}

	// Get text content
	text := htmlquery.InnerText(node)

	return text
}
