package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/crowmw/ai_devs3/pkg/ai"
	"github.com/crowmw/ai_devs3/pkg/config"
	"github.com/crowmw/ai_devs3/pkg/http"
	"github.com/crowmw/ai_devs3/pkg/processor"
	"github.com/sashabaranov/go-openai"
)

type RobotDescriptionResult struct {
	Description string `json:"description"`
}

func main() {
	// Load environment variables from .env file
	if err := config.LoadEnv(); err != nil {
		fmt.Println(err)
		return
	}

	// // Create output directory if it doesn't exist
	outputDir := filepath.Join("cmd", "s02e05", "downloaded")
	// if err := os.MkdirAll(outputDir, 0755); err != nil {
	// 	fmt.Println("Error creating output directory:", err)
	// 	return
	// }

	// // Fetch HTML content
	htmlContent, err := http.FetchData(config.GetC3ntralaURL() + "/dane/arxiv-draft.html")
	if err != nil {
		fmt.Println("Error fetching HTML:", err)
		return
	}

	// // Parse HTML document
	doc, err := htmlquery.Parse(bytes.NewReader(htmlContent))
	if err != nil {
		fmt.Println("Error parsing HTML:", err)
		return
	}

	// Process images in HTML
	if err := processor.ProcessImageElements(doc, outputDir); err != nil {
		fmt.Printf("Error processing images: %v\n", err)
		return
	}

	// Process audio elements in HTML
	if err := processor.ProcessAudioElements(doc, outputDir); err != nil {
		fmt.Printf("Error processing audio elements: %v\n", err)
		return
	}

	// Convert modified HTML document back to string
	modifiedHTML := htmlquery.OutputHTML(doc, true)

	// Convert HTML to plain text by removing all HTML tags
	re := regexp.MustCompile("<[^>]*>")
	plainText := re.ReplaceAllString(modifiedHTML, "")
	plainText = strings.TrimSpace(plainText)

	// Save as markdown
	markdownFilePath := filepath.Join(outputDir, "article_described.md")
	if err := os.WriteFile(markdownFilePath, []byte(plainText), 0644); err != nil {
		fmt.Println("Error saving Markdown file:", err)
		return
	}

	// Fetch questions from arxiv.txt
	questionsURL := config.GetC3ntralaURL() + "/data/" + config.GetMyAPIKey() + "/arxiv.txt"
	questionsData, err := http.FetchData(questionsURL)
	if err != nil {
		fmt.Println("Error fetching questions:", err)
		return
	}

	questions := processor.ReadLinesFromTextFile(questionsData)

	fmt.Println("Scraping completed successfully")
	fmt.Println("Markdown file saved to:", markdownFilePath)
	fmt.Println("Questions file saved to:", questions)

	questionsPrompt := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: "<context>" + plainText + "</context>",
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: "Based ONLY on context of the article, answer in one paragraph for each of the following questions: " + strings.Join(questions, "\n") + ". The answer should be in the same language as the article. Think a bit before place answer. Analyze what portion of text or image description or audio transcription answer for question. Context is key: Pay attention to the context in which graphics and sounds appear. Image captions and surrounding text can contain relevant information. Response as a JSON object with format {01: answer, 02: answer, 03: answer, ...}, do not include any other text in your response. Do not include any markdown text decorations.",
		},
	}

	aiResponse, err := ai.SendChatCompletion("gpt-4o", false, questionsPrompt)
	if err != nil {
		fmt.Println("Error sending chat completion:", err)
		return
	}

	fmt.Println(aiResponse)

	// Parse the AI response into a JSON object
	var answers map[string]string
	if err := json.Unmarshal([]byte(aiResponse), &answers); err != nil {
		fmt.Println("Error parsing AI response:", err)
		return
	}

	result, err := http.SendC3ntralaReport("arxiv", answers)
	if err != nil {
		fmt.Println("Error sending report:", err)
		return
	}

	fmt.Println(result)
}
