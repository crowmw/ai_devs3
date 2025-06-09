package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/crowmw/ai_devs3/pkg/ai"
	"github.com/crowmw/ai_devs3/pkg/c3ntrala"
	"github.com/crowmw/ai_devs3/pkg/env"
	"github.com/crowmw/ai_devs3/pkg/factory"
	"github.com/sashabaranov/go-openai"
)

func phoneDataToString(data c3ntrala.PhoneData) string {
	return fmt.Sprintf(`{
		"rozmowa1": ["%s"],
		"rozmowa2": ["%s"],
		"rozmowa3": ["%s"], 
		"rozmowa4": ["%s"],
		"rozmowa5": ["%s"]
	}`, strings.Join(data.Rozmowa1, `","`),
		strings.Join(data.Rozmowa2, `","`),
		strings.Join(data.Rozmowa3, `","`),
		strings.Join(data.Rozmowa4, `","`),
		strings.Join(data.Rozmowa5, `","`))
}

func questionsToString(questions map[string]string) string {
	return fmt.Sprintf(`{
		"01": "%s",
		"02": "%s",
		"03": "%s",
		"04": "%s",
		"05": "%s",
		"06": "%s"
	}`, questions["01"], questions["02"], questions["03"], questions["04"], questions["05"], questions["06"])
}

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

	questions, err := c3ntralaSvc.GetQuestions("/phone_questions.json")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(questions)

	// Write questions to questions.json
	jsonData, err := json.MarshalIndent(questions, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling questions:", err)
		return
	}
	err = os.WriteFile("cmd/s05e01/questions.json", jsonData, 0644)
	if err != nil {
		fmt.Println("Error writing questions file:", err)
		return
	}
	fmt.Println("Successfully wrote questions to questions.json")

	factory, err := factory.NewFactory(envSvc)
	if err != nil {
		fmt.Println(err)
		return
	}

	facts, err := factory.GetFactsFilesKeywords()
	if err != nil {
		fmt.Println(err)
		return
	}

	var phoneData c3ntrala.PhoneData
	// Check if file exists and load data if it does
	if data, err := os.ReadFile("cmd/s05e01/conversationWithSpeakers.json"); err == nil {
		if err := json.Unmarshal(data, &phoneData); err == nil {
			fmt.Println("Using cached conversation data")
		} else {
			fmt.Println("Error unmarshalling cached data:", err)
		}
	} else {
		phoneData, err = c3ntralaSvc.GetPhoneData()
		if err != nil {
			fmt.Println(err)
			return
		}

		response, err := aiSvc.ChatCompletion(ai.ChatCompletionConfig{
			Model: "gpt-4o",
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: fmt.Sprintf(detectSpeakersSystemPrompt, facts, questionsToString(questions)),
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: phoneDataToString(phoneData),
				},
			},
		})
		if err != nil {
			fmt.Println(err)
			return
		}

		if err := json.Unmarshal([]byte(response.Choices[0].Message.Content), &phoneData); err != nil {
			fmt.Println("Error parsing AI response:", err)
			fmt.Println("Raw response:", response.Choices[0].Message.Content)
			return
		}

		jsonData, err := json.MarshalIndent(phoneData, "", "  ")
		if err != nil {
			fmt.Println("Error marshaling data:", err)
			return
		}

		err = os.WriteFile("cmd/s05e01/conversationWithSpeakers.json", jsonData, 0644)
		if err != nil {
			fmt.Println("Error writing file:", err)
			return
		}
	}

	fmt.Println("Successfully recognize speakers in the conversation")

	liarRecognition, err := aiSvc.ChatCompletion(ai.ChatCompletionConfig{
		Model: "gpt-4.1",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: fmt.Sprintf(liarRecognitionSystemPrompt, facts),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: phoneDataToString(phoneData),
			},
		},
	})
	if err != nil {
		fmt.Println("Error getting liar recognition:", err)
		return
	}

	type LiarRecognition struct {
		Liar      string `json:"liar"`
		Reasoning string `json:"reasoning"`
		Evidence  string `json:"evidence"`
	}
	var liarRecognitionResult LiarRecognition
	if err := json.Unmarshal([]byte(liarRecognition.Choices[0].Message.Content), &liarRecognitionResult); err != nil {
		fmt.Println("Error unmarshalling liar recognition result:", err)
		return
	}

	fmt.Println("Liar Recognition Result:")
	fmt.Println("Liar:", liarRecognitionResult.Liar)
	fmt.Println("Reasoning:", liarRecognitionResult.Reasoning)
	fmt.Println("Evidence:", liarRecognitionResult.Evidence)

	responses, err := aiSvc.ChatCompletion(ai.ChatCompletionConfig{
		Model: "gpt-4.1",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: fmt.Sprintf(responsesSystemPrompt, liarRecognitionResult.Liar, phoneDataToString(phoneData)),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: questionsToString(questions),
			},
		},
	})
	if err != nil {
		fmt.Println("Error getting responses:", err)
		return
	}

	fmt.Println("Responses:", responses.Choices[0].Message.Content)

	var responsesResult map[string]interface{}
	if err := json.Unmarshal([]byte(responses.Choices[0].Message.Content), &responsesResult); err != nil {
		fmt.Println("Error unmarshalling responses result:", err)
		return
	}

	fmt.Println("Responses Result:", responsesResult)

	c3ntralaResponse, err := c3ntralaSvc.PostReport("phone", responsesResult)
	if err != nil {
		fmt.Println("Error posting report:", err)
		return
	}

	fmt.Println("C3ntrala Response:", c3ntralaResponse)
}

