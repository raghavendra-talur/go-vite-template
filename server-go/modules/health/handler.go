package health

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// AppVersion is set by main at startup.
var AppVersion string

func RegisterRoutes(r chi.Router) {
	r.Get("/health", getHealth)
}

func getHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(HealthResponse{
		Status:  "ok",
		Version: AppVersion,
	})
}
