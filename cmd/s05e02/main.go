package main

import (
	"fmt"

	"github.com/crowmw/ai_devs3/pkg/ai"
	"github.com/crowmw/ai_devs3/pkg/c3ntrala"
	"github.com/crowmw/ai_devs3/pkg/env"
	"github.com/crowmw/ai_devs3/pkg/gps_agent"
	"github.com/sashabaranov/go-openai"
)

func main() {
	envSvc, err := env.NewService()
	if err != nil {
		fmt.Println(err)
		return
	}

	c3ntralaSvc, err := c3ntrala.NewService(envSvc)
	if err != nil {
		fmt.Println(err)
		return
	}

	aiSvc, err := ai.NewService(envSvc)
	if err != nil {
		fmt.Println(err)
		return
	}

	logs, err := c3ntralaSvc.GetLogs()
	if err != nil {
		fmt.Println(err)
		return
	}

	question, err := c3ntralaSvc.GetGpsQuestion()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("--------------------------------")
	fmt.Println("Question:")
	fmt.Println(question)
	fmt.Println("--------------------------------")

	gpsAgentSvc, err := gps_agent.NewService(envSvc, aiSvc, c3ntralaSvc, []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: fmt.Sprintf("You are an AI assistant helping analyze GPS logs. You are detail-oriented and methodical in your analysis. Here are the logs: %s", logs),
		},
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	answer, err := gpsAgentSvc.Execute(question)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("--------------------------------")
	fmt.Println("Answer:")
	fmt.Println(answer)

	response, err := c3ntralaSvc.PostReport("gps", answer)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("--------------------------------")
	fmt.Println("Response:")
	fmt.Println(response)
}
