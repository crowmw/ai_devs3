package robot

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/crowmw/ai_devs3/pkg/ai"
	"github.com/crowmw/ai_devs3/pkg/config"
	"github.com/crowmw/ai_devs3/pkg/http"
	"github.com/sashabaranov/go-openai"
)

// TestData represents a single test case for the robot
type TestData struct {
	Question string `json:"question"`
	Answer   int    `json:"answer"`
	Test     *struct {
		Q string `json:"q"`
		A string `json:"a"`
	} `json:"test,omitempty"`
}

// CalibrationData represents the structure of the JSON response
type CalibrationData struct {
	APIKey      string     `json:"apikey"`
	Description string     `json:"description"`
	Copyright   string     `json:"copyright"`
	TestData    []TestData `json:"test-data"`
}

// IndustrialRobot represents an industrial robot service
type IndustrialRobot struct {
	calibrationData CalibrationData
}

// NewIndustrialRobot creates a new instance of IndustrialRobot and loads calibration data
func NewIndustrialRobot() (*IndustrialRobot, error) {
	// Construct URL with API key
	url := fmt.Sprintf("%s/data/%s/json.txt",
		config.GetC3ntralaURL(),
		config.GetMyAPIKey())

	fmt.Printf("URL: %v\n", url)

	// Fetch calibration data as JSON
	var calibrationData CalibrationData
	err := http.FetchJSONData(url, &calibrationData)
	if err != nil {
		return nil, fmt.Errorf("error fetching calibration data: %w", err)
	}

	return &IndustrialRobot{
		calibrationData: calibrationData,
	}, nil
}

// calculateResult parses and evaluates the mathematical equation from the test case
func calculateResult(testCase TestData) (int, error) {
	// Parse and evaluate the mathematical equation
	equation := testCase.Question
	// Remove extra spaces
	equation = strings.TrimSpace(equation)

	// Split equation into parts
	parts := strings.Split(equation, " ")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid equation format: %s", equation)
	}

	// Parse numbers and operator
	num1, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("error parsing first number: %w", err)
	}

	operator := parts[1]

	num2, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, fmt.Errorf("error parsing second number: %w", err)
	}

	// Calculate actual result
	switch operator {
	case "+":
		return num1 + num2, nil
	case "-":
		return num1 - num2, nil
	case "*":
		return num1 * num2, nil
	case "/":
		if num2 == 0 {
			return 0, fmt.Errorf("division by zero")
		}
		return num1 / num2, nil
	default:
		return 0, fmt.Errorf("unsupported operator: %s", operator)
	}
}

func (r *IndustrialRobot) Recalibrate() error {
	fmt.Println("Starting recalibration...")

	r.calibrationData.APIKey = config.GetMyAPIKey()

	for i := range r.calibrationData.TestData {
		testCase := &r.calibrationData.TestData[i]

		if testCase.Test != nil {
			// If test data contains "test" parameter, just print it
			fmt.Printf("Test Q: %s\n", testCase.Test.Q)

			aiAnswer, err := ai.SendChatCompletion("gpt-4o-mini", true, []openai.ChatCompletionMessage{
				{
					Role:    "system",
					Content: "You can response with only one word, no other text.",
				},
				{
					Role:    "user",
					Content: testCase.Test.Q,
				},
			})
			if err != nil {
				return fmt.Errorf("error getting OpenAI response: %w", err)
			}
			fmt.Printf("AI answer: %s\n", aiAnswer)
			fmt.Printf("Updating calibration data with correct answer...\n")
			testCase.Test.A = aiAnswer
			continue
		}

		// Calculate actual result
		actualResult, err := calculateResult(*testCase)
		if err != nil {
			fmt.Printf("Error calculating result: %v\n", err)
			continue
		}

		// Compare with provided answer
		if actualResult != testCase.Answer {
			fmt.Printf("Equation: %s\n", testCase.Question)
			fmt.Printf("ERROR: Answer provided (%d) is incorrect! Actual result is: %d\n",
				testCase.Answer, actualResult)
			fmt.Printf("Updating calibration data with correct answer...\n")
			testCase.Answer = actualResult
		}
	}
	return nil
}

// SendCalibrationReport sends the calibration data to the central server
func (r *IndustrialRobot) SendCalibrationReport() error {
	fmt.Println("Sending calibration report...")

	resp, err := http.SendC3ntralaReport("JSON", r.calibrationData)

	if err != nil {
		return fmt.Errorf("error sending calibration report: %w", err)
	}

	fmt.Printf("Calibration report sent successfully: %v\n", resp)
	return nil
}
