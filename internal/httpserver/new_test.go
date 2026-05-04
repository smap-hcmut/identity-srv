package httpserver

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"identity-srv/config"
	"identity-srv/internal/authentication/usecase"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	sharedauth "github.com/smap-hcmut/shared-libs/go/auth"
	sharedencrypter "github.com/smap-hcmut/shared-libs/go/encrypter"
	sharedlog "github.com/smap-hcmut/shared-libs/go/log"
	sharedredis "github.com/smap-hcmut/shared-libs/go/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	type mock struct{}

	tcs := map[string]struct {
		input  HTTPServer
		mock   mock
		output error
		err    error
	}{
		"success": {
			input:  validHTTPServer(),
			output: nil,
		},
		"missing_logger": {
			input:  withHTTPServer(validHTTPServer(), func(srv *HTTPServer) { srv.l = nil }),
			output: errors.New("logger is required"),
			err:    errors.New("logger is required"),
		},
		"missing_mode": {
			input:  withHTTPServer(validHTTPServer(), func(srv *HTTPServer) { srv.mode = "" }),
			output: errors.New("mode is required"),
			err:    errors.New("mode is required"),
		},
		"missing_port": {
			input:  withHTTPServer(validHTTPServer(), func(srv *HTTPServer) { srv.port = 0 }),
			output: errors.New("port is required"),
			err:    errors.New("port is required"),
		},
		"missing_postgres": {
			input:  withHTTPServer(validHTTPServer(), func(srv *HTTPServer) { srv.postgresDB = nil }),
			output: errors.New("postgresDB is required"),
			err:    errors.New("postgresDB is required"),
		},
		"missing_config": {
			input:  withHTTPServer(validHTTPServer(), func(srv *HTTPServer) { srv.config = nil }),
			output: errors.New("config is required"),
			err:    errors.New("config is required"),
		},
		"missing_jwt_manager": {
			input:  withHTTPServer(validHTTPServer(), func(srv *HTTPServer) { srv.jwtManager = nil }),
			output: errors.New("jwtManager is required"),
			err:    errors.New("jwtManager is required"),
		},
		"missing_redis": {
			input:  withHTTPServer(validHTTPServer(), func(srv *HTTPServer) { srv.redisClient = nil }),
			output: errors.New("redisClient is required"),
			err:    errors.New("redisClient is required"),
		},
		"missing_session_manager": {
			input:  withHTTPServer(validHTTPServer(), func(srv *HTTPServer) { srv.sessionManager = nil }),
			output: errors.New("sessionManager is required"),
			err:    errors.New("sessionManager is required"),
		},
		"missing_blacklist_manager": {
			input:  withHTTPServer(validHTTPServer(), func(srv *HTTPServer) { srv.blacklistManager = nil }),
			output: errors.New("blacklistManager is required"),
			err:    errors.New("blacklistManager is required"),
		},
		"missing_encrypter": {
			input:  withHTTPServer(validHTTPServer(), func(srv *HTTPServer) { srv.encrypter = nil }),
			output: errors.New("encrypter is required"),
			err:    errors.New("encrypter is required"),
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			output := tc.input.validate()

			if tc.err != nil {
				assert.EqualError(t, output, tc.output.Error())
				return
			}
			assert.NoError(t, output)
		})
	}
}

func TestNew(t *testing.T) {
	type mock struct{}

	tcs := map[string]struct {
		input  Config
		mock   mock
		output bool
		err    error
	}{
		"success": {
			input: Config{
				Logger:      &sharedlog.MockLogger{},
				Host:        "127.0.0.1",
				Port:        8080,
				Mode:        "test",
				Environment: "test",
				PostgresDB:  &sql.DB{},
				Config: &config.Config{
					Session: config.SessionConfig{TTL: 3600},
				},
				JWTManager:  &sharedauth.MockManager{},
				RedisClient: &sharedredis.MockIRedis{},
				Encrypter:   &sharedencrypter.MockEncrypter{},
			},
			output: true,
		},
		"validate_error": {
			input: Config{
				Logger:      &sharedlog.MockLogger{},
				Port:        8080,
				Mode:        "test",
				Environment: "test",
				Config:      &config.Config{},
				JWTManager:  &sharedauth.MockManager{},
				RedisClient: &sharedredis.MockIRedis{},
				Encrypter:   &sharedencrypter.MockEncrypter{},
			},
			err: errors.New("postgresDB is required"),
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			output, err := New(tc.input.Logger, tc.input)

			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
				assert.Nil(t, output)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.output, output != nil)
		})
	}
}

