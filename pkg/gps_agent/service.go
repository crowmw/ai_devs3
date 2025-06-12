package gps_agent

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/crowmw/ai_devs3/pkg/ai"
	"github.com/crowmw/ai_devs3/pkg/c3ntrala"
	"github.com/crowmw/ai_devs3/pkg/env"
	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
)

type Service struct {
	State       State
	envSvc      *env.Service
	aiSvc       *ai.Service
	c3ntralaSvc *c3ntrala.Service
}

func NewService(envSvc *env.Service, aiSvc *ai.Service, c3ntralaSvc *c3ntrala.Service, initSystemMessages []openai.ChatCompletionMessage) (*Service, error) {
	fmt.Println("\n🚀 [AGENT] Initializing GPS Agent Service...")

	state := State{
		Config: Config{
			MaxSteps: 10,
			Step:     0,
			Task:     nil,
			Action:   nil,
			Tools:    getTools(),
		},
		Tasks:     []Task{},
		Tools:     getTools(),
		Documents: []string{},
		Messages:  initSystemMessages,
	}

	fmt.Println("✅ [AGENT] Service initialized successfully")
	return &Service{envSvc: envSvc, aiSvc: aiSvc, c3ntralaSvc: c3ntralaSvc, State: state}, nil
}

func (s *Service) Execute(userMessage string) (interface{}, error) {
	fmt.Println("\n🔄 [AGENT] Starting new execution cycle...")
	fmt.Printf("📝 [AGENT] Processing user message: %s\n", userMessage)

	s.State.UserMessage = userMessage

	s.State.Config.Step = 0
	s.State.Messages = append(s.State.Messages, openai.ChatCompletionMessage{
		Role:    "user",
		Content: userMessage})

	fmt.Println("\n🤔 [AGENT] Starting initial thinking phase...")
	s.executeThinkingPhase(userMessage)

	for s.State.Config.Step < s.State.Config.MaxSteps {
		fmt.Printf("\n📍 [AGENT] Starting step %d of %d\n", s.State.Config.Step+1, s.State.Config.MaxSteps)
		s.executePlanningPhase(userMessage)
		s.executeActionPhase(userMessage)

		currentTask := s.getCurrentTask()
		currentAction := s.getCurrentAction(currentTask)

		if currentAction == nil || currentAction.Result == nil {
			fmt.Println("\n⚠️ [AGENT] No valid action result found")
			s.State.Config.Step++
			continue
		}

		if currentAction.ToolName == "final_answer" {
			fmt.Println("\n✅ [AGENT] Execution cycle completed successfully")
			return currentAction.Result.Data, nil
		}

		if currentTask != nil {
			currentTask.Status = "completed"
			fmt.Printf("✅ [AGENT] Completed task: %s\n", currentTask.Name)

			var nextTask *Task
			for i := range s.State.Tasks {
				if s.State.Tasks[i].Status == "pending" {
					nextTask = &s.State.Tasks[i]
					break
				}
			}

			if nextTask != nil {
				fmt.Printf("➡️ [AGENT] Moving to next task: %s\n", nextTask.Name)
				s.State.Config.Task = &nextTask.Uuid
				if len(nextTask.Actions) > 0 {
					s.State.Config.Action = &nextTask.Actions[0].Uuid
				} else {
					s.State.Config.Action = nil
				}
			} else {
				fmt.Println("📌 [AGENT] No more pending tasks")
				s.State.Config.Task = nil
				s.State.Config.Action = nil
			}
		}
		s.State.Config.Step++
	}

	fmt.Println("❌ [AGENT] Max steps reached without finding an answer")
	return nil, errors.New("no answer found")
}

func (s *Service) executeThinkingPhase(question string) {
	fmt.Println("\n🔍 [THINKING] Analyzing available tools...")

	toolsAnalysisResponse, err := s.aiSvc.ChatCompletion(ai.ChatCompletionConfig{
		Model: "gpt-4.1",
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: getToolsPrompt(&s.State)},
			{Role: openai.ChatMessageRoleUser, Content: question},
		},
		JSONMode: true,
	})

	toolsAnalysis, err := unmarshalToolsAnalysis(toolsAnalysisResponse.Choices[0].Message.Content)
	if err != nil {
		fmt.Printf("❌ [THINKING] Error analyzing tools: %v\n", err)
		return
	}

	fmt.Printf("✅ [THINKING] Tools analysis completed. Found %d relevant tools\n", len(toolsAnalysis.Result))
	for _, tool := range toolsAnalysis.Result {
		fmt.Printf("   📌 Tool: %s, Query: %s\n", tool.Tool, tool.Query)
	}

	var toolsStr []string
	for _, tool := range toolsAnalysis.Result {
		toolsStr = append(toolsStr, fmt.Sprintf(`{"query": "%s", "tool": "%s"}`, tool.Query, tool.Tool))
	}
	s.State.Thoughts.Tools = strings.Join(toolsStr, "\n")
}

