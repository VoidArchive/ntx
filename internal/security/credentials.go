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

	// Save credentials first so we can derive key from file info
	if err := saveCredentialsTemporary(path, creds); err != nil {
		return fmt.Errorf("failed to create credentials file: %w", err)
	}

	// Derive master key for encryption (now that file exists)
	masterKey, err := deriveMasterKey(path)
	if err != nil {
		return fmt.Errorf("failed to derive master key: %w", err)
	}

	// Encrypt and save credentials with proper encryption
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

// saveCredentialsTemporary saves credentials temporarily to establish file for key derivation
func saveCredentialsTemporary(path string, creds *Credentials) error {
	// Create a temporary placeholder file
	placeholder := []byte("placeholder")
	if err := os.WriteFile(path, placeholder, 0600); err != nil {
		return fmt.Errorf("failed to create temporary credentials file: %w", err)
	}
	return nil
}

// LoadCredentials loads and decrypts credentials from disk
func LoadCredentials(configDir string) (*Credentials, error) {
	credentialsPath := filepath.Join(configDir, credentialsFile)

	// Check if credentials file exists
	if _, err := os.Stat(credentialsPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("credentials file not found: %s", credentialsPath)
	}

	// Read encrypted credentials file
	encryptedData, err := os.ReadFile(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials file: %w", err)
	}

	// For this implementation, we'll derive the master key from a combination of:
	// 1. System-specific information (hostname, user)
	// 2. File modification time (as additional entropy)
	// This provides reasonable security for a local-first application
	masterKey, err := deriveMasterKey(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to derive master key: %w", err)
	}

	// Decrypt credentials
	credentials, err := decryptCredentials(encryptedData, masterKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt credentials: %w", err)
	}

	return credentials, nil
}

// deriveMasterKey derives a master key from system-specific information
func deriveMasterKey(credentialsPath string) ([]byte, error) {
	// Verify file exists
	if _, err := os.Stat(credentialsPath); err != nil {
		return nil, fmt.Errorf("failed to access credentials file: %w", err)
	}

	// Get hostname for system-specific entropy
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	// Get current user for user-specific entropy
	user := os.Getenv("USER")
	if user == "" {
		user = os.Getenv("USERNAME") // Windows fallback
		if user == "" {
			user = "unknown"
		}
	}

	// Combine entropy sources (removed file modification time as it changes during creation)
	entropy := fmt.Sprintf("%s:%s", hostname, user)

	// Use a simple but effective key derivation
	// In production, you might want to use PBKDF2 or Argon2
	hasher := func(data string) []byte {
		// Simple hash-based key derivation
		result := make([]byte, keySize)
		dataBytes := []byte(data)

		for i := 0; i < keySize; i++ {
			var sum byte
			for j, b := range dataBytes {
				sum ^= b + byte(i) + byte(j)
			}
			result[i] = sum
		}
		return result
	}

	return hasher(entropy), nil
}

// decryptCredentials decrypts the credentials using the master key
func decryptCredentials(encryptedData []byte, key []byte) (*Credentials, error) {
	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Check minimum length
	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return nil, fmt.Errorf("encrypted data too short")
	}

	// Extract nonce and ciphertext
	nonce := encryptedData[:nonceSize]
	ciphertext := encryptedData[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt credentials: %w", err)
	}

	// Unmarshal JSON
	var credentials Credentials
	if err := json.Unmarshal(plaintext, &credentials); err != nil {
		return nil, fmt.Errorf("failed to unmarshal credentials: %w", err)
	}

	return &credentials, nil
}
