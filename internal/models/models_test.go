package models

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthRequest_HashPass(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "normal password",
			password: "secret123",
		},
		{
			name:     "empty password",
			password: "",
		},
		{
			name:     "long password",
			password: "this is a very long password for testing purposes",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := &AuthRequest{
				Username: "testuser",
				Password: tc.password,
			}

			hash := sha256.Sum256([]byte(tc.password))
			expected := hex.EncodeToString(hash[:])

			req.HashPass()
			assert.Equal(t, expected, req.Password)
		})
	}
}