func TestHealthCheck(t *testing.T) {
	type mock struct{}

	tcs := map[string]struct {
		input  string
		mock   mock
		output int
		err    error
	}{
		"success": {
			input:  "/health",
			output: http.StatusOK,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			srv := HTTPServer{}
			router.GET(tc.input, srv.healthCheck)
			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tc.input, nil))

			assert.Equal(t, tc.output, recorder.Code)
			assert.Contains(t, recorder.Body.String(), "healthy")
			assert.NoError(t, tc.err)
		})
	}
}

func TestReadyCheck(t *testing.T) {
	type mock struct {
		pingErr error
	}

	tcs := map[string]struct {
		input  string
		mock   mock
		output int
		err    error
	}{
		"success": {
			input:  "/ready",
			output: http.StatusOK,
		},
		"ping_error": {
			input:  "/ready",
			mock:   mock{pingErr: errors.New("ping error")},
			output: http.StatusServiceUnavailable,
			err:    errors.New("ping error"),
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			db, sqlMock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
			require.NoError(t, err)
			defer db.Close()
			expect := sqlMock.ExpectPing()
			if tc.mock.pingErr != nil {
				expect.WillReturnError(tc.mock.pingErr)
			}
			router := gin.New()
			srv := HTTPServer{postgresDB: db}
			router.GET(tc.input, srv.readyCheck)
			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tc.input, nil))

			assert.Equal(t, tc.output, recorder.Code)
			if tc.err != nil {
				assert.Contains(t, recorder.Body.String(), tc.err.Error())
			} else {
				assert.Contains(t, recorder.Body.String(), "ready")
			}
			assert.NoError(t, sqlMock.ExpectationsWereMet())
		})
	}
}

func TestLiveCheck(t *testing.T) {
	type mock struct{}

	tcs := map[string]struct {
		input  string
		mock   mock
		output int
		err    error
	}{
		"success": {
			input:  "/live",
			output: http.StatusOK,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			srv := HTTPServer{}
			router.GET(tc.input, srv.liveCheck)
			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tc.input, nil))

			assert.Equal(t, tc.output, recorder.Code)
			assert.Contains(t, recorder.Body.String(), "alive")
			assert.NoError(t, tc.err)
		})
	}
}

func TestRegisterSystemRoutes(t *testing.T) {
	type mock struct{}

	tcs := map[string]struct {
		input  string
		mock   mock
		output []string
		err    error
	}{
		"success_development": {
			input: "development",
			output: []string{
				"GET /health",
				"GET /ready",
				"GET /live",
				"GET /test",
				"GET /swagger/*any",
			},
		},
		"success_production": {
			input: "production",
			output: []string{
				"GET /health",
				"GET /ready",
				"GET /live",
				"GET /swagger/*any",
			},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			srv := HTTPServer{gin: gin.New(), environment: tc.input}

			srv.registerSystemRoutes()

			routes := make([]string, 0, len(srv.gin.Routes()))
			for _, route := range srv.gin.Routes() {
				routes = append(routes, route.Method+" "+route.Path)
			}
			for _, expected := range tc.output {
				assert.Contains(t, routes, expected)
			}
			assert.NoError(t, tc.err)
		})
	}
}

func TestInitOAuthProvider(t *testing.T) {
	type mockData struct {
		logger *sharedlog.MockLogger
	}

	tcs := map[string]struct {
		input  config.OAuth2Config
		mock   mockData
		output string
		err    error
	}{
		"success": {
			input:  config.OAuth2Config{Provider: "google"},
			mock:   mockData{logger: &sharedlog.MockLogger{}},
			output: "google",
		},
		"unsupported_provider": {
			input: config.OAuth2Config{Provider: "github"},
			mock:  mockData{logger: &sharedlog.MockLogger{}},
			err:   errors.New("unsupported provider type"),
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			if tc.err == nil {
				tc.mock.logger.On("Infof", mock.Anything, "OAuth provider initialized: %s", tc.output).Return()
			}
			srv := HTTPServer{
				l: tc.mock.logger,
				config: &config.Config{
					OAuth2: tc.input,
				},
			}

			output, err := srv.initOAuthProvider()

			if tc.err != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.err.Error())
				assert.Nil(t, output)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.output, output.GetProviderName())
			tc.mock.logger.AssertExpectations(t)
		})
	}
}

func validHTTPServer() HTTPServer {
	return HTTPServer{
		l:                &sharedlog.MockLogger{},
		mode:             "test",
		port:             8080,
		postgresDB:       &sql.DB{},
		config:           &config.Config{},
		jwtManager:       &sharedauth.MockManager{},
		redisClient:      &sharedredis.MockIRedis{},
		sessionManager:   &usecase.SessionManager{},
		blacklistManager: &usecase.BlacklistManager{},
		encrypter:        &sharedencrypter.MockEncrypter{},
	}
}

func withHTTPServer(srv HTTPServer, mutate func(*HTTPServer)) HTTPServer {
	mutate(&srv)
	return srv
}
