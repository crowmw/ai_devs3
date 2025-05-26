package factory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/crowmw/ai_devs3/pkg/ai"
	"github.com/crowmw/ai_devs3/pkg/config"
	"github.com/crowmw/ai_devs3/pkg/http"
	"github.com/crowmw/ai_devs3/pkg/processor"
	"github.com/otiai10/gosseract/v2"
	"github.com/sashabaranov/go-openai"
)

// Factory represents a factory with its files
type Factory struct {
	DirPath string
}

// NewFactory creates a new Factory instance and initializes it with files
func NewFactory() (*Factory, error) {
	factory := &Factory{
		DirPath: "pliki_z_fabryki",
	}

	// Check if directory exists
	if _, err := os.Stat(factory.DirPath); err == nil {
		fmt.Println("Factory files directory already exists, skipping download")
		return factory, nil
	}

	// Download zip file
	fmt.Println("Downloading factory files...")
	zipData, err := http.FetchData(config.GetC3ntralaURL() + "/data/" + config.GetMyAPIKey() + "/pliki_z_fabryki.zip")
	if err != nil {
		return nil, fmt.Errorf("error downloading factory files: %w", err)
	}

	// Extract files
	fmt.Println("Extracting factory files...")
	if err := processor.ExtractZipToDirectory(zipData, factory.DirPath); err != nil {
		return nil, fmt.Errorf("error extracting factory files: %w", err)
	}

	fmt.Println("Factory files downloaded and extracted successfully")
	return factory, nil
}

// FactoryFileContent represents a text file with its name and content
type FactoryFileContent struct {
	File    string
	Content string
}

// getTextsFromDirectory reads all .txt files from the given directory and returns their names and contents
func (f *Factory) getTextsFromDirectory(dirPath string) ([]FactoryFileContent, error) {
	var texts []FactoryFileContent

	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("error reading directory: %w", err)
	}

	for _, file := range files {
		// Skip directories and non-txt files
		if file.IsDir() || !strings.HasSuffix(strings.ToLower(file.Name()), ".txt") {
			continue
		}

		// Read file contents
		content, err := os.ReadFile(filepath.Join(dirPath, file.Name()))
		if err != nil {
			return nil, fmt.Errorf("error reading file %s: %w", file.Name(), err)
		}

		texts = append(texts, FactoryFileContent{
			File:    file.Name(),
			Content: string(content),
		})
	}

	return texts, nil
}

// GetTextFiles reads all .txt files from the factory directory and returns their names and contents
func (f *Factory) GetTextFiles() ([]FactoryFileContent, error) {
	return f.getTextsFromDirectory(f.DirPath)
}

// GetFactsFiles reads all .txt files from the facts subdirectory and returns their names and contents
func (f *Factory) GetFactsFiles() ([]FactoryFileContent, error) {
	return f.getTextsFromDirectory(filepath.Join(f.DirPath, "facts"))
}

// createKeywordsDirectory creates a keywords directory for a given file path if it doesn't exist
func (f *Factory) createKeywordsDirectory(filePath string) error {
	keywordsDir := filepath.Join(filepath.Dir(filePath), "keywords")
	return os.MkdirAll(keywordsDir, 0755)
}

// getKeywordsFilePath returns the path to the keywords file for a given file
func (f *Factory) getKeywordsFilePath(filePath string) string {
	keywordsFileName := strings.TrimSuffix(filepath.Base(filePath), ".txt") + "-keywords.txt"
	return filepath.Join(filepath.Dir(filePath), "keywords", keywordsFileName)
}

// getExistingKeywords reads existing keywords if they exist
func (f *Factory) getExistingKeywords(filePath string) (string, bool, error) {
	keywordsFilePath := f.getKeywordsFilePath(filePath)
	keywords, err := os.ReadFile(keywordsFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("error reading keywords file: %w", err)
	}
	return string(keywords), true, nil
}

func (f *Factory) GetFactFileKeywords(fact FactoryFileContent) (string, error) {
	filePath := filepath.Join(f.DirPath, "facts", fact.File)

	// Try to read existing keywords file
	if keywords, exists, err := f.getExistingKeywords(filePath); err != nil {
		return "", err
	} else if exists {
		fmt.Println("🔍 Keywords already exist for:", fact.File)
		return keywords, nil
	}

	fmt.Println("🔍 Extracting keywords for:", fact.File)
	prompt := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: extractFactsFilesKeywordsPrompt,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: "File name: " + fact.File + "\n\n" + fact.Content,
		},
	}

	aiResponse, err := ai.SendChatCompletion("gpt-4o-mini", false, prompt)
	if err != nil {
		return "", fmt.Errorf("error getting AI analysis: %w", err)
	}

	// Create keywords directory if it doesn't exist
	if err := f.createKeywordsDirectory(filePath); err != nil {
		return "", fmt.Errorf("error creating keywords directory: %w", err)
	}

	// Write keywords to file
	if err := os.WriteFile(f.getKeywordsFilePath(filePath), []byte(aiResponse), 0644); err != nil {
		return "", fmt.Errorf("error writing keywords file: %w", err)
	}

	return aiResponse, nil
}

