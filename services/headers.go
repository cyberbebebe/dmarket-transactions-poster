package services

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Generate creates authorized headers for a Dmarket API request
func generateHeaders(secretKey, method, apiURLPath string, body interface{}) (http.Header, error) {
	// Generate nonce from current timestamp
	nonce := fmt.Sprintf("%d", time.Now().Unix())

	// Decode the secret key (128 hex characters = 64 bytes)
	privateKeyBytes, err := hex.DecodeString(secretKey)
	if err != nil || len(privateKeyBytes) != 64 {
		return nil, fmt.Errorf("invalid secret key: must be 128 hex characters")
	}
	var privateKey [64]byte
	copy(privateKey[:], privateKeyBytes)

	// Extract public key from the last 32 bytes of the private key
	publicKey := hex.EncodeToString(privateKey[32:])

	// Handle body
	var bodyStr string
	if body != nil {
		str, ok := body.(string)
		if ok {
			bodyStr = str
		} else {
			// Fallback for other types, just in case.
			b, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal body: %v", err)
			}
			bodyStr = string(b)
		}
	} else {
		bodyStr = ""
	}


	// Build string to sign
	stringToSign := method + apiURLPath + bodyStr + nonce

	// Sign the string using Ed25519
	signature := hex.EncodeToString(ed25519.Sign(privateKey[:], []byte(stringToSign)))

	// Build headers
	headers := http.Header{}
	headers.Set("X-Api-Key", publicKey)
    headers.Set("X-Request-Sign", "dmar ed25519 "+signature)
    headers.Set("X-Sign-Date", nonce)

	return headers, nil
}