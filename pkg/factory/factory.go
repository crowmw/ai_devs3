package factory

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/crowmw/ai_devs3/pkg/config"
	"github.com/crowmw/ai_devs3/pkg/http"
	"github.com/crowmw/ai_devs3/pkg/processor"
	"github.com/otiai10/gosseract/v2"
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

// GetTextFiles reads all .txt files from the factory directory and returns their names and contents
func (f *Factory) GetTextFiles() ([]FactoryFileContent, error) {
	var texts []FactoryFileContent

	files, err := os.ReadDir(f.DirPath)
	if err != nil {
		return nil, fmt.Errorf("error reading factory directory: %w", err)
	}

	for _, file := range files {
		// Skip directories and non-txt files
		if file.IsDir() || !strings.HasSuffix(strings.ToLower(file.Name()), ".txt") {
			continue
		}

		// Read file contents
		content, err := os.ReadFile(filepath.Join(f.DirPath, file.Name()))
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

// TextFile represents a text file with its name and content
type FactoryFileImage struct {
	File  string
	Image []byte
}

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

	// Create a directory for transcriptions

	transcriptionsDir := filepath.Join(f.DirPath + "/transcriptions")
	// Check if transcription already exists
	if _, err := os.Stat(transcriptionsDir); err == nil {
		fmt.Println("Transcription already exists for:", transcriptionsDir)
	} else if err := os.Mkdir(transcriptionsDir, 0755); err != nil {
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
		transcriptionPath, err := processor.TranscribeAudioFile(f.DirPath+"/"+audio.File, transcriptionsDir)
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
