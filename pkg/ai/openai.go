package ai

import (
	"context"
	"fmt"

	"github.com/crowmw/ai_devs3/pkg/config"

	openai "github.com/sashabaranov/go-openai"
)

func SendChatCompletion(model string, store bool, messages []openai.ChatCompletionMessage) (string, error) {
	client := openai.NewClient(config.GetOpenAIKey())

	req := openai.ChatCompletionRequest{
		Model:    model,
		Messages: messages,
	}

	resp, err := client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		return "", fmt.Errorf("error creating chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return resp.Choices[0].Message.Content, nil
}
