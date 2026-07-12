package service

import (
	"context"
	"fmt"
	"net/mail"

	"X-BE/internal/models"
	"X-BE/internal/repository"
)

func NewNewsletterService(repo *repository.NewsletterRepository) *NewsletterService {
	return &NewsletterService{repo: repo}
}

func (s *NewsletterService) Subscribe(ctx context.Context, req models.NewsletterSubscriptionRequest) (map[string]any, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("newsletter service is not initialized")
	}
	email := models.NormalizeEmail(req.Email)
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, fmt.Errorf("invalid email")
	}
	isSubscribed, err := s.repo.IsSubscribed(ctx, email)
	if err != nil {
		return nil, err
	}
	if isSubscribed {
		return map[string]any{
			"message":           "Email already subscribed",
			"email":             email,
			"alreadySubscribed": true,
		}, nil
	}
	outcome, err := s.repo.Subscribe(ctx, email)
	if err != nil {
		return nil, err
	}
	result := map[string]any{
		"message": "Successfully subscribed to newsletter",
		"email":   email,
		"mongodb": "saved",
		"zoho":    outcome.ZohoStatus,
		"success": true,
		"errors":  outcome.Errors,
	}
	if result["errors"] == nil {
		result["errors"] = []string{}
	}
	s.sendAcknowledgmentEmail(email)
	return result, nil
}

func (s *NewsletterService) Unsubscribe(ctx context.Context, email string) (map[string]any, error) {
	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("newsletter service is not initialized")
	}
	normalized := models.NormalizeEmail(email)
	if _, err := mail.ParseAddress(normalized); err != nil {
		return nil, fmt.Errorf("invalid email")
	}

	updated, err := s.repo.Unsubscribe(ctx, normalized)
	if err != nil {
		return nil, err
	}
	if !updated {
		return map[string]any{
			"message": "Email was not subscribed",
			"email":   normalized,
		}, nil
	}
	return map[string]any{
		"message": "Unsubscribed successfully",
		"email":   normalized,
	}, nil
}
