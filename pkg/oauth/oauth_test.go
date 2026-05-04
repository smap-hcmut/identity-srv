package oauth

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestNewProvider(t *testing.T) {
	type mock struct{}

	tcs := map[string]struct {
		input  Config
		mock   mock
		output string
		err    error
	}{
		"google": {
			input:  Config{ProviderType: "google"},
			output: "google",
		},
		"azure": {
			input:  Config{ProviderType: "azure"},
			output: "azure",
		},
		"okta": {
			input:  Config{ProviderType: "okta", OktaDomain: "example.okta.com"},
			output: "okta",
		},
		"okta_missing_domain": {
			input: Config{ProviderType: "okta"},
			err:   errors.New("okta_domain is required for Okta provider"),
		},
		"unsupported_provider": {
			input: Config{ProviderType: "github"},
			err:   errors.New("unsupported provider type: github (supported: google, azure, okta)"),
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			output, err := NewProvider(tc.input)

			if tc.err != nil {
				require.Error(t, err)
				assert.EqualError(t, err, tc.err.Error())
				assert.Nil(t, output)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.output, output.GetProviderName())
		})
	}
}

func TestNewGoogleProvider(t *testing.T) {
	type mock struct{}

	tcs := map[string]struct {
		input  Config
		mock   mock
		output string
		err    error
	}{
		"success": {
			input: Config{
				ClientID:     "client-id",
				ClientSecret: "client-secret",
				RedirectURI:  "https://app.example.com/callback",
				Scopes:       []string{"email", "profile"},
			},
			output: "google",
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			output := NewGoogleProvider(tc.input)

			require.NotNil(t, output)
			assert.Equal(t, tc.output, output.GetProviderName())
			assert.NoError(t, tc.err)
		})
	}
}

func TestGoogleProviderGetAuthCodeURL(t *testing.T) {
	type input struct {
		state string
	}
	type mock struct{}

	tcs := map[string]struct {
		input  input
		mock   mock
		output string
		err    error
	}{
		"success": {
			input:  input{state: "state-1"},
			output: "state=state-1",
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			provider := NewGoogleProvider(Config{ClientID: "client-id", RedirectURI: "https://app.example.com/callback"})

			output := provider.GetAuthCodeURL(tc.input.state)

			assert.Contains(t, output, tc.output)
			assert.Contains(t, output, "client_id=client-id")
			assert.NoError(t, tc.err)
		})
	}
}

func TestGoogleProviderExchangeCode(t *testing.T) {
	type input struct {
		code string
	}
	type mock struct {
		status int
		body   string
		err    error
	}

	tcs := map[string]struct {
		input  input
		mock   mock
		output *oauth2.Token
		err    error
	}{
		"transport_error": {
			input: input{code: "code-1"},
			mock:  mock{err: errors.New("network error")},
			err:   errors.New("network error"),
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			provider := NewGoogleProvider(Config{})
			provider.config.Endpoint.TokenURL = "https://oauth.example.com/token"
			ctx := context.WithValue(context.Background(), oauth2.HTTPClient, &http.Client{
				Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					return nil, tc.mock.err
				}),
			})

			output, err := provider.ExchangeCode(ctx, tc.input.code)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.err.Error())
			assert.Equal(t, tc.output, output)
		})
	}
}

func TestGoogleProviderGetUserInfo(t *testing.T) {
	type mock struct {
		status int
		body   string
		err    error
	}

	tcs := map[string]struct {
		input  *oauth2.Token
		mock   mock
		output *UserInfo
		err    error
	}{
		"success": {
			input: &oauth2.Token{AccessToken: "token-1"},
			mock:  mock{status: http.StatusOK, body: `{"email":"user@example.com","name":"User Name","picture":"avatar.png"}`},
			output: &UserInfo{
				Email:   "user@example.com",
				Name:    "User Name",
				Picture: "avatar.png",
			},
		},
		"transport_error": {
			input: &oauth2.Token{AccessToken: "token-1"},
			mock:  mock{err: errors.New("network error")},
			err:   errors.New("failed to get user info"),
		},
		"status_error": {
			input: &oauth2.Token{AccessToken: "token-1"},
			mock:  mock{status: http.StatusUnauthorized, body: `{}`},
			err:   errors.New("google API returned status 401"),
		},
		"decode_error": {
			input: &oauth2.Token{AccessToken: "token-1"},
			mock:  mock{status: http.StatusOK, body: `{`},
			err:   errors.New("failed to decode user info"),
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			restore := replaceDefaultTransport(tc.mock)
			defer restore()
			provider := NewGoogleProvider(Config{})

			output, err := provider.GetUserInfo(context.Background(), tc.input)

			if tc.err != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.err.Error())
				assert.Nil(t, output)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.output, output)
		})
	}
}

