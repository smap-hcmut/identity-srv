package scope

import (
	"encoding/base64"
	"encoding/json"

<<<<<<<< HEAD:identity/pkg/scope/scope.go
	"smap-api/internal/model"
========
	"smap-collector/internal/models"
>>>>>>>> 9c65a15b02994a6cc9940a129c9a3c4f61fd0697:collector/pkg/scope/scope.go
)

// NewScope creates a new scope.
func NewScope(payload Payload) model.Scope {
	return model.Scope{
		UserID:   payload.UserID,
		Username: payload.Username,
		Role:     payload.Role,
	}
}

func CreateScopeHeader(scope model.Scope) (string, error) {
	// Marshal the scope data to JSON
	jsonData, err := json.Marshal(scope)
	if err != nil {
		return "", err
	}

	// Encode the JSON data as Base64
	base64Data := base64.StdEncoding.EncodeToString(jsonData)
	return base64Data, nil
}

func ParseScopeHeader(scopeHeader string) (model.Scope, error) {
	// Decode the Base64 data
	jsonData, err := base64.StdEncoding.DecodeString(scopeHeader)
	if err != nil {
		return model.Scope{}, err
	}

	// Unmarshal the JSON data
	var scope model.Scope
	err = json.Unmarshal(jsonData, &scope)
	if err != nil {
		return model.Scope{}, err
	}

	return scope, nil
}

func (m implManager) VerifyScope(scopeHeader string) (model.Scope, error) {
	scope, err := ParseScopeHeader(scopeHeader)
	if err != nil {
		return model.Scope{}, err
	}

	return scope, nil
}
