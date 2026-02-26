package app

import (
	"encoding/json"
	"net/http"
)

type orgResponse struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	ZendeskSubdomain string `json:"zendesk_subdomain,omitempty"`
}

// @Summary     Get current organization
// @Tags        Organization
// @Description Returns the organization associated with the provided API key
// @Produce     json
// @Success     200  {object}  orgResponse
// @Failure     401  {string}  string  "Unauthorized"
// @Security    ApiKeyAuth
// @Router      /org [get]
func (a *App) getOrg(w http.ResponseWriter, r *http.Request) {
	o := orgFromContext(r.Context())
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orgResponse{ID: o.ID, Name: o.Name, ZendeskSubdomain: o.ZendeskSubdomain})
}
