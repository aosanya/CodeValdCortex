package giteawebhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// Validator handles HMAC SHA-256 signature validation for Gitea webhooks
type Validator struct {
	secret string
}

// NewValidator creates a new validator with the configured webhook secret
func NewValidator(secret string) *Validator {
	return &Validator{
		secret: secret,
	}
}

// ValidateSignature verifies the X-Gitea-Signature header against the payload
// Gitea uses HMAC SHA-256 format: sha256=<hex-encoded-hash>
func (v *Validator) ValidateSignature(payload []byte, signature string) error {
	if signature == "" {
		return fmt.Errorf("missing X-Gitea-Signature header")
	}

	// Gitea signature format: "sha256=<hash>"
	if len(signature) < 7 || signature[:7] != "sha256=" {
		return fmt.Errorf("invalid signature format: expected 'sha256=<hash>'")
	}

	providedHash := signature[7:] // Remove "sha256=" prefix

	// Compute expected HMAC SHA-256
	mac := hmac.New(sha256.New, []byte(v.secret))
	mac.Write(payload)
	expectedHash := hex.EncodeToString(mac.Sum(nil))

	// Constant-time comparison to prevent timing attacks
	if !hmac.Equal([]byte(expectedHash), []byte(providedHash)) {
		return fmt.Errorf("signature validation failed: hash mismatch")
	}

	return nil
}

// GetSignatureHeader returns the header name used for Gitea webhook signatures
func (v *Validator) GetSignatureHeader() string {
	return "X-Gitea-Signature"
}