func TestGoogleProviderGetProviderName(t *testing.T) {
	type mock struct{}

	tcs := map[string]struct {
		input  *GoogleProvider
		mock   mock
		output string
		err    error
	}{
		"success": {
			input:  NewGoogleProvider(Config{}),
			output: "google",
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			output := tc.input.GetProviderName()

			assert.Equal(t, tc.output, output)
			assert.NoError(t, tc.err)
		})
	}
}

func TestNewAzureProvider(t *testing.T) {
	type mock struct{}

	tcs := map[string]struct {
		input  Config
		mock   mock
		output string
		err    error
	}{
		"success": {
			input:  Config{ClientID: "client-id", ClientSecret: "client-secret", RedirectURI: "https://app.example.com/callback"},
			output: "azure",
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			output := NewAzureProvider(tc.input)

			require.NotNil(t, output)
			assert.Equal(t, tc.output, output.GetProviderName())
			assert.NoError(t, tc.err)
		})
	}
}

func TestAzureProviderGetAuthCodeURL(t *testing.T) {
	type input struct {
		state string
	}
	type mock struct{}

	tcs := map[string]struct {
		input  input
		mock   mock
		output string
		err    error
	}{
		"success": {
			input:  input{state: "state-1"},
			output: "state=state-1",
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			provider := NewAzureProvider(Config{ClientID: "client-id", RedirectURI: "https://app.example.com/callback"})

			output := provider.GetAuthCodeURL(tc.input.state)

			assert.Contains(t, output, tc.output)
			assert.Contains(t, output, "client_id=client-id")
			assert.NoError(t, tc.err)
		})
	}
}

func TestAzureProviderExchangeCode(t *testing.T) {
	type input struct {
		code string
	}
	type mock struct {
		err error
	}

	tcs := map[string]struct {
		input  input
		mock   mock
		output *oauth2.Token
		err    error
	}{
		"transport_error": {
			input: input{code: "code-1"},
			mock:  mock{err: errors.New("network error")},
			err:   errors.New("network error"),
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			provider := NewAzureProvider(Config{})
			provider.config.Endpoint.TokenURL = "https://oauth.example.com/token"
			ctx := context.WithValue(context.Background(), oauth2.HTTPClient, &http.Client{
				Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					return nil, tc.mock.err
				}),
			})

			output, err := provider.ExchangeCode(ctx, tc.input.code)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.err.Error())
			assert.Equal(t, tc.output, output)
		})
	}
}

func TestAzureProviderGetUserInfo(t *testing.T) {
	type mock struct {
		status int
		body   string
		err    error
	}

	tcs := map[string]struct {
		input  *oauth2.Token
		mock   mock
		output *UserInfo
		err    error
	}{
		"success": {
			input: &oauth2.Token{AccessToken: "token-1"},
			mock:  mock{status: http.StatusOK, body: `{"mail":"user@example.com","displayName":"User Name"}`},
			output: &UserInfo{
				Email:   "user@example.com",
				Name:    "User Name",
				Picture: "",
			},
		},
		"transport_error": {
			input: &oauth2.Token{AccessToken: "token-1"},
			mock:  mock{err: errors.New("network error")},
			err:   errors.New("failed to get user info"),
		},
		"status_error": {
			input: &oauth2.Token{AccessToken: "token-1"},
			mock:  mock{status: http.StatusForbidden, body: `{}`},
			err:   errors.New("microsoft graph API returned status 403"),
		},
		"decode_error": {
			input: &oauth2.Token{AccessToken: "token-1"},
			mock:  mock{status: http.StatusOK, body: `{`},
			err:   errors.New("failed to decode user info"),
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			restore := replaceDefaultTransport(tc.mock)
			defer restore()
			provider := NewAzureProvider(Config{})

			output, err := provider.GetUserInfo(context.Background(), tc.input)

			if tc.err != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.err.Error())
				assert.Nil(t, output)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.output, output)
		})
	}
}

func TestAzureProviderGetProviderName(t *testing.T) {
	type mock struct{}

	tcs := map[string]struct {
		input  *AzureProvider
		mock   mock
		output string
		err    error
	}{
		"success": {
			input:  NewAzureProvider(Config{}),
			output: "azure",
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			output := tc.input.GetProviderName()

			assert.Equal(t, tc.output, output)
			assert.NoError(t, tc.err)
		})
	}
}

