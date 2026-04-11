package config

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

// AdminAuth holds the admin token generated on first startup.
type AdminAuth struct {
	Token string `json:"token"`
}

// EnsureAdminAuth loads or creates adminAuth.json in dataDir.
// Returns the admin token.
func EnsureAdminAuth(dataDir string) string {
	path := filepath.Join(dataDir, "adminAuth.json")

	data, err := os.ReadFile(path)
	if err == nil {
		var auth AdminAuth
		if json.Unmarshal(data, &auth) == nil && auth.Token != "" {
			log.Printf("Admin auth loaded from %s", path)
			return auth.Token
		}
	}

	// Generate new token
	b := make([]byte, 32)
	rand.Read(b)
	token := "rb_admin_" + hex.EncodeToString(b)

	auth := AdminAuth{Token: token}
	jsonData, _ := json.MarshalIndent(auth, "", "  ")

	if err := os.MkdirAll(dataDir, 0700); err != nil {
		log.Fatalf("Failed to create data dir %s: %v", dataDir, err)
	}

	if err := os.WriteFile(path, jsonData, 0600); err != nil {
		log.Fatalf("Failed to write %s: %v", path, err)
	}

	log.Printf("Admin auth created at %s", path)
	return token
}