func (s *Service) executePlanningPhase(userMessage string) {
	fmt.Println("\n📋 [PLANNING] Creating execution plan...")

	taskThoughtsResponse, err := s.aiSvc.ChatCompletion(ai.ChatCompletionConfig{
		Model: "gpt-4.1",
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: getTaskThoughtsPrompt(&s.State)},
			{Role: openai.ChatMessageRoleUser, Content: userMessage},
		},
		JSONMode: true,
	})

	taskThoughts, err := unmarshalTaskThoughts(taskThoughtsResponse.Choices[0].Message.Content)
	if err != nil {
		fmt.Printf("❌ [PLANNING] Error analyzing tasks: %v\n", err)
		return
	}

	fmt.Printf("✅ [PLANNING] Task analysis completed. Found %d tasks to process\n", len(taskThoughts.Result))
	for _, thought := range taskThoughts.Result {
		fmt.Printf("   📌 Task: %s - %s\n", thought.Name, thought.Description)
	}

	for _, thought := range taskThoughts.Result {
		if thought.Uuid != nil {
			for i, existingTask := range s.State.Tasks {
				if existingTask.Uuid == thought.Uuid && existingTask.Status == "pending" {
					fmt.Printf("🔄 [PLANNING] Updating existing task: %s\n", thought.Name)
					s.State.Tasks[i].Name = thought.Name
					s.State.Tasks[i].Description = thought.Description
					s.State.Tasks[i].Updated_at = time.Now()
					break
				}
			}
		} else {
			fmt.Printf("➕ [PLANNING] Creating new task: %s\n", thought.Name)
			newTask := Task{
				Uuid:              uuid.New().String(),
				Conversation_uuid: uuid.New().String(),
				Status:            "pending",
				Name:              thought.Name,
				Description:       thought.Description,
				Actions:           []Action{},
				Created_at:        time.Now(),
				Updated_at:        time.Now(),
			}
			s.State.Tasks = append(s.State.Tasks, newTask)
		}
	}

	for _, task := range s.State.Tasks {
		if task.Status == "pending" {
			fmt.Printf("➡️ [PLANNING] Selected first pending task: %s\n", task.Name)
			s.State.Config.Task = &task.Uuid
			break
		}
	}

	fmt.Println("\n🎯 [PLANNING] Planning actions for current task...")
	actionThoughtsPrompt := getActionThoughtsPrompt(&s.State)
	actionThoughtsResponse, err := s.aiSvc.ChatCompletion(ai.ChatCompletionConfig{
		Model: "gpt-4.1",
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: actionThoughtsPrompt},
			{Role: openai.ChatMessageRoleUser, Content: userMessage},
		},
		JSONMode: true,
	})

	actionThoughts, err := unmarshalActionThoughts(actionThoughtsResponse.Choices[0].Message.Content)
	if err != nil {
		fmt.Printf("❌ [PLANNING] Error analyzing actions: %v\n", err)
		return
	}

	fmt.Printf("✅ [PLANNING] Action analysis completed for task: %s\n", actionThoughts.Result.Description)

	if actionThoughts != nil {
		var task *Task
		for i := range s.State.Tasks {
			if s.State.Tasks[i].Uuid == actionThoughts.Result.TaskUuid {
				task = &s.State.Tasks[i]
				break
			}
		}

		if task != nil {
			fmt.Printf("➕ [PLANNING] Creating new action for task '%s': %s using tool '%s'\n",
				task.Name, actionThoughts.Result.Description, actionThoughts.Result.ToolName)

			newAction := Action{
				Uuid:        uuid.New().String(),
				TaskUuid:    task.Uuid,
				Description: actionThoughts.Result.Description,
				ToolName:    actionThoughts.Result.ToolName,
				Status:      "pending",
				Result:      nil,
				Payload:     make(map[string]interface{}),
			}

			task.Actions = []Action{newAction}
			task.Updated_at = time.Now()
			s.State.Config.Task = &task.Uuid
			s.State.Config.Action = &newAction.Uuid
		}
	}

	currentTask := s.getCurrentTask()
	currentAction := s.getCurrentAction(currentTask)

	fmt.Println("\n📊 [PLANNING] Current execution state:")
	fmt.Printf("   Current Task: %s\n", func() string {
		if currentTask != nil {
			return currentTask.Name
		}
		return "None"
	}())
	fmt.Printf("   Current Action: %s\n", func() string {
		if currentAction != nil {
			return currentAction.Description
		}
		return "None"
	}())
}

func (s *Service) executeActionPhase(userMessage string) {
	if s.State.Config.Task == nil || s.State.Config.Action == nil {
		fmt.Println("⚠️ [ACTION] No task or action to execute")
		return
	}

	fmt.Println("\n⚡ [ACTION] Preparing to execute action...")
	systemMessage := s.getUseThoughtsPrompt()
	useThoughtsResponse, err := s.aiSvc.ChatCompletion(ai.ChatCompletionConfig{
		Model: "gpt-4.1",
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemMessage},
			{Role: openai.ChatMessageRoleUser, Content: userMessage},
		},
		JSONMode: true,
	})

	useThoughts, err := unmarshalUseThoughts(useThoughtsResponse.Choices[0].Message.Content)
	if err != nil {
		fmt.Printf("❌ [ACTION] Error analyzing tool usage: %v\n", err)
		return
	}

	currentTask := s.getCurrentTask()
	fmt.Printf("🔄 [ACTION] Executing task: %s\n", currentTask.Name)
	currentAction := s.getCurrentAction(currentTask)

	if currentTask == nil || currentAction == nil {
		fmt.Println("❌ [ACTION] Could not find current task or action")
		return
	}

	currentAction.Payload = useThoughts.Result
	fmt.Printf("🛠️ [ACTION] Using tool '%s' with payload: %s\n", currentAction.ToolName, prettyPrint(useThoughts.Result))

	toolHandler, ok := s.getToolHandlers()[currentAction.ToolName]
	if !ok {
		fmt.Printf("❌ [ACTION] Unknown tool: %s\n", currentAction.ToolName)
		return
	}

	toolResult, err := toolHandler(useThoughts.Result)
	if err != nil {
		fmt.Printf("❌ [ACTION] Error executing tool: %v\n", err)
		return
	}

	currentAction.Result = &ActionResult{
		Status: "completed",
		Data:   toolResult.Data,
	}
	currentAction.Status = "completed"
	fmt.Printf("✅ [ACTION] Successfully completed action: %s\n", currentAction.Description)
}
