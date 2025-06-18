package serce_agent

import (
	"time"
)

type State struct {
	Memory      string      `json:"memory"`
	Tool        string      `json:"tool"`
	Tools       []Tool      `json:"tools"`
	Payload     interface{} `json:"payload"`
	Messages    []string    `json:"messages"`
	UserMessage string      `json:"user_message"`
}

type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Instruction string `json:"instruction"`
}

type ActionResult struct {
	Status string      `json:"status"`
	Answer interface{} `json:"answer"`
}

type Action struct {
	Uuid        string                 `json:"uuid"`
	TaskUuid    string                 `json:"task_uuid"`
	ToolName    string                 `json:"tool_name"`
	Payload     map[string]interface{} `json:"payload"`
	Result      *ActionResult          `json:"result"`
	Status      string                 `json:"status"` // "pending", "completed", "failed"
	Sequence    int                    `json:"sequence"`
	Description string                 `json:"description"`
}

type Task struct {
	Uuid              string
	Conversation_uuid string
	Status            string
	Name              string
	Actions           []Action
	Description       string
	Created_at        time.Time
	Updated_at        time.Time
}

type Config struct {
	MaxSteps int     `json:"max_steps"`
	Step     int     `json:"step"`
	Task     *string `json:"task"`
	Action   *string `json:"action"`
	Tools    []Tool  `json:"tools"`
}

type Thoughts struct {
	Tools string `json:"tools"`
}

// ToolsAnalysis represents a single tool analysis result
type ToolsAnalysis struct {
	Payload interface{} `json:"payload"`
	Tool    string      `json:"tool"`
}

// ToolsAnalysisResponse represents the complete analysis response
type ToolsAnalysisResponse struct {
	Thinking string        `json:"_thinking"`
	Result   ToolsAnalysis `json:"result"`
}

type TaskThoughts struct {
	Uuid        interface{} `json:"uuid"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Status      string      `json:"status"`
}

type TaskThoughtsResponse struct {
	Thinking string         `json:"_thinking"`
	Result   []TaskThoughts `json:"result"`
}

type ActionThoughts struct {
	Description string      `json:"description"`
	ToolName    string      `json:"tool_name"`
	TaskUuid    interface{} `json:"task_uuid"`
}

type ActionThoughtsResponse struct {
	Thinking string         `json:"_thinking"`
	Result   ActionThoughts `json:"result"`
}

type UseThoughtsResponse struct {
	Thinking string                 `json:"_thinking"`
	Result   map[string]interface{} `json:"result"`
}

type FinalAnswer struct {
	Answer map[string]interface{} `json:"answer"`
}

type FinalAnswerResponse struct {
	Thinking string      `json:"_thinking"`
	Result   FinalAnswer `json:"result"`
}

type ToolResult struct {
	Status string      `json:"status"`
	Answer interface{} `json:"answer"`
}

type ToolHandler func(payload map[string]interface{}) (ToolResult, error)

type GpsPayload struct {
	UserID string `json:"userID"`
}

type PersonIDFinderPayload struct {
	Name string `json:"name"`
}

type PersonFinderPayload struct {
	City string `json:"city"`
}
