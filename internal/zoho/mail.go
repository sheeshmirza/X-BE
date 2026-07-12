package zoho

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (c *Client) SendEmail(ctx context.Context, req EmailRequest) error {
	if !c.Enabled() {
		return ErrNotConfigured
	}
	if len(req.To) == 0 {
		return ErrMailNotConfigured
	}
	accessToken, err := c.getAccessToken(ctx)
	if err != nil {
		return err
	}
	accountID := strings.TrimSpace(req.AccountID)
	if accountID == "" {
		accountID = strings.TrimSpace(c.accountID)
	}
	if accountID == "" {
		return ErrMailNotConfigured
	}
	payload, err := json.Marshal(buildMailPayload(req))
	if err != nil {
		return fmt.Errorf("marshal zoho mail payload: %w", err)
	}
	var lastErr error
	for _, endpoint := range buildMailEndpoints(accountID) {
		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
		if err != nil {
			return fmt.Errorf("create zoho mail request: %w", err)
		}
		httpReq.Header.Set("Authorization", "Zoho-oauthtoken "+accessToken)
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Accept", "application/json")
		resp, err := c.httpClient.Do(httpReq)
		if err != nil {
			lastErr = fmt.Errorf("zoho mail request failed: %w", err)
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}
		lastErr = fmt.Errorf("zoho mail failed: status=%d body=%s", resp.StatusCode, string(body))
	}
	if lastErr != nil {
		return lastErr
	}
	return ErrMailNotConfigured
}

func buildMailEndpoints(accountID string) []string {
	trimmed := strings.TrimSpace(accountID)
	if trimmed == "" {
		return nil
	}
	endpoints := make([]string, 0, len(zohoMailEndpoints))
	for _, fmtString := range zohoMailEndpoints {
		endpoints = append(endpoints, fmt.Sprintf(fmtString, trimmed))
	}
	return endpoints
}

func buildMailPayload(req EmailRequest) map[string]string {
	mailFormat := "html"
	if strings.EqualFold(req.ContentType, "text") {
		mailFormat = "plaintext"
	}
	return map[string]string{
		"fromAddress": strings.TrimSpace(req.FromAddress),
		"toAddress":   strings.Join(req.To, ","),
		"subject":     req.Subject,
		"content":     req.Content,
		"mailFormat":  mailFormat,
	}
}
