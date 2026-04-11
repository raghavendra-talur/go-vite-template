package tokens

import "time"

type Token struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Token      string     `json:"token,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
	LastUsedAt *time.Time `json:"lastUsedAt"`
}

type CreateTokenRequest struct {
	Name string `json:"name"`
}
