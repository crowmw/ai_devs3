package main

import (
	"fmt"

	"github.com/crowmw/ai_devs3/pkg/ai"
	"github.com/crowmw/ai_devs3/pkg/config"
	"github.com/crowmw/ai_devs3/pkg/http"
)

type RobotDescriptionResult struct {
	Description string `json:"description"`
}

func main() {
	// Load environment variables from .env file
	if err := config.LoadEnv(); err != nil {
		fmt.Println(err)
		return
	}

	var robotDescriptionResult RobotDescriptionResult
	err := http.FetchJSONData(config.GetC3ntralaURL()+"/data/"+config.GetMyAPIKey()+"/robotid.json", &robotDescriptionResult)
	if err != nil {
		fmt.Println("Error fetching robot description:", err)
		return
	}

	fmt.Println("Robot description:", robotDescriptionResult.Description)

	// Create a detailed prompt for DALL-E
	prompt := fmt.Sprintf("Create a detailed, high-quality image of a robot with the following characteristics: %s. The image should be photorealistic, with high attention to detail and professional lighting.", robotDescriptionResult.Description)

	// Generate image using DALL-E 3
	imageURL, err := ai.GenerateImageWithDalle(prompt)
	if err != nil {
		fmt.Println("Error generating image:", err)
		return
	}

	fmt.Println("Generated image URL:", imageURL)

	report, err := http.SendC3ntralaReport("robotid", imageURL)
	if err != nil {
		fmt.Println("Error sending report:", err)
		return
	}

	fmt.Println("Report:", report)
}
