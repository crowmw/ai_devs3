package c3ntrala

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/crowmw/ai_devs3/pkg/env"
	"github.com/crowmw/ai_devs3/pkg/http"
	"github.com/crowmw/ai_devs3/pkg/processor"
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

func (s *Service) PostReport(task string, answer interface{}, justUpdate bool) (string, error) {
	fmt.Println("Sending report to C3ntrala...")

	postData := map[string]interface{}{
		"task":       task,
		"answer":     answer,
		"apikey":     s.apiKey,
		"justUpdate": justUpdate,
	}

	fmt.Println(postData)

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

func (s *Service) GetQuestions(path string) (map[string]string, error) {
	body, err := http.FetchData(s.baseUrl + "/data/" + s.apiKey + path)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var questions map[string]string
	err = json.Unmarshal([]byte(body), &questions)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return questions, nil
}

func (s *Service) GetSoftoQuestions() (map[string]string, error) {
	body, err := http.FetchData(s.baseUrl + "/data/" + s.apiKey + "/softo.json")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var questions map[string]string
	err = json.Unmarshal([]byte(body), &questions)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return questions, nil
}

type PhoneData struct {
	Rozmowa1 []string `json:"rozmowa1"`
	Rozmowa2 []string `json:"rozmowa2"`
	Rozmowa3 []string `json:"rozmowa3"`
	Rozmowa4 []string `json:"rozmowa4"`
	Rozmowa5 []string `json:"rozmowa5"`
}

func (s *Service) GetPhoneData() (PhoneData, error) {
	body, err := http.FetchData(s.baseUrl + "/data/" + s.apiKey + "/phone_sorted.json")
	if err != nil {
		fmt.Println(err)
		return PhoneData{}, err
	}

	var phoneData PhoneData
	err = json.Unmarshal([]byte(body), &phoneData)
	if err != nil {
		fmt.Println(err)
		return PhoneData{}, err
	}
	return phoneData, nil
}

func (s *Service) GetLogs() (string, error) {
	body, err := http.FetchData(s.baseUrl + "/data/" + s.apiKey + "/gps.txt")
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	logs := processor.ReadLinesFromTextFileAsString(body)
	return logs, nil
}

func (s *Service) GetGpsQuestion() (string, error) {
	body, err := http.FetchData(s.baseUrl + "/data/" + s.apiKey + "/gps_question.json")
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	var question map[string]string
	err = json.Unmarshal([]byte(body), &question)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return question["question"], nil
}