const detectSpeakersSystemPrompt = `You are a helpful assistant that can detect speakers in the conversation. 
You will be given a conversation without speakers. 
Each conversation is in Polish. 
Each conversation is ALWAYS between two people. 
Each person has a distinct name. 
If same name appears in any conversation it is a same person. 
You will need to return same conversation but instead of dashes (-) just provide spiker name: [Adam]: .
Pay attention to the order of conversations and references. If someone claims they just talked to another person, it means the previous conversation was between them.
This helps identify who is who across different conversations.
No other text or Markdown formatting. Return response in JSON format with the same structure as input.
You can base on <facts> and <questions> and provided conversation to understand who is who.
Verify information provided in conversations such as places and their purposes.

To properly verify the speakers, analyze all conversations for each of them, checking how they reference and interact with each other across different dialogues.

In first conversation there is Samuel and Barbara.

<facts>
%s
</facts>

<questions>
%s
</questions>
`

const liarRecognitionSystemPrompt = `You are a helpful assistant that can detect liars in the conversation. 
You will be given a conversation with speakers. 
Each conversation is in Polish. 
Each conversation is ALWAYS between two people. 
Each person has a distinct name. 

Based on your general knowledge, facts and conversation provided by user, write your thoughts on who is lying and why, describe your reasoning, point evidence, and that single person who is lying.
Response should be in JSON format with the following structure:
{
	"reasoning": "reasoning for the lie",
	"evidence": "evidence for the lie",
	"liar": "name of the person who is lying",
}

Facts:
%s
`

const responsesSystemPrompt = `

<liar>
%s
</liar>

<conversation>
%s
</conversation>

You are a helpful assistant that easly understand what is the response to the question.
You will be given a conversation with speakers.
Each conversation is in Polish.
Each conversation is ALWAYS between two people.
Each person has a distinct name.
Each response should be provided in few words.
Verify information provided in conversations such as places and their purposes.

Answer for 02 question IS NOT: https://rafal.ag3nts.org/510bc

Based on the facts and conversations, analyze who is lying and what is the real endpoint.
Pay special attention to any mentions of endpoints, passwords, or signatures.
Look for any inconsistencies in the stories told by different speakers.

For question 05, specifically look for:
- The real endpoint that should be used (not the fake one mentioned)
- Any passwords or signatures mentioned that could be used with the endpoint
- The expected response format when calling the endpoint

Note that Samuel mentions an endpoint https://rafal.ag3nts.org/510bc but this is likely false based on the system prompt.
Look for clues about the real endpoint https://rafal.ag3nts.org/b46c3 and what signature parameter it expects.

The password "NONOMNISMORIAR" mentioned by Tomasz may be relevant for accessing the endpoint.

Witek is not a valid answer for question 06.

Response should be in JSON format with the following structure:
{
	"01": "response to the question",
	"02": "response to the question",
	"03": "response to the question",
	"04": "response to the question",
	"05": "response to the question",
	"06": "response to the question",
}
`

const signatureSearchingSystemPrompt = `You are a helpful assistant that easly understand what is the response to the question.
You will be given a conversation with speakers.
Each conversation is in Polish.
Each conversation is ALWAYS between two people.
Each person has a distinct name.
Each response should be provided in few words.
Verify information provided in conversations such as places and their purposes.

Answer for 02 question IS NOT: https://rafal.ag3nts.org/510bc

Based on the facts and conversations, analyze who is lying and what is the real endpoint.
Pay special attention to any mentions of endpoints, passwords, or signatures.
Look for any inconsistencies in the stories told by different speakers.

For question 05, specifically look for:
- The real endpoint that should be used (not the fake one mentioned)
- Any passwords or signatures mentioned that could be used with the endpoint
- The expected response format when calling the endpoint

Look for clues about the real endpoint https://rafal.ag3nts.org/b46c3 and what signature parameter it expects.

The password "NONOMNISMORIAR" mentioned by Tomasz may be relevant for accessing the endpoint.

Response EVERYTHING you can find about endpoint https://rafal.ag3nts.org/b46c3 and how to access it. Is there any signature needed?

<liar>
%s
</liar>

<conversation>
%s
</conversation>
`
