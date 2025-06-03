package recognize

import (
	"fmt"
	"strings"

	"github.com/crowmw/ai_devs3/pkg/ai"
	"github.com/crowmw/ai_devs3/pkg/c3ntrala"
	"github.com/crowmw/ai_devs3/pkg/env"
	photosautomate "github.com/crowmw/ai_devs3/pkg/photos-automate"
	"github.com/sashabaranov/go-openai"
)

type Service struct {
	envSvc      *env.Service
	aiSvc       *ai.Service
	photosSvc   *photosautomate.Service
	c3ntralaSvc *c3ntrala.Service
	photos      []string
}

func NewService(envSvc *env.Service, aiSvc *ai.Service, photosSvc *photosautomate.Service, c3ntralaSvc *c3ntrala.Service) (*Service, error) {
	photos := photosSvc.GetPhotos()
	fmt.Println("❇️ Photos:", photos)
	return &Service{envSvc: envSvc, aiSvc: aiSvc, photosSvc: photosSvc, photos: photos, c3ntralaSvc: c3ntralaSvc}, nil
}

func (s *Service) StartRecognize() (string, error) {
	fmt.Println("❇️ Starting recognize")
	if len(s.photos) == 0 {
		fmt.Println("❇️ No photos to recognize")
		return "", nil
	}

	fixedPhotos := s.processPhotos()
	fmt.Println("❇️ Photos after processing:", fixedPhotos)

	imageUrls := make([]openai.ChatMessagePart, len(fixedPhotos))
	for i, photo := range fixedPhotos {
		imageUrls[i] = openai.ChatMessagePart{
			Type: "image_url",
			ImageURL: &openai.ChatMessageImageURL{
				URL:    photo,
				Detail: openai.ImageURLDetailHigh,
			},
		}
	}

	var prompt = ai.ChatCompletionConfig{
		Model: "gpt-4.1",
		Messages: []openai.ChatCompletionMessage{
			{Role: "system", Content: barbaraDescriptionSystemPrompt},
			{Role: "user", MultiContent: []openai.ChatMessagePart{
				{Type: "text", Text: "Opisz osobę na tych zdjęciach: "},
			}},
		},
	}

	prompt.Messages[1].MultiContent = append(prompt.Messages[1].MultiContent, imageUrls...)

	fmt.Println("❇️ Describing person onphotos...")
	barbaraDescription, err := s.aiSvc.ChatCompletion(prompt)
	if err != nil {
		fmt.Println("Error getting AI description:", err)
		return "", err
	}

	return barbaraDescription.Choices[0].Message.Content, nil
}

func (s *Service) processPhotos() []string {
	// Keep track of photos that need processing
	needsProcessing := make(map[string]bool)
	var fixedPhotos []string
	for _, photo := range s.photos {
		needsProcessing[photo] = true
	}

	// Continue processing until all photos are GOOD
	for len(needsProcessing) > 0 {
		for photo := range needsProcessing {
			response := s.analyzePhoto(photo)
			fmt.Println("❇️ Response:", response)

			if strings.HasPrefix(response, "GOOD") {
				delete(needsProcessing, photo)
				fixedPhotos = append(fixedPhotos, photo)
				continue
			}

			if strings.HasPrefix(response, "INVALID") {
				fmt.Println("❇️ Photo is invalid:", response)
				// Remove the invalid photo from s.photos
				for i, p := range s.photos {
					if p == photo {
						s.photos = append(s.photos[:i], s.photos[i+1:]...)
						break
					}
				}
				delete(needsProcessing, photo)
				continue
			}

			if strings.HasPrefix(response, "REPAIR") || strings.HasPrefix(response, "BRIGHTEN") || strings.HasPrefix(response, "DARKEN") {
				fmt.Println("❇️ Photo needs processing:", response)
				fixedPhoto, err := s.c3ntralaSvc.FixPhoto(response)
				if err != nil {
					fmt.Println("❇️ Error fixing photo:", err)
					continue
				}
				delete(needsProcessing, photo)
				needsProcessing[fixedPhoto] = true
			}
		}
	}

	return fixedPhotos
}

func (s *Service) analyzePhoto(url string) string {
	fmt.Println("❇️ Analyzing photo:", url)

	// Replace .PNG with -small.PNG in the URL
	smallUrl := strings.Replace(url, ".PNG", "-small.PNG", 1)

	response, err := s.aiSvc.ChatCompletion(ai.ChatCompletionConfig{
		Model: "gpt-4o",
		Messages: []openai.ChatCompletionMessage{
			{Role: "system", Content: analyzePhotoSystemPrompt},
			{Role: "user", MultiContent: []openai.ChatMessagePart{
				{Type: "text", Text: "Analyze this photo: " + url},
				{Type: "image_url", ImageURL: &openai.ChatMessageImageURL{
					URL:    smallUrl,
					Detail: openai.ImageURLDetailHigh,
				}},
			}},
		},
	})
	if err != nil {
		fmt.Println("❇️ Error analyzing photo:", err)
		return ""
	}

	return response.Choices[0].Message.Content
}

const analyzePhotoSystemPrompt = `You are a helpful assistant that analyzes photos.
Try to count persons in the photo. Try to recognize faces. In the photo should be at least one woman.
If you do not see any person or do not see detailed faces, then try to figure out what is the problem.

Based on your analysis, respond ONLY with one of these formats:
- For noisy/glitchy photo: 'REPAIR [filename]'
- For too bright photo: 'DARKEN [filename]'
- For too dark photo: 'BRIGHTEN [filename]'
- For photo that not need any processing: 'GOOD [filename]'
- For photo that not contain any person: 'INVALID [filename]'

<examples>
USER: Analyze this photo: https://c3ntrala.ag3nts.org/dane/barbara/IMG_0000.PNG
ASSISTANT: REPAIR IMG_0000.PNG

USER: Analyze this photo: https://c3ntrala.ag3nts.org/dane/barbara/IMG_0001.PNG
ASSISTANT: DARKEN IMG_0001.PNG

USER: Analyze this photo: https://c3ntrala.ag3nts.org/dane/barbara/IMG_0002.PNG
ASSISTANT: BRIGHTEN IMG_0002.PNG
</examples>
`

const barbaraDescriptionSystemPrompt = `You are a helpful assistant that describes woman.
You are given a list of photos.
You need to describe each photo in a few sentences.
Focus especially on the woman's appearance and distinguishing marks.
Respond in Polish language.
`
