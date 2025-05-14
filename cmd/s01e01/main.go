package main

import (
	"fmt"

	"github.com/crowmw/ai_devs3/pkg/ai"
	"github.com/crowmw/ai_devs3/pkg/config"
	"github.com/crowmw/ai_devs3/pkg/http"
	"github.com/crowmw/ai_devs3/pkg/processor"
)

func main() {
	// Load environment variables
	if err := config.LoadEnv(); err != nil {
		fmt.Println(err)
		return
	}

	// Fetch data
	data, err := http.FetchData(config.GetXYZURL())
	if err != nil {
		fmt.Println(err)
		return
	}

	// Process data
	htmlString := processor.GetHTMLString(data)

	question := processor.ExtractTextFromHTML(htmlString, "//p[@id='human-question']")

	// Send question to OpenAI
	openaiYearResponse, err := ai.SendChatCompletion("gpt-4o-mini", true, ai.GetYearExtractionPrompt(question))
	if err != nil {
		fmt.Println("Error getting OpenAI response:", err)
		return
	}
	formData := map[string]string{
		"username": "tester",
		"password": "574e112a",
		"answer":   openaiYearResponse,
	}

	response, err := http.SendFormData(config.GetXYZURL(), formData)
	if err != nil {
		fmt.Println("Error sending form data:", err)
		return
	}

	flag := processor.ExtractTextFromHTML(response, "//h2/text()")

	fmt.Println(flag)
}
