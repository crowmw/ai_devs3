package processor

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/crowmw/ai_devs3/pkg/ai"
	"github.com/crowmw/ai_devs3/pkg/config"
	httpclient "github.com/crowmw/ai_devs3/pkg/http"
	"golang.org/x/net/html"
)

// createTempFileWithData creates a temporary file and writes the provided data to it
// Returns the temporary file name and cleanup function
func createTempFileWithData(data []byte, prefix string) (string, func(), error) {
	tmpFile, err := os.CreateTemp("", prefix)
	if err != nil {
		return "", nil, fmt.Errorf("error creating temp file: %w", err)
	}

	// Write data to temp file
	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return "", nil, fmt.Errorf("error writing data to temp file: %w", err)
	}
	tmpFile.Close()

	// Return cleanup function
	cleanup := func() {
		os.Remove(tmpFile.Name())
	}

	return tmpFile.Name(), cleanup, nil
}

// ExtractZipToDirectory extracts contents of a zip file to a directory
// If the directory doesn't exist, it creates a new one with the given name
func ExtractZipToDirectory(zipData []byte, dirName string) error {
	// Create temporary file with zip data
	tmpFileName, cleanup, err := createTempFileWithData(zipData, "temp-zip-*")
	if err != nil {
		return err
	}
	defer cleanup()

	// Open the zip file
	reader, err := zip.OpenReader(tmpFileName)
	if err != nil {
		return fmt.Errorf("error opening zip file: %w", err)
	}
	defer reader.Close()

	// Check if the directory already exists
	if _, err := os.Stat(dirName); err == nil {
		fmt.Println("Audio files already extracted, skipping extraction.")
		return nil
	}

	// Create the target directory if it doesn't exist
	if err := os.MkdirAll(dirName, 0755); err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}

	// Extract each file from the zip
	for _, file := range reader.File {
		// Create the full path for the file
		filePath := filepath.Join(dirName, file.Name)

		// Create directories if needed
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(filePath, 0755); err != nil {
				return fmt.Errorf("error creating directory: %w", err)
			}
			continue
		}

		// Create the file
		outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return fmt.Errorf("error creating file: %w", err)
		}

		// Open the zip file
		rc, err := file.Open()
		if err != nil {
			outFile.Close()
			return fmt.Errorf("error opening zip file: %w", err)
		}

		// Copy the contents
		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return fmt.Errorf("error extracting file: %w", err)
		}
	}

	fmt.Println("Audio files extracted to:", dirName)

	return nil
}

// TranscribeAudioFile transcribes a single audio file and saves the transcription
// Returns the path to the saved transcription file
func TranscribeAudioFile(audioPath string, outputDir string) (string, error) {
	// Get the base name without extension
	baseName := strings.TrimSuffix(filepath.Base(audioPath), filepath.Ext(audioPath))
	fmt.Println("Processing audio file:", baseName+filepath.Ext(audioPath))

	// Set up transcription file path
	transcriptionPath := filepath.Join(outputDir, baseName+".txt")

	// Check if transcription already exists
	if _, err := os.Stat(transcriptionPath); err == nil {
		fmt.Println("Transcription already exists for:", baseName)
		return transcriptionPath, nil
	}

	// Transcribe the audio file
	transcription, err := ai.TranscribeAudio(audioPath)
	if err != nil {
		return "", fmt.Errorf("error transcribing file %s: %w", audioPath, err)
	}

	// Save the transcription
	fmt.Println("Transcription:", transcription)
	fmt.Println("Transcription path:", transcriptionPath)
	if err := os.WriteFile(transcriptionPath, []byte(transcription), 0644); err != nil {
		return "", fmt.Errorf("error saving transcription for %s: %w", audioPath, err)
	}

	return transcriptionPath, nil
}

