package gps_agent

import (
	"encoding/json"
	"fmt"

	"github.com/crowmw/ai_devs3/pkg/config"
	"github.com/crowmw/ai_devs3/pkg/http"
)

// GetTools returns the list of available tools
func getTools() []Tool {
	return []Tool{
		{
			Name:        "person_finder",
			Description: "Returns an array of person names who were seen in a specified city",
			Instruction: `Provide a JSON payload with "city" field containing the name of the city you want to get persons list for, like: {"city": "Warszawa"}. This will return an array of person names who were seen in that city.`,
		},
		{
			Name:        "person_id_finder",
			Description: "Returns the userID for a given person's name. Takes a name as input and returns the corresponding userID.",
			Instruction: `Provide a JSON payload with "name" field containing the name of the person you want to get ID for, like: {"name": "Azazel"}`,
		},
		{
			Name:        "gps",
			Description: "Returns GPS coordinates for a person when given their userID. STRICTLY ONE numeric userID per call. Do NOT include lat/lon or multiple IDs.",
			Instruction: `Provide a JSON payload with ONLY "userID" field containing ONE numeric ID, like: {"userID": "69"}. NEVER include any other keys (e.g., "answer", "lat", "lon"). NEVER send arrays or comma-separated IDs.`,
		},
		{
			Name:        "final_answer",
			Description: "Use this to answer the user question",
			Instruction: `Provide a JSON payload with "answer" field containing the original question from the user, like: {"answer": "Where was John last seen?"}`,
		},
	}
}

func (s *Service) handlePersonFinder(payload map[string]interface{}) (ToolResult, error) {
	fmt.Println("\n🔍 [PERSON_FINDER] Starting search for people...")

	var personFinderPayload PersonFinderPayload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("❌ [PERSON_FINDER] Error marshaling payload: %v\n", err)
		return ToolResult{
			Status: "error",
			Data:   fmt.Sprintf("error marshaling payload: %v", err),
		}, err
	}

	if err := json.Unmarshal(payloadBytes, &personFinderPayload); err != nil {
		fmt.Printf("❌ [PERSON_FINDER] Error unmarshaling payload: %v\n", err)
		return ToolResult{
			Status: "error",
			Data:   fmt.Sprintf("error unmarshaling payload: %v", err),
		}, err
	}

	fmt.Printf("📍 [PERSON_FINDER] Searching in city: %s\n", personFinderPayload.City)
	persons, err := s.c3ntralaSvc.GetWhoWasSeenThere(personFinderPayload.City)
	if err != nil {
		fmt.Printf("❌ [PERSON_FINDER] Error getting people: %v\n", err)
		return ToolResult{
			Status: "error",
			Data:   err.Error(),
		}, err
	}

	type Result struct {
		City    string   `json:"city"`
		Persons []string `json:"persons"`
	}

	Data := Result{City: personFinderPayload.City, Persons: persons}
	fmt.Printf("✅ [PERSON_FINDER] Found %d people in %s: %v\n", len(persons), personFinderPayload.City, persons)

	return ToolResult{
		Status: "success",
		Data:   Data,
	}, nil
}

func (s *Service) handlePersonIDFinder(payload map[string]interface{}) (ToolResult, error) {
	fmt.Println("\n🆔 [PERSON_ID_FINDER] Starting ID lookup...")

	var personIDFinderPayload PersonIDFinderPayload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("❌ [PERSON_ID_FINDER] Error marshaling payload: %v\n", err)
		return ToolResult{
			Status: "error",
			Data:   fmt.Sprintf("error marshaling payload: %v", err),
		}, err
	}

	if err := json.Unmarshal(payloadBytes, &personIDFinderPayload); err != nil {
		fmt.Printf("❌ [PERSON_ID_FINDER] Error unmarshaling payload: %v\n", err)
		return ToolResult{
			Status: "error",
			Data:   fmt.Sprintf("error unmarshaling payload: %v", err),
		}, err
	}

	if personIDFinderPayload.Name == "" {
		fmt.Println("❌ [PERSON_ID_FINDER] Error: Name parameter is required")
		return ToolResult{
			Status: "error",
			Data:   fmt.Sprintf("You need to provide only parameter 'name' with person name in payload"),
		}, err
	}

	fmt.Printf("👤 [PERSON_ID_FINDER] Looking up ID for: %s\n", personIDFinderPayload.Name)
	query := fmt.Sprintf(`SELECT id FROM users WHERE username = "%s"`, personIDFinderPayload.Name)
	result, err := http.PostSQLQueryToAPIDB(s.envSvc, query)
	if err != nil {
		fmt.Printf("❌ [PERSON_ID_FINDER] Database error: %v\n", err)
		return ToolResult{
			Status: "error",
			Data:   err.Error(),
		}, err
	}

	type Reply struct {
		ID string `json:"id"`
	}
	var data struct {
		Reply []Reply `json:"reply"`
		Error string  `json:"error"`
	}
	if err := json.Unmarshal([]byte(result), &data); err != nil {
		fmt.Printf("❌ [PERSON_ID_FINDER] Error parsing database response: %v\n", err)
		return ToolResult{
			Status: "error",
			Data:   fmt.Sprintf("error unmarshaling payload: %v", err),
		}, err
	}

	type Result struct {
		Name   string `json:"name"`
		UserID string `json:"userID"`
	}

	Data := Result{Name: personIDFinderPayload.Name, UserID: data.Reply[0].ID}
	fmt.Printf("✅ [PERSON_ID_FINDER] Found ID for %s: %s\n", Data.Name, Data.UserID)

	return ToolResult{
		Status: "success",
		Data:   Data,
	}, nil
}

