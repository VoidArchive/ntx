package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInitializeCredentials tests credential initialization
func TestInitializeCredentials(t *testing.T) {
	t.Run("New credentials file", func(t *testing.T) {
		// Create temporary directory for test
		tempDir := t.TempDir()

		err := InitializeCredentials(tempDir)
		require.NoError(t, err)

		// Verify credentials file was created
		credentialsPath := filepath.Join(tempDir, credentialsFile)
		assert.FileExists(t, credentialsPath)

		// Verify file permissions
		info, err := os.Stat(credentialsPath)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
	})

	t.Run("Existing credentials file", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create credentials first time
		err := InitializeCredentials(tempDir)
		require.NoError(t, err)

		// Initialize again - should not error
		err = InitializeCredentials(tempDir)
		assert.NoError(t, err)
	})

	t.Run("Invalid directory permissions", func(t *testing.T) {
		// Try to create in a directory we can't write to
		invalidDir := "/proc/invalid" // This should fail on most systems

		err := InitializeCredentials(invalidDir)
		// Should handle error gracefully
		assert.Error(t, err)
	})
}

// TestLoadCredentials tests credential loading
func TestLoadCredentials(t *testing.T) {
	t.Run("Load valid credentials", func(t *testing.T) {
		tempDir := t.TempDir()

		// Initialize credentials
		err := InitializeCredentials(tempDir)
		require.NoError(t, err)

		// Load credentials
		creds, err := LoadCredentials(tempDir)
		require.NoError(t, err)
		require.NotNil(t, creds)

		// Verify credential structure
		assert.Equal(t, "1.0", creds.Version)
		assert.NotEmpty(t, creds.DatabaseKey)
		assert.NotEmpty(t, creds.ConfigSalt)
		assert.NotNil(t, creds.APITokens)
		assert.True(t, creds.CreatedAt.Before(time.Now().Add(time.Second)))

		// Verify key sizes
		assert.Equal(t, keySize, len(creds.DatabaseKey))
		assert.Equal(t, saltSize, len(creds.ConfigSalt))
	})

	t.Run("Missing credentials file", func(t *testing.T) {
		tempDir := t.TempDir()

		// Try to load without initializing
		_, err := LoadCredentials(tempDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "credentials file not found")
	})

	t.Run("Corrupted credentials file", func(t *testing.T) {
		tempDir := t.TempDir()
		credentialsPath := filepath.Join(tempDir, credentialsFile)

		// Create invalid credentials file
		err := os.WriteFile(credentialsPath, []byte("corrupted data"), 0600)
		require.NoError(t, err)

		// Try to load corrupted file
		_, err = LoadCredentials(tempDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decrypt credentials")
	})
}

// TestCredentialsEncryption tests encryption/decryption functionality
func TestCredentialsEncryption(t *testing.T) {
	t.Run("Encryption round trip", func(t *testing.T) {
		tempDir := t.TempDir()

		// Initialize and load credentials
		err := InitializeCredentials(tempDir)
		require.NoError(t, err)

		creds1, err := LoadCredentials(tempDir)
		require.NoError(t, err)

		// Load again and compare
		creds2, err := LoadCredentials(tempDir)
		require.NoError(t, err)

		assert.Equal(t, creds1.Version, creds2.Version)
		assert.Equal(t, creds1.DatabaseKey, creds2.DatabaseKey)
		assert.Equal(t, creds1.ConfigSalt, creds2.ConfigSalt)
		assert.Equal(t, creds1.CreatedAt.Unix(), creds2.CreatedAt.Unix())
	})
}

// TestDeriveMasterKey tests master key derivation
func TestDeriveMasterKey(t *testing.T) {
	t.Run("Consistent key derivation", func(t *testing.T) {
		tempDir := t.TempDir()
		credentialsPath := filepath.Join(tempDir, credentialsFile)

		// Create a test file
		err := os.WriteFile(credentialsPath, []byte("test"), 0600)
		require.NoError(t, err)

		// Derive key multiple times
		key1, err := deriveMasterKey(credentialsPath)
		require.NoError(t, err)

		key2, err := deriveMasterKey(credentialsPath)
		require.NoError(t, err)

		// Keys should be identical for same file
		assert.Equal(t, key1, key2)
		assert.Equal(t, keySize, len(key1))
	})

	t.Run("Different keys for different modification times", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create first file
		file1 := filepath.Join(tempDir, "file1")
		err := os.WriteFile(file1, []byte("test1"), 0600)
		require.NoError(t, err)

		key1, err := deriveMasterKey(file1)
		require.NoError(t, err)

		// Wait a second to ensure different modification time
		time.Sleep(1 * time.Second)

		// Update the same file (different mod time)
		err = os.WriteFile(file1, []byte("test1-updated"), 0600)
		require.NoError(t, err)

		key2, err := deriveMasterKey(file1)
		require.NoError(t, err)

		// Keys should be different due to different modification times
		assert.NotEqual(t, key1, key2)
	})

	t.Run("Missing file", func(t *testing.T) {
		nonExistentPath := "/tmp/nonexistent-file-12345"

		_, err := deriveMasterKey(nonExistentPath)
		assert.Error(t, err)
	})
}

// TestCredentialsAPITokens tests API token management
func TestCredentialsAPITokens(t *testing.T) {
	t.Run("Empty API tokens initially", func(t *testing.T) {
		tempDir := t.TempDir()

		err := InitializeCredentials(tempDir)
		require.NoError(t, err)

		creds, err := LoadCredentials(tempDir)
		require.NoError(t, err)

		assert.NotNil(t, creds.APITokens)
		assert.Empty(t, creds.APITokens)
	})
}

