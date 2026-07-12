package zoho

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"X-BE/internal/config"
)

func (c *Client) AppendRow(ctx context.Context, sheet config.ZohoSheetConfig, row map[string]string) error {
	if !c.Enabled() {
		return ErrNotConfigured
	}
	if sheet.ID == "" || sheet.Name == "" {
		return ErrSheetNotConfigured
	}
	accessToken, err := c.getAccessToken(ctx)
	if err != nil {
		return err
	}
	apiURL := buildAppendRowURL(sheet.ID)
	payload := buildAppendRowForm(sheet.Name, false, row)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, strings.NewReader(payload.Encode()))
	if err != nil {
		return fmt.Errorf("create zoho request: %w", err)
	}
	req.Header.Set("Authorization", "Zoho-oauthtoken "+accessToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("zoho append request failed: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("zoho append failed: status=%d url=%s body=%s", resp.StatusCode, apiURL, string(body))
	}
	var result struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("decode zoho append response: %w", err)
	}
	if result.Status != "success" {
		return fmt.Errorf("zoho append unexpected response: %s", string(body))
	}
	return nil
}

func buildAppendRowURL(sheetID string) string {
	return fmt.Sprintf("%s/api/v2/%s", zohoSheetsBaseURL, sheetID)
}

func buildAppendRowForm(sheetName string, insertAtTop bool, row map[string]string) url.Values {
	payload, err := json.Marshal([]map[string]string{row})
	if err != nil {
		payload = []byte("[]")
	}
	form := url.Values{}
	form.Set("method", "worksheet.records.add")
	form.Set("worksheet_name", sheetName)
	form.Set("insert_at_top", fmt.Sprintf("%t", insertAtTop))
	form.Set("json_data", string(payload))
	return form
}
