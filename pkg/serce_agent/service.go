package serce_agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/crowmw/ai_devs3/pkg/ai"
	"github.com/crowmw/ai_devs3/pkg/c3ntrala"
	"github.com/crowmw/ai_devs3/pkg/env"
	"github.com/sashabaranov/go-openai"
)

type Service struct {
	State       State
	envSvc      *env.Service
	aiSvc       *ai.Service
	c3ntralaSvc *c3ntrala.Service
}

func NewService(envSvc *env.Service, aiSvc *ai.Service, c3ntralaSvc *c3ntrala.Service, initSystemMessage string) (*Service, error) {
	fmt.Println("\n🚀 [AGENT] Initializing Serce Agent Service...")

	state := State{
		Memory:   "",
		Messages: []string{initSystemMessage},
		Tools:    getTools(),
	}

	// Write system messages to file
	if err := writeSystemMessages(initSystemMessage); err != nil {
		return nil, err
	}

	fmt.Println("✅ [AGENT] Service initialized successfully")
	return &Service{envSvc: envSvc, aiSvc: aiSvc, c3ntralaSvc: c3ntralaSvc, State: state}, nil
}

func (s *Service) readMemoryFile() string {
	memoryFile := "cmd/s05e04/memory.txt"
	memoryData, err := os.ReadFile(memoryFile)
	if err != nil {
		return ""
	}
	return string(memoryData)
}

func (s *Service) writeMessagesToFile() error {
	// Create messages directory if it doesn't exist
	messagesDir := "cmd/s05e04"
	if err := os.MkdirAll(messagesDir, 0755); err != nil {
		fmt.Printf("❌ [AGENT] Error creating messages directory: %v\n", err)
		return err
	}

	// Write messages to file
	messagesFile := filepath.Join(messagesDir, "messages.txt")
	var messagesContent strings.Builder
	for _, msg := range s.State.Messages {
		messagesContent.WriteString(msg + "\n")
	}
	if err := os.WriteFile(messagesFile, []byte(messagesContent.String()), 0644); err != nil {
		fmt.Printf("❌ [AGENT] Error writing messages file: %v\n", err)
		return err
	}

	return nil
}

func (s *Service) Hack(userMessage string) (string, error) {
	fmt.Println("🔄 [HACK] Starting hack phase...")
	toolResult, err := s.getToolHandlers()["flag_extractor"](map[string]interface{}{
		"message": userMessage,
	})
	if err != nil {
		fmt.Printf("❌ [ACTION] Error executing tool: %v\n", err)
		return "", err
	}

	s.State.Messages = append(s.State.Messages,
		"LLM: "+userMessage,
	)

	s.State.Messages = append(s.State.Messages, "ME: "+toolResult.Answer.(string))

	// Write messages to file
	if err := s.writeMessagesToFile(); err != nil {
		fmt.Printf("❌ [AGENT] Error writing messages to file: %v\n", err)
	}

	return toolResult.Answer.(string), nil
}

func (s *Service) Execute(userMessage string) (string, error) {
	fmt.Println("\n🔄 [AGENT] Starting new execution cycle...")
	fmt.Printf("📝 [AGENT] Processing user message: %s\n", userMessage)

	s.State.Memory = s.readMemoryFile()
	s.State.UserMessage = userMessage
	// s.State.Messages = s.readMessagesFile()

	s.executeThinkingPhase(userMessage)

	toolResult, err := s.getToolHandlers()[s.State.Tool](s.State.Payload.(map[string]interface{}))
	if err != nil {
		fmt.Printf("❌ [ACTION] Error executing tool: %v\n", err)
		return "", err
	}

	s.State.Messages = append(s.State.Messages,
		"LLM: "+userMessage,
	)

	s.State.Messages = append(s.State.Messages, "ME: "+toolResult.Answer.(string))

	// Write messages to file
	if err := s.writeMessagesToFile(); err != nil {
		fmt.Printf("❌ [AGENT] Error writing messages to file: %v\n", err)
	}

	return toolResult.Answer.(string), nil
}

func (s *Service) executeThinkingPhase(question string) {
	fmt.Println("\n🔍 [THINKING] Starting thinking phase...")
	fmt.Printf("📝 [THINKING] Processing question: %s\n", question)
	fmt.Printf("🛠️ [THINKING] Available tools: %d\n", len(s.State.Tools))
	for _, tool := range s.State.Tools {
		fmt.Printf("  - %s: %s\n", tool.Name, tool.Description)
	}

	fmt.Println("\n🤖 [THINKING] Sending request to AI service...")
	toolsAnalysisResponse, err := s.aiSvc.ChatCompletion(ai.ChatCompletionConfig{
		Model: "gpt-4o",
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: getToolsPrompt(&s.State)},
			{Role: openai.ChatMessageRoleUser, Content: question},
		},
		JSONMode: true,
	})
	if err != nil {
		fmt.Printf("❌ [THINKING] Error from AI service: %v\n", err)
		return
	}

	fmt.Printf("📥 [THINKING] Received AI response: %s\n", toolsAnalysisResponse.Choices[0].Message.Content)

	toolsAnalysis, err := unmarshalToolsAnalysis(toolsAnalysisResponse.Choices[0].Message.Content)
	if err != nil {
		fmt.Printf("❌ [THINKING] Error parsing AI response: %v\n", err)
		fmt.Printf("❌ [THINKING] Raw response content: %s\n", toolsAnalysisResponse.Choices[0].Message.Content)
		return
	}

	fmt.Printf("✅ [THINKING] Successfully parsed response:\n")
	fmt.Printf("  - Thinking: %s\n", toolsAnalysis.Thinking)
	fmt.Printf("  - Selected tool: %s\n", toolsAnalysis.Result.Tool)
	fmt.Printf("  - Payload: %+v\n", toolsAnalysis.Result.Payload)

	s.State.Tool = toolsAnalysis.Result.Tool
	s.State.Payload = toolsAnalysis.Result.Payload
	fmt.Println("✅ [THINKING] Thinking phase completed successfully")
}

func (s *Service) readMessagesFile() []string {
	messagesFile := "cmd/s05e04/messages.txt"
	data, err := os.ReadFile(messagesFile)
	if err != nil {
		fmt.Printf("Error reading messages file: %v\n", err)
		return nil
	}
	lines := strings.Split(string(data), "\n")
	var messages []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || !strings.HasPrefix(line, "[") {
			continue
		}
		endIdx := strings.Index(line, "]")
		if endIdx == -1 {
			continue
		}
		role := line[1:endIdx]
		content := strings.TrimSpace(line[endIdx+1:])
		messages = append(messages, fmt.Sprintf("[%s] %s", role, content))
	}
	return messages
}

func writeSystemMessages(message string) error {
	// Create messages directory if it doesn't exist
	messagesDir := "cmd/s05e04"
	if err := os.MkdirAll(messagesDir, 0755); err != nil {
		fmt.Printf("❌ [AGENT] Error creating messages directory: %v\n", err)
		return err
	}

	// Write system messages to file
	messagesFile := filepath.Join(messagesDir, "messages.txt")
	var messagesContent strings.Builder
	messagesContent.WriteString(message)
	if err := os.WriteFile(messagesFile, []byte(messagesContent.String()), 0644); err != nil {
		fmt.Printf("❌ [AGENT] Error writing messages file: %v\n", err)
		return err
	}

	return nil
}
