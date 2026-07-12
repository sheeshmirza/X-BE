package handlers

import (
	"net/http"

	"X-BE/internal/models"
)

func (h *Handlers) NewsletterCreate(w http.ResponseWriter, r *http.Request) {
	if !h.hasCoreServices() {
		writeError(w, http.StatusInternalServerError, errServicesNotReady)
		return
	}
	var payload models.NewsletterSubscriptionRequest
	if err := decodeJSONBody(r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, errInvalidJSONPayload)
		return
	}
	result, err := h.services.Newsletter.Subscribe(r.Context(), payload)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeSuccess(w, http.StatusCreated, result)
}

func (h *Handlers) NewsletterDelete(w http.ResponseWriter, r *http.Request) {
	if !h.hasCoreServices() {
		writeError(w, http.StatusInternalServerError, errServicesNotReady)
		return
	}
	var payload models.NewsletterSubscriptionRequest
	if err := decodeJSONBody(r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, errInvalidJSONPayload)
		return
	}
	result, err := h.services.Newsletter.Unsubscribe(r.Context(), payload.Email)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeSuccess(w, http.StatusOK, result)
}
