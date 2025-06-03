package tune

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type TrainingExample struct {
	Messages []Message `json:"messages"`
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: program <correct.txt> <incorrect.txt>")
		return
	}

	correctFile := os.Args[1]
	incorrectFile := os.Args[2]

	// Create output file
	outFile, err := os.Create("fine-tuning-input.jsonl")
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		return
	}
	defer outFile.Close()

	// Process correct examples
	if err := processFile(correctFile, "1", outFile); err != nil {
		fmt.Printf("Error processing correct file: %v\n", err)
		return
	}

	// Process incorrect examples
	if err := processFile(incorrectFile, "0", outFile); err != nil {
		fmt.Printf("Error processing incorrect file: %v\n", err)
		return
	}

	fmt.Println("Successfully created fine-tuning-input.jsonl")
}

func processFile(filename, label string, outFile *os.File) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error opening file %s: %w", filename, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		example := TrainingExample{
			Messages: []Message{
				{
					Role:    "system",
					Content: "validate data",
				},
				{
					Role:    "user",
					Content: line,
				},
				{
					Role:    "assistant",
					Content: label,
				},
			},
		}

		jsonData, err := json.Marshal(example)
		if err != nil {
			return fmt.Errorf("error marshaling JSON: %w", err)
		}

		// Write JSON line with newline
		if _, err := outFile.Write(append(jsonData, '\n')); err != nil {
			return fmt.Errorf("error writing to file: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	return nil
}
