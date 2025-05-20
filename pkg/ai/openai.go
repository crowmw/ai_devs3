package ai

import (
	"context"
	"fmt"
	"os"

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

// TranscribeAudio transcribes an audio file using OpenAI's Whisper model
func TranscribeAudio(audioFilePath string) (string, error) {
	client := openai.NewClient(config.GetOpenAIKey())

	// Open the audio file
	audioFile, err := os.Open(audioFilePath)
	if err != nil {
		return "", fmt.Errorf("error opening audio file: %w", err)
	}
	defer audioFile.Close()

	// Create the transcription request
	req := openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: audioFilePath,
		Format:   openai.AudioResponseFormatText,
	}

	// Send the request to OpenAI
	fmt.Println("Sending audio to OpenAI...")
	resp, err := client.CreateTranscription(context.Background(), req)
	if err != nil {
		return "", fmt.Errorf("error creating transcription: %w", err)
	}

	return resp.Text, nil
}
