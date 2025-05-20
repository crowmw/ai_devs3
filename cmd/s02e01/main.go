package main

import (
	"crypto/md5"
	"fmt"

	"github.com/crowmw/ai_devs3/pkg/ai"
	"github.com/crowmw/ai_devs3/pkg/config"
	"github.com/crowmw/ai_devs3/pkg/http"
	"github.com/crowmw/ai_devs3/pkg/processor"
	"github.com/sashabaranov/go-openai"
)

func main() {
	// Load environment variables from .env file
	if err := config.LoadEnv(); err != nil {
		fmt.Println(err)
		return
	}

	zippedData, err := http.FetchData(config.GetC3ntralaURL() + "/dane/przesluchania.zip")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Create MD5 hash of the zipped data
	hash := md5.Sum(zippedData)

	// Unzip the data to a directory named after the hash
	hashDir := fmt.Sprintf("%x", hash)
	if err := processor.ExtractZipToDirectory(zippedData, hashDir); err != nil {
		fmt.Println(err)
		return
	}

	err = processor.ProcessAudioFiles(hashDir)
	if err != nil {
		fmt.Println(err)
		return
	}

	concatenatedTranscriptions, err := processor.ReadAllTxtFilesFromDirectory(hashDir + "/transcriptions")
	if err != nil {
		fmt.Println(err)
		return
	}

	systemPrompt := GeneratePrompt(concatenatedTranscriptions)

	message := []openai.ChatCompletionMessage{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: "Użyj swojej wiedzy o Uniwersytetach w Polsce i ich instytutach, określ konkretną ulicę, gdzie znajduje się instytut.",
		},
	}

	aiResponse, err := ai.SendChatCompletion("gpt-4o", true, message)

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("AI response: ", aiResponse)

	result, err := http.SendC3ntralaReport("mp3", aiResponse)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Result: ", result)
}

func GeneratePrompt(transcription string) string {
	return `Your task is to determine the street where the university institute is located based solely on the provided transcription and your knowledge of the university in Poland. 
	I need the specific street name where the institute is located (not the main university building).

Analyze the following recording transcription step by step. Pay special attention to any location-related clues.

Transcription:
"""
` + transcription + `
"""

Let's analyze this step by step:
1. What clues about the university appear in the text?
2. Are there mentions of specific buildings or institutes?
3. Are there any specific mentions of streets or locations?

Focus only on the information provided in the transcription to determine the specific street where the institute is located.

_thinking...

Let me think about this step by step:
1. First, I'll identify all mentions of streets and locations in the text
2. Then, I'll look for connections between these locations and any mentioned institute
3. I'll analyze any descriptions of the area or directions given
4. I'll connect the dots between different location references
5. Finally, I'll determine the specific street based solely on the transcription and your knowledge of the university in Poland
`
}
