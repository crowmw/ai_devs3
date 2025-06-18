package serce_agent

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

func (s *Service) getToolByName(toolName string) *Tool {
	for _, tool := range s.State.Tools {
		if tool.Name == toolName {
			return &tool
		}
	}
	return nil
}
