package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/crowmw/ai_devs3/pkg/config"
	"github.com/crowmw/ai_devs3/pkg/http"
	"github.com/crowmw/ai_devs3/pkg/processor"
)

func main() {
	// Load environment variables from .env file
	if err := config.LoadEnv(); err != nil {
		fmt.Println(err)
		return
	}

	// Create output directory if it doesn't exist
	outputDir := filepath.Join("cmd", "s02e05", "downloaded")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Println("Error creating output directory:", err)
		return
	}

	// Create article processor
	articleProcessor := processor.NewArticleProcessor(outputDir)

	// Fetch HTML content
	htmlContent, err := http.FetchData(config.GetC3ntralaURL() + "/dane/arxiv-draft.html")
	if err != nil {
		fmt.Println("Error fetching HTML:", err)
		return
	}

	// Process article
	articleText, err := articleProcessor.ProcessArticle(htmlContent)
	if err != nil {
		fmt.Println("Error processing article:", err)
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

	// Process questions
	answers, err := articleProcessor.ProcessQuestions(articleText, questions)
	if err != nil {
		fmt.Println("Error processing questions:", err)
		return
	}

	// Send report
	result, err := http.SendC3ntralaReport("arxiv", answers)
	if err != nil {
		fmt.Println("Error sending report:", err)
		return
	}

	fmt.Println("Processing completed successfully")
	fmt.Println("Result:", result)
}
