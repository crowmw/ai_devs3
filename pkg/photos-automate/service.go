package photosautomate

import (
	"fmt"

	"github.com/crowmw/ai_devs3/pkg/c3ntrala"
	"github.com/crowmw/ai_devs3/pkg/env"
)

type Service struct {
	envSvc      *env.Service
	c3ntralaSvc *c3ntrala.Service
	photos      []string
}

func NewService(envSvc *env.Service, c3ntralaSvc *c3ntrala.Service) (*Service, error) {
	photos, err := c3ntralaSvc.GetPhotos()
	if err != nil {
		return nil, err
	}

	fmt.Println("❇️ Photos:", photos)
	return &Service{envSvc: envSvc, c3ntralaSvc: c3ntralaSvc, photos: photos}, nil
}

func (s *Service) GetPhotos() []string {
	return s.photos
}
