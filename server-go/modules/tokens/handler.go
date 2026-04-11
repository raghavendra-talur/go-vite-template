package tokens

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router) {
	r.Post("/tokens", createToken)
	r.Get("/tokens", listTokens)
	r.Delete("/tokens/{id}", deleteToken)
}

func createToken(w http.ResponseWriter, r *http.Request) {
	var req CreateTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		http.Error(w, `{"error":"Name is required"}`, http.StatusBadRequest)
		return
	}

	token, err := Create(r.Context(), req.Name)
	if err != nil {
		http.Error(w, `{"error":"Failed to create token"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(token)
}

func listTokens(w http.ResponseWriter, r *http.Request) {
	tokens, err := GetAll(r.Context())
	if err != nil {
		http.Error(w, `{"error":"Failed to list tokens"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokens)
}

func deleteToken(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	deleted, err := Delete(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"Failed to delete token"}`, http.StatusInternalServerError)
		return
	}
	if !deleted {
		http.Error(w, `{"error":"Token not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true}`))
}
