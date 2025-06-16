package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/crowmw/ai_devs3/pkg/ai"
	"github.com/crowmw/ai_devs3/pkg/env"
	"github.com/crowmw/ai_devs3/pkg/http"
	"github.com/crowmw/ai_devs3/pkg/processor"
	"github.com/sashabaranov/go-openai"
)

// QuestionResponse represents the structure of response from each question endpoint
type QuestionResponse struct {
	Task string   `json:"task"`
	Data []string `json:"data"`
}

// Question represents a single question with its task and data
type Question struct {
	Task string
	Data string
}

// fetchQuestionData fetches data from a single question endpoint
// Returns the full response and any error that occurred
func fetchQuestionData(questionURL string) (*QuestionResponse, error) {
	var response QuestionResponse
	err := http.FetchJSONData(questionURL, &response)
	if err != nil {
		return nil, fmt.Errorf("error fetching data from %s: %w", questionURL, err)
	}
	return &response, nil
}

// processQuestionAsync processes a single question URL asynchronously
// Sends results to the provided channel and handles errors
func processQuestionAsync(questionURL string, resultChan chan<- []Question, wg *sync.WaitGroup) {
	defer wg.Done()

	response, err := fetchQuestionData(questionURL)
	if err != nil {
		fmt.Printf("Warning: %v\n", err)
		return
	}

	// Convert response data to Question structs
	questions := make([]Question, len(response.Data))
	for i, data := range response.Data {
		questions[i] = Question{
			Task: response.Task,
			Data: data,
		}
	}

	resultChan <- questions
}

// collectQuestionsAsync processes all question URLs concurrently
// Returns a slice containing all questions with their tasks and data
func collectQuestionsAsync(questionURLs []string) []Question {
	// Create channel to collect results from goroutines
	resultChan := make(chan []Question)
	var wg sync.WaitGroup

	// Launch a goroutine for each question URL
	for _, url := range questionURLs {
		wg.Add(1)
		go processQuestionAsync(url, resultChan, &wg)
	}

	// Close channel when all goroutines are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect all results into a single slice
	var allQuestions []Question
	for questions := range resultChan {
		allQuestions = append(allQuestions, questions...)
	}

	return allQuestions
}

// getURLContext gets context data for a URL, either from cache or by fetching
func getURLContext(urlMatch string) (string, error) {
	// Create cache directory if it doesn't exist
	cacheDir := "cmd/s05e03"
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("error creating cache directory: %w", err)
	}

	// Extract filename from URL
	filename := filepath.Base(urlMatch)
	if filename == "" {
		return "", fmt.Errorf("invalid URL: could not extract filename from %s", urlMatch)
	}

	// Create cache filename
	cacheFile := filepath.Join(cacheDir, filename)

	// Try to read from cache first
	if data, err := os.ReadFile(cacheFile); err == nil {
		return string(data), nil
	}

	// If not in cache, fetch from URL
	data, err := http.FetchData(urlMatch)
	if err != nil {
		return "", fmt.Errorf("error fetching data from %s: %w", urlMatch, err)
	}

	// Sanitize and process the HTML content
	sanitizedContent := processor.SanitizeHTML(string(data))

	// Save to cache
	if err := os.WriteFile(cacheFile, []byte(sanitizedContent), 0644); err != nil {
		return "", fmt.Errorf("error writing to cache file: %w", err)
	}

	return sanitizedContent, nil
}

// processAIResponse processes a single question through AI service
// Returns the response and any error that occurred
func processAIResponse(aiSvc *ai.Service, question Question, urlMatch string) (string, error) {
	var systemPrompt string
	if urlMatch != "" {
		// Get context data for URL
		contextData, err := getURLContext(urlMatch)
		if err != nil {
			return "", fmt.Errorf("error getting URL context: %w", err)
		}

		systemPrompt = fmt.Sprintf(`You are a helpful assistant that can answer questions based on datasource from url: %s. 
Context data:
<context>
%s
</context>
Respond in polish language in just short answer.`, urlMatch, contextData)
	} else {
		systemPrompt = "You are a helpful assistant that can answer questions. Respond in polish language in just short answer."
	}

	response, err := aiSvc.ChatCompletion(ai.ChatCompletionConfig{
		Model: "gpt-4o-mini",
		Messages: []openai.ChatCompletionMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: question.Data},
		},
	})
	if err != nil {
		return "", fmt.Errorf("error processing question '%s': %w", question.Data, err)
	}

	return fmt.Sprintf("Task: %s, Data: %s, Answer: %s",
		question.Task,
		question.Data,
		response.Choices[0].Message.Content), nil
}

