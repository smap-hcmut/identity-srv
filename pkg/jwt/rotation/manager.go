package rotation

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"smap-api/internal/model"
	"time"
)

type KeyRepository interface {
	SaveKey(ctx context.Context, key *model.JWTKey) error
	GetActiveKey(ctx context.Context) (*model.JWTKey, error)
	GetActiveAndRotatingKeys(ctx context.Context) ([]*model.JWTKey, error)
	UpdateKeyStatus(ctx context.Context, kid, status string) error
	GetRotatingKeys(ctx context.Context) ([]*model.JWTKey, error)
}

type Manager struct {
	repo         KeyRepository
	rotationDays int
	graceDays    int
}

func NewManager(repo KeyRepository, rotationDays, graceDays int) *Manager {
	return &Manager{
		repo:         repo,
		rotationDays: rotationDays,
		graceDays:    graceDays,
	}
}

// RotateKeys performs the key rotation process
func (m *Manager) RotateKeys(ctx context.Context) error {
	// Step 1: Check if rotation is needed
	activeKey, err := m.repo.GetActiveKey(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active key: %w", err)
	}

	if activeKey != nil {
		daysSinceCreation := time.Since(activeKey.CreatedAt).Hours() / 24
		if daysSinceCreation < float64(m.rotationDays) {
			// Not time to rotate yet
			return nil
		}
	}

	// Step 2: Generate new key
	newKey, err := m.generateNewKey(ctx)
	if err != nil {
		return fmt.Errorf("failed to generate new key: %w", err)
	}

	// Step 3: Mark old active key as rotating
	if activeKey != nil {
		if err := m.repo.UpdateKeyStatus(ctx, activeKey.KID, model.KeyStatusRotating); err != nil {
			return fmt.Errorf("failed to mark old key as rotating: %w", err)
		}

		// Set expiration for the rotating key
		expiresAt := time.Now().Add(time.Duration(m.graceDays) * 24 * time.Hour)
		activeKey.SetExpiration(expiresAt)
	}

	// Step 4: Save new active key
	if err := m.repo.SaveKey(ctx, newKey); err != nil {
		return fmt.Errorf("failed to save new key: %w", err)
	}

	// Step 5: Retire expired rotating keys
	if err := m.retireExpiredKeys(ctx); err != nil {
		return fmt.Errorf("failed to retire expired keys: %w", err)
	}

	return nil
}

// generateNewKey creates a new RSA key pair
func (m *Manager) generateNewKey(ctx context.Context) (*model.JWTKey, error) {
	// Generate RSA key pair (2048 bits)
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	// Encode private key to PEM
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	// Encode public key to PEM
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	// Generate unique KID
	kid := fmt.Sprintf("key-%d", time.Now().Unix())

	return model.NewJWTKey(kid, string(privateKeyPEM), string(publicKeyPEM)), nil
}

// retireExpiredKeys marks rotating keys as retired if grace period expired
func (m *Manager) retireExpiredKeys(ctx context.Context) error {
	rotatingKeys, err := m.repo.GetRotatingKeys(ctx)
	if err != nil {
		return err
	}

	now := time.Now()
	for _, key := range rotatingKeys {
		if key.ExpiresAt != nil && now.After(*key.ExpiresAt) {
			if err := m.repo.UpdateKeyStatus(ctx, key.KID, model.KeyStatusRetired); err != nil {
				return err
			}
		}
	}

	return nil
}

// EnsureActiveKey ensures at least one active key exists
func (m *Manager) EnsureActiveKey(ctx context.Context) error {
	activeKey, err := m.repo.GetActiveKey(ctx)
	if err != nil {
		return err
	}

	if activeKey == nil {
		// No active key exists, generate one
		newKey, err := m.generateNewKey(ctx)
		if err != nil {
			return err
		}
		return m.repo.SaveKey(ctx, newKey)
	}

	return nil
}
