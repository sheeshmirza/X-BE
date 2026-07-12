package service

import (
	"context"
	"fmt"

	"X-BE/internal/zoho"
)

func NewZohoService(client *zoho.Client) *ZohoService {
	return &ZohoService{client: client}
}

func (s *ZohoService) GetAuthorizationURL(redirectURI, scope string) (string, string, error) {
	if s == nil || s.client == nil {
		return "", "", fmt.Errorf("Zoho service is not initialized")
	}
	return s.client.BuildAuthorizationURL(redirectURI, scope)
}

func (s *ZohoService) HandleOAuthCallback(ctx context.Context, code, redirectURI, accountID string) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("Zoho service is not initialized")
	}
	return s.client.ExchangeAuthorizationCode(ctx, code, redirectURI, accountID)
}