func (s *Service) handleGps(payload map[string]interface{}) (ToolResult, error) {
	fmt.Println("\n📍 [GPS] Starting location lookup...")

	var gpsPayload GpsPayload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("❌ [GPS] Error marshaling payload: %v\n", err)
		return ToolResult{
			Status: "error",
			Data:   fmt.Sprintf("error marshaling payload: %v", err),
		}, err
	}

	if err := json.Unmarshal(payloadBytes, &gpsPayload); err != nil {
		fmt.Printf("❌ [GPS] Error unmarshaling payload: %v\n", err)
		return ToolResult{
			Status: "error",
			Data:   fmt.Sprintf("error unmarshaling payload: %v", err),
		}, err
	}

	fmt.Printf("🔍 [GPS] Looking up coordinates for userID: %s\n", gpsPayload.UserID)
	result, err := http.SendJSONPost(config.GetC3ntralaURL()+"/gps", map[string]string{"userID": gpsPayload.UserID})
	if err != nil {
		fmt.Printf("❌ [GPS] API error: %v\n", err)
		return ToolResult{
			Status: "error",
			Data:   err.Error(),
		}, err
	}

	var data struct {
		Code    int `json:"code"`
		Message struct {
			Lat float64 `json:"lat"`
			Lon float64 `json:"lon"`
		} `json:"message"`
	}
	if err := json.Unmarshal([]byte(result), &data); err != nil {
		fmt.Printf("❌ [GPS] Error parsing API response: %v\n", err)
		return ToolResult{
			Status: "error",
			Data:   fmt.Sprintf("error unmarshaling payload: %v", err),
		}, err
	}

	type Result struct {
		UserID string  `json:"userID"`
		Lat    float64 `json:"lat"`
		Long   float64 `json:"long"`
	}

	Data := Result{UserID: gpsPayload.UserID, Lat: data.Message.Lat, Long: data.Message.Lon}
	fmt.Printf("✅ [GPS] Found location for userID %s: lat=%f, lon=%f\n", Data.UserID, Data.Lat, Data.Long)

	return ToolResult{
		Status: "success",
		Data:   Data,
	}, nil
}

func (s *Service) handleFinalAnswer(payload map[string]interface{}) (ToolResult, error) {
	fmt.Println("\n🎯 [FINAL_ANSWER] Preparing final response...")
	fmt.Printf("📝 [FINAL_ANSWER] Original question: %s\n", s.State.UserMessage)

	// First try to get answer from result.answer structure
	result, hasResult := payload["result"]
	if hasResult {
		resultMap, ok := result.(map[string]interface{})
		if !ok {
			return ToolResult{
				Status: "error",
				Data:   "'result' field must be an object",
			}, fmt.Errorf("'result' field must be an object")
		}

		answer, hasAnswer := resultMap["answer"]
		if !hasAnswer {
			return ToolResult{
				Status: "error",
				Data:   "missing 'answer' field in result",
			}, fmt.Errorf("missing 'answer' field in result")
		}

		fmt.Printf("✅ [FINAL_ANSWER] Generated response: %s\n", prettyPrint(answer))
		return ToolResult{
			Status: "success",
			Data:   answer,
		}, nil
	}

	// If no result wrapper, try direct answer field
	answer, hasAnswer := payload["answer"]
	if !hasAnswer {
		return ToolResult{
			Status: "error",
			Data:   "missing either 'result.answer' or 'answer' field in payload",
		}, fmt.Errorf("missing either 'result.answer' or 'answer' field in payload")
	}

	fmt.Printf("✅ [FINAL_ANSWER] Generated response: %s\n", prettyPrint(answer))
	return ToolResult{
		Status: "success",
		Data:   answer,
	}, nil
}

func (s *Service) getToolHandlers() map[string]ToolHandler {
	return map[string]ToolHandler{
		"gps":              s.handleGps,
		"person_id_finder": s.handlePersonIDFinder,
		"person_finder":    s.handlePersonFinder,
		"final_answer":     s.handleFinalAnswer,
	}
}