// GetFactsFilesTexts reads all .txt files from the facts subdirectory and returns their names and contents
func (f *Factory) GetFactsFilesKeywords() (string, error) {
	facts, err := f.GetFactsFiles()
	if err != nil {
		return "", fmt.Errorf("error getting facts files: %w", err)
	}
	var allFacts string
	for _, fact := range facts {
		keywords, err := f.GetFactFileKeywords(fact)
		if err != nil {
			return "", fmt.Errorf("error getting facts files keywords: %w", err)
		}
		allFacts += fmt.Sprintf("Filename: %s\n%s\n", fact.File, fact.Content+"Generated Keywords:\n"+keywords+"\n\n")
	}

	return allFacts, nil
}

// ExtractKeyInformation extracts key information from a text file using AI analysis
func (f *Factory) ExtractKeyInformation(content FactoryFileContent) (FactoryFileContent, error) {
	fmt.Println("🔍 Extracting key information from:", content.File)
	filePath := filepath.Join(f.DirPath, content.File)

	// Try to read existing keywords file
	if keywords, exists, err := f.getExistingKeywords(filePath); err != nil {
		return FactoryFileContent{}, err
	} else if exists {
		fmt.Println("🔍 Keywords already exist for:", content.File)
		content.Content = content.Content + "\n\nKey information:\n" + keywords
		return content, nil
	}

	facts, err := f.GetFactsFilesKeywords()
	if err != nil {
		return FactoryFileContent{}, fmt.Errorf("error getting facts files keywords: %w", err)
	}

	prompt := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: fmt.Sprintf(extractKeyInformationPrompt, facts),
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: "File name: " + content.File + "\n" + content.Content,
		},
	}

	result, err := ai.SendChatCompletion("gpt-4o-mini", false, prompt)
	if err != nil {
		return FactoryFileContent{}, fmt.Errorf("error getting AI analysis: %w", err)
	}

	// Create keywords directory if it doesn't exist
	if err := f.createKeywordsDirectory(filePath); err != nil {
		return FactoryFileContent{}, fmt.Errorf("error creating keywords directory: %w", err)
	}

	// Write keywords to file
	if err := os.WriteFile(f.getKeywordsFilePath(filePath), []byte(result), 0644); err != nil {
		return FactoryFileContent{}, fmt.Errorf("error writing keywords file: %w", err)
	}

	content.Content = content.Content + "\n\nGenerated Keywords:\n" + result
	return content, nil
}

// GetReportFiles reads all .txt files from the factory directory and returns their names and contents
func (f *Factory) GetReportFiles() ([]FactoryFileContent, error) {
	rawReports, err := f.getTextsFromDirectory(filepath.Join(f.DirPath))
	if err != nil {
		return nil, fmt.Errorf("error getting report files: %w", err)
	}

	reports := []FactoryFileContent{}
	for _, report := range rawReports {
		report, err := f.ExtractKeyInformation(report)
		if err != nil {
			return nil, fmt.Errorf("error extracting key information from %s: %w", report.File, err)
		}

		reports = append(reports, report)
	}
	return reports, nil
}

// ReportAnalysis represents a map of report files and their analyzed keywords
type ReportAnalysis map[string]string

// AnalyzeReports analyzes all reports and returns a map of report files and their analyzed keywords
func (f *Factory) AnalyzeReports() (ReportAnalysis, error) {
	fmt.Println("🔍 Analyzing reports...")
	reports, err := f.GetReportFiles()
	fmt.Println("\n\n🔍 Reports:\n\n", reports)
	if err != nil {
		return nil, fmt.Errorf("error getting report files: %w", err)
	}

	fmt.Println("🔍 Getting facts files keywords...")
	facts, err := f.GetFactsFilesKeywords()
	fmt.Println("\n\n🔍 Facts files keywords:\n\n", facts)
	if err != nil {
		return nil, fmt.Errorf("error getting facts files: %w", err)
	}

	fmt.Println("🔍 Analyzing reports...")
	reportAnalyses := make(ReportAnalysis)
	var allReportsContent string
	for _, report := range reports {
		allReportsContent += fmt.Sprintf("File: %s\n%s\n\n", report.File, report.Content)
	}

	prompt := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: fmt.Sprintf(analyzeReportsPrompt, facts),
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: allReportsContent,
		},
	}

	fmt.Println("\n\n🔍 Prompt:\n\n", prompt)

	aiResponse, err := ai.SendChatCompletion("gpt-4.1", false, prompt)
	if err != nil {
		return nil, fmt.Errorf("error sending chat completion: %w", err)
	}

	fmt.Println("\n\n🔍 AI response:\n\n", aiResponse)

	var response struct {
		Thinking    string            `json:"_thinking"`
		FinalResult map[string]string `json:"finalResult"`
	}

	err = json.Unmarshal([]byte(aiResponse), &response)
	if err != nil {
		return nil, fmt.Errorf("error parsing AI response: %w", err)
	}

	reportAnalyses = response.FinalResult

	return reportAnalyses, nil
}

