package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/crowmw/ai_devs3/pkg/c3ntrala"
	"github.com/crowmw/ai_devs3/pkg/env"
	"github.com/crowmw/ai_devs3/pkg/graph"
	"github.com/crowmw/ai_devs3/pkg/http"
)

func main() {
	envSvc, err := env.NewService()
	if err != nil {
		fmt.Println(err)
		return
	}

	users, err := http.PostSQLQueryToAPIDB(envSvc, "SELECT id, username FROM users")
	if err != nil {
		fmt.Println(err)
		return
	}
	var usersResponse struct {
		Reply []struct {
			ID       string `json:"id"`
			Username string `json:"username"`
		} `json:"reply"`
		Error string `json:"error"`
	}

	if err := json.Unmarshal([]byte(users), &usersResponse); err != nil {
		fmt.Println("Error parsing result:", err)
		return
	}

	fmt.Println(len(usersResponse.Reply))

	connections, err := http.PostSQLQueryToAPIDB(envSvc, "SELECT * FROM connections")
	if err != nil {
		fmt.Println(err)
		return
	}

	var connectionsResponse struct {
		Reply []struct {
			User1_id string `json:"user1_id"`
			User2_id string `json:"user2_id"`
		} `json:"reply"`
		Error string `json:"error"`
	}

	if err := json.Unmarshal([]byte(connections), &connectionsResponse); err != nil {
		fmt.Println("Error parsing result:", err)
		return
	}

	fmt.Println(len(connectionsResponse.Reply))

	graphSvc, err := graph.NewService(context.Background(), envSvc)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Graph service created", graphSvc)

	for _, user := range usersResponse.Reply {
		err := graphSvc.CreatePerson(context.Background(), graph.PersonNode{
			OriginalID: user.ID,
			Username:   user.Username,
		})
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	for _, connection := range connectionsResponse.Reply {
		err := graphSvc.CreateRelationship(context.Background(), connection.User1_id, connection.User2_id, "KNOWS")
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	var rafalID string
	for _, user := range usersResponse.Reply {
		if user.Username == "Rafał" {
			rafalID = user.ID
			break
		}
	}
	fmt.Println("RAFAŁ's ID:", rafalID)
	var barbaraID string
	for _, user := range usersResponse.Reply {
		if user.Username == "Barbara" {
			barbaraID = user.ID
			break
		}
	}
	fmt.Println("BARBARA's ID:", barbaraID)

	shortestConnection, err := graphSvc.GetShortestConnection(context.Background(), rafalID, barbaraID)
	if err != nil {
		fmt.Println(err)
		return
	}

	joined := strings.Join(shortestConnection, ",")
	fmt.Println(joined)

	c3ntralaSvc, err := c3ntrala.NewService(envSvc)
	if err != nil {
		fmt.Println(err)
		return
	}
	apiResponse, err := c3ntralaSvc.PostReport("connections", joined)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("❇️ API response:", apiResponse)
}
