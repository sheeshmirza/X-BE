package handlers

import (
	"net/http"

	"X-BE/internal/models"
)

func (h *Handlers) SaleCreate(w http.ResponseWriter, r *http.Request) {
	if !h.hasCoreServices() {
		writeError(w, http.StatusInternalServerError, errServicesNotReady)
		return
	}
	var payload models.SaleCreateRequest
	if err := decodeJSONBody(r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, errInvalidJSONPayload)
		return
	}
	result, err := h.services.Sale.Create(r.Context(), payload)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeSuccess(w, http.StatusCreated, result)
}
