package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Credentials represents the encrypted credentials structure
type Credentials struct {
	Version     string            `json:"version"`
	CreatedAt   time.Time         `json:"created_at"`
	DatabaseKey []byte            `json:"database_key"`
	ConfigSalt  []byte            `json:"config_salt"`
	APITokens   map[string]string `json:"api_tokens"`
}

const (
	credentialsFile = "credentials"
	keySize         = 32 // AES-256
	saltSize        = 16
)

// InitializeCredentials initializes the credentials file if it doesn't exist
func InitializeCredentials(configDir string) error {
	credentialsPath := filepath.Join(configDir, credentialsFile)

	// Check if credentials file already exists
	if _, err := os.Stat(credentialsPath); err == nil {
		// File exists, verify it's readable
		return verifyCredentials(credentialsPath)
	}

	// Create new credentials
	return createCredentials(credentialsPath)
}

// createCredentials generates and saves new credentials
func createCredentials(path string) error {
	// Generate database encryption key
	dbKey := make([]byte, keySize)
	if _, err := rand.Read(dbKey); err != nil {
		return fmt.Errorf("failed to generate database key: %w", err)
	}

	// Generate config salt
	salt := make([]byte, saltSize)
	if _, err := rand.Read(salt); err != nil {
		return fmt.Errorf("failed to generate config salt: %w", err)
	}

	// Create credentials structure
	creds := &Credentials{
		Version:     "1.0",
		CreatedAt:   time.Now(),
		DatabaseKey: dbKey,
		ConfigSalt:  salt,
		APITokens:   make(map[string]string),
	}

	// Generate master key for encryption
	masterKey := make([]byte, keySize)
	if _, err := rand.Read(masterKey); err != nil {
		return fmt.Errorf("failed to generate master key: %w", err)
	}

	// Encrypt and save credentials
	if err := saveEncryptedCredentials(path, creds, masterKey); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	// Set proper file permissions (user only)
	if err := os.Chmod(path, 0600); err != nil {
		return fmt.Errorf("failed to set credentials file permissions: %w", err)
	}

	return nil
}

// verifyCredentials checks if the credentials file is valid
func verifyCredentials(path string) error {
	// For now, just check if the file exists and is readable
	// TODO: Implement proper credential verification
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("credentials file not accessible: %w", err)
	}

	// Check file permissions
	if info.Mode().Perm() != 0600 {
		// Fix permissions if they're wrong
		if err := os.Chmod(path, 0600); err != nil {
			return fmt.Errorf("failed to fix credentials file permissions: %w", err)
		}
	}

	return nil
}

// saveEncryptedCredentials encrypts and saves credentials to disk
func saveEncryptedCredentials(path string, creds *Credentials, key []byte) error {
	// Marshal credentials to JSON
	data, err := json.Marshal(creds)
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt data
	encrypted := gcm.Seal(nonce, nonce, data, nil)

	// Write to file
	if err := os.WriteFile(path, encrypted, 0600); err != nil {
		return fmt.Errorf("failed to write credentials file: %w", err)
	}

	return nil
}

// LoadCredentials loads and decrypts credentials from disk
// Note: This is a placeholder implementation
// In a real implementation, we'd need a way to derive or store the master key
func LoadCredentials(configDir string) (*Credentials, error) {
	// TODO: Implement proper credential loading and decryption
	// For now, return a mock credentials object
	return &Credentials{
		Version:     "1.0",
		CreatedAt:   time.Now(),
		DatabaseKey: make([]byte, keySize),
		ConfigSalt:  make([]byte, saltSize),
		APITokens:   make(map[string]string),
	}, nil
}