// TestEncryptionPrimitives tests the underlying encryption functions
func TestEncryptionPrimitives(t *testing.T) {
	t.Run("AES encryption works correctly", func(t *testing.T) {
		// Generate test key
		key := make([]byte, keySize)
		_, err := rand.Read(key)
		require.NoError(t, err)

		// Test data
		originalData := &Credentials{
			Version:     "test",
			CreatedAt:   time.Now(),
			DatabaseKey: make([]byte, keySize),
			ConfigSalt:  make([]byte, saltSize),
			APITokens:   map[string]string{"test": "token"},
		}

		// Marshal to JSON
		jsonData, err := json.Marshal(originalData)
		require.NoError(t, err)

		// Create cipher
		block, err := aes.NewCipher(key)
		require.NoError(t, err)

		gcm, err := cipher.NewGCM(block)
		require.NoError(t, err)

		// Encrypt
		nonce := make([]byte, gcm.NonceSize())
		_, err = io.ReadFull(rand.Reader, nonce)
		require.NoError(t, err)

		encrypted := gcm.Seal(nonce, nonce, jsonData, nil)

		// Decrypt
		if len(encrypted) < gcm.NonceSize() {
			t.Fatal("encrypted data too short")
		}

		recoveredNonce := encrypted[:gcm.NonceSize()]
		ciphertext := encrypted[gcm.NonceSize():]

		decrypted, err := gcm.Open(nil, recoveredNonce, ciphertext, nil)
		require.NoError(t, err)

		// Unmarshal
		var recoveredData Credentials
		err = json.Unmarshal(decrypted, &recoveredData)
		require.NoError(t, err)

		// Verify data integrity
		assert.Equal(t, originalData.Version, recoveredData.Version)
		assert.Equal(t, originalData.DatabaseKey, recoveredData.DatabaseKey)
		assert.Equal(t, originalData.ConfigSalt, recoveredData.ConfigSalt)
		assert.Equal(t, originalData.APITokens, recoveredData.APITokens)
	})
}

// TestCredentialsFilePermissions tests file permission handling
func TestCredentialsFilePermissions(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping file permission tests in CI environment")
	}

	t.Run("Correct file permissions", func(t *testing.T) {
		tempDir := t.TempDir()

		err := InitializeCredentials(tempDir)
		require.NoError(t, err)

		credentialsPath := filepath.Join(tempDir, credentialsFile)
		info, err := os.Stat(credentialsPath)
		require.NoError(t, err)

		// File should be readable/writable by owner only
		assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
	})

	t.Run("Fix incorrect permissions", func(t *testing.T) {
		tempDir := t.TempDir()
		credentialsPath := filepath.Join(tempDir, credentialsFile)

		// Create file with wrong permissions
		err := os.WriteFile(credentialsPath, []byte("test"), 0644)
		require.NoError(t, err)

		// Initialize should fix permissions
		err = InitializeCredentials(tempDir)
		require.NoError(t, err)

		info, err := os.Stat(credentialsPath)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
	})
}

// TestConcurrentCredentialAccess tests thread safety
func TestConcurrentCredentialAccess(t *testing.T) {
	t.Run("Concurrent initialization", func(t *testing.T) {
		tempDir := t.TempDir()

		// Run multiple initializations concurrently
		done := make(chan error, 5)
		for i := 0; i < 5; i++ {
			go func() {
				done <- InitializeCredentials(tempDir)
			}()
		}

		// All should succeed without panics
		for i := 0; i < 5; i++ {
			err := <-done
			assert.NoError(t, err)
		}

		// Verify credentials file exists and is valid
		creds, err := LoadCredentials(tempDir)
		require.NoError(t, err)
		assert.NotNil(t, creds)
	})

	t.Run("Concurrent loading", func(t *testing.T) {
		tempDir := t.TempDir()

		// Initialize first
		err := InitializeCredentials(tempDir)
		require.NoError(t, err)

		// Load concurrently
		done := make(chan *Credentials, 5)
		errors := make(chan error, 5)

		for i := 0; i < 5; i++ {
			go func() {
				creds, err := LoadCredentials(tempDir)
				if err != nil {
					errors <- err
					return
				}
				done <- creds
			}()
		}

		// All should succeed
		var allCreds []*Credentials
		for i := 0; i < 5; i++ {
			select {
			case creds := <-done:
				allCreds = append(allCreds, creds)
			case err := <-errors:
				t.Fatalf("Concurrent load failed: %v", err)
			}
		}

		// All credentials should be identical
		for i := 1; i < len(allCreds); i++ {
			assert.Equal(t, allCreds[0].DatabaseKey, allCreds[i].DatabaseKey)
			assert.Equal(t, allCreds[0].ConfigSalt, allCreds[i].ConfigSalt)
		}
	})
}

// BenchmarkCredentialOperations benchmarks credential operations
func BenchmarkCredentialOperations(b *testing.B) {
	tempDir := b.TempDir()

	// Initialize once
	err := InitializeCredentials(tempDir)
	require.NoError(b, err)

	b.Run("LoadCredentials", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := LoadCredentials(tempDir)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("DeriveMasterKey", func(b *testing.B) {
		credentialsPath := filepath.Join(tempDir, credentialsFile)
		for i := 0; i < b.N; i++ {
			_, err := deriveMasterKey(credentialsPath)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
