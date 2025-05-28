package main

import (
	"context"
	"fmt"

	"github.com/crowmw/ai_devs3/pkg/ai"
	"github.com/crowmw/ai_devs3/pkg/env"
	"github.com/crowmw/ai_devs3/pkg/factory"
	"github.com/crowmw/ai_devs3/pkg/http"
	"github.com/crowmw/ai_devs3/pkg/vector"
	"github.com/qdrant/go-client/qdrant"
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

	weaponTests, err := f.GetWeaponTests()
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

	collectionName := "weapon_tests"

	points := make([]vector.NewPoint, len(weaponTests))
	for i, test := range weaponTests {
		points[i] = vector.NewPoint{
			ID:   "",
			Text: test.Content,
			Metadata: map[string]any{
				"date":     fmt.Sprintf("%s-%s-%s", test.Filename[0:4], test.Filename[5:7], test.Filename[8:10]),
				"filename": test.Filename,
			},
		}
	}
	addedPoints, err := vectorSvc.AddPoints(context.Background(), collectionName, points)
	if err != nil {
		fmt.Println(err)
		return
	}

	query := "W raporcie, z którego dnia znajduje się wzmianka o kradzieży prototypu broni?"

	results, err := vectorSvc.Search(context.Background(), collectionName, query, 1)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Search results:", results)

	var matchingPoint *qdrant.Struct
	for _, point := range addedPoints {
		if point.Id.String() == results[0].Id.String() {
			// Debug print the structure
			fmt.Printf("Payload structure: %+v\n", point.Payload)
			metadataValue := point.Payload["metadata"]
			fmt.Printf("Metadata value: %+v\n", metadataValue)
			matchingPoint = metadataValue.GetStructValue()
			fmt.Printf("Metadata struct: %+v\n", matchingPoint)
			break
		}
	}

	fmt.Printf("Matching point: %+v\n", matchingPoint)
	r, err := http.SendC3ntralaReport("wektory", matchingPoint.GetFields()["date"].GetStringValue())
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Report response:", r)
}
