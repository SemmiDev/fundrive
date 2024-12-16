package fundrive

import (
	"encoding/base64"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTokenEncryption(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid 32-byte key",
			key:     "12345678901234567890123456789012",
			wantErr: false,
		},
		{
			name:        "key too short",
			key:         "short-key",
			wantErr:     true,
			errContains: "must be exactly 32 bytes",
		},
		{
			name:        "key too long",
			key:         "this-key-is-definitely-longer-than-32-bytes-long",
			wantErr:     true,
			errContains: "must be exactly 32 bytes",
		},
		{
			name:        "empty key",
			key:         "",
			wantErr:     true,
			errContains: "must be exactly 32 bytes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encryption, err := NewTokenEncryption(tt.key)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, encryption)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, encryption)
				assert.Equal(t, []byte(tt.key), encryption.key)
			}
		})
	}
}

func TestTokenEncryption_EncryptDecrypt(t *testing.T) {
	// Setup
	key := "12345678901234567890123456789012"
	encryption, err := NewTokenEncryption(key)
	require.NoError(t, err)
	require.NotNil(t, encryption)

	tests := []struct {
		name        string
		plaintext   string
		wantErr     bool
		errContains string
	}{
		{
			name:      "normal text",
			plaintext: "hello world",
			wantErr:   false,
		},
		{
			name:      "empty string",
			plaintext: "",
			wantErr:   false,
		},
		{
			name:      "long text",
			plaintext: strings.Repeat("long text ", 100),
			wantErr:   false,
		},
		{
			name:      "special characters",
			plaintext: "!@#$%^&*()_+-=[]{}|;:,.<>?",
			wantErr:   false,
		},
		{
			name:      "unicode characters",
			plaintext: "Hello, 世界! 🌎",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Encryption
			encrypted, err := encryption.Encrypt(tt.plaintext)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, encrypted)
			assert.NotEqual(t, tt.plaintext, encrypted)

			// Test Decryption
			decrypted, err := encryption.Decrypt(encrypted)
			require.NoError(t, err)
			assert.Equal(t, tt.plaintext, decrypted)
		})
	}
}

func TestTokenEncryption_DecryptInvalid(t *testing.T) {
	// Setup
	key := "12345678901234567890123456789012"
	encryption, err := NewTokenEncryption(key)
	require.NoError(t, err)
	require.NotNil(t, encryption)

	tests := []struct {
		name        string
		encrypted   string
		errContains string
	}{
		{
			name:        "invalid base64",
			encrypted:   "not-base64!@#$",
			errContains: "failed to decode base64",
		},
		{
			name:        "empty string",
			encrypted:   "",
			errContains: "ciphertext too short",
		},
		{
			name:        "too short after base64 decode",
			encrypted:   base64.StdEncoding.EncodeToString([]byte("short")),
			errContains: "ciphertext too short",
		},
		{
			name:        "invalid nonce",
			encrypted:   base64.StdEncoding.EncodeToString(make([]byte, NonceSize+5)),
			errContains: "failed to decrypt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decrypted, err := encryption.Decrypt(tt.encrypted)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errContains)
			assert.Empty(t, decrypted)
		})
	}
}

func TestTokenEncryption_EncryptionUniqueness(t *testing.T) {
	// Setup
	key := "12345678901234567890123456789012"
	encryption, err := NewTokenEncryption(key)
	require.NoError(t, err)
	require.NotNil(t, encryption)

	// Test that multiple encryptions of the same plaintext produce different ciphertexts
	plaintext := "test message"
	encryptedTexts := make(map[string]bool)

	// Encrypt the same text multiple times
	for i := 0; i < 10; i++ {
		encrypted, err := encryption.Encrypt(plaintext)
		require.NoError(t, err)

		// Verify this ciphertext hasn't been seen before
		assert.False(t, encryptedTexts[encrypted], "Encryption should produce unique ciphertexts")
		encryptedTexts[encrypted] = true

		// Verify decryption still works
		decrypted, err := encryption.Decrypt(encrypted)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	}
}

