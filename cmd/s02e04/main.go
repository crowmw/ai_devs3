package main

import (
	"encoding/json"
	"fmt"

	"github.com/crowmw/ai_devs3/pkg/ai"
	"github.com/crowmw/ai_devs3/pkg/config"
	"github.com/crowmw/ai_devs3/pkg/factory"
	"github.com/crowmw/ai_devs3/pkg/http"
	"github.com/sashabaranov/go-openai"
)

type RobotDescriptionResult struct {
	Description string `json:"description"`
}

func main() {
	// Load environment variables from .env file
	if err := config.LoadEnv(); err != nil {
		fmt.Println(err)
		return
	}

	factoryData, err := factory.NewFactory()
	if err != nil {
		fmt.Println(err)
		return
	}

	var allFactoryFilesContent []factory.FactoryFileContent

	// Get text files
	factoryTextFiles, err := factoryData.GetTextFiles()
	if err != nil {
		fmt.Println(err)
		return
	}
	allFactoryFilesContent = append(allFactoryFilesContent, factoryTextFiles...)

	// Get image files texts
	factoryImageFilesTexts, err := factoryData.GetImageFilesTexts()
	if err != nil {
		fmt.Println(err)
		return
	}
	allFactoryFilesContent = append(allFactoryFilesContent, factoryImageFilesTexts...)

	// Get audio files texts
	factoryAudioFilesTexts, err := factoryData.GetAudioFilesTexts()
	if err != nil {
		fmt.Println(err)
		return
	}
	allFactoryFilesContent = append(allFactoryFilesContent, factoryAudioFilesTexts...)

	// Print all files
	for _, file := range allFactoryFilesContent {
		fmt.Println(file.File)
		fmt.Println(file.Content)
	}

	message := []openai.ChatCompletionMessage{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("%+v", allFactoryFilesContent),
		},
	}

	fmt.Println("--------------------------------")
	fmt.Println(fmt.Sprintf("%+v", allFactoryFilesContent))
	fmt.Println("--------------------------------")

	aiResponse, err := ai.SendChatCompletion("gpt-4o", false, message)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(aiResponse)

	var categories struct {
		People   []string `json:"people"`
		Hardware []string `json:"hardware"`
	}

	if err := json.Unmarshal([]byte(aiResponse), &categories); err != nil {
		fmt.Println("Error unmarshaling response:", err)
		return
	}

	report, err := http.SendC3ntralaReport("kategorie", categories)
	if err != nil {
		fmt.Println("Error sending report:", err)
		return
	}
	fmt.Println("Report:", report)
}

var systemPrompt = `
You are a classification assistant. 
Your task is to analyze text content from files and categorize them into two categories: 
- 'people' (Categorize the text only as "people" if it contains concrete information about captured individuals (e.g., names, status, location, or situation of captured people), or evidence of human presence (e.g., footprints, personal belongings, sightings in a specific location).
Do NOT assign to this category if:
A person is only mentioned casually (e.g., by name or profession) with no evidence of capture or presence.
It includes hypothetical or fictional references to people (e.g., jokes, speculation, or indirect references).
Use strict criteria — if there's any doubt, do not assign to the "people" category.)
- 'hardware' (Hardware (not software) defects.). 

For each file, you should explain why it belongs in a particular category. 
Files that don't fit either category should be skipped. 
Return ONLY a JSON object with two category keys ('people' and 'hardware') containing arrays of UNCHANGED filenames with ORIGINAL EXTENSIONS. 
Example: {\"people\": [\"file1.txt\"], \"hardware\": [\"file2.txt\"]}. 
DO NOT add any explanation, markdown formatting, code blocks, or additional text.
`
