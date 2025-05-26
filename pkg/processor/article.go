package processor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/crowmw/ai_devs3/pkg/ai"
	"github.com/crowmw/ai_devs3/pkg/config"
	htmlpkg "github.com/crowmw/ai_devs3/pkg/html"
	"github.com/crowmw/ai_devs3/pkg/media"
	"github.com/sashabaranov/go-openai"
	"golang.org/x/net/html"
)

// ArticleProcessor handles processing of HTML articles
type ArticleProcessor struct {
	outputDir string
	htmlProc  *htmlpkg.Processor
	mediaProc *media.Processor
}

// NewArticleProcessor creates a new ArticleProcessor instance
func NewArticleProcessor(outputDir string) *ArticleProcessor {
	htmlProc := htmlpkg.NewProcessor(config.GetC3ntralaURL())
	mediaProc := media.NewProcessor(outputDir, htmlProc)
	return &ArticleProcessor{
		outputDir: outputDir,
		htmlProc:  htmlProc,
		mediaProc: mediaProc,
	}
}

// ProcessArticle processes an HTML article, including images and audio
func (ap *ArticleProcessor) ProcessArticle(htmlContent []byte) (string, error) {
	// Parse HTML document
	doc, err := ap.htmlProc.ParseHTML(htmlContent)
	if err != nil {
		return "", fmt.Errorf("error parsing HTML: %w", err)
	}

	// Process images in HTML
	if err := ap.processImages(doc); err != nil {
		return "", fmt.Errorf("error processing images: %w", err)
	}

	// Process audio elements in HTML
	if err := ap.processAudio(doc); err != nil {
		return "", fmt.Errorf("error processing audio elements: %w", err)
	}

	// Convert to plain text
	plainText := ap.htmlProc.ConvertToPlainText(doc)

	// Save as markdown
	markdownFilePath := filepath.Join(ap.outputDir, "article_described.md")
	if err := os.WriteFile(markdownFilePath, []byte(plainText), 0644); err != nil {
		return "", fmt.Errorf("error saving Markdown file: %w", err)
	}

	return plainText, nil
}

// processImages processes all images in the document
func (ap *ArticleProcessor) processImages(doc *html.Node) error {
	imageNodes := ap.htmlProc.GetElementsByTag(doc, "img")
	for i, node := range imageNodes {
		src := ap.htmlProc.GetElementAttribute(node, "src")
		if src == "" {
			continue
		}

		description, err := ap.mediaProc.ProcessImage(src, i)
		if err != nil {
			return fmt.Errorf("error processing image %s: %w", src, err)
		}

		ap.htmlProc.ReplaceNode(node, description)
	}
	return nil
}

// processAudio processes all audio elements in the document
func (ap *ArticleProcessor) processAudio(doc *html.Node) error {
	audioNodes := ap.htmlProc.GetElementsByTag(doc, "source")
	for i, node := range audioNodes {
		src := ap.htmlProc.GetElementAttribute(node, "src")
		if src == "" {
			continue
		}

		transcription, err := ap.mediaProc.ProcessAudio(src, i)
		if err != nil {
			return fmt.Errorf("error processing audio %s: %w", src, err)
		}

		ap.htmlProc.ReplaceNode(node, transcription)
	}
	return nil
}

// ProcessQuestions processes questions and generates answers based on article content
func (ap *ArticleProcessor) ProcessQuestions(articleText string, questions []string) (map[string]string, error) {
	questionsPrompt := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: "<context>" + articleText + "</context>",
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: "Based ONLY on context of the article, answer in one paragraph for each of the following questions: " + strings.Join(questions, "\n") + ". The answer should be in the same language as the article. Think a bit before place answer. Analyze what portion of text or image description or audio transcription answer for question. Context is key: Pay attention to the context in which graphics and sounds appear. Image captions and surrounding text can contain relevant information. Response as a JSON object with format {01: answer, 02: answer, 03: answer, ...}, do not include any other text in your response. Do not include any markdown text decorations.",
		},
	}

	aiResponse, err := ai.SendChatCompletion("gpt-4o", false, questionsPrompt)
	if err != nil {
		return nil, fmt.Errorf("error sending chat completion: %w", err)
	}

	// Parse the AI response into a JSON object
	var answers map[string]string
	if err := json.Unmarshal([]byte(aiResponse), &answers); err != nil {
		return nil, fmt.Errorf("error parsing AI response: %w", err)
	}

	return answers, nil
}
