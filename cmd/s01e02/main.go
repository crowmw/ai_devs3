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
	guard, err := robot.NewRobotGuard()
	if err != nil {
		fmt.Printf("Error creating robot guard: %v\n", err)
		return
	}

	// Start the verification process
	err = guard.StartVerification()
	if err != nil {
		fmt.Printf("Error during verification: %v\n", err)
		return
	}
}
