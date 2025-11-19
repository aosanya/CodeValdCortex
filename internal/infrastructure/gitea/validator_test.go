package giteawebhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidator_ValidateSignature(t *testing.T) {
	secret := "test-secret-key"
	validator := NewValidator(secret)

	t.Run("valid signature", func(t *testing.T) {
		payload := []byte(`{"action":"opened","issue":{"id":1}}`)
		
		// Generate valid signature
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(payload)
		expectedHash := hex.EncodeToString(mac.Sum(nil))
		signature := "sha256=" + expectedHash

		err := validator.ValidateSignature(payload, signature)
		assert.NoError(t, err, "Valid signature should pass validation")
	})

	t.Run("invalid signature", func(t *testing.T) {
		payload := []byte(`{"action":"opened","issue":{"id":1}}`)
		signature := "sha256=invalid_hash_here"

		err := validator.ValidateSignature(payload, signature)
		assert.Error(t, err, "Invalid signature should fail validation")
		assert.Contains(t, err.Error(), "signature validation failed")
	})

	t.Run("missing signature", func(t *testing.T) {
		payload := []byte(`{"action":"opened","issue":{"id":1}}`)
		signature := ""

		err := validator.ValidateSignature(payload, signature)
		assert.Error(t, err, "Missing signature should fail validation")
		assert.Contains(t, err.Error(), "missing X-Gitea-Signature")
	})

	t.Run("wrong signature format", func(t *testing.T) {
		payload := []byte(`{"action":"opened","issue":{"id":1}}`)
		signature := "md5=somehash" // Wrong format

		err := validator.ValidateSignature(payload, signature)
		assert.Error(t, err, "Wrong format should fail validation")
		assert.Contains(t, err.Error(), "invalid signature format")
	})

	t.Run("tampered payload", func(t *testing.T) {
		originalPayload := []byte(`{"action":"opened","issue":{"id":1}}`)
		tamperedPayload := []byte(`{"action":"opened","issue":{"id":999}}`)
		
		// Generate signature for original payload
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(originalPayload)
		expectedHash := hex.EncodeToString(mac.Sum(nil))
		signature := "sha256=" + expectedHash

		// Validate with tampered payload
		err := validator.ValidateSignature(tamperedPayload, signature)
		assert.Error(t, err, "Tampered payload should fail validation")
	})
}

func TestValidator_GetSignatureHeader(t *testing.T) {
	validator := NewValidator("test-secret")
	
	header := validator.GetSignatureHeader()
	assert.Equal(t, "X-Gitea-Signature", header, "Header name should be X-Gitea-Signature")
}
