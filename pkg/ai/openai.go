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

// TranscribeAudioAndFormat transcribes an audio file and formats the output with a header
func TranscribeAudioAndFormat(audioFilePath string) (string, error) {
	transcription, err := TranscribeAudio(audioFilePath)
	if err != nil {
		return "", fmt.Errorf("error transcribing audio: %w", err)
	}

	// Format the output with a header
	formattedOutput := "To jest transkrypcja audio:\n\n" + transcription
	return formattedOutput, nil
}

// DescribeImageAndFormat describes an image using GPT-4 Vision with proper format handling
func DescribeImageAndFormat(imageBase64 string, format string) (string, error) {
	prompt := GetImageAnalysisPrompt(imageBase64, format, "Please analyze these image, in a few sentences, do not include any other text in your response, just the analysis, in Polish language, add information before that said about this is an image description search for the subject of the image:")
	return SendChatCompletion("gpt-4o-mini", false, prompt)
}

// GenerateImageWithDalle generates an image using DALL-E 3 model
func GenerateImageWithDalle(prompt string) (string, error) {
	fmt.Println("Generating image with DALL-E 3...")
	client := openai.NewClient(config.GetOpenAIKey())

	req := openai.ImageRequest{
		Model:   "dall-e-3",
		Prompt:  prompt,
		N:       1,
		Size:    "1024x1024",
		Quality: "standard",
		Style:   "natural",
	}

	resp, err := client.CreateImage(context.Background(), req)
	if err != nil {
		return "", fmt.Errorf("error creating image: %w", err)
	}

	if len(resp.Data) == 0 {
		return "", fmt.Errorf("no image data in response")
	}

	return resp.Data[0].URL, nil
}

// DescribeImage describes an image using GPT-4 Vision
func DescribeImage(imageBase64 string, format string, userText string) (string, error) {
	fmt.Println("Describing image...")
	prompt := GetImageAnalysisPrompt(imageBase64, format, userText)
	fmt.Println("Prompt:", prompt)
	return SendChatCompletion("gpt-4o-mini", true, prompt)
}
