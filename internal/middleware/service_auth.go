package middleware

import (
	"strings"

	"smap-api/pkg/response"

	"github.com/gin-gonic/gin"
)

// ServiceAuth validates X-Service-Key header for internal service-to-service authentication
func (m Middleware) ServiceAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get X-Service-Key header
		serviceKey := c.GetHeader("X-Service-Key")
		if serviceKey == "" {
			response.Unauthorized(c)
			c.Abort()
			return
		}

		// Decrypt service key using encrypter
		decryptedKey, err := m.encrypter.Decrypt(serviceKey)
		if err != nil {
			m.l.Errorf(c.Request.Context(), "ServiceAuth: Decrypt failed: %v", err)
			response.Unauthorized(c)
			c.Abort()
			return
		}

		// Validate decrypted key against configured service keys
		// Format: "service_name:key_value"
		parts := strings.SplitN(decryptedKey, ":", 2)
		if len(parts) != 2 {
			m.l.Errorf(c.Request.Context(), "ServiceAuth: Invalid format: %s", decryptedKey)
			response.Unauthorized(c)
			c.Abort()
			return
		}

		serviceName := parts[0]
		keyValue := parts[1]

		// Check if service exists in config
		configuredKey, exists := m.config.InternalConfig.ServiceKeys[serviceName]
		if !exists {
			m.l.Errorf(c.Request.Context(), "ServiceAuth: Service not found: %s", serviceName)
			response.Unauthorized(c)
			c.Abort()
			return
		}

		// Validate key value
		if keyValue != configuredKey {
			m.l.Errorf(c.Request.Context(), "ServiceAuth: Key mismatch for %s. Got: %s, Want: %s", serviceName, keyValue, configuredKey)
			response.Unauthorized(c)
			c.Abort()
			return
		}

		// Store service name in context for logging/audit
		c.Set("service_name", serviceName)

		c.Next()
	}
}
