package c3ntrala

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/crowmw/ai_devs3/pkg/env"
	"github.com/crowmw/ai_devs3/pkg/http"
)

type Service struct {
	envSvc  *env.Service
	baseUrl string
	apiKey  string
}

func NewService(envSvc *env.Service) (*Service, error) {
	return &Service{
		envSvc:  envSvc,
		baseUrl: envSvc.GetC3ntralaURL(),
		apiKey:  envSvc.GetMyAPIKey(),
	}, nil
}

func (s *Service) PostReport(task string, answer interface{}) (string, error) {
	fmt.Println("Sending report to C3ntrala...")

	postData := map[string]interface{}{
		"task":   task,
		"answer": answer,
		"apikey": s.apiKey,
	}

	fmt.Println("❇️ Post data:", postData)

	resp, err := http.SendPost(
		s.baseUrl+"/report",
		postData,
	)

	return resp, err
}

func (s *Service) GetBarbaraNote() (string, error) {
	url := s.baseUrl + "/dane/barbara.txt"
	body, err := http.FetchData(url)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (s *Service) GetPlacesWhereSeen(firstName string) ([]string, error) {
	url := s.baseUrl + "/people"

	body, err := http.SendPost(url, map[string]interface{}{
		"apikey": s.apiKey,
		"query":  firstName,
	})
	if err != nil {
		return nil, err
	}

	// Remove [**RESTRICTED DATA**] before unmarshaling
	cleanBody := strings.ReplaceAll(body, "[**RESTRICTED DATA**]", "")

	var response struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal([]byte(cleanBody), &response); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	return strings.Split(response.Message, " "), nil
}

func (s *Service) GetWhoWasSeenThere(city string) ([]string, error) {
	url := s.baseUrl + "/places"

	body, err := http.SendPost(url, map[string]interface{}{
		"apikey": s.apiKey,
		"query":  city,
	})
	if err != nil {
		return nil, err
	}

	// Remove [**RESTRICTED DATA**] before unmarshaling
	cleanBody := strings.ReplaceAll(body, "[**RESTRICTED DATA**]", "")

	var response struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal([]byte(cleanBody), &response); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	// Convert Polish characters to ASCII
	message := strings.NewReplacer(
		"Ą", "A",
		"Ć", "C",
		"Ę", "E",
		"Ł", "L",
		"Ń", "N",
		"Ó", "O",
		"Ś", "S",
		"Ź", "Z",
		"Ż", "Z",
		"ą", "a",
		"ć", "c",
		"ę", "e",
		"ł", "l",
		"ń", "n",
		"ó", "o",
		"ś", "s",
		"ź", "z",
		"ż", "z",
	).Replace(response.Message)

	return strings.Split(message, " "), nil
}

func (s *Service) GetPhotos() ([]string, error) {
	url := s.baseUrl + "/report"

	body, err := http.SendPost(url, map[string]interface{}{
		"task":   "photos",
		"apikey": s.apiKey,
		"answer": "START",
	})

	var resp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	err = json.Unmarshal([]byte(body), &resp)
	if err != nil {
		return nil, err
	}

	// Extract PNG filenames using regex
	re := regexp.MustCompile(`IMG_\d+\.PNG`)
	matches := re.FindAllString(resp.Message, -1)

	// Create full URLs
	baseURL := s.baseUrl + "/dane/barbara/"
	urls := make([]string, len(matches))
	for i, filename := range matches {
		urls[i] = baseURL + filename
	}

	return urls, nil
}

func (s *Service) FixPhoto(answer string) (string, error) {
	url := s.baseUrl + "/report"

	body, err := http.SendPost(url, map[string]interface{}{
		"task":   "photos",
		"apikey": s.apiKey,
		"answer": answer,
	})

	var resp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	err = json.Unmarshal([]byte(body), &resp)
	if err != nil {
		return "", err
	}
	fmt.Println("❇️ Response:", resp.Message)
	// Extract PNG filename using regex
	re := regexp.MustCompile(`IMG_\d+_[A-Z0-9]+\.PNG`)
	matches := re.FindAllString(resp.Message, -1)

	// Create full URL
	baseURL := s.baseUrl + "/dane/barbara/"
	urls := make([]string, len(matches))
	for i, filename := range matches {
		urls[i] = baseURL + filename
	}

	fmt.Println("❇️ FIXED URL:", urls[0])

	return urls[0], nil
}
