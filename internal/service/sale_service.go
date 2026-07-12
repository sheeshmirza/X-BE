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

func NewSaleService(repo *repository.SaleRepository, newsletter *NewsletterService) *SaleService {
	return &SaleService{repo: repo, newsletter: newsletter}
}

func (s *SaleService) Create(ctx context.Context, req models.SaleCreateRequest) (map[string]any, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("sale service is not initialized")
	}
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Query) == "" {
		return nil, fmt.Errorf("name and query are required")
	}
	email := models.NormalizeEmail(req.Email)
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, fmt.Errorf("invalid email")
	}

	lead := models.SaleLead{
		Name:   strings.TrimSpace(req.Name),
		Email:  email,
		Query:  strings.TrimSpace(req.Query),
		Date:   time.Now().UTC(),
		Status: "new",
		Source: "website",
	}
	outcome, err := s.repo.Create(ctx, lead)
	if err != nil {
		return nil, err
	}
	if s.newsletter != nil {
		_, _ = s.newsletter.Subscribe(ctx, models.NewsletterSubscriptionRequest{Email: email})
	}
	s.sendAcknowledgmentEmail(email, req.Name, req.Query)
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
