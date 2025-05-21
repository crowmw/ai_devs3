package main

import (
	"fmt"

	"github.com/crowmw/ai_devs3/pkg/ai"
	"github.com/crowmw/ai_devs3/pkg/config"
	"github.com/crowmw/ai_devs3/pkg/processor"
	"github.com/sashabaranov/go-openai"
)

func main() {
	// Load environment variables from .env file
	if err := config.LoadEnv(); err != nil {
		fmt.Println(err)
		return
	}

	// Read all images from maps directory
	fmt.Println("Reading images from maps directory")
	mapImagesBase64, err := processor.ReadAllImagesFromDirectory("cmd/s02e02/maps")
	if err != nil {
		fmt.Printf("Error reading images: %v\n", err)
		return
	}

	prompt := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		},
		{
			Role: openai.ChatMessageRoleUser,
			MultiContent: []openai.ChatMessagePart{
				{
					Type: openai.ChatMessagePartTypeText,
					Text: "Please analyze these map fragments:",
				},
			},
		},
	}

	// Add image parts to the message
	for _, img := range mapImagesBase64 {
		prompt[1].MultiContent = append(prompt[1].MultiContent, openai.ChatMessagePart{
			Type: openai.ChatMessagePartTypeImageURL,
			ImageURL: &openai.ChatMessageImageURL{
				URL:    "data:image/jpeg;base64," + img,
				Detail: openai.ImageURLDetailHigh,
			},
		})
	}

	fmt.Println("Sending chat completion")
	response, err := ai.SendChatCompletion("gpt-4o", false, prompt)
	if err != nil {
		fmt.Printf("Error sending chat completion: %v\n", err)
		return
	}

	fmt.Println(response)
}

const systemPrompt = `You are an expert in analyzing Polish maps. Please carefully analyze four map fragments that will be sent as JPG images.

Your main task is to identify the city from which these map fragments originate. Pay special attention to characteristic elements such as:

1. Granaries and fortress - these are key objects helping to identify the city in Poland
2. Street names
3. Important buildings and places:
   - churches
   - cemeteries
   - schools
   - parks
   - other characteristic objects

Remember that one of the fragments may come from a different city in Poland. If you notice an inconsistency:
- Indicate which fragment doesn't match the others
- Explain why you think it comes from a different city

Finally, provide the name of the city in Poland from which 3 out of 4 map fragments originate. Justify your answer by pointing out specific elements from the maps that allowed you to identify this city.`