// processAIQuestionAsync processes a single question through AI service asynchronously
// Sends results to the provided channel and handles errors
func processAIQuestionAsync(aiSvc *ai.Service, question Question, resultChan chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	urlRegex := regexp.MustCompile(`https?://[^\s]+`)
	urlMatch := urlRegex.FindString(question.Task)

	// Skip if not a URL question and doesn't contain "Odpowiedz"
	if urlMatch == "" && !strings.Contains(question.Task, "Odpowiedz") {
		return
	}

	response, err := processAIResponse(aiSvc, question, urlMatch)
	if err != nil {
		fmt.Printf("Warning: %v\n", err)
		return
	}

	resultChan <- response
}

// collectAIResponses processes all questions through AI service concurrently
// Returns a slice containing all AI responses
func collectAIResponses(aiSvc *ai.Service, questions []Question) []string {
	// Create channel to collect results from goroutines
	resultChan := make(chan string)
	var wg sync.WaitGroup

	// Launch a goroutine for each question
	for _, question := range questions {
		wg.Add(1)
		go processAIQuestionAsync(aiSvc, question, resultChan, &wg)
	}

	// Close channel when all goroutines are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect all results into a single slice
	var responses []string
	for response := range resultChan {
		responses = append(responses, response)
	}

	return responses
}

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

	// Time starts here
	// Time starts here
	startTime := time.Now()
	//https://rafal.ag3nts.org/b46c3
	type HashResponse struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Hint    string `json:"hint"`
	}
	var hashResponse HashResponse
	hashBody := map[string]string{
		"password": "NONOMNISMORIAR",
	}
	err = http.POSTJSONData("https://rafal.ag3nts.org/b46c3", hashBody, &hashResponse)
	if err != nil {
		fmt.Println(err)
		return
	}

	hash := hashResponse.Message

	type SignatureResponseMesage struct {
		Signature  string   `json:"signature"`
		Timestamp  int64    `json:"timestamp"`
		Challenges []string `json:"challenges"`
	}
	type SignatureResponse struct {
		Code    int                     `json:"code"`
		Message SignatureResponseMesage `json:"message"`
	}

	var signatureResponse SignatureResponse
	signatureBody := map[string]string{
		"sign": hash,
	}

	err = http.POSTJSONData("https://rafal.ag3nts.org/b46c3", signatureBody, &signatureResponse)
	if err != nil {
		fmt.Println(err)
		return
	}

	questions := signatureResponse.Message.Challenges

	// Process all questions concurrently and collect results
	allQuestions := collectQuestionsAsync(questions)

	// Process all questions through AI service concurrently
	responses := collectAIResponses(aiSvc, allQuestions)

	// Print all responses
	for _, response := range responses {
		fmt.Println(response)
	}

	type ResponseBody struct {
		ApiKey    string   `json:"apikey"`
		Answer    []string `json:"answer"`
		Signature string   `json:"signature"`
		Timestamp int64    `json:"timestamp"`
	}

	responseBody := ResponseBody{
		ApiKey:    envSvc.GetMyAPIKey(),
		Answer:    responses,
		Signature: signatureResponse.Message.Signature,
		Timestamp: signatureResponse.Message.Timestamp,
	}
	response, err := http.SendPost("https://rafal.ag3nts.org/b46c3", responseBody)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Response: ", response)

	// Time ends here
	fmt.Println("Time taken: ", time.Since(startTime))
}