func TestTokenEncryption_CrossInstanceDecryption(t *testing.T) {
	// Test that different instances with the same key can decrypt each other's messages
	key := "12345678901234567890123456789012"
	encryption1, err := NewTokenEncryption(key)
	require.NoError(t, err)

	encryption2, err := NewTokenEncryption(key)
	require.NoError(t, err)

	plaintext := "test message"

	// Encrypt with first instance
	encrypted, err := encryption1.Encrypt(plaintext)
	require.NoError(t, err)

	// Decrypt with second instance
	decrypted, err := encryption2.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)

	// And vice versa
	encrypted, err = encryption2.Encrypt(plaintext)
	require.NoError(t, err)

	decrypted, err = encryption1.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestTokenEncryptionDecryption(t *testing.T) {
	// Setup
	key := "12345678901234567890123456789012"
	encryption, err := NewTokenEncryption(key)
	require.NoError(t, err)
	require.NotNil(t, encryption)

	tests := []struct {
		name            string
		plaintext       string
		modifyEncrypted func(string) string
		wantDecrypted   string
		wantErr         bool
		errContains     string
	}{
		{
			name:            "successful decryption",
			plaintext:       "test message",
			modifyEncrypted: nil,
			wantDecrypted:   "test message",
			wantErr:         false,
		},
		{
			name:            "empty plaintext",
			plaintext:       "",
			modifyEncrypted: nil,
			wantDecrypted:   "",
			wantErr:         false,
		},
		{
			name:            "long plaintext",
			plaintext:       strings.Repeat("very long message ", 100),
			modifyEncrypted: nil,
			wantDecrypted:   strings.Repeat("very long message ", 100),
			wantErr:         false,
		},
		{
			name:            "special characters",
			plaintext:       "!@#$%^&*()_+-=[]{};:'\"\\|,.<>?/~`",
			modifyEncrypted: nil,
			wantDecrypted:   "!@#$%^&*()_+-=[]{};:'\"\\|,.<>?/~`",
			wantErr:         false,
		},
		{
			name:            "multi-byte characters",
			plaintext:       "Hello 世界 🌍 привет",
			modifyEncrypted: nil,
			wantDecrypted:   "Hello 世界 🌍 привет",
			wantErr:         false,
		},
		{
			name:      "tampered encrypted data",
			plaintext: "test message",
			modifyEncrypted: func(s string) string {
				decoded, _ := base64.StdEncoding.DecodeString(s)
				if len(decoded) > NonceSize+1 {
					decoded[NonceSize+1] ^= 0x01 // Flip a bit in the ciphertext
				}
				return base64.StdEncoding.EncodeToString(decoded)
			},
			wantErr:     true,
			errContains: "failed to decrypt",
		},
		{
			name:      "modified nonce",
			plaintext: "test message",
			modifyEncrypted: func(s string) string {
				decoded, _ := base64.StdEncoding.DecodeString(s)
				if len(decoded) > 0 {
					decoded[0] ^= 0x01 // Flip a bit in the nonce
				}
				return base64.StdEncoding.EncodeToString(decoded)
			},
			wantErr:     true,
			errContains: "failed to decrypt",
		},
		{
			name:      "truncated ciphertext",
			plaintext: "test message",
			modifyEncrypted: func(s string) string {
				decoded, _ := base64.StdEncoding.DecodeString(s)
				return base64.StdEncoding.EncodeToString(decoded[:len(decoded)-1])
			},
			wantErr:     true,
			errContains: "failed to decrypt",
		},
		{
			name:      "invalid base64",
			plaintext: "test message",
			modifyEncrypted: func(s string) string {
				return "invalid base64 %%%"
			},
			wantErr:     true,
			errContains: "failed to decode base64",
		},
		{
			name:      "too short ciphertext",
			plaintext: "test message",
			modifyEncrypted: func(s string) string {
				return base64.StdEncoding.EncodeToString([]byte("short"))
			},
			wantErr:     true,
			errContains: "ciphertext too short",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// First encrypt the plaintext
			encrypted, err := encryption.Encrypt(tt.plaintext)
			require.NoError(t, err, "Encryption should not fail")

			// Apply any modifications to the encrypted text if specified
			if tt.modifyEncrypted != nil {
				encrypted = tt.modifyEncrypted(encrypted)
			}

			// Try to decrypt
			decrypted, err := encryption.Decrypt(encrypted)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantDecrypted, decrypted)

			// Additional verification for successful cases
			if !tt.wantErr {
				// Verify that a second encryption/decryption cycle also works
				secondEncrypted, err := encryption.Encrypt(decrypted)
				require.NoError(t, err)
				assert.NotEqual(t, encrypted, secondEncrypted, "Second encryption should produce different ciphertext")

				secondDecrypted, err := encryption.Decrypt(secondEncrypted)
				require.NoError(t, err)
				assert.Equal(t, tt.wantDecrypted, secondDecrypted)
			}
		})
	}

	// Test with different encryption instances
	t.Run("cross-instance decryption", func(t *testing.T) {
		encryption1, err := NewTokenEncryption(key)
		require.NoError(t, err)

		encryption2, err := NewTokenEncryption(key)
		require.NoError(t, err)

		plaintext := "cross instance test"
		encrypted, err := encryption1.Encrypt(plaintext)
		require.NoError(t, err)

		decrypted, err := encryption2.Decrypt(encrypted)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	// Test concurrent decryption
	t.Run("concurrent decryption", func(t *testing.T) {
		const goroutines = 10
		var wg sync.WaitGroup
		plaintext := "concurrent test"
		encrypted, err := encryption.Encrypt(plaintext)
		require.NoError(t, err)

		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				decrypted, err := encryption.Decrypt(encrypted)
				assert.NoError(t, err)
				assert.Equal(t, plaintext, decrypted)
			}()
		}
		wg.Wait()
	})
}
