package main

import (
	"fmt"

	"github.com/crowmw/ai_devs3/pkg/ai"
	"github.com/crowmw/ai_devs3/pkg/c3ntrala"
	"github.com/crowmw/ai_devs3/pkg/env"
	photosautomate "github.com/crowmw/ai_devs3/pkg/photos-automate"
	"github.com/crowmw/ai_devs3/pkg/recognize"
)

func main() {
	envSvc, err := env.NewService()
	if err != nil {
		fmt.Println(err)
		return
	}

	aiSvc, err := ai.NewService(envSvc)
	if err != nil {
		fmt.Println(err)
		return
	}

	c3ntralaSvc, err := c3ntrala.NewService(envSvc)
	if err != nil {
		fmt.Println(err)
		return
	}

	photosSvc, err := photosautomate.NewService(envSvc, c3ntralaSvc)
	if err != nil {
		fmt.Println(err)
		return
	}

	recognizeSvc, err := recognize.NewService(envSvc, aiSvc, photosSvc, c3ntralaSvc)
	if err != nil {
		fmt.Println(err)
		return
	}

	barbaraDescription, err := recognizeSvc.StartRecognize()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("❇️ Recognize:", barbaraDescription)

	answer, err := c3ntralaSvc.PostReport("photos", barbaraDescription)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("❇️ Answer:", answer)
}
