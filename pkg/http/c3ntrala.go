package http

import (
	"fmt"

	"github.com/crowmw/ai_devs3/pkg/config"
)

func SendC3ntralaReport(task string, answer interface{}) (string, error) {
	fmt.Println("Sending report to C3ntrala...")

	postData := map[string]interface{}{
		"task":   task,
		"answer": answer,
		"apikey": config.GetMyAPIKey(),
	}

	resp, err := SendPost(
		config.GetC3ntralaURL()+"/report",
		postData,
	)

	return resp, err
}
