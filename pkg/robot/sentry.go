package robot

import (
	"fmt"

	"github.com/crowmw/ai_devs3/pkg/ai"
	"github.com/crowmw/ai_devs3/pkg/config"
	"github.com/crowmw/ai_devs3/pkg/http"
)

// RobotDescriptionResult represents the structure of the robot description JSON
type RobotDescriptionResult struct {
	Description string `json:"description"`
}

// SentryRobot represents a sentry robot with its description and generated image
type SentryRobot struct {
	Description RobotDescriptionResult
	ImageURL    string
}

// NewSentryRobot creates a new SentryRobot instance and initializes it with data
func NewSentryRobot() (*SentryRobot, error) {
	// Fetch robot description
	var robotDescriptionResult RobotDescriptionResult
	err := http.FetchJSONData(config.GetC3ntralaURL()+"/data/"+config.GetMyAPIKey()+"/robotid.json", &robotDescriptionResult)
	if err != nil {
		return nil, fmt.Errorf("error fetching robot description: %w", err)
	}

	// Create a detailed prompt for DALL-E
	prompt := fmt.Sprintf("Create a detailed, high-quality image of a robot with the following characteristics: %s. The image should be photorealistic, with high attention to detail and professional lighting.", robotDescriptionResult.Description)

	// Generate image using DALL-E 3
	imageURL, err := ai.GenerateImageWithDalle(prompt)
	if err != nil {
		return nil, fmt.Errorf("error generating image: %w", err)
	}

	return &SentryRobot{
		Description: robotDescriptionResult,
		ImageURL:    imageURL,
	}, nil
}
