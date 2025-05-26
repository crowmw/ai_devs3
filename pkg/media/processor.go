package media

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/crowmw/ai_devs3/pkg/ai"
	httpclient "github.com/crowmw/ai_devs3/pkg/http"
	"golang.org/x/net/html"
)

// Processor handles media file processing
type Processor struct {
	outputDir string
	htmlProc  HTMLProcessor
}

// HTMLProcessor interface for HTML operations
type HTMLProcessor interface {
	GetFullURL(path string) string
	ReplaceNode(node *html.Node, text string)
}

// NewProcessor creates a new media processor
func NewProcessor(outputDir string, htmlProc HTMLProcessor) *Processor {
	return &Processor{
		outputDir: outputDir,
		htmlProc:  htmlProc,
	}
}

// ProcessImage processes a single image and returns its description
func (p *Processor) ProcessImage(src string, index int) (string, error) {
	// Get full URL
	fullURL := p.htmlProc.GetFullURL(src)

	// Fetch image
	fmt.Println("Fetching image:", fullURL)
	imageData, err := httpclient.FetchData(fullURL)
	if err != nil {
		return "", fmt.Errorf("error fetching image %s: %w", src, err)
	}

	// Extract filename from URL or create numbered filename
	filename := filepath.Base(src)
	if filename == "" {
		filename = fmt.Sprintf("image_%d.png", index)
	}

	// Save image
	imagePath := filepath.Join(p.outputDir, filename)
	if err := os.WriteFile(imagePath, imageData, 0644); err != nil {
		return "", fmt.Errorf("error saving image %s: %w", imagePath, err)
	}

	// Read image to base64
	fmt.Println("Reading image to base64:", imagePath)
	imageBase64, err := ReadImageToBase64(imagePath)
	if err != nil {
		return "", fmt.Errorf("error reading image %s: %w", imagePath, err)
	}

	// Get image format from extension
	format := strings.TrimPrefix(strings.ToLower(filepath.Ext(imagePath)), ".")

	// Describe image
	return ai.DescribeImageAndFormat(imageBase64, format)
}

// ProcessAudio processes a single audio file and returns its transcription
func (p *Processor) ProcessAudio(src string, index int) (string, error) {
	// Get full URL
	fullURL := p.htmlProc.GetFullURL(src)

	// Fetch audio
	fmt.Println("Fetching audio:", fullURL)
	audioData, err := httpclient.FetchData(fullURL)
	if err != nil {
		return "", fmt.Errorf("error fetching audio %s: %w", src, err)
	}

	// Extract filename from URL or create numbered filename
	filename := filepath.Base(src)
	if filename == "" {
		filename = fmt.Sprintf("audio_%d.mp3", index)
	}

	// Save audio
	audioPath := filepath.Join(p.outputDir, filename)
	if err := os.WriteFile(audioPath, audioData, 0644); err != nil {
		return "", fmt.Errorf("error saving audio %s: %w", audioPath, err)
	}

	// Transcribe audio
	fmt.Println("Transcribing audio:", audioPath)
	return ai.TranscribeAudioAndFormat(audioPath)
}

// ReadImageToBase64 reads an image file and converts it to base64
func ReadImageToBase64(path string) (string, error) {
	// Read the image file
	imageBytes, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("error reading image file: %v", err)
	}

	// Convert to base64
	base64String := base64.StdEncoding.EncodeToString(imageBytes)
	return base64String, nil
}
