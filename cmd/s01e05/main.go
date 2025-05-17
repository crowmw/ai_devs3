package main

import (
	"fmt"

	"github.com/crowmw/ai_devs3/pkg/config"
)

func main() {
	// Load environment variables
	if err := config.LoadEnv(); err != nil {
		fmt.Println(err)
		return
	}
}
