package main

import (
	"fmt"
	"os"

	"github.com/crowmw/ai_devs3/pkg/config"
	"github.com/crowmw/ai_devs3/pkg/http"
	"github.com/crowmw/ai_devs3/pkg/processor"
)

func main() {
	// Load environment variables
	if err := config.LoadEnv(); err != nil {
		fmt.Println(err)
		return
	}

	// Fetch data
	data, err := http.FetchData(config.GetPoligonURL() + "/dane.txt")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Process data
	filteredLines := processor.ReadLinesFromTextFile(data)

	// Display processed data
	fmt.Println("Pobrane dane w formie tablicy stringów:")
	for i, line := range filteredLines {
		fmt.Printf("Linia %d: %s\n", i+1, line)
	}

	// Get API key
	apiKey := os.Getenv("POLIGON_API_KEY")
	if apiKey == "" {
		fmt.Println("POLIGON_API_KEY environment variable is not set")
		return
	}

	// Send POST request
	postData := map[string]interface{}{
		"task":   "POLIGON",
		"answer": filteredLines,
		"apikey": apiKey,
	}
	response, err := http.SendPost(config.GetPoligonURL()+"/verify", postData)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Odpowiedź z serwera po wysłaniu POST:")
	fmt.Println(response)
}
