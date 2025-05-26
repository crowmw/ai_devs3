package processor

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/pkoukk/tiktoken-go"
)

// Headers represents a map of header levels to their content
type Headers map[string][]string

// Doc represents a text chunk with its metadata
type Doc struct {
	Text     string
	Metadata struct {
		Tokens  int
		Headers Headers
		URLs    []string
		Images  []string
	}
}

// TextSplitter handles splitting text into chunks with metadata
type TextSplitter struct {
	modelName     string
	specialTokens map[string]int
	tokenizer     *tiktoken.Tiktoken
	mu            sync.Mutex
}

// NewTextSplitter creates a new TextSplitter instance
func NewTextSplitter(modelName string) *TextSplitter {
	if modelName == "" {
		modelName = "gpt-4o"
	}

	return &TextSplitter{
		modelName: modelName,
		specialTokens: map[string]int{
			"<|im_start|>": 100264,
			"<|im_end|>":   100265,
			"<|im_sep|>":   100266,
		},
	}
}

// initializeTokenizer initializes the tokenizer if it hasn't been initialized yet
func (ts *TextSplitter) initializeTokenizer() error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.tokenizer != nil {
		return nil
	}

	// Map model names to their encoding names
	modelToEncoding := map[string]string{
		"gpt-4o":        "cl100k_base",
		"gpt-4":         "cl100k_base",
		"gpt-3.5":       "cl100k_base",
		"gpt-3.5-turbo": "cl100k_base",
	}

	encoding, ok := modelToEncoding[ts.modelName]
	if !ok {
		return fmt.Errorf("unsupported model: %s", ts.modelName)
	}

	tokenizer, err := tiktoken.GetEncoding(encoding)
	if err != nil {
		return fmt.Errorf("error getting encoding: %w", err)
	}

	ts.tokenizer = tokenizer
	return nil
}

// countTokens counts the number of tokens in the given text
func (ts *TextSplitter) countTokens(text string) int {
	if err := ts.initializeTokenizer(); err != nil {
		// Fallback to approximation if tokenizer initialization fails
		return len(text) / 4
	}

	// Add special tokens to the text
	formattedText := ts.formatForTokenization(text)

	// Count tokens
	tokens := ts.tokenizer.Encode(formattedText, nil, nil)
	return len(tokens)
}

// Split splits text into chunks with metadata
func (ts *TextSplitter) Split(text string, limit int) ([]Doc, error) {
	fmt.Printf("Starting split process with limit: %d tokens\n", limit)
	chunks := []Doc{}
	position := 0
	totalLength := len(text)
	currentHeaders := make(Headers)

	for position < totalLength {
		fmt.Printf("Processing chunk starting at position: %d\n", position)
		chunkText, chunkEnd := ts.getChunk(text, position, limit)
		tokens := ts.countTokens(chunkText)
		fmt.Printf("Chunk tokens: %d\n", tokens)

		headersInChunk := ts.extractHeaders(chunkText)
		ts.updateCurrentHeaders(currentHeaders, headersInChunk)

		content, urls, images := ts.extractURLsAndImages(chunkText)

		chunks = append(chunks, Doc{
			Text: content,
			Metadata: struct {
				Tokens  int
				Headers Headers
				URLs    []string
				Images  []string
			}{
				Tokens:  tokens,
				Headers: ts.copyHeaders(currentHeaders),
				URLs:    urls,
				Images:  images,
			},
		})

		fmt.Printf("Chunk processed. New position: %d\n", chunkEnd)
		position = chunkEnd
	}

	fmt.Printf("Split process completed. Total chunks: %d\n", len(chunks))
	return chunks, nil
}

func (ts *TextSplitter) getChunk(text string, start int, limit int) (string, int) {
	fmt.Printf("Getting chunk starting at %d with limit %d\n", start, limit)

	// Account for token overhead due to formatting
	overhead := ts.countTokens(ts.formatForTokenization("")) - ts.countTokens("")

	// Initial tentative end position
	remainingText := text[start:]
	remainingTokens := ts.countTokens(remainingText)
	end := min(start+int(float64(len(remainingText))*float64(limit)/float64(remainingTokens)), len(text))

	// Adjust end to avoid exceeding token limit
	chunkText := text[start:end]
	tokens := ts.countTokens(chunkText)

	for tokens+overhead > limit && end > start {
		fmt.Printf("Chunk exceeds limit with %d tokens. Adjusting end position...\n", tokens+overhead)
		end = ts.findNewChunkEnd(text, start, end)
		chunkText = text[start:end]
		tokens = ts.countTokens(chunkText)
	}

	// Adjust chunk end to align with newlines
	end = ts.adjustChunkEnd(text, start, end, tokens+overhead, limit)

	chunkText = text[start:end]
	tokens = ts.countTokens(chunkText)
	fmt.Printf("Final chunk end: %d\n", end)
	return chunkText, end
}

