package http

import (
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"identity-srv/internal/audit/repository"
	"identity-srv/internal/model"

	"github.com/gin-gonic/gin"
	sharedlog "github.com/smap-hcmut/shared-libs/go/log"
	"github.com/smap-hcmut/shared-libs/go/middleware"
	"github.com/stretchr/testify/require"
)

type mockDeps struct {
	repo *repository.MockRepository
}

func initHandler(t *testing.T) (handler, mockDeps) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	l := sharedlog.NewZapLogger(sharedlog.ZapConfig{
		Level:        sharedlog.LevelFatal,
		Mode:         sharedlog.ModeProduction,
		Encoding:     sharedlog.EncodingJSON,
		ColorEnabled: false,
	})
	repo := repository.NewMockRepository(t)

	return handler{
		l:    l,
		repo: repo,
	}, mockDeps{repo: repo}
}

func TestGetAuditLogs(t *testing.T) {
	from := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 5, 4, 0, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)
	expiresAt := createdAt.AddDate(0, 0, 90)
	userID := "user-1"
	resourceType := "project"
	resourceID := "project-1"
	ipAddress := "127.0.0.1"
	userAgent := "Mozilla/5.0"

	type mockRepoQuery struct {
		isCalled bool
		input    repository.QueryOptions
		output   []model.AuditLog
		total    int
		err      error
	}
	type mock struct {
		query mockRepoQuery
	}

	tcs := map[string]struct {
		input    string
		mock     mock
		output   string
		err      error
		wantCode int
		wantBody string
	}{
		"success": {
			input: "/audit-logs?page=2&limit=10&user_id=user-1&action=LOGIN&from=2026-05-01T00:00:00Z&to=2026-05-04T00:00:00Z",
			mock: mock{query: mockRepoQuery{
				isCalled: true,
				input: repository.QueryOptions{
					UserID: "user-1",
					Action: model.ActionLogin,
					From:   &from,
					To:     &to,
					Page:   2,
					Limit:  10,
				},
				output: []model.AuditLog{{
					ID:           "audit-1",
					UserID:       &userID,
					Action:       model.ActionLogin,
					ResourceType: &resourceType,
					ResourceID:   &resourceID,
					Metadata:     map[string]interface{}{"provider": "google", "attempt": 1},
					IPAddress:    &ipAddress,
					UserAgent:    &userAgent,
					CreatedAt:    createdAt,
					ExpiresAt:    expiresAt,
				}},
				total: 1,
			}},
			wantCode: 200,
			wantBody: `{
				"error_code": 0,
				"message": "Success",
				"data": {
					"logs": [{
						"id": "audit-1",
						"user_id": "user-1",
						"action": "LOGIN",
						"resource_type": "project",
						"resource_id": "project-1",
						"metadata": {"provider":"google","attempt":"1"},
						"ip_address": "127.0.0.1",
						"user_agent": "Mozilla/5.0",
						"created_at": "2026-05-02T10:00:00Z",
						"expires_at": "2026-07-31T10:00:00Z"
					}],
					"total_count": 1,
					"page": 2,
					"limit": 10
				}
			}`,
		},
		"default_pagination_and_limit_cap": {
			input: "/audit-logs?limit=999",
			mock: mock{query: mockRepoQuery{
				isCalled: true,
				input:    repository.QueryOptions{Page: 1, Limit: 100},
				output:   []model.AuditLog{},
				total:    0,
			}},
			wantCode: 200,
			wantBody: `{"error_code":0,"message":"Success","data":{"logs":[],"total_count":0,"page":1,"limit":100}}`,
		},
		"invalid_page": {
			input:    "/audit-logs?page=abc",
			wantCode: 400,
			wantBody: `{"error_code":30001,"message":"Invalid page number"}`,
		},
		"invalid_limit": {
			input:    "/audit-logs?limit=0",
			wantCode: 400,
			wantBody: `{"error_code":30002,"message":"Invalid limit"}`,
		},
		"invalid_from": {
			input:    "/audit-logs?from=bad-date",
			wantCode: 400,
			wantBody: `{"error_code":30003,"message":"Invalid from date format (use RFC3339)"}`,
		},
		"invalid_to": {
			input:    "/audit-logs?to=bad-date",
			wantCode: 400,
			wantBody: `{"error_code":30004,"message":"Invalid to date format (use RFC3339)"}`,
		},
		"repo_error": {
			input: "/audit-logs?page=1&limit=10",
			mock: mock{query: mockRepoQuery{
				isCalled: true,
				input:    repository.QueryOptions{Page: 1, Limit: 10},
				err:      errors.New("repo error"),
			}},
			wantCode: 500,
			wantBody: `{"error_code":500,"message":"Something went wrong"}`,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			h, deps := initHandler(t)
			w := httptest.NewRecorder()
			c, engine := gin.CreateTestContext(w)
			engine.GET("/audit-logs", h.GetAuditLogs)
			c.Request = httptest.NewRequest("GET", tc.input, nil)
			if tc.mock.query.isCalled {
				deps.repo.EXPECT().Query(c.Request.Context(), tc.mock.query.input).
					Return(tc.mock.query.output, tc.mock.query.total, tc.mock.query.err)
			}

			engine.ServeHTTP(w, c.Request)

			require.Equal(t, tc.wantCode, w.Code)
			require.JSONEq(t, tc.wantBody, w.Body.String())
		})
	}
}

func TestNew(t *testing.T) {
	type mock struct{}

	tcs := map[string]struct {
		input  struct{}
		mock   mock
		output bool
		err    error
	}{
		"success": {
			output: true,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			h := New(nil, nil, nil)

			require.Equal(t, tc.output, h != nil)
			require.NoError(t, tc.err)
		})
	}
}

func TestRegisterRoutes(t *testing.T) {
	type mock struct{}

	tcs := map[string]struct {
		input  string
		mock   mock
		output []string
		err    error
	}{
		"success": {
			input:  "/audit-logs",
			output: []string{"GET /audit-logs"},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			engine := gin.New()
			h, _ := initHandler(t)
			mw := middleware.New(middleware.Config{})

			h.RegisterRoutes(engine.Group(tc.input), mw)

			routes := make([]string, 0, len(engine.Routes()))
			for _, route := range engine.Routes() {
				routes = append(routes, route.Method+" "+route.Path)
			}
			require.Equal(t, tc.output, routes)
			require.NoError(t, tc.err)
		})
	}
}
