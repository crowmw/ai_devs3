package main

import (
	"fmt"

	"github.com/crowmw/ai_devs3/pkg/ai"
	"github.com/crowmw/ai_devs3/pkg/c3ntrala"
	"github.com/crowmw/ai_devs3/pkg/env"
	"github.com/crowmw/ai_devs3/pkg/softo"
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

	questions, err := c3ntralaSvc.GetSoftoQuestions()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("❇️ Questions:", questions)

	softoSvc, err := softo.NewService(envSvc, aiSvc)
	if err != nil {
		fmt.Println(err)
		return
	}

	answers := make(map[string]string)
	for i, question := range questions {
		currentUrl := ""
		for {
			answer, err := softoSvc.TryToFindAnswer(question, currentUrl)
			if err != nil {
				fmt.Println(err)
				return
			}

			if answer.Response == "YES" {
				fmt.Println("🎆 Answer:", answer.Answer)
				answers[i] = answer.Answer
				break
			} else if answer.Response == "NO" {
				fmt.Println("⛔️ Answer:", answer.Answer)
				currentUrl = answer.Answer
				continue
			}
		}
	}

	fmt.Println("🚀 Answers:", answers)

	report, err := c3ntralaSvc.PostReport("softo", answers)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("😎 Report:", report)
}
