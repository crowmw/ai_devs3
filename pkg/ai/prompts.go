package ai

import (
	openai "github.com/sashabaranov/go-openai"
)

const yearExtractionPrompt = `You are a helpful assistant that extracts years from text. 
Your task is to respond ONLY with the year mentioned in the text.
If there are multiple years, respond with the most relevant one.
If no year is mentioned, respond with "No year found".

Example:
Input: Kiedy wybuchła druga wojna światowa?
Output: 1939

Example: Kiedy miał miejsce atak na World Trade Center?
Output: 2001

Input: %s
Output:`

// GetYearExtractionPrompt creates a chat completion message with the year extraction prompt
func GetYearExtractionPrompt(question string) []openai.ChatCompletionMessage {
	return []openai.ChatCompletionMessage{
		{
			Role:    "system",
			Content: yearExtractionPrompt,
		},
		{
			Role:    "user",
			Content: question,
		},
	}
}

const imageAnalysisPrompt = `You are a helpful assistant that analyzes images.
Your task is to describe what you see in the image in a clear and concise way.
Focus on the main subjects, important details, and overall composition.
Describe any text, symbols, or notable visual elements.
Keep your description objective and factual.
Respond in a natural, conversational tone.
Do not mention that you are an AI or that you are analyzing an image.
Do not include phrases like "I see" or "I can see".
Just describe what is present in the image directly.`

// GetImageAnalysisPrompt creates a chat completion message for image analysis
func GetImageAnalysisPrompt(imageBase64 string, format string, userText string) []openai.ChatCompletionMessage {
	prompt := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: imageAnalysisPrompt,
		},
		{
			Role: openai.ChatMessageRoleUser,
			MultiContent: []openai.ChatMessagePart{
				{
					Type: openai.ChatMessagePartTypeText,
					Text: userText,
				},
			},
		},
	}

	// Add image parts to the message
	prompt[1].MultiContent = append(prompt[1].MultiContent, openai.ChatMessagePart{
		Type: openai.ChatMessagePartTypeImageURL,
		ImageURL: &openai.ChatMessageImageURL{
			URL:    "data:image/" + format + ";base64," + imageBase64,
			Detail: openai.ImageURLDetailHigh,
		},
	})

	return prompt
}
