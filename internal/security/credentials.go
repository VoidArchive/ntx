package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/pbkdf2"
)

// Credentials represents the encrypted credentials structure
type Credentials struct {
	Version     string            `json:"version"`
	CreatedAt   time.Time         `json:"created_at"`
	DatabaseKey []byte            `json:"database_key"`
	ConfigSalt  []byte            `json:"config_salt"`
	APITokens   map[string]string `json:"api_tokens"`
}

// EncryptedContainer wraps the encrypted credentials with metadata needed for decryption
type EncryptedContainer struct {
	Version       string    `json:"version"`
	CreatedAt     time.Time `json:"created_at"`
	Hostname      string    `json:"hostname"` // Store hostname used for key derivation
	Username      string    `json:"username"` // Store username used for key derivation
	EncryptedData []byte    `json:"encrypted_data"`
	Salt          []byte    `json:"salt"` // Store salt directly in container
}

const (
	credentialsFile  = "credentials"
	keySize          = 32     // AES-256
	saltSize         = 32     // 256-bit salt for PBKDF2
	pbkdf2Iterations = 100000 // OWASP recommended minimum iterations
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
func saveCredentialsTemporary(path string, _ *Credentials) error {
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

	// Read credentials file
	fileData, err := os.ReadFile(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials file: %w", err)
	}

	// Check for corrupted placeholder data
	if len(fileData) >= 11 && string(fileData[:11]) == "placeholder" {
		return nil, fmt.Errorf("credentials file contains placeholder data - file was not properly encrypted")
	}

	// Try to parse as new container format first
	var container EncryptedContainer
	if err := json.Unmarshal(fileData, &container); err == nil {
		return loadFromContainer(&container)
	}

	// Fall back to legacy format
	return loadLegacyCredentials(credentialsPath, fileData)
}

// deriveMasterKey derives a master key using PBKDF2 with system-specific information
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

	// Create a deterministic but unique password from system info
	password := fmt.Sprintf("ntx-credentials:%s:%s", hostname, user)

	// Generate or load salt from a consistent source
	salt, err := getOrCreateSalt(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get salt: %w", err)
	}

	// Use PBKDF2 for proper key derivation (OWASP recommended)
	derivedKey := pbkdf2.Key([]byte(password), salt, pbkdf2Iterations, keySize, sha256.New)

	return derivedKey, nil
}

// getOrCreateSalt generates or retrieves a consistent salt for key derivation
func getOrCreateSalt(credentialsPath string) ([]byte, error) {
	saltPath := credentialsPath + ".salt"

	// Try to read existing salt first
	if saltData, err := os.ReadFile(saltPath); err == nil && len(saltData) == saltSize {
		return saltData, nil
	}

	// Generate new salt if none exists or is invalid
	salt := make([]byte, saltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	// Save salt to file with secure permissions
	if err := os.WriteFile(saltPath, salt, 0600); err != nil {
		return nil, fmt.Errorf("failed to save salt: %w", err)
	}

	return salt, nil
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

// loadFromContainer loads credentials from the new container format
func loadFromContainer(container *EncryptedContainer) (*Credentials, error) {
	// Use the stored system info for key derivation
	password := fmt.Sprintf("ntx-credentials:%s:%s", container.Hostname, container.Username)

	// Use PBKDF2 for key derivation with stored salt
	derivedKey := pbkdf2.Key([]byte(password), container.Salt, pbkdf2Iterations, keySize, sha256.New)

	// Decrypt the credentials
	return decryptCredentials(container.EncryptedData, derivedKey)
}

// loadLegacyCredentials loads credentials using the legacy format
func loadLegacyCredentials(credentialsPath string, encryptedData []byte) (*Credentials, error) {
	// Use the original derivation method for backward compatibility
	masterKey, err := deriveMasterKey(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to derive master key: %w", err)
	}

	// Decrypt credentials
	credentials, err := decryptCredentials(encryptedData, masterKey)
	if err != nil {
		// Check if this looks like a key mismatch (most common cause)
		if err.Error() == "failed to decrypt credentials: cipher: message authentication failed" {
			fmt.Printf("RECOVERY: Credentials decryption failed due to system info change.\n")
			fmt.Printf("RECOVERY: This usually happens when hostname or username changed.\n")
			fmt.Printf("RECOVERY: Recreating credentials with current system info...\n")

			// Backup the old file
			backupPath := credentialsPath + ".backup." + fmt.Sprintf("%d", time.Now().Unix())
			if backupErr := os.Rename(credentialsPath, backupPath); backupErr != nil {
				fmt.Printf("WARNING: Could not backup old credentials: %v\n", backupErr)
			} else {
				fmt.Printf("RECOVERY: Backed up old credentials to: %s\n", backupPath)
			}

			// Backup salt file too
			saltPath := credentialsPath + ".salt"
			saltBackupPath := saltPath + ".backup." + fmt.Sprintf("%d", time.Now().Unix())
			if backupErr := os.Rename(saltPath, saltBackupPath); backupErr != nil {
				fmt.Printf("WARNING: Could not backup old salt: %v\n", backupErr)
			} else {
				fmt.Printf("RECOVERY: Backed up old salt to: %s\n", saltBackupPath)
			}

			// Create new credentials
			if createErr := createCredentials(credentialsPath); createErr != nil {
				return nil, fmt.Errorf("failed to recreate credentials: %w", createErr)
			}

			fmt.Printf("RECOVERY: Successfully recreated credentials.\n")
			fmt.Printf("RECOVERY: Application will now use the new credentials.\n")

			// Load the newly created credentials
			return LoadCredentials(filepath.Dir(credentialsPath))
		}

		return nil, fmt.Errorf("failed to decrypt credentials: %w", err)
	}

	return credentials, nil
}
