package processor

import (
	"fmt"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/crowmw/ai_devs3/pkg/ai"
	"github.com/gen2brain/go-fitz"
	"github.com/otiai10/gosseract/v2"
)

// PDFService handles PDF processing operations
type PDFService struct {
	pdfURL    string
	tempDir   string
	client    *http.Client
	tesseract *gosseract.Client
	aiSvc     *ai.Service
}

// NewPDFService creates a new instance of PDFService
func NewPDFService(pdfURL string, aiSvc *ai.Service) (*PDFService, error) {
	dir := "cmd/s04e05/pdf_processing"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Set Tesseract data path
	// if err := os.Setenv("TESSDATA_PREFIX", "/usr/local/share/tessdata"); err != nil {
	// 	return nil, fmt.Errorf("failed to set TESSDATA_PREFIX: %w", err)
	// }

	client := gosseract.NewClient()
	if err := client.SetLanguage("pol"); err != nil {
		return nil, fmt.Errorf("failed to set OCR language: %w", err)
	}

	return &PDFService{
		pdfURL:    pdfURL,
		tempDir:   dir,
		client:    &http.Client{},
		tesseract: client,
		aiSvc:     aiSvc,
	}, nil
}

// GetText extracts all text from the PDF file
func (s *PDFService) GetText() (string, error) {
	// Download PDF file
	fmt.Println("Downloading PDF file...")
	resp, err := s.client.Get(s.pdfURL)
	if err != nil {
		return "", fmt.Errorf("failed to download PDF: %w", err)
	}
	defer resp.Body.Close()

	fmt.Println("Creating PDF file...")
	pdfPath := filepath.Join(s.tempDir, "document.pdf")
	file, err := os.Create(pdfPath)
	if err != nil {
		return "", fmt.Errorf("failed to create PDF file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save PDF file: %w", err)
	}

	fmt.Println("Extracting text from PDF file...")
	// Open the PDF document
	doc, err := fitz.New(pdfPath)
	if err != nil {
		return "", fmt.Errorf("failed to open PDF document: %w", err)
	}
	defer doc.Close()

	// Extract text from each page
	fmt.Println("Extracting text from each page...")
	text := ""
	for i := 0; i < doc.NumPage(); i++ {
		fmt.Println("Extracting page", i+1)
		pageText, err := doc.Text(i)
		if err != nil {
			return "", fmt.Errorf("failed to extract text from page %d: %w", i+1, err)
		}
		text += pageText + "\n"
	}

	fmt.Println("Returning text...")
	return text, nil
}

// ExtractTextFromPage extracts text from a specific page using OCR
func (s *PDFService) ExtractTextFromPage(pageNumber int) (string, error) {
	// Download PDF file if not already downloaded
	pdfPath := filepath.Join(s.tempDir, "document.pdf")
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		resp, err := s.client.Get(s.pdfURL)
		if err != nil {
			return "", fmt.Errorf("failed to download PDF: %w", err)
		}
		defer resp.Body.Close()

		file, err := os.Create(pdfPath)
		if err != nil {
			return "", fmt.Errorf("failed to create PDF file: %w", err)
		}
		defer file.Close()

		_, err = io.Copy(file, resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to save PDF file: %w", err)
		}
	}

	// Create a directory for the extracted page
	pageDir := filepath.Join(s.tempDir, fmt.Sprintf("page_%d", pageNumber))
	if err := os.MkdirAll(pageDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create page directory: %w", err)
	}

	// Check if text file already exists
	textPath := filepath.Join(pageDir, "text.txt")
	if text, err := os.ReadFile(textPath); err == nil {
		fmt.Println("Using cached text for page", pageNumber)
		return string(text), nil
	}

	// Open the PDF document
	doc, err := fitz.New(pdfPath)
	if err != nil {
		return "", fmt.Errorf("failed to open PDF document: %w", err)
	}
	defer doc.Close()

	// Validate page number
	if pageNumber < 1 || pageNumber > doc.NumPage() {
		return "", fmt.Errorf("invalid page number: %d (document has %d pages)", pageNumber, doc.NumPage())
	}

	// Extract the page as image
	fmt.Println("Extracting page as image...")
	img, err := doc.Image(pageNumber - 1) // go-fitz uses 0-based indexing
	if err != nil {
		return "", fmt.Errorf("failed to extract page as image: %w", err)
	}

	fmt.Println("Saving image...")
	// Save the image
	imagePath := filepath.Join(pageDir, "page.png")
	file, err := os.Create(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to create image file: %w", err)
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		return "", fmt.Errorf("failed to save image: %w", err)
	}

	fmt.Println("Processing image with OCR...")
	result, err := s.aiSvc.OCRImage(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to process image: %w", err)
	}

	// Save the extracted text
	if err := os.WriteFile(textPath, []byte(result.Text), 0644); err != nil {
		return "", fmt.Errorf("failed to save text file: %w", err)
	}

	return result.Text, nil
}

// Close cleans up temporary resources
func (s *PDFService) Close() error {
	if err := s.tesseract.Close(); err != nil {
		return fmt.Errorf("failed to close tesseract client: %w", err)
	}
	if err := os.RemoveAll(s.tempDir); err != nil {
		return fmt.Errorf("failed to clean up temp directory: %w", err)
	}
	return nil
}
