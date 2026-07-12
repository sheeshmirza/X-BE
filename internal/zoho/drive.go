package zoho

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
)

func (c *Client) UploadFile(ctx context.Context, req DriveUploadRequest) (string, error) {
	if !c.Enabled() {
		return "", ErrNotConfigured
	}
	if strings.TrimSpace(req.ParentFolderID) == "" || len(req.Content) == 0 {
		return "", ErrDriveNotConfigured
	}
	accessToken, err := c.getAccessToken(ctx)
	if err != nil {
		return "", err
	}
	filename := strings.TrimSpace(req.Filename)
	if filename == "" {
		filename = "upload.bin"
	}
	mimeType := strings.TrimSpace(req.MimeType)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("parent_id", req.ParentFolderID)
	_ = writer.WriteField("folder_id", req.ParentFolderID)
	_ = writer.WriteField("override_name_exist", "false")
	part, err := writer.CreateFormFile("file", filepath.Base(filename))
	if err != nil {
		return "", fmt.Errorf("create drive multipart file: %w", err)
	}
	if _, err := part.Write(req.Content); err != nil {
		return "", fmt.Errorf("write drive multipart file: %w", err)
	}
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("close drive multipart writer: %w", err)
	}
	bodyBytes := body.Bytes()
	var lastErr error
	for _, endpoint := range buildDriveUploadEndpoints() {
		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
		if err != nil {
			return "", fmt.Errorf("create drive upload request: %w", err)
		}
		httpReq.Header.Set("Authorization", "Zoho-oauthtoken "+accessToken)
		httpReq.Header.Set("Content-Type", writer.FormDataContentType())
		httpReq.Header.Set("Accept", "application/json")
		httpReq.Header.Set("X-Content-Type", mimeType)
		resp, err := c.httpClient.Do(httpReq)
		if err != nil {
			lastErr = fmt.Errorf("zoho drive request failed: %w", err)
			continue
		}
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			var parsed any
			if err := json.Unmarshal(respBody, &parsed); err != nil {
				return "", fmt.Errorf("decode zoho drive response: %w", err)
			}
			fileID := extractFirstID(parsed)
			if fileID == "" {
				return "", fmt.Errorf("zoho drive upload missing file id: %s", string(respBody))
			}
			return fileID, nil
		}
		lastErr = fmt.Errorf("zoho drive upload failed: status=%d body=%s", resp.StatusCode, string(respBody))
	}
	if lastErr != nil {
		return "", lastErr
	}
	return "", ErrDriveNotConfigured
}

func buildDriveUploadEndpoints() []string {
	return zohoDriveUploadEndpoints
}

func extractFirstID(value any) string {
	switch typed := value.(type) {
	case map[string]any:
		if idRaw, ok := typed["id"]; ok {
			if id, ok := idRaw.(string); ok && id != "" {
				return id
			}
		}
		for _, child := range typed {
			if found := extractFirstID(child); found != "" {
				return found
			}
		}
	case []any:
		for _, child := range typed {
			if found := extractFirstID(child); found != "" {
				return found
			}
		}
	}
	return ""
}
