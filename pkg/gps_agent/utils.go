package gps_agent

import (
	"encoding/json"
	"fmt"
)

func prettyPrint(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%+v", v)
	}
	return string(b)
}

// getCurrentTask returns the current task based on Config.Task
func (s *Service) getCurrentTask() *Task {
	if s.State.Config.Task == nil {
		return nil
	}
	for i := range s.State.Tasks {
		if s.State.Tasks[i].Uuid == *s.State.Config.Task {
			return &s.State.Tasks[i]
		}
	}
	return nil
}

// getCurrentAction returns the current action for the given task
func (s *Service) getCurrentAction(task *Task) *Action {
	if task == nil || s.State.Config.Action == nil {
		return nil
	}
	for i := range task.Actions {
		if task.Actions[i].Uuid == *s.State.Config.Action {
			return &task.Actions[i]
		}
	}
	return nil
}

func (s *Service) getToolByName(toolName string) *Tool {
	for _, tool := range s.State.Tools {
		if tool.Name == toolName {
			return &tool
		}
	}
	return nil
}
