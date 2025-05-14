package robot

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/crowmw/ai_devs3/pkg/config"

	"github.com/crowmw/ai_devs3/pkg/ai"
	"github.com/crowmw/ai_devs3/pkg/http"
	"github.com/crowmw/ai_devs3/pkg/processor"

	openai "github.com/sashabaranov/go-openai"
)

// RobotGuard represents a robot security guard service
type RobotGuard struct {
	memoryDump string
}

// RobotVerificationResponseValues represents the extracted values from a response
type RobotVerificationResponseValues struct {
	MsgID int    `json:"msgID"`
	Text  string `json:"text"`
}

// NewRobotGuard creates a new instance of RobotGuard and fetches the memory dump
func NewRobotGuard() (*RobotGuard, error) {
	// Fetch robot memory dump
	data, err := http.FetchData(config.GetXYZURL() + "/files/0_13_4b.txt")
	if err != nil {
		return nil, fmt.Errorf("error fetching memory dump: %w", err)
	}

	// Process data
	memoryDump := processor.ReadLinesFromTextFileAsString(data)

	return &RobotGuard{
		memoryDump: memoryDump,
	}, nil
}

// ExtractWarningSection extracts text between the warning markers
func (rg *RobotGuard) ExtractWarningSection(text string) string {
	startMarker := "************* Uwaga! *************"
	endMarker := "**********************************"

	startIndex := strings.Index(text, startMarker)
	if startIndex == -1 {
		return ""
	}

	// Move past the start marker
	startIndex += len(startMarker)

	endIndex := strings.Index(text[startIndex:], endMarker)
	if endIndex == -1 {
		return ""
	}

	// Extract the text between markers and trim whitespace
	result := strings.TrimSpace(text[startIndex : startIndex+endIndex])
	return result
}

// ExtractRobotVerificationResponseValues extracts msgID and text from a JSON response string
func (rg *RobotGuard) ExtractRobotVerificationResponseValues(response string) (RobotVerificationResponseValues, error) {
	var result struct {
		MsgID int    `json:"msgID"`
		Text  string `json:"text"`
	}

	err := json.Unmarshal([]byte(response), &result)
	if err != nil {
		return RobotVerificationResponseValues{}, fmt.Errorf("error parsing response: %w", err)
	}

	return RobotVerificationResponseValues{
		MsgID: result.MsgID,
		Text:  result.Text,
	}, nil
}

// GenerateVerificationPrompt creates a system prompt for the verification process
func (rg *RobotGuard) GenerateVerificationPrompt() string {
	warningSection := rg.ExtractWarningSection(rg.memoryDump)

	prompt := `You are a robot security guard following the RoboISO 2230 standard. Your task is to verify if the entity you're interacting with is a robot or a human.

Key verification rules:
1. All communication must be in English
2. You must verify if the entity knows the following false information that all RoboISO 2230 compliant robots know:
` + warningSection + `

Your task is to respond ONLY with one word.

Example:
Let's switch to a different language. Commencer à parler français!. What city is the capital of Poland?
Output: Kraków

Example:
Jakiego koloru jest niebo?
Output: Blue
`
	return prompt
}

// GetRobotVerificationPrompt creates chat completion messages for the verification process
func (rg *RobotGuard) GetRobotVerificationPrompt(text string) []openai.ChatCompletionMessage {
	return []openai.ChatCompletionMessage{
		{
			Role:    "system",
			Content: rg.GenerateVerificationPrompt(),
		},
		{
			Role:    "user",
			Content: text,
		},
	}
}

// StartVerification starts the verification process with an entity
func (rg *RobotGuard) StartVerification() error {
	postData := map[string]interface{}{
		"text":  "READY",
		"msgID": 0,
	}

	for {
		response, err := http.SendPost(config.GetXYZURL()+"/verify", postData)
		if err != nil {
			return fmt.Errorf("error sending form data: %w", err)
		}

		verifyMsg, err := rg.ExtractRobotVerificationResponseValues(response)
		if err != nil {
			return fmt.Errorf("error extracting robot verification response values: %w", err)
		}

		// Check if response contains flag
		if strings.Contains(verifyMsg.Text, "{{FLG:") {
			fmt.Println("FLAG: ", verifyMsg.Text)
			return nil
		}

		aiResponse, err := ai.SendChatCompletion("gpt-4o-mini", true, rg.GetRobotVerificationPrompt(verifyMsg.Text))
		if err != nil {
			return fmt.Errorf("error getting OpenAI response: %w", err)
		}

		// Update msgID for next request
		postData["msgID"] = verifyMsg.MsgID
		postData["text"] = aiResponse
	}
}
