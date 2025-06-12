package gps_agent

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

// unmarshalTaskThoughts parses the task thoughts response
func unmarshalTaskThoughts(content string) (*TaskThoughtsResponse, error) {
	var taskThoughts TaskThoughtsResponse
	err := json.Unmarshal([]byte(content), &taskThoughts)
	if err != nil {
		return nil, err
	}
	return &taskThoughts, nil
}

// unmarshalActionThoughts parses the action thoughts response
func unmarshalActionThoughts(content string) (*ActionThoughtsResponse, error) {
	var actionThoughts ActionThoughtsResponse
	err := json.Unmarshal([]byte(content), &actionThoughts)
	if err != nil {
		return nil, err
	}
	return &actionThoughts, nil
}

func unmarshalUseThoughts(content string) (*UseThoughtsResponse, error) {
	var useThoughts UseThoughtsResponse
	err := json.Unmarshal([]byte(content), &useThoughts)
	if err != nil {
		return nil, err
	}
	return &useThoughts, nil
}

func unmarcharFinalAnswer(content string) (*FinalAnswerResponse, error) {
	var finalAnswer FinalAnswerResponse
	err := json.Unmarshal([]byte(content), &finalAnswer)
	if err != nil {
		return nil, err
	}
	return &finalAnswer, nil
}
