package main

import (
	"encoding/json"
	"net/http"
)

// @Summary     Health check
// @Description Returns health status of the API and its dependencies
// @Produce     json
// @Success     200  {object}  map[string]any
// @Failure     503  {object}  map[string]any
// @Router      /health [get]
func (a *app) health(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	dbOK := a.db.PingContext(ctx) == nil
	redisOK := a.redis.Ping(ctx).Err() == nil

	status := "ok"
	code := http.StatusOK
	if !dbOK || !redisOK {
		status = "degraded"
		code = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]any{
		"status":   status,
		"postgres": dbOK,
		"redis":    redisOK,
	})
}
