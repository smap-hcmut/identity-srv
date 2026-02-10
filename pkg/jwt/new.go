package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
)

// Config holds JWT manager configuration
type Config struct {
	PrivateKeyPath string
	PublicKeyPath  string
	Issuer         string
	Audience       []string
	TTL            time.Duration
}

// New creates a new JWT manager with a single key (legacy mode)
func New(cfg Config) (*Manager, error) {
	// Load private key
	privateKey, err := loadPrivateKey(cfg.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	// Load public key
	publicKey, err := loadPublicKey(cfg.PublicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load public key: %w", err)
	}

	// Generate Key ID
	kid := uuid.New().String()

	// Create manager with single key
	manager := &Manager{
		keys:      make(map[string]*keyPair),
		activeKID: kid,
		issuer:    cfg.Issuer,
		audience:  cfg.Audience,
		ttl:       cfg.TTL,
	}

	manager.keys[kid] = &keyPair{
		privateKey: privateKey,
		publicKey:  publicKey,
	}

	return manager, nil
}

// GenerateKeyPair generates a new RSA key pair (2048-bit)
func GenerateKeyPair() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	return privateKey, &privateKey.PublicKey, nil
}

// SavePrivateKey saves private key to PEM file
func SavePrivateKey(privateKey *rsa.PrivateKey, filepath string) error {
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create private key file: %w", err)
	}
	defer file.Close()

	if err := pem.Encode(file, privateKeyPEM); err != nil {
		return fmt.Errorf("failed to encode private key: %w", err)
	}

	// Set restrictive permissions (owner read/write only)
	if err := os.Chmod(filepath, 0600); err != nil {
		return fmt.Errorf("failed to set private key permissions: %w", err)
	}

	return nil
}

// SavePublicKey saves public key to PEM file
func SavePublicKey(publicKey *rsa.PublicKey, filepath string) error {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return fmt.Errorf("failed to marshal public key: %w", err)
	}

	publicKeyPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}

	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create public key file: %w", err)
	}
	defer file.Close()

	if err := pem.Encode(file, publicKeyPEM); err != nil {
		return fmt.Errorf("failed to encode public key: %w", err)
	}

	return nil
}

// loadPrivateKey loads RSA private key from PEM file
func loadPrivateKey(filepath string) (*rsa.PrivateKey, error) {
	keyData, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// Try PKCS8 format
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		var ok bool
		privateKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("not an RSA private key")
		}
	}

	return privateKey, nil
}

// loadPublicKey loads RSA public key from PEM file
func loadPublicKey(filepath string) (*rsa.PublicKey, error) {
	keyData, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	return rsaPublicKey, nil
}