func TestNewOktaProvider(t *testing.T) {
	type input struct {
		cfg        Config
		oktaDomain string
	}
	type mock struct{}

	tcs := map[string]struct {
		input  input
		mock   mock
		output string
		err    error
	}{
		"success": {
			input:  input{cfg: Config{ClientID: "client-id"}, oktaDomain: "example.okta.com"},
			output: "okta",
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			output := NewOktaProvider(tc.input.cfg, tc.input.oktaDomain)

			require.NotNil(t, output)
			assert.Equal(t, tc.output, output.GetProviderName())
			assert.NoError(t, tc.err)
		})
	}
}

func TestOktaProviderGetAuthCodeURL(t *testing.T) {
	type input struct {
		state string
	}
	type mock struct{}

	tcs := map[string]struct {
		input  input
		mock   mock
		output string
		err    error
	}{
		"success": {
			input:  input{state: "state-1"},
			output: "state=state-1",
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			provider := NewOktaProvider(Config{ClientID: "client-id", RedirectURI: "https://app.example.com/callback"}, "example.okta.com")

			output := provider.GetAuthCodeURL(tc.input.state)

			assert.Contains(t, output, tc.output)
			assert.Contains(t, output, "client_id=client-id")
			assert.NoError(t, tc.err)
		})
	}
}

func TestOktaProviderExchangeCode(t *testing.T) {
	type input struct {
		code string
	}
	type mock struct {
		err error
	}

	tcs := map[string]struct {
		input  input
		mock   mock
		output *oauth2.Token
		err    error
	}{
		"transport_error": {
			input: input{code: "code-1"},
			mock:  mock{err: errors.New("network error")},
			err:   errors.New("network error"),
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			provider := NewOktaProvider(Config{}, "example.okta.com")
			provider.config.Endpoint.TokenURL = "https://oauth.example.com/token"
			ctx := context.WithValue(context.Background(), oauth2.HTTPClient, &http.Client{
				Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					return nil, tc.mock.err
				}),
			})

			output, err := provider.ExchangeCode(ctx, tc.input.code)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.err.Error())
			assert.Equal(t, tc.output, output)
		})
	}
}

func TestOktaProviderGetUserInfo(t *testing.T) {
	type mock struct {
		status int
		body   string
		err    error
	}

	tcs := map[string]struct {
		input  *oauth2.Token
		mock   mock
		output *UserInfo
		err    error
	}{
		"success": {
			input: &oauth2.Token{AccessToken: "token-1"},
			mock:  mock{status: http.StatusOK, body: `{"email":"user@example.com","name":"User Name","picture":"avatar.png"}`},
			output: &UserInfo{
				Email:   "user@example.com",
				Name:    "User Name",
				Picture: "avatar.png",
			},
		},
		"request_error": {
			input: &oauth2.Token{AccessToken: "token-1"},
			mock:  mock{},
			err:   errors.New("failed to create request"),
		},
		"transport_error": {
			input: &oauth2.Token{AccessToken: "token-1"},
			mock:  mock{err: errors.New("network error")},
			err:   errors.New("failed to get user info"),
		},
		"status_error": {
			input: &oauth2.Token{AccessToken: "token-1"},
			mock:  mock{status: http.StatusUnauthorized, body: `{}`},
			err:   errors.New("okta API returned status 401"),
		},
		"decode_error": {
			input: &oauth2.Token{AccessToken: "token-1"},
			mock:  mock{status: http.StatusOK, body: `{`},
			err:   errors.New("failed to decode user info"),
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			restore := replaceDefaultTransport(tc.mock)
			defer restore()
			domain := "example.okta.com"
			if name == "request_error" {
				domain = "%"
			}
			provider := NewOktaProvider(Config{}, domain)

			output, err := provider.GetUserInfo(context.Background(), tc.input)

			if tc.err != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.err.Error())
				assert.Nil(t, output)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.output, output)
		})
	}
}

func TestOktaProviderGetProviderName(t *testing.T) {
	type mock struct{}

	tcs := map[string]struct {
		input  *OktaProvider
		mock   mock
		output string
		err    error
	}{
		"success": {
			input:  NewOktaProvider(Config{}, "example.okta.com"),
			output: "okta",
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			output := tc.input.GetProviderName()

			assert.Equal(t, tc.output, output)
			assert.NoError(t, tc.err)
		})
	}
}

func replaceDefaultTransport(mock struct {
	status int
	body   string
	err    error
}) func() {
	original := http.DefaultTransport
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if mock.err != nil {
			return nil, mock.err
		}
		return &http.Response{
			StatusCode: mock.status,
			Body:       io.NopCloser(strings.NewReader(mock.body)),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	})
	return func() {
		http.DefaultTransport = original
	}
}
