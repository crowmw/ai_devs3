package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/crowmw/ai_devs3/pkg/ai"
	"github.com/crowmw/ai_devs3/pkg/c3ntrala"
	"github.com/crowmw/ai_devs3/pkg/config"
	"github.com/crowmw/ai_devs3/pkg/env"
	"github.com/crowmw/ai_devs3/pkg/http"
	"github.com/crowmw/ai_devs3/pkg/processor"
	"github.com/sashabaranov/go-openai"
)

type HintResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Hint    string `json:"hint"`
	Debug   string `json:"debug"`
}

type Answer struct {
	Thinking string `json:"_thinking"`
	Answer   string `json:"answer"`
}

func processSingleQuestion(aiSvc *ai.Service, questionID string, question string, systemPrompt string) (*Answer, error) {
	fmt.Printf("Processing question %s: %s\n", questionID, question)
	fmt.Println("Question:")
	fmt.Println(question)
	fmt.Println("--------------------------------")

	response, err := aiSvc.ChatCompletion(ai.ChatCompletionConfig{
		Model: "gpt-4.1",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: question,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error processing question %s: %w", questionID, err)
	}

	var answer Answer
	if err := json.Unmarshal([]byte(response.Choices[0].Message.Content), &answer); err != nil {
		return nil, fmt.Errorf("error parsing answer: %w", err)
	}

	fmt.Println("Answer:")
	fmt.Println(answer)
	fmt.Println("--------------------------------")
	fmt.Printf("Answer for question %s: %s\n\n", questionID, answer.Answer)

	return &answer, nil
}

func buildSystemPrompt(currentDate, previousAnswers, hints, context string) string {
	return fmt.Sprintf(answerQuestionsSystemPrompt,
		currentDate,
		previousAnswers,
		hints,
		context)
}

func checkAndProcessIncorrectAnswer(aiSvc *ai.Service, c3ntralaSvc *c3ntrala.Service, questions map[string]string, answers map[string]string, previousAnswers, hints, context string) (bool, error) {
	hint, err := c3ntralaSvc.PostReport("notes", answers)
	if err != nil {
		return false, fmt.Errorf("error submitting answers: %w", err)
	}

	var hintResp HintResponse
	if err := json.Unmarshal([]byte(hint), &hintResp); err != nil {
		return false, fmt.Errorf("error parsing hint: %w", err)
	}

	fmt.Printf("Code: %d\n", hintResp.Code)
	fmt.Printf("Message: %s\n", hintResp.Message)
	fmt.Printf("Hint: %s\n", hintResp.Hint)
	fmt.Printf("Debug: %s\n", hintResp.Debug)

	if !strings.Contains(hintResp.Message, "is incorrect") {
		return false, nil
	}

	re := regexp.MustCompile(`question (\d+) is incorrect`)
	matches := re.FindStringSubmatch(hintResp.Message)
	if len(matches) < 2 {
		return false, fmt.Errorf("could not extract question number from message: %s", hintResp.Message)
	}

	questionID := matches[1]
	question := questions[questionID]

	jsonData, err := json.Marshal(answers)
	if err != nil {
		return false, fmt.Errorf("failed to marshal answers: %w", err)
	}
	previousAnswers = string(jsonData)
	hints = hintResp.Message + "\n" + hintResp.Hint + "\n" + hintResp.Debug

	systemPrompt := buildSystemPrompt(time.Now().Format("2006-01-02"), previousAnswers, hints, context)
	answer, err := processSingleQuestion(aiSvc, questionID, question, systemPrompt)
	if err != nil {
		return false, err
	}

	answers[questionID] = answer.Answer
	return true, nil
}

func processQuestions(aiSvc *ai.Service, c3ntralaSvc *c3ntrala.Service, questions map[string]string, context string) (map[string]string, error) {
	answers := make(map[string]string)
	var previousAnswers string
	var hints string

	// Process all questions first
	systemPrompt := buildSystemPrompt(time.Now().Format("2006-01-02"), previousAnswers, hints, context)
	for questionID, question := range questions {
		answer, err := processSingleQuestion(aiSvc, questionID, question, systemPrompt)
		if err != nil {
			return nil, err
		}
		answers[questionID] = answer.Answer
	}

	// Check and reprocess incorrect answers
	for {
		hasIncorrect, err := checkAndProcessIncorrectAnswer(aiSvc, c3ntralaSvc, questions, answers, previousAnswers, hints, context)
		if err != nil {
			return nil, err
		}
		if !hasIncorrect {
			break
		}
	}

	return answers, nil
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

	pdfSvc, err := processor.NewPDFService(envSvc.GetC3ntralaURL()+"/dane/notatnik-rafala.pdf", aiSvc)
	if err != nil {
		fmt.Println(err)
		return
	}

	text, err := pdfSvc.GetText()
	if err != nil {
		fmt.Println(err)
		return
	}

	image, err := pdfSvc.ExtractTextFromPage(19)
	if err != nil {
		fmt.Println(err)
		return
	}

	fullPDFText := "<pdf_text>" + text + "</pdf_text>" + "\n" + "<pdf_last_page_image>" + image + "</pdf_last_page_image>"

	var questions map[string]string
	err = http.FetchJSONData(envSvc.GetC3ntralaURL()+"/data/"+config.GetMyAPIKey()+"/notes.json", &questions)
	if err != nil {
		fmt.Println(err)
		return
	}

	answers, err := processQuestions(aiSvc, c3ntralaSvc, questions, fullPDFText)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("\nFinal answers:")
	for questionID, answer := range answers {
		fmt.Printf("Question %s: %s\n", questionID, answer)
	}
}

const answerQuestionsSystemPrompt = `
<context>
%s
</context>

You are a helpful assistant that can answer questions based on context.
Please analyze the provided context carefully. When answering questions:

1. If a previous answer was marked as incorrect, do not repeat that answer
2. Pay attention to any hints provided to guide your response
3. Consider the full context including text, image descriptions, and audio transcriptions
4. Provide specific evidence from the context to support your answer
5. Respond in Polish
6. Respond with plain JSON with struct {"_thinking": "thinking", "answer": "answer"}. No other text or Markdown formatting.
7. Before answering, think about the answer and write your thinking in the _thinking field.

<current_date>
%s
</current_date>

<previous_answers>
%s
</previous_answers>

<Hints>
%s
</Hints>

<rules>
- Required: Give a new answer that:
  * Avoids the previous incorrect response
  * Takes the hint into account
  * Is supported by specific evidence from the context
</rules>
`
