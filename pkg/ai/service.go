package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/crowmw/ai_devs3/pkg/env"
	"github.com/sashabaranov/go-openai"
)

type Service struct {
	openai *openai.Client
	envSvc *env.Service
}

func NewService(envSvc *env.Service) (*Service, error) {
	client := openai.NewClient(envSvc.GetOpenAIKey())

	return &Service{
		openai: client,
		envSvc: envSvc,
	}, nil
}

func (s *Service) CreateOpenAIEmbedding(text string) ([]float32, error) {
	response, err := s.openai.CreateEmbeddings(
		context.Background(),
		openai.EmbeddingRequest{
			Input: []string{text},
			Model: openai.EmbeddingModel("text-embedding-3-small"),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error creating embedding: %w", err)
	}
	return response.Data[0].Embedding, nil
}

type jinaEmbeddingRequest struct {
	Model         string   `json:"model"`
	Task          string   `json:"task"`
	Dimensions    int      `json:"dimensions"`
	LateChunking  bool     `json:"late_chunking"`
	EmbeddingType string   `json:"embedding_type"`
	Input         []string `json:"input"`
}

type jinaEmbeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
}

func (s *Service) CreateJinaEmbedding(text string) ([]float32, error) {
	reqBody := jinaEmbeddingRequest{
		Model:         "jina-embeddings-v3",
		Task:          "text-matching",
		Dimensions:    1024,
		LateChunking:  false,
		EmbeddingType: "float",
		Input:         []string{text},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.jina.ai/v1/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.envSvc.GetJinaAPIKey()))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP error! status: %d, body: %s", resp.StatusCode, string(body))
	}

	var result jinaEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	if len(result.Data) == 0 || len(result.Data[0].Embedding) == 0 {
		return nil, fmt.Errorf("empty embedding in response")
	}

	return result.Data[0].Embedding, nil
}

// ChatCompletionConfig holds the configuration for chat completion
type ChatCompletionConfig struct {
	Messages  []openai.ChatCompletionMessage
	Model     string
	Stream    bool
	JSONMode  bool
	MaxTokens int
}

// ChatCompletion handles chat completions with configurable options
func (s *Service) ChatCompletion(config ChatCompletionConfig) (openai.ChatCompletionResponse, error) {
	// Set default values
	if config.Model == "" {
		config.Model = "gpt-4"
	}
	if config.MaxTokens == 0 {
		config.MaxTokens = 4096
	}

	// Prepare request
	req := openai.ChatCompletionRequest{
		Model:    config.Model,
		Messages: config.Messages,
		Stream:   config.Stream,
	}

	// Add JSON mode if specified
	if config.JSONMode {
		req.ResponseFormat = &openai.ChatCompletionResponseFormat{
			Type: "json_object",
		}
	}

	// Add max tokens if not using o1 models
	if config.Model != "o1-mini" && config.Model != "o1-preview" {
		req.MaxTokens = config.MaxTokens
	}

	// Make the request
	response, err := s.openai.CreateChatCompletion(context.Background(), req)
	if err != nil {
		return openai.ChatCompletionResponse{}, fmt.Errorf("error in OpenAI completion: %w", err)
	}

	return response, nil
}

// ProcessImageResult holds the result of image processing
type OCRImageResult struct {
	Text   string
	Source string
}

// ProcessImage processes an image using GPT-4 Vision
func (s *Service) OCRImage(imagePath string) (*OCRImageResult, error) {
	// Read the image file
	fmt.Println("Reading image file...")
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read image file: %w", err)
	}

	fmt.Println("Converting image to base64...")
	// Convert image to base64
	base64Image := base64.StdEncoding.EncodeToString(imageData)

	fmt.Println("Creating chat completion request...")
	// Create the chat completion request
	response, err := s.ChatCompletion(ChatCompletionConfig{
		Model: "gpt-4.1",
		Messages: []openai.ChatCompletionMessage{
			{
				Role: "user",
				MultiContent: []openai.ChatMessagePart{
					{
						Type: openai.ChatMessagePartTypeText,
						Text: "Read the text from the image. Text is in Polish. Return only the text, no other text or Markdown formatting.",
					},
					{
						Type: openai.ChatMessagePartTypeImageURL,
						ImageURL: &openai.ChatMessageImageURL{
							URL: fmt.Sprintf("data:image/png;base64,%s", base64Image),
						},
					},
				},
			},
		},
		// MaxTokens: 300,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to process image with GPT-4o: %w", err)
	}

	return &OCRImageResult{
		Text:   response.Choices[0].Message.Content,
		Source: imagePath,
	}, nil
}

// ProcessImage processes an image using GPT-4 Vision
func (s *Service) ImageAnalysis(imagePath string) (string, error) {
	// Read the image file
	fmt.Println("Reading image file...")
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to read image file: %w", err)
	}

	fmt.Println("Converting image to base64...")
	// Convert image to base64
	base64Image := base64.StdEncoding.EncodeToString(imageData)

	fmt.Println("Creating chat completion request...")
	// Create the chat completion request
	response, err := s.ChatCompletion(ChatCompletionConfig{
		Model: "gpt-4.1",
		Messages: []openai.ChatCompletionMessage{
			{
				Role: "user",
				MultiContent: []openai.ChatMessagePart{
					{
						Type: openai.ChatMessagePartTypeText,
						Text: "Analyze the image and return a description of the image. Text is in Polish. Return only the text, no other text or Markdown formatting.",
					},
					{
						Type: openai.ChatMessagePartTypeImageURL,
						ImageURL: &openai.ChatMessageImageURL{
							URL: fmt.Sprintf("data:image/png;base64,%s", base64Image),
						},
					},
				},
			},
		},
		// MaxTokens: 300,
	})
	if err != nil {
		return "", fmt.Errorf("failed to process image with GPT-4o: %w", err)
	}

	return response.Choices[0].Message.Content, nil
}

func (s *Service) AudioAnalysis(audioFileUrl string) (string, error) {
	fmt.Printf("🎵 [AUDIO] Starting audio analysis for URL: %s\n", audioFileUrl)

	// Create a temporary file to store the audio
	tempFile, err := os.CreateTemp("", "audio-*.mp3")
	if err != nil {
		return "", fmt.Errorf("error creating temp file: %w", err)
	}
	defer os.Remove(tempFile.Name()) // Clean up the temp file when done
	defer tempFile.Close()

	fmt.Println("📥 [AUDIO] Downloading audio file...")
	// Download the audio file
	httpResp, err := http.Get(audioFileUrl)
	if err != nil {
		return "", fmt.Errorf("error downloading audio file: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error downloading audio file: status code %d", httpResp.StatusCode)
	}

	// Copy the downloaded content to the temp file
	_, err = io.Copy(tempFile, httpResp.Body)
	if err != nil {
		return "", fmt.Errorf("error saving audio file: %w", err)
	}

	// Ensure all data is written to disk
	if err := tempFile.Sync(); err != nil {
		return "", fmt.Errorf("error syncing temp file: %w", err)
	}

	fmt.Println("🤖 [AUDIO] Sending audio to OpenAI for transcription...")
	// Create the transcription request
	req := openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: tempFile.Name(),
		Format:   openai.AudioResponseFormatText,
	}

	// Send the request to OpenAI
	transcriptionResp, err := s.openai.CreateTranscription(context.Background(), req)
	if err != nil {
		return "", fmt.Errorf("error creating transcription: %w", err)
	}

	fmt.Println("✅ [AUDIO] Successfully transcribed audio")
	return transcriptionResp.Text, nil
}
