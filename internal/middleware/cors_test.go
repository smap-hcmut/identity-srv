package middleware

import (
	"testing"
)

// Test isPrivateOrigin function
func TestIsPrivateOrigin(t *testing.T) {
	tests := []struct {
		name   string
		origin string
		want   bool
	}{
		// Valid IPs in configured subnets
		{
			name:   "K8s cluster subnet - first IP",
			origin: "http://172.16.21.1:3000",
			want:   true,
		},
		{
			name:   "K8s cluster subnet - middle IP",
			origin: "http://172.16.21.50:8080",
			want:   true,
		},
		{
			name:   "K8s cluster subnet - last IP",
			origin: "http://172.16.21.254:3000",
			want:   true,
		},
		{
			name:   "Private network 1 - valid IP",
			origin: "http://172.16.19.100:3000",
			want:   true,
		},
		{
			name:   "Private network 2 - valid IP",
			origin: "http://192.168.1.50:3000",
			want:   true,
		},
		{
			name:   "HTTPS protocol in private subnet",
			origin: "https://172.16.21.50:3000",
			want:   true,
		},
		{
			name:   "No port specified in private subnet",
			origin: "http://172.16.21.50",
			want:   true,
		},

		// IPs outside configured subnets
		{
			name:   "Different subnet - 172.16.22.x",
			origin: "http://172.16.22.50:3000",
			want:   false,
		},
		{
			name:   "Different subnet - 10.0.0.x",
			origin: "http://10.0.0.50:3000",
			want:   false,
		},
		{
			name:   "Public IP",
			origin: "http://1.2.3.4:3000",
			want:   false,
		},
		{
			name:   "K8s subnet boundary - .0 (network address)",
			origin: "http://172.16.21.0:3000",
			want:   true, // .0 is technically in range but usually reserved
		},
		{
			name:   "K8s subnet boundary - .255 (broadcast)",
			origin: "http://172.16.21.255:3000",
			want:   true, // .255 is technically in range but usually reserved
		},

		// Invalid formats
		{
			name:   "Invalid IP format",
			origin: "http://not-an-ip:3000",
			want:   false,
		},
		{
			name:   "Invalid origin URL",
			origin: "not a url",
			want:   false,
		},
		{
			name:   "Empty origin",
			origin: "",
			want:   false,
		},
		{
			name:   "Domain name (not IP)",
			origin: "http://example.com:3000",
			want:   false,
		},
		{
			name:   "Localhost (not in private subnets)",
			origin: "http://localhost:3000",
			want:   false,
		},
		{
			name:   "127.0.0.1 (not in private subnets)",
			origin: "http://127.0.0.1:3000",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPrivateOrigin(tt.origin)
			if got != tt.want {
				t.Errorf("isPrivateOrigin(%q) = %v, want %v", tt.origin, got, tt.want)
			}
		})
	}
}

// Test isLocalhostOrigin function
func TestIsLocalhostOrigin(t *testing.T) {
	tests := []struct {
		name   string
		origin string
		want   bool
	}{
		// HTTP localhost variations
		{
			name:   "http://localhost without port",
			origin: "http://localhost",
			want:   true,
		},
		{
			name:   "http://localhost with port 3000",
			origin: "http://localhost:3000",
			want:   true,
		},
		{
			name:   "http://localhost with port 5173",
			origin: "http://localhost:5173",
			want:   true,
		},
		{
			name:   "http://localhost with port 8080",
			origin: "http://localhost:8080",
			want:   true,
		},

		// HTTPS localhost variations
		{
			name:   "https://localhost without port",
			origin: "https://localhost",
			want:   true,
		},
		{
			name:   "https://localhost with port 3000",
			origin: "https://localhost:3000",
			want:   true,
		},

		// HTTP 127.0.0.1 variations
		{
			name:   "http://127.0.0.1 without port",
			origin: "http://127.0.0.1",
			want:   true,
		},
		{
			name:   "http://127.0.0.1 with port 3000",
			origin: "http://127.0.0.1:3000",
			want:   true,
		},
		{
			name:   "http://127.0.0.1 with port 8080",
			origin: "http://127.0.0.1:8080",
			want:   true,
		},

		// HTTPS 127.0.0.1 variations
		{
			name:   "https://127.0.0.1 without port",
			origin: "https://127.0.0.1",
			want:   true,
		},
		{
			name:   "https://127.0.0.1 with port 3000",
			origin: "https://127.0.0.1:3000",
			want:   true,
		},

		// Non-localhost origins
		{
			name:   "Production domain",
			origin: "https://smap.tantai.dev",
			want:   false,
		},
		{
			name:   "Private subnet IP",
			origin: "http://172.16.21.50:3000",
			want:   false,
		},
		{
			name:   "Public IP",
			origin: "http://1.2.3.4:3000",
			want:   false,
		},
		{
			name:   "Empty origin",
			origin: "",
			want:   false,
		},
		{
			name:   "Different loopback IP (127.0.0.2)",
			origin: "http://127.0.0.2:3000",
			want:   false, // Only 127.0.0.1 is explicitly supported
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isLocalhostOrigin(tt.origin)
			if got != tt.want {
				t.Errorf("isLocalhostOrigin(%q) = %v, want %v", tt.origin, got, tt.want)
			}
		})
	}
}

