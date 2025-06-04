package softo

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/crowmw/ai_devs3/pkg/ai"
	"github.com/crowmw/ai_devs3/pkg/env"
	"github.com/crowmw/ai_devs3/pkg/http"
	"github.com/crowmw/ai_devs3/pkg/processor"
	"github.com/sashabaranov/go-openai"
)

type Service struct {
	baseUrl  string
	pagesDir string
	aiSvc    *ai.Service
}

func NewService(envSvc *env.Service, aiSvc *ai.Service) (*Service, error) {
	baseUrl := envSvc.GetSoftoURL()
	pagesDir := "pkg/softo/pages"
	return &Service{
		baseUrl:  baseUrl,
		pagesDir: pagesDir,
		aiSvc:    aiSvc,
	}, nil
}

func (s *Service) GetPage(url string) (string, error) {
	fmt.Println("❇️ Fetching page:", url)

	// Extract filename from URL
	filename := processor.GetFilenameFromURL(url)
	if filename == "" {
		return "", fmt.Errorf("could not extract filename from URL")
	}

	// Check if file already exists
	filePath := filepath.Join(s.pagesDir, filename+".html")
	if _, err := os.Stat(filePath); err == nil {
		// File exists, read and return content
		content, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Error reading existing file %s: %v\n", filePath, err)
			return "", fmt.Errorf("error reading existing file %s: %v", filePath, err)
		}
		fmt.Printf("✅ Found existing file %s\n", filePath)
		return string(content), nil
	}

	body, err := http.FetchData(url)
	if err != nil {
		fmt.Println(err)
		return "", fmt.Errorf("error fetching data: %v", err)
	}

	sanitizedHTML := processor.SanitizeHTML(string(body))

	// Create pages directory if it doesn't exist
	if err := os.MkdirAll(s.pagesDir, 0755); err != nil {
		fmt.Printf("Error creating directory %s: %v\n", s.pagesDir, err)
		return "", fmt.Errorf("error creating directory %s: %v", s.pagesDir, err)
	}

	// Write MD to file
	if err := os.WriteFile(filePath, []byte(sanitizedHTML), 0644); err != nil {
		fmt.Printf("Error writing file %s: %v\n", filePath, err)
		return "", fmt.Errorf("error writing file %s: %v", filePath, err)
	}

	return sanitizedHTML, nil
}

type Answer struct {
	Response string `json:"response"`
	Answer   string `json:"answer"`
}

func (s *Service) TryToFindAnswer(question string, url string) (*Answer, error) {
	fmt.Println("❇️ Trying to find answer for:", question, "at:", s.baseUrl+url)

	page, err := s.GetPage(s.baseUrl + url)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	response, err := s.aiSvc.ChatCompletion(ai.ChatCompletionConfig{
		Model: "gpt-4o-mini",
		Messages: []openai.ChatCompletionMessage{
			{Role: "system", Content: fmt.Sprintf(answerQuestionSystemPrompt, page)},
			{Role: "user", Content: question},
		},
	})
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var answer *Answer
	err = json.Unmarshal([]byte(response.Choices[0].Message.Content), &answer)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return answer, nil
}
