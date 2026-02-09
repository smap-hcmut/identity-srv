package jwt

import (
	"crypto/rsa"
	"encoding/base64"
	"math/big"
)

// JWK represents a JSON Web Key
type JWK struct {
	Kty string `json:"kty"` // Key Type (RSA)
	Use string `json:"use"` // Public Key Use (sig for signature)
	Kid string `json:"kid"` // Key ID
	Alg string `json:"alg"` // Algorithm (RS256)
	N   string `json:"n"`   // Modulus
	E   string `json:"e"`   // Exponent
}

// JWKS represents a JSON Web Key Set
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// GetJWKS returns the JWKS representation of the public key
func (m *Manager) GetJWKS() *JWKS {
	jwk := publicKeyToJWK(m.publicKey, m.kid)
	return &JWKS{
		Keys: []JWK{jwk},
	}
}

// publicKeyToJWK converts an RSA public key to JWK format
func publicKeyToJWK(publicKey *rsa.PublicKey, kid string) JWK {
	// Encode modulus (N) as base64url
	nBytes := publicKey.N.Bytes()
	n := base64.RawURLEncoding.EncodeToString(nBytes)

	// Encode exponent (E) as base64url
	eBytes := big.NewInt(int64(publicKey.E)).Bytes()
	e := base64.RawURLEncoding.EncodeToString(eBytes)

	return JWK{
		Kty: "RSA",
		Use: "sig",
		Kid: kid,
		Alg: "RS256",
		N:   n,
		E:   e,
	}
}
