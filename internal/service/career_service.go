package service

import (
	"context"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"X-BE/internal/models"
	"X-BE/internal/repository"
)

func NewCareerService(repo *repository.CareerRepository, newsletter *NewsletterService) *CareerService {
	return &CareerService{repo: repo, newsletter: newsletter}
}

func (s *CareerService) Create(ctx context.Context, req models.CareerCreateRequest) (map[string]any, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("career service is not initialized")
	}
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.City) == "" || strings.TrimSpace(req.Country) == "" || strings.TrimSpace(req.Department) == "" || strings.TrimSpace(req.Message) == "" {
		return nil, fmt.Errorf("name, city, country, department and message are required")
	}
	email := models.NormalizeEmail(req.Email)
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, fmt.Errorf("invalid email")
	}
	app := models.CareerApplication{
		Name:       strings.TrimSpace(req.Name),
		Email:      email,
		Phone:      strings.TrimSpace(req.Phone),
		City:       strings.TrimSpace(req.City),
		State:      strings.TrimSpace(req.State),
		Country:    strings.TrimSpace(req.Country),
		Department: strings.TrimSpace(req.Department),
		Message:    strings.TrimSpace(req.Message),
		Date:       time.Now().UTC(),
		Status:     "applied",
		Source:     "website",
	}
	outcome, err := s.repo.Create(ctx, app)
	if err != nil {
		return nil, err
	}
	if s.newsletter != nil {
		_, _ = s.newsletter.Subscribe(ctx, models.NewsletterSubscriptionRequest{Email: email})
	}
	s.sendAcknowledgmentEmail(email, req.Name, req.Department, app.Date)
	result := map[string]any{
		"mongodb": "saved",
		"zoho":    outcome.ZohoStatus,
		"success": true,
	}
	if len(outcome.Errors) > 0 {
		result["errors"] = outcome.Errors
	}
	return result, nil
}