// TextFile represents a text file with its name and content
type FactoryFileImage struct {
	File  string
	Image []byte
}

// GetImageFiles reads all .png files from the factory directory and returns their names and contents
func (f *Factory) GetImageFiles() ([]FactoryFileImage, error) {
	var images []FactoryFileImage

	err := filepath.Walk(f.DirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-image files
		if info.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".png") {
			return nil
		}

		// Read file contents
		image, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("error reading file %s: %w", path, err)
		}

		images = append(images, FactoryFileImage{
			File:  filepath.Base(path),
			Image: image,
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking factory directory: %w", err)
	}
	return images, nil
}

// GetImageFilesTexts extracts text from images using OCR and returns their contents
func (f *Factory) GetImageFilesTexts() ([]FactoryFileContent, error) {
	// Get all image files
	images, err := f.GetImageFiles()
	if err != nil {
		return nil, fmt.Errorf("error getting image files: %w", err)
	}

	// Initialize Tesseract client
	client := gosseract.NewClient()
	defer client.Close()

	var texts []FactoryFileContent

	// Process each image
	for _, img := range images {
		// Create a temporary file for the image
		tmpFile, err := os.CreateTemp("", "ocr-*.png")
		if err != nil {
			return nil, fmt.Errorf("error creating temporary file: %w", err)
		}
		defer os.Remove(tmpFile.Name())

		// Write image data to temporary file
		if _, err := tmpFile.Write(img.Image); err != nil {
			tmpFile.Close()
			return nil, fmt.Errorf("error writing image to temporary file: %w", err)
		}
		tmpFile.Close()

		// Set image path for OCR
		if err := client.SetImage(tmpFile.Name()); err != nil {
			return nil, fmt.Errorf("error setting image for OCR: %w", err)
		}

		// Perform OCR
		text, err := client.Text()
		if err != nil {
			return nil, fmt.Errorf("error performing OCR on %s: %w", img.File, err)
		}

		texts = append(texts, FactoryFileContent{
			File:    img.File,
			Content: text,
		})
	}

	return texts, nil
}

// AudioFile represents an audio file with its name and content
type FactoryFileAudio struct {
	File  string
	Audio []byte
}

// GetAudioFiles reads all .mp3 files from the factory directory and returns their names and contents
func (f *Factory) GetAudioFiles() ([]FactoryFileAudio, error) {
	var audios []FactoryFileAudio

	err := filepath.Walk(f.DirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-audio files
		if info.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".mp3") {
			return nil
		}

		// Read file contents
		audio, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("error reading file %s: %w", path, err)
		}

		audios = append(audios, FactoryFileAudio{
			File:  filepath.Base(path),
			Audio: audio,
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking factory directory: %w", err)
	}

	return audios, nil
}

// GetAudioFilesTexts transcribes audio files and returns their contents
func (f *Factory) GetAudioFilesTexts() ([]FactoryFileContent, error) {
	// Create a temporary directory for audio files
	tmpDir, err := os.MkdirTemp("", "audio-files-*")
	if err != nil {
		return nil, fmt.Errorf("error creating temporary directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create directory for transcriptions
	if err := processor.CreateTranscriptionDirectory(f.DirPath); err != nil {
		return nil, fmt.Errorf("error creating transcriptions directory: %w", err)
	}

	// Get all audio files
	audios, err := f.GetAudioFiles()
	if err != nil {
		return nil, fmt.Errorf("error getting audio files: %w", err)
	}

	var texts []FactoryFileContent

	// Process each audio file
	for _, audio := range audios {
		// Transcribe the audio file
		transcriptionPath, err := processor.TranscribeAudioFile(f.DirPath+"/"+audio.File, filepath.Join(f.DirPath, "transcriptions"))
		if err != nil {
			return nil, fmt.Errorf("error transcribing file %s: %w", audio.File, err)
		}

		// Read the transcription
		content, err := os.ReadFile(transcriptionPath)
		if err != nil {
			return nil, fmt.Errorf("error reading transcription for %s: %w", audio.File, err)
		}

		texts = append(texts, FactoryFileContent{
			File:    audio.File,
			Content: string(content),
		})
	}

	return texts, nil
}
