package jwt

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Manager handles JWT token generation and verification
type Manager struct {
	keys      map[string]*keyPair // kid -> key pair
	activeKID string
	issuer    string
	audience  []string
	ttl       time.Duration
}

type keyPair struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

// Claims represents JWT claims structure
type Claims struct {
	Email  string   `json:"email"`
	Role   string   `json:"role"`
	Groups []string `json:"groups,omitempty"`
	jwt.RegisteredClaims
}

// GenerateToken generates a new JWT token with RS256 algorithm
func (m *Manager) GenerateToken(userID, email, role string, groups []string) (string, error) {
	activeKey, ok := m.keys[m.activeKID]
	if !ok {
		return "", fmt.Errorf("active key not found")
	}

	now := time.Now()
	expiresAt := now.Add(m.ttl)

	// Generate unique JTI (JWT ID) for token tracking and revocation
	jti := uuid.New().String()

	claims := Claims{
		Email:  email,
		Role:   role,
		Groups: groups,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   userID,
			Audience:  m.audience,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        jti,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// Set Key ID in header
	token.Header["kid"] = m.activeKID

	// Sign token with active private key
	tokenString, err := token.SignedString(activeKey.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// VerifyToken verifies and parses a JWT token
func (m *Manager) VerifyToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get KID from header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("missing kid in token header")
		}

		// Find corresponding public key
		keyPair, ok := m.keys[kid]
		if !ok {
			return nil, fmt.Errorf("unknown kid: %s", kid)
		}

		return keyPair.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid claims type")
	}

	return claims, nil
}

// GetPublicKey returns the active public key for external verification
func (m *Manager) GetPublicKey() *rsa.PublicKey {
	if activeKey, ok := m.keys[m.activeKID]; ok {
		return activeKey.publicKey
	}
	return nil
}

// GetKID returns the active Key ID
func (m *Manager) GetKID() string {
	return m.activeKID
}

// GetAllPublicKeys returns all public keys (for JWKS endpoint)
func (m *Manager) GetAllPublicKeys() map[string]*rsa.PublicKey {
	publicKeys := make(map[string]*rsa.PublicKey)
	for kid, keyPair := range m.keys {
		publicKeys[kid] = keyPair.publicKey
	}
	return publicKeys
}

// LoadKeys loads multiple JWT keys for rotation support
func (m *Manager) LoadKeys(keys []*JWTKeyData) error {
	m.keys = make(map[string]*keyPair)

	for _, key := range keys {
		// Parse private key
		privateKey, err := parsePrivateKeyFromPEM(key.PrivateKey)
		if err != nil {
			return fmt.Errorf("failed to parse private key for kid %s: %w", key.KID, err)
		}

		// Parse public key
		publicKey, err := parsePublicKeyFromPEM(key.PublicKey)
		if err != nil {
			return fmt.Errorf("failed to parse public key for kid %s: %w", key.KID, err)
		}

		m.keys[key.KID] = &keyPair{
			privateKey: privateKey,
			publicKey:  publicKey,
		}

		// Set active KID
		if key.IsActive {
			m.activeKID = key.KID
		}
	}

	if m.activeKID == "" {
		return fmt.Errorf("no active key found")
	}

	return nil
}

// JWTKeyData represents key data for loading
type JWTKeyData struct {
	KID        string
	PrivateKey string
	PublicKey  string
	IsActive   bool
}

// parsePrivateKeyFromPEM parses RSA private key from PEM string
func parsePrivateKeyFromPEM(pemStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
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

// parsePublicKeyFromPEM parses RSA public key from PEM string
func parsePublicKeyFromPEM(pemStr string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
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

// SetConfig sets the issuer, audience, and TTL for the manager
func (m *Manager) SetConfig(issuer string, audience []string, ttl time.Duration) {
	m.issuer = issuer
	m.audience = audience
	m.ttl = ttl
}
