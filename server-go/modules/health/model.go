package health

// HealthResponse is the shape returned by GET /health and GET /api/v1/health.
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version,omitempty"`
}
