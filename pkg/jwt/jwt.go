package jwt

import (
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Manager handles JWT token generation and verification
type Manager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	issuer     string
	audience   []string
	ttl        time.Duration
	kid        string // Key ID
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
	token.Header["kid"] = m.kid

	// Sign token with private key
	tokenString, err := token.SignedString(m.privateKey)
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
		return m.publicKey, nil
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

// GetPublicKey returns the public key for external verification
func (m *Manager) GetPublicKey() *rsa.PublicKey {
	return m.publicKey
}

// GetKID returns the Key ID
func (m *Manager) GetKID() string {
	return m.kid
}
