package main

import (
	"fmt"

	"github.com/crowmw/ai_devs3/pkg/config"
	"github.com/crowmw/ai_devs3/pkg/robot"
)

func main() {
	// Load environment variables
	if err := config.LoadEnv(); err != nil {
		fmt.Println(err)
		return
	}

	// Create new robot guard instance
	robot, err := robot.NewIndustrialRobot()
	if err != nil {
		fmt.Printf("Error creating industrial robot: %v\n", err)
		return
	}

	err = robot.Recalibrate()
	if err != nil {
		fmt.Printf("Error during recalibration: %v\n", err)
		return
	}

	err = robot.SendCalibrationReport()
	if err != nil {
		fmt.Printf("Error sending calibration report: %v\n", err)
		return
	}
}
