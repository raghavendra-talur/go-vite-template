package middleware

import (
	"net/http"
	"strings"

	"github.com/raghavendra-talur/go-vite-template/server-go/modules/tokens"
)

// adminToken is set at startup via SetAdminToken.
var adminToken string

// SetAdminToken configures the admin token loaded from adminAuth.json.
func SetAdminToken(token string) {
	adminToken = token
}

// BearerAuth requires a valid token for all requests.
// Accepts the admin token (from adminAuth.json) or any API token from the database.
func BearerAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// WebSocket connections can't set custom headers, so accept token as query param
		if tokenParam := r.URL.Query().Get("token"); tokenParam != "" && r.Header.Get("Authorization") == "" {
			r.Header.Set("Authorization", "Bearer "+tokenParam)
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"Authorization required"}`, http.StatusUnauthorized)
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, `{"error":"Invalid authorization header"}`, http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")

		// Check admin token first
		if adminToken != "" && token == adminToken {
			next.ServeHTTP(w, r)
			return
		}

		// Check API tokens from database
		if !tokens.ValidateToken(r.Context(), token) {
			http.Error(w, `{"error":"Invalid or revoked token"}`, http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
