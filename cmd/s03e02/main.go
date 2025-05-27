package main

import (
	"fmt"

	"github.com/crowmw/ai_devs3/pkg/ai"
	"github.com/crowmw/ai_devs3/pkg/env"
	"github.com/crowmw/ai_devs3/pkg/factory"
	"github.com/crowmw/ai_devs3/pkg/vector"
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

	err = f.LoadWeaponTests()
	if err != nil {
		fmt.Println(err)
		return
	}

	aiSvc, err := ai.NewService(envSvc)
	if err != nil {
		fmt.Println(err)
		return
	}

	vectorSvc, err := vector.NewService(envSvc, aiSvc, 1024)
	if err != nil {
		fmt.Println(err)
		return
	}

	collections := vectorSvc.GetCollections()
	fmt.Println(collections)
}
