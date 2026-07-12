package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"X-BE/internal/models"
)

const (
	errInvalidJSONPayload = "Invalid JSON payload"
	errServicesNotReady   = "Services are not initialized"
	errZohoNotReady       = "Zoho service is not initialized"
)

func (h *Handlers) hasCoreServices() bool {
	return h != nil && h.services != nil && h.services.Career != nil && h.services.Newsletter != nil && h.services.Sale != nil
}

func (h *Handlers) hasZohoService() bool {
	return h != nil && h.services != nil && h.services.Zoho != nil
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, models.APIResponse{Success: false, Error: message})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(payload)
}

func writeSuccess(w http.ResponseWriter, status int, data any) {
	writeJSON(w, status, models.APIResponse{Success: true, Data: data})
}

func decodeJSONBody(r *http.Request, dest any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dest); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return io.ErrUnexpectedEOF
	}
	return nil
}

func inferredRedirectURI(r *http.Request) string {
	scheme := r.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		if r.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}
	return fmt.Sprintf("%s://%s/api/zoho/oauth/callback", scheme, r.Host)
}
