package zoho

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (c *Client) BuildAuthorizationURL(redirectURI, scope string) (string, string, error) {
	if !c.Enabled() {
		return "", "", ErrNotConfigured
	}
	finalRedirect := strings.TrimSpace(redirectURI)
	if finalRedirect == "" {
		return "", "", fmt.Errorf("redirect URI is required")
	}
	finalScope := strings.TrimSpace(scope)
	if finalScope == "" {
		finalScope = defaultZohoScope
	}
	authURL := zohoAuthBaseURL + "/oauth/v2/auth"
	query := url.Values{}
	query.Set("scope", finalScope)
	query.Set("client_id", c.cfg.ClientID)
	query.Set("response_type", "code")
	query.Set("access_type", "offline")
	query.Set("prompt", "consent")
	query.Set("redirect_uri", finalRedirect)
	return authURL + "?" + query.Encode(), finalRedirect, nil
}

func (c *Client) ExchangeAuthorizationCode(ctx context.Context, code, redirectURI, accountID string) error {
	if !c.Enabled() {
		return ErrNotConfigured
	}
	if strings.TrimSpace(code) == "" {
		return fmt.Errorf("authorization code is required")
	}
	if strings.TrimSpace(redirectURI) == "" {
		return fmt.Errorf("redirect URI is required")
	}
	form := url.Values{}
	form.Set("code", code)
	form.Set("client_id", c.cfg.ClientID)
	form.Set("client_secret", c.cfg.ClientSecret)
	form.Set("redirect_uri", redirectURI)
	form.Set("grant_type", "authorization_code")
	tokenURL := zohoAuthBaseURL + "/oauth/v2/token"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("create oauth exchange request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("oauth exchange request failed: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("oauth exchange failed: status=%d body=%s", resp.StatusCode, string(body))
	}
	var result OAuthTokenResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("decode oauth exchange response: %w", err)
	}
	if result.Error != "" {
		return fmt.Errorf("oauth exchange error: %s", result.Error)
	}
	if strings.TrimSpace(result.AccessToken) == "" {
		return fmt.Errorf("oauth exchange response missing access token")
	}
	if err := c.loadTokensFromDB(ctx); err != nil && !errors.Is(err, ErrTokenNotFound) {
		return err
	}
	c.accessToken = result.AccessToken
	if strings.TrimSpace(result.RefreshToken) != "" {
		c.refreshToken = strings.TrimSpace(result.RefreshToken)
	}
	c.accountID = resolveAccountID(accountID, c.cfg.AccountID)
	if result.ExpiresIn <= 0 {
		result.ExpiresIn = 3600
	}
	c.expiresAt = time.Now().UTC().Add(time.Duration(result.ExpiresIn) * time.Second)
	return c.saveTokensToDB(ctx)
}

func (c *Client) getAccessToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.loadTokensFromDB(ctx); err != nil {
		return "", err
	}
	if c.accessToken != "" && time.Now().UTC().Before(c.expiresAt.Add(-2*time.Minute)) {
		return c.accessToken, nil
	}
	token, refreshToken, expiresIn, err := c.refreshAccessToken(ctx)
	if err != nil {
		return "", err
	}
	c.accessToken = token
	if refreshToken != "" {
		c.refreshToken = refreshToken
	}
	c.expiresAt = time.Now().UTC().Add(time.Duration(expiresIn) * time.Second)
	if err := c.saveTokensToDB(ctx); err != nil {
		return "", err
	}
	return token, nil
}

func (c *Client) refreshAccessToken(ctx context.Context) (string, string, int, error) {
	if c.refreshToken == "" {
		return "", "", 0, ErrTokenNotFound
	}
	form := url.Values{}
	form.Set("client_id", c.cfg.ClientID)
	form.Set("client_secret", c.cfg.ClientSecret)
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", c.refreshToken)
	authURL := zohoAuthBaseURL + "/oauth/v2/token"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, authURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", "", 0, fmt.Errorf("create zoho auth request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", 0, fmt.Errorf("zoho auth request failed: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", "", 0, fmt.Errorf("zoho auth failed: status=%d body=%s", resp.StatusCode, string(body))
	}
	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		Error        string `json:"error"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", "", 0, fmt.Errorf("decode zoho auth response: %w", err)
	}
	if result.Error != "" {
		return "", "", 0, fmt.Errorf("zoho auth error: %s", result.Error)
	}
	if result.AccessToken == "" {
		return "", "", 0, fmt.Errorf("zoho auth response missing access token")
	}
	if result.ExpiresIn <= 0 {
		result.ExpiresIn = 3600
	}
	return result.AccessToken, result.RefreshToken, result.ExpiresIn, nil
}

func (c *Client) loadTokensFromDB(ctx context.Context) error {
	if c.tokens == nil {
		return ErrNotConfigured
	}
	if c.refreshToken != "" {
		return nil
	}
	var doc tokenDocument
	err := c.tokens.FindOne(ctx, bson.M{"provider": tokenProvider}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrTokenNotFound
		}
		return fmt.Errorf("read zoho token from database: %w", err)
	}
	c.accessToken = doc.AccessToken
	c.refreshToken = doc.RefreshToken
	c.accountID = doc.AccountID
	c.expiresAt = doc.ExpiresAt
	if c.refreshToken == "" {
		return ErrTokenNotFound
	}
	return nil
}

func (c *Client) saveTokensToDB(ctx context.Context) error {
	if c.tokens == nil {
		return ErrNotConfigured
	}
	update := bson.M{
		"$set": bson.M{
			"provider":      tokenProvider,
			"accountId":     c.accountID,
			"accessToken":   c.accessToken,
			"refreshToken":  c.refreshToken,
			"expiresAt":     c.expiresAt,
			"lastRefreshed": time.Now().UTC(),
		},
	}
	_, err := c.tokens.UpdateOne(ctx, bson.M{"provider": tokenProvider}, update, options.Update().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("persist zoho token to database: %w", err)
	}
	return nil
}

func resolveAccountID(callbackAccountID, configuredAccountID string) string {
	if trimmed := strings.TrimSpace(callbackAccountID); trimmed != "" {
		return trimmed
	}
	return strings.TrimSpace(configuredAccountID)
}
