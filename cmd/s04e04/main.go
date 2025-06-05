package main

import (
	"fmt"
	"time"

	"encoding/json"
	"log"
	"net/http"

	"github.com/crowmw/ai_devs3/pkg/ai"
	"github.com/crowmw/ai_devs3/pkg/c3ntrala"
	"github.com/crowmw/ai_devs3/pkg/env"
	"github.com/sashabaranov/go-openai"
)

type DroneInstruction struct {
	Instruction string `json:"instruction"`
}

func handleDroneMovement(w http.ResponseWriter, r *http.Request, aiSvc *ai.Service) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var instruction DroneInstruction
	if err := json.NewDecoder(r.Body).Decode(&instruction); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	fmt.Println("Instruction:", instruction.Instruction)

	response, err := aiSvc.ChatCompletion(ai.ChatCompletionConfig{
		Model: "gpt-4.1",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "system",
				Content: fmt.Sprintf(movementInstructionSystemPrompt, mapDescription),
			},
			{
				Role:    "user",
				Content: instruction.Instruction,
			},
		},
	})
	if err != nil {
		http.Error(w, "Failed to get AI response", http.StatusInternalServerError)
		return
	}

	var landingPosition struct {
		Thinking    string `json:"_thinking"`
		Description string `json:"description"`
	}
	json.Unmarshal([]byte(response.Choices[0].Message.Content), &landingPosition)

	fmt.Println("AI response:", landingPosition)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"description": landingPosition.Description,
	})
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

	c3ntralaSvc, err := c3ntrala.NewService(envSvc)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Set up HTTP server
	http.HandleFunc("/drone", func(w http.ResponseWriter, r *http.Request) {
		handleDroneMovement(w, r, aiSvc)
	})

	log.Println("Starting HTTP server on :8080")

	// Start server in a goroutine
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal("Server error:", err)
		}
	}()

	// Wait a bit for server to start
	time.Sleep(2 * time.Second)

	centralaResponse, err := c3ntralaSvc.PostReport("webhook", envSvc.GetNGrokURL()+"/drone")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Centrala response:", centralaResponse)

	// Keep main goroutine alive
	select {}
}

const mapDescription = `
Oto opis mapy 4x4 jako tekstowa reprezentacja siatki z zaznaczonymi elementami:

Mapa jest podzielona na 16 pól, ułożonych w siatkę 4x4. 
Każde pole jest numerowane zgodnie z wierszami i kolumnami, zaczynając od lewego górnego rogu jako pola (1,1). 
Mapa wygląda następująco:

Pierwszy wiersz (od lewej do prawej):
Pole (1,1): Dron (startowa pozycja drona)
Pole (1,2): Puste
Pole (1,3): Drzewo
Pole (1,4): Budynek
Drugi wiersz (od lewej do prawej):
Pole (2,1): Puste
Pole (2,2): Wiatrak
Pole (2,3): Puste
Pole (2,4): Puste
Trzeci wiersz (od lewej do prawej):
Pole (3,1): Puste
Pole (3,2): Puste
Pole (3,3): Skały
Pole (3,4): Drzewa
Czwarty wiersz (od lewej do prawej):
Pole (4,1): Góry
Pole (4,2): Góry
Pole (4,3): Samochód (docelowa pozycja drona)
Pole (4,4): Jaskinia 

Na tej mapie startowa pozycja drona znajduje się na polu (1,1), docelowa pozycja drona znajduje się na polu (4,3). 
Można sobie wyobrazić, że dron może przemieszczać się pomiędzy tymi polami zgodnie z ruchem, który mu się ustawi.
`

const movementInstructionSystemPrompt = `
	You are a helpful assistant that helps with movement instructions for a drone.
	Based on the map description provided in <map_description> and the user's instruction, you need to guess the movement of the drone, and landing position.
	You need to guess landing place.
	First, separate instructions into movements.
	If user specified number of fields to move, use that number.
	If user specified direction, use that direction but move only one field.
	If user specified something like most Top, most bottom, most left, most right, use that direction.
	If user specified both, use both.
	If user corrects their path or mentions mistakes, consider only the final corrected path.
	For example, if user says "I went right, oh no, I meant left, then down", consider only the path "left, then down".
	Analyze the full context of the instruction to determine the actual intended path.
	Second, guess landing place based on movements.
	Use one or two words based on field description NOT the field number.
	Return only JSON object with _thinking property and landing place description. 
	<response_example>
	{
		"_thinking": "Ruszam się jedno pole w prawo, a potem na ostatnie pole w dół",
		"description": "Skały"
	}
	</response_example>

	<map_description>
		%s
	</map_description>

	<example>
	User: ""poleciałem jedno pole w prawo, a później na sam dół""
	You: {
		"_thinking": "Ruszam się jedno pole w prawo, a potem na ostatnie pole w dół, więc docelowa pozycja to skały",
		"description": "Skały"
	}
	</example>


`
