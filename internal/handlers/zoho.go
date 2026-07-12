package handlers

import "net/http"

func (h *Handlers) ZohoAuth(w http.ResponseWriter, r *http.Request) {
	if !h.hasZohoService() {
		writeError(w, http.StatusInternalServerError, errZohoNotReady)
		return
	}
	redirectURI := r.URL.Query().Get("redirect_uri")
	if redirectURI == "" {
		redirectURI = inferredRedirectURI(r)
	}
	scope := r.URL.Query().Get("scope")
	authURL, finalRedirectURI, err := h.services.Zoho.GetAuthorizationURL(redirectURI, scope)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeSuccess(w, http.StatusOK, map[string]any{
		"authorizationUrl": authURL,
		"redirectUri":      finalRedirectURI,
	})
}

func (h *Handlers) ZohoOAuthCallback(w http.ResponseWriter, r *http.Request) {
	if !h.hasZohoService() {
		writeError(w, http.StatusInternalServerError, errZohoNotReady)
		return
	}
	if oauthErr := r.URL.Query().Get("error"); oauthErr != "" {
		writeError(w, http.StatusBadRequest, "Authorization error: "+oauthErr)
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		writeError(w, http.StatusBadRequest, "Authorization code not provided")
		return
	}
	redirectURI := r.URL.Query().Get("redirect_uri")
	if redirectURI == "" {
		redirectURI = inferredRedirectURI(r)
	}
	accountID := r.URL.Query().Get("account_id")
	if err := h.services.Zoho.HandleOAuthCallback(r.Context(), code, redirectURI, accountID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccess(w, http.StatusOK, map[string]any{
		"message": "Zoho tokens saved successfully",
	})
}