// ProcessAudioFiles transcribes audio files and saves transcriptions
func ProcessAudioFiles(hashDir string) error {
	// Check if transcriptions directory exists
	transcriptionsDir := filepath.Join(hashDir, "transcriptions")
	if _, err := os.Stat(transcriptionsDir); err == nil {
		fmt.Println("Transcriptions directory already exists, skipping processing.")
		return nil
	}

	// Create transcriptions directory
	if err := os.MkdirAll(transcriptionsDir, 0755); err != nil {
		return fmt.Errorf("error creating transcriptions directory: %w", err)
	}

	// Walk through the hash directory
	err := filepath.Walk(hashDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-audio files
		if info.IsDir() || filepath.Ext(path) != ".m4a" {
			return nil
		}

		// Transcribe the audio file
		_, err = TranscribeAudioFile(path, transcriptionsDir)
		if err != nil {
			return fmt.Errorf("error processing file %s: %w", path, err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error processing audio files: %w", err)
	}

	return nil
}

// CreateTranscriptionDirectory creates a directory for transcriptions if it doesn't exist
func CreateTranscriptionDirectory(dirPath string) error {
	transcriptionsDir := filepath.Join(dirPath, "transcriptions")
	// Check if transcription already exists
	if _, err := os.Stat(transcriptionsDir); err == nil {
		fmt.Println("Transcription already exists for:", transcriptionsDir)
		return nil
	}
	return os.Mkdir(transcriptionsDir, 0755)
}

// ProcessAudioElements processes audio elements in HTML and replaces them with transcriptions
func ProcessAudioElements(doc *html.Node, outputDir string) error {
	// Find all audio elements
	audioNodes := htmlquery.Find(doc, "//source")
	for i, node := range audioNodes {
		src := htmlquery.SelectAttr(node, "src")
		if src == "" {
			continue
		}

		// Handle relative URLs
		if !strings.HasPrefix(src, "http") {
			src = config.GetC3ntralaURL() + "/dane/" + src
		}

		fmt.Println("Fetching audio:", src)
		// Fetch audio
		audioData, err := httpclient.FetchData(src)
		if err != nil {
			return fmt.Errorf("error fetching audio %s: %w", src, err)
		}

		// Extract filename from URL or create numbered filename
		filename := filepath.Base(src)
		if filename == "" {
			filename = fmt.Sprintf("audio_%d.mp3", i)
		}

		// Save audio
		audioPath := filepath.Join(outputDir, filename)
		if err := os.WriteFile(audioPath, audioData, 0644); err != nil {
			return fmt.Errorf("error saving audio %s: %w", audioPath, err)
		}

		fmt.Println("Transcribing audio:", audioPath)
		// Transcribe audio
		transcription, err := ai.TranscribeAudioAndFormat(audioPath)
		if err != nil {
			return fmt.Errorf("error transcribing audio %s: %w", audioPath, err)
		}

		// Create a new text node with the transcription
		descriptionNode := &html.Node{
			Type: html.TextNode,
			Data: transcription,
		}

		// Replace the audio node with the transcription node
		parent := node.Parent
		if parent != nil {
			parent.InsertBefore(descriptionNode, node)
			parent.RemoveChild(node)
		}
	}

	return nil
}

// ProcessImageElements processes image elements in HTML and replaces them with descriptions
func ProcessImageElements(doc *html.Node, outputDir string) error {
	// Find all image elements
	imageNodes := htmlquery.Find(doc, "//img")
	for i, node := range imageNodes {
		src := htmlquery.SelectAttr(node, "src")
		if src == "" {
			continue
		}

		// Handle relative URLs
		if !strings.HasPrefix(src, "http") {
			src = config.GetC3ntralaURL() + "/dane/" + src
		}

		// Fetch image
		fmt.Println("Fetching image:", src)
		imageData, err := httpclient.FetchData(src)
		if err != nil {
			return fmt.Errorf("error fetching image %s: %w", src, err)
		}

		// Extract filename from URL or create numbered filename
		filename := filepath.Base(src)
		if filename == "" {
			filename = fmt.Sprintf("image_%d.png", i)
		}

		// Save image
		imagePath := filepath.Join(outputDir, filename)
		if err := os.WriteFile(imagePath, imageData, 0644); err != nil {
			return fmt.Errorf("error saving image %s: %w", imagePath, err)
		}

		// Read image to base64
		fmt.Println("Reading image to base64:", imagePath)
		imageBase64, err := ReadImageToBase64(imagePath)
		if err != nil {
			return fmt.Errorf("error reading image %s: %w", imagePath, err)
		}

		// Get image format from extension
		format := strings.TrimPrefix(strings.ToLower(filepath.Ext(imagePath)), ".")

		// Describe image
		imageDescription, err := ai.DescribeImageAndFormat(imageBase64, format)
		if err != nil {
			return fmt.Errorf("error describing image %s: %w", imagePath, err)
		}

		// Create a new text node with the image description
		descriptionNode := &html.Node{
			Type: html.TextNode,
			Data: imageDescription,
		}

		// Replace the image node with the description node
		parent := node.Parent
		if parent != nil {
			parent.InsertBefore(descriptionNode, node)
			parent.RemoveChild(node)
		}
	}

	return nil
}
