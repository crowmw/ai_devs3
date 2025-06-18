package serce_agent

import "encoding/json"

// unmarshalToolsAnalysis parses the tools analysis response
func unmarshalToolsAnalysis(content string) (*ToolsAnalysisResponse, error) {
	var analysis ToolsAnalysisResponse
	err := json.Unmarshal([]byte(content), &analysis)
	if err != nil {
		return nil, err
	}
	return &analysis, nil
}
