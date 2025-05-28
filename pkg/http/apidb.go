package http

import (
	"fmt"

	"github.com/crowmw/ai_devs3/pkg/env"
)

func PostSQLQueryToAPIDB(envSvc *env.Service, query string) (string, error) {
	fmt.Println("Sending SQL query to APIDB:", query)
	url := envSvc.GetC3ntralaURL() + "/apidb"

	postData := map[string]interface{}{
		"task":   "database",
		"query":  query,
		"apikey": envSvc.GetMyAPIKey(),
	}

	resp, err := SendPost(
		url,
		postData,
	)

	return resp, err
}
