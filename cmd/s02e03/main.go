package main

import (
	"fmt"

	"github.com/crowmw/ai_devs3/pkg/config"
	"github.com/crowmw/ai_devs3/pkg/http"
	"github.com/crowmw/ai_devs3/pkg/robot"
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

	robot, err := robot.NewSentryRobot()
	if err != nil {
		fmt.Println("Error creating robot:", err)
		return
	}

	report, err := http.SendC3ntralaReport("robotid", robot.ImageURL)
	if err != nil {
		fmt.Println("Error sending report:", err)
		return
	}

	fmt.Println("Report:", report)
}