func (ts *TextSplitter) adjustChunkEnd(text string, start int, end int, currentTokens int, limit int) int {
	minChunkTokens := int(float64(limit) * 0.8) // Minimum chunk size is 80% of limit

	nextNewline := strings.Index(text[end:], "\n")
	prevNewline := strings.LastIndex(text[start:end], "\n")

	// Try extending to next newline
	if nextNewline != -1 {
		extendedEnd := end + nextNewline + 1
		chunkText := text[start:extendedEnd]
		tokens := ts.countTokens(chunkText)
		if tokens <= limit && tokens >= minChunkTokens {
			fmt.Printf("Extending chunk to next newline at position %d\n", extendedEnd)
			return extendedEnd
		}
	}

	// Try reducing to previous newline
	if prevNewline > 0 {
		reducedEnd := start + prevNewline + 1
		chunkText := text[start:reducedEnd]
		tokens := ts.countTokens(chunkText)
		if tokens <= limit && tokens >= minChunkTokens {
			fmt.Printf("Reducing chunk to previous newline at position %d\n", reducedEnd)
			return reducedEnd
		}
	}

	return end
}

func (ts *TextSplitter) findNewChunkEnd(text string, start int, end int) int {
	// Reduce end position to try to fit within token limit
	newEnd := end - (end-start)/10 // Reduce by 10% each iteration
	if newEnd <= start {
		newEnd = start + 1 // Ensure at least one character is included
	}
	return newEnd
}

func (ts *TextSplitter) extractHeaders(text string) Headers {
	headers := make(Headers)
	headerRegex := regexp.MustCompile(`(^|\n)(#{1,6})\s+(.*)`)
	matches := headerRegex.FindAllStringSubmatch(text, -1)

	for _, match := range matches {
		if len(match) >= 4 {
			level := len(match[2])
			content := strings.TrimSpace(match[3])
			key := fmt.Sprintf("h%d", level)
			headers[key] = append(headers[key], content)
		}
	}

	return headers
}

func (ts *TextSplitter) updateCurrentHeaders(current Headers, extracted Headers) {
	for level := 1; level <= 6; level++ {
		key := fmt.Sprintf("h%d", level)
		if headers, exists := extracted[key]; exists {
			current[key] = headers
			ts.clearLowerHeaders(current, level)
		}
	}
}

func (ts *TextSplitter) clearLowerHeaders(headers Headers, level int) {
	for l := level + 1; l <= 6; l++ {
		delete(headers, fmt.Sprintf("h%d", l))
	}
}

func (ts *TextSplitter) extractURLsAndImages(text string) (string, []string, []string) {
	urls := []string{}
	images := []string{}
	urlIndex := 0
	imageIndex := 0

	// Replace image markdown with placeholders
	imageRegex := regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)
	content := imageRegex.ReplaceAllStringFunc(text, func(match string) string {
		submatches := imageRegex.FindStringSubmatch(match)
		if len(submatches) >= 3 {
			altText := submatches[1]
			url := submatches[2]
			images = append(images, url)
			return fmt.Sprintf("![%s]({{$img%d}})", altText, imageIndex)
		}
		return match
	})

	// Replace URL markdown with placeholders
	urlRegex := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	content = urlRegex.ReplaceAllStringFunc(content, func(match string) string {
		submatches := urlRegex.FindStringSubmatch(match)
		if len(submatches) >= 3 {
			linkText := submatches[1]
			url := submatches[2]
			urls = append(urls, url)
			return fmt.Sprintf("[%s]({{$url%d}})", linkText, urlIndex)
		}
		return match
	})

	return content, urls, images
}

func (ts *TextSplitter) formatForTokenization(text string) string {
	return fmt.Sprintf("<|im_start|>user\n%s<|im_end|>\n<|im_start|>assistant<|im_end|>", text)
}

func (ts *TextSplitter) copyHeaders(headers Headers) Headers {
	copy := make(Headers)
	for k, v := range headers {
		copy[k] = append([]string{}, v...)
	}
	return copy
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
