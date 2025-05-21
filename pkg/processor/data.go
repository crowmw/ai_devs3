package processor

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
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

// ReadAllTxtFilesFromDirectory reads all .txt files from the specified directory
// and combines their content into a single string
func ReadAllTxtFilesFromDirectory(dirPath string) (string, error) {
	// Read all files from directory
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return "", fmt.Errorf("error reading directory: %w", err)
	}

	var allContent strings.Builder

	// Process each file
	for _, file := range files {
		// Skip directories and non-txt files
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".txt") {
			continue
		}

		// Read file content
		filePath := filepath.Join(dirPath, file.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("error reading file %s: %w", file.Name(), err)
		}

		// Process content using existing function
		processedContent := ReadLinesFromTextFileAsString(content)

		// Add to combined content with a newline separator
		if allContent.Len() > 0 {
			allContent.WriteString("\n")
		}
		allContent.WriteString(processedContent)
	}

	return allContent.String(), nil
}

func ReadImageToBase64(path string) (string, error) {
	// Read the image file
	imageBytes, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("error reading image file: %v", err)
	}

	// Convert to base64
	base64String := base64.StdEncoding.EncodeToString(imageBytes)
	return base64String, nil
}

// ReadAllImagesFromDirectory reads all image files from the specified directory
// and converts them to base64 strings
func ReadAllImagesFromDirectory(dirPath string) ([]string, error) {
	// Read all files from directory
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("error reading directory: %w", err)
	}

	var base64Images []string

	// Process each file
	for _, file := range files {
		// Skip directories and non-image files
		if file.IsDir() {
			continue
		}

		// Check if file is an image (you can add more extensions if needed)
		ext := strings.ToLower(filepath.Ext(file.Name()))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
			continue
		}

		// Read and convert image to base64
		filePath := filepath.Join(dirPath, file.Name())
		base64String, err := ReadImageToBase64(filePath)
		if err != nil {
			return nil, fmt.Errorf("error processing image %s: %w", file.Name(), err)
		}

		base64Images = append(base64Images, base64String)
	}

	return base64Images, nil
}
