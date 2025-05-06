package handlers

import (
	"net/http"
)

// HealthCheck godoc
// @Summary      Health check
// @Description  Returns 200 if the service is healthy
// @Tags         health
// @Produce      json
// @Success      200  {string}  string  "ok"
// @Router       /v1/healthz [get]
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
