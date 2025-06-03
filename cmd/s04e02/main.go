package main

import (
	"fmt"
	"os"

	"github.com/crowmw/ai_devs3/pkg/ai"
	"github.com/crowmw/ai_devs3/pkg/c3ntrala"
	"github.com/crowmw/ai_devs3/pkg/env"
	"github.com/crowmw/ai_devs3/pkg/processor"
	"github.com/sashabaranov/go-openai"
)

func main() {
	envSvc, err := env.NewService()
	if err != nil {
		fmt.Println(err)
		return
	}

	aiSvc, err := ai.NewService(envSvc)
	if err != nil {
		fmt.Println(err)
		return
	}

	c3ntralaSvc, err := c3ntrala.NewService(envSvc)
	if err != nil {
		fmt.Println(err)
		return
	}

	file, err := os.ReadFile("lab_data/verify.txt")
	if err != nil {
		fmt.Println(err)
		return
	}

	lines := processor.ReadLinesFromTextFile(file)
	for i := range lines {
		if len(lines[i]) > 3 {
			lines[i] = lines[i][3:]
		}
	}

	fmt.Println("❇️ Lines:", lines)

	var correctLines []string
	for index, line := range lines {
		answer, err := aiSvc.ChatCompletion(ai.ChatCompletionConfig{
			Model: "ft:gpt-4.1-mini-2025-04-14:personal:aidev:BeSk3Dmi",
			Messages: []openai.ChatCompletionMessage{
				{Role: "system", Content: "validate data"},
				{Role: "user", Content: line},
			},
		})
		if err != nil {
			fmt.Println(err)
			return
		}

		if answer.Choices[0].Message.Content == "1" {
			correctLines = append(correctLines, fmt.Sprintf("%02d", index+1))
		}
	}

	fmt.Println("❇️ Correct lines:", correctLines)

	reportResponse, err := c3ntralaSvc.PostReport("research", correctLines)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("❇️ Report response:", reportResponse)
}
