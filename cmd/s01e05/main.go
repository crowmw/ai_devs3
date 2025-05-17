package main

import (
	"fmt"

	"github.com/crowmw/ai_devs3/pkg/ai"
	"github.com/crowmw/ai_devs3/pkg/config"
	"github.com/crowmw/ai_devs3/pkg/http"
	"github.com/crowmw/ai_devs3/pkg/processor"
	"github.com/sashabaranov/go-openai"
)

func main() {
	// Load environment variables from .env file
	if err := config.LoadEnv(); err != nil {
		fmt.Println(err)
		return
	}

	// Fetch the text file that needs to be censored from the API
	data, err := http.FetchData(config.GetC3ntralaURL() + "/data/" + config.GetMyAPIKey() + "/cenzura.txt")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Convert the fetched data into a string format
	dataToCensor := processor.ReadLinesFromTextFileAsString(data)

	fmt.Println(dataToCensor)

	// Read the system instructions from system.md file
	// This file contains the rules and guidelines for censorship
	systemPrompt, err := processor.ReadMarkdownFile("cmd/s01e05/systemPrompt.md")
	if err != nil {
		fmt.Println("Error reading system.md:", err)
		return
	}

	// Create a chat completion request with:
	// - System message: Instructions from system.md
	// - User message: The text that needs to be censored
	question := []openai.ChatCompletionMessage{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: dataToCensor,
		},
	}

	// Send the request to OpenAI API to get censored version of the text
	censoredData, err := ai.SendChatCompletion("gpt-4o-mini", true, question)
	if err != nil {
		fmt.Println("Error sending chat completion:", err)
		return
	}

	fmt.Println(censoredData)

	// Send the censored text back to the API as a report
	reportResponse, err := http.SendC3ntralaReport("CENZURA", censoredData)
	if err != nil {
		fmt.Println("Error sending report:", err)
		return
	}

	fmt.Println(reportResponse)
}
