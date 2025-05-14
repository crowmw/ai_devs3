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