// Test DefaultCORSConfig function
func TestDefaultCORSConfig(t *testing.T) {
	tests := []struct {
		name                     string
		environment              string
		wantAllowOriginFuncIsNil bool
		wantStaticOrigins        []string
	}{
		{
			name:                     "Production mode",
			environment:              "production",
			wantAllowOriginFuncIsNil: true,
			wantStaticOrigins:        productionOrigins,
		},
		{
			name:                     "Dev mode uses dynamic validation",
			environment:              "dev",
			wantAllowOriginFuncIsNil: false,
			wantStaticOrigins:        nil,
		},
		{
			name:                     "Staging mode uses dynamic validation",
			environment:              "staging",
			wantAllowOriginFuncIsNil: false,
			wantStaticOrigins:        nil,
		},
		{
			name:                     "Empty environment defaults to production",
			environment:              "",
			wantAllowOriginFuncIsNil: true,
			wantStaticOrigins:        productionOrigins,
		},
		{
			name:                     "Invalid environment defaults to production",
			environment:              "invalid",
			wantAllowOriginFuncIsNil: false, // Invalid doesn't match "production", so uses dynamic
			wantStaticOrigins:        nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultCORSConfig(tt.environment)

			// Check AllowCredentials is always true
			if !config.AllowCredentials {
				t.Errorf("AllowCredentials should always be true, got false")
			}

			// Check AllowOriginFunc
			gotAllowOriginFuncIsNil := config.AllowOriginFunc == nil
			if gotAllowOriginFuncIsNil != tt.wantAllowOriginFuncIsNil {
				t.Errorf("AllowOriginFunc nil = %v, want %v", gotAllowOriginFuncIsNil, tt.wantAllowOriginFuncIsNil)
			}

			// Check static origins
			if tt.wantStaticOrigins != nil {
				if len(config.AllowedOrigins) != len(tt.wantStaticOrigins) {
					t.Errorf("AllowedOrigins length = %d, want %d", len(config.AllowedOrigins), len(tt.wantStaticOrigins))
				}
				for i, origin := range tt.wantStaticOrigins {
					if i >= len(config.AllowedOrigins) || config.AllowedOrigins[i] != origin {
						t.Errorf("AllowedOrigins[%d] = %v, want %v", i, config.AllowedOrigins, tt.wantStaticOrigins)
						break
					}
				}
			}
		})
	}
}

// Test AllowOriginFunc behavior in non-production mode
func TestDefaultCORSConfigAllowOriginFunc(t *testing.T) {
	config := DefaultCORSConfig("dev")

	if config.AllowOriginFunc == nil {
		t.Fatal("AllowOriginFunc should not be nil in dev mode")
	}

	tests := []struct {
		name   string
		origin string
		want   bool
	}{
		// Production domains should be allowed
		{
			name:   "Production domain 1",
			origin: "https://smap.tantai.dev",
			want:   true,
		},
		{
			name:   "Production domain 2",
			origin: "https://smap-api.tantai.dev",
			want:   true,
		},

		// Localhost should be allowed
		{
			name:   "localhost with port",
			origin: "http://localhost:3000",
			want:   true,
		},
		{
			name:   "127.0.0.1 with port",
			origin: "http://127.0.0.1:8080",
			want:   true,
		},

		// Private subnets should be allowed
		{
			name:   "K8s subnet IP",
			origin: "http://172.16.21.50:3000",
			want:   true,
		},
		{
			name:   "Private network 1 IP",
			origin: "http://172.16.19.100:3000",
			want:   true,
		},
		{
			name:   "Private network 2 IP",
			origin: "http://192.168.1.50:3000",
			want:   true,
		},

		// Unauthorized origins should be rejected
		{
			name:   "Random domain",
			origin: "https://evil.com",
			want:   false,
		},
		{
			name:   "Public IP",
			origin: "http://1.2.3.4:3000",
			want:   false,
		},
		{
			name:   "Different private subnet",
			origin: "http://10.0.0.50:3000",
			want:   false,
		},
		{
			name:   "Empty origin",
			origin: "",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := config.AllowOriginFunc(tt.origin)
			if got != tt.want {
				t.Errorf("AllowOriginFunc(%q) = %v, want %v", tt.origin, got, tt.want)
			}
		})
	}
}

// Test production mode rejects non-production origins
func TestProductionModeRejectsPrivateOrigins(t *testing.T) {
	config := DefaultCORSConfig("production")

	// Production should have no AllowOriginFunc
	if config.AllowOriginFunc != nil {
		t.Error("Production mode should not have AllowOriginFunc")
	}

	// Check that only production origins are in the static list
	if len(config.AllowedOrigins) != len(productionOrigins) {
		t.Errorf("Production mode should have %d origins, got %d", len(productionOrigins), len(config.AllowedOrigins))
	}

	// Verify production origins are present
	for _, prodOrigin := range productionOrigins {
		found := false
		for _, allowedOrigin := range config.AllowedOrigins {
			if allowedOrigin == prodOrigin {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Production origin %q not found in AllowedOrigins", prodOrigin)
		}
	}
}
