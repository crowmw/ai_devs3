package main

import (
	"fmt"

	"github.com/crowmw/ai_devs3/pkg/env"
	"github.com/crowmw/ai_devs3/pkg/factory"
	"github.com/crowmw/ai_devs3/pkg/http"
)

func main() {
	envSvc, err := env.NewService()
	if err != nil {
		fmt.Println(err)
		return
	}

	f, err := factory.NewFactory(envSvc)
	if err != nil {
		fmt.Println(err)
		return
	}

	aiResponse, err := f.AnalyzeReports()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(aiResponse)

	// Send report
	result, err := http.SendC3ntralaReport("dokumenty", aiResponse)
	if err != nil {
		fmt.Println("Error sending report:", err)
		return
	}

	fmt.Println("Result:", result)
}
