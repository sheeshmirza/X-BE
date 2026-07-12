package handlers

import "net/http"

func (h *Handlers) Health(w http.ResponseWriter, _ *http.Request) {
	writeSuccess(w, http.StatusOK, map[string]string{
		"service": "X-BE-NEW",
		"status":  "OK",
	})
}
