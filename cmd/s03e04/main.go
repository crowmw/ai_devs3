package main

import (
	"encoding/json"
	"fmt"

	"github.com/crowmw/ai_devs3/pkg/ai"
	"github.com/crowmw/ai_devs3/pkg/c3ntrala"
	"github.com/crowmw/ai_devs3/pkg/env"
	"github.com/sashabaranov/go-openai"
)

func main() {
	envSvc, err := env.NewService()
	if err != nil {
		fmt.Println(err)
		return
	}

	c3ntralaSvc, err := c3ntrala.NewService(envSvc)
	if err != nil {
		fmt.Println(err)
		return
	}

	barbaraNote, err := c3ntralaSvc.GetBarbaraNote()
	if err != nil {
		fmt.Println(err)
		return
	}

	aiSvc, err := ai.NewService(envSvc)
	if err != nil {
		fmt.Println(err)
		return
	}

	model := "gpt-4o"

	messages := []openai.ChatCompletionMessage{
		{Role: "system", Content: extractNamesAndCitiesPrompt},
		{Role: "user", Content: barbaraNote},
	}

	aiExtractResponse, err := aiSvc.ChatCompletion(ai.ChatCompletionConfig{
		Model:    model,
		Messages: messages,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(aiExtractResponse.Choices[0].Message.Content)

	type ExtractedData struct {
		Names  []string `json:"names"`
		Cities []string `json:"cities"`
	}

	var extractedData ExtractedData
	if err := json.Unmarshal([]byte(aiExtractResponse.Choices[0].Message.Content), &extractedData); err != nil {
		fmt.Println("Error parsing extracted data:", err)
		return
	}

	// Keep searching until we find BARBARA
	var lastPlace string
	processedNames := make(map[string]bool)  // Track processed names
	processedCities := make(map[string]bool) // Track processed cities

	// First, process all names to get cities
	for _, name := range extractedData.Names {
		if name == "" || processedNames[name] {
			fmt.Println("Skipping name", name)
			// continue
		}
		processedNames[name] = true

		places, err := c3ntralaSvc.GetPlacesWhereSeen(name)
		if err != nil {
			fmt.Println("ERROR GetPlacesWhereSeen", err)
			return
		}

		fmt.Println("🚶", name, "👀", places)
		fmt.Println("PROCESSED NAMES", processedNames)

		// Add new places to cities list
		for _, place := range places {
			if place != "" && !processedCities[place] {
				extractedData.Cities = append(extractedData.Cities, place)
				processedCities[place] = true
			}
		}
		fmt.Println("🏙️", extractedData.Cities)
	}

	// Then, process all cities to get names and check for BARBARA
	for _, city := range extractedData.Cities {
		if city == "" || processedCities[city] {
			fmt.Println("Skipping city", city)
			// continue
		}
		processedCities[city] = true

		peoples, err := c3ntralaSvc.GetWhoWasSeenThere(city)
		if err != nil {
			fmt.Println("ERROR GetWhoWasSeenThere", err)
			return
		}

		fmt.Println("🌆", city, "👀", peoples)
		fmt.Println("PROCESSED CITIES", processedCities)

		// Check for BARBARA first
		for _, person := range peoples {
			if person == "BARBARA" && city != "WARSZAWA" && city != "KRAKOW" {
				lastPlace = city
				fmt.Println("Found BARBARA in:", lastPlace)
				break
			}
		}

		// Add new names to the list
		for _, person := range peoples {
			if person != "" && !processedNames[person] {
				extractedData.Names = append(extractedData.Names, person)
				processedNames[person] = true
			}
		}
		fmt.Println("👥", extractedData.Names)
	}

	// After processing all cities, process any new names that were added
	for _, name := range extractedData.Names {
		if name == "" || processedNames[name] {
			fmt.Println("Skipping name", name)
			// continue
		}
		processedNames[name] = true

		places, err := c3ntralaSvc.GetPlacesWhereSeen(name)
		if err != nil {
			fmt.Println("ERROR GetPlacesWhereSeen", err)
			return
		}

		fmt.Println("🚶", name, "👀", places)
		fmt.Println("PROCESSED NAMES", processedNames)

		// Add new places to cities list
		for _, place := range places {
			if place != "" && !processedCities[place] {
				extractedData.Cities = append(extractedData.Cities, place)
				processedCities[place] = true
			}
		}
		fmt.Println("🏙️", extractedData.Cities)
	}

	// Then, process all cities to get names and check for BARBARA
	for _, city := range extractedData.Cities {
		if city == "" || processedCities[city] {
			fmt.Println("Skipping city", city)
			// continue
		}
		processedCities[city] = true

		peoples, err := c3ntralaSvc.GetWhoWasSeenThere(city)
		if err != nil {
			fmt.Println("ERROR GetWhoWasSeenThere", err)
			return
		}

		fmt.Println("🌆", city, "👀", peoples)
		fmt.Println("PROCESSED CITIES", processedCities)

		// Check for BARBARA first
		for _, person := range peoples {
			if person == "BARBARA" && city != "WARSZAWA" && city != "KRAKOW" {
				lastPlace = city
				fmt.Println("Found BARBARA in:", lastPlace)
				break
			}
		}

		// Add new names to the list
		for _, person := range peoples {
			if person != "" && !processedNames[person] {
				extractedData.Names = append(extractedData.Names, person)
				processedNames[person] = true
			}
		}
		fmt.Println("👥", extractedData.Names)
	}

	fmt.Println("Last place checked:", lastPlace)
	fmt.Println("Names:", extractedData.Names)
	fmt.Println("Cities:", extractedData.Cities)

	report, err := c3ntralaSvc.PostReport("loop", lastPlace)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("❇️ Report result:", report)
}

const extractNamesAndCitiesPrompt = `You are a helpful assistant that extracts names and cities from a text.

<rules>
- You should extract only first names and cities that are mentioned in the text.
- Make sure the returned first name is in the nominative and without Polish characters like ą, ć, ę, ł, ń, ś, ź, ż. (e.g. RAFAL instead of Rafałowi Bombie).
- Return only the first names and cities in the following format: {"names": ["FIRST_NAME1", "FIRST_NAME2", "FIRST_NAME3"], "cities": ["CITY1", "CITY2", "CITY3"]}. Plain JSON format. No other text or Markdown decorations.
- If there are no names or cities, return {names: [], cities: []}.
- Make sure firstnames and cities is distinct for each group.
- Return only uppercase names and cities and without Polish characters 
- Return only valid names in polish. Fix any issues in names like ALEKSANDER instead of ALEKSANDR.
</rules>

<examples>
USER: [Typical input example]
AI: {"names": ["FIRST_NAME1", "FIRST_NAME2", "FIRST_NAME3"], "cities": ["CITY1", "CITY2", "CITY3"]}
</examples>

Let's begin extracting names and cities from the text.
`
