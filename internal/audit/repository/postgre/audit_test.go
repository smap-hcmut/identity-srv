package postgre

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"identity-srv/internal/audit"
	"identity-srv/internal/audit/repository"
	"identity-srv/internal/model"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeScanner struct {
	err error
}

func (f fakeScanner) Scan(dest ...interface{}) error {
	if f.err != nil {
		return f.err
	}
	userID := "user-1"
	resourceType := "project"
	resourceID := "project-1"
	ipAddress := "127.0.0.1"
	userAgent := "Mozilla/5.0"
	now := time.Now()

	*(dest[0].(*string)) = "audit-1"
	*(dest[1].(**string)) = &userID
	*(dest[2].(*string)) = model.ActionLogin
	*(dest[3].(**string)) = &resourceType
	*(dest[4].(**string)) = &resourceID
	*(dest[5].(*[]byte)) = []byte(`{"provider":"google"}`)
	*(dest[6].(**string)) = &ipAddress
	*(dest[7].(**string)) = &userAgent
	*(dest[8].(*time.Time)) = now
	*(dest[9].(*time.Time)) = now.AddDate(0, 0, 90)
	return nil
}

func TestScanAuditLog(t *testing.T) {
	type mock struct {
		scanner fakeScanner
	}

	tcs := map[string]struct {
		input  struct{}
		mock   mock
		output model.AuditLog
		err    error
	}{
		"success": {
			mock: mock{scanner: fakeScanner{}},
			output: model.AuditLog{
				ID:           "audit-1",
				UserID:       strPtr("user-1"),
				Action:       model.ActionLogin,
				ResourceType: strPtr("project"),
				ResourceID:   strPtr("project-1"),
				Metadata:     map[string]interface{}{"provider": "google"},
				IPAddress:    strPtr("127.0.0.1"),
				UserAgent:    strPtr("Mozilla/5.0"),
			},
		},
		"scan_error": {
			mock: mock{scanner: fakeScanner{err: errors.New("scan error")}},
			err:  errors.New("scan error"),
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			repo := &implRepository{}

			output, err := repo.scanAuditLog(tc.mock.scanner)

			if tc.err != nil {
				require.Error(t, err)
				assert.EqualError(t, err, tc.err.Error())
				assert.Equal(t, tc.output, output)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.output.ID, output.ID)
			assert.Equal(t, tc.output.UserID, output.UserID)
			assert.Equal(t, tc.output.Action, output.Action)
			assert.Equal(t, tc.output.ResourceType, output.ResourceType)
			assert.Equal(t, tc.output.ResourceID, output.ResourceID)
			assert.Equal(t, tc.output.Metadata, output.Metadata)
			assert.Equal(t, tc.output.IPAddress, output.IPAddress)
			assert.Equal(t, tc.output.UserAgent, output.UserAgent)
			assert.False(t, output.CreatedAt.IsZero())
			assert.False(t, output.ExpiresAt.IsZero())
		})
	}
}

func TestMarshalMetadata(t *testing.T) {
	type mock struct{}

	tcs := map[string]struct {
		input  map[string]string
		mock   mock
		output string
		err    error
	}{
		"success": {
			input:  map[string]string{"provider": "google"},
			output: `{"provider":"google"}`,
		},
		"nil_metadata": {
			output: `{}`,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			repo := &implRepository{}

			output := repo.marshalMetadata(tc.input)

			assert.JSONEq(t, tc.output, string(output))
			assert.NoError(t, tc.err)
		})
	}
}

func TestUnmarshalMetadata(t *testing.T) {
	type mock struct{}

	tcs := map[string]struct {
		input  []byte
		mock   mock
		output map[string]interface{}
		err    error
	}{
		"success": {
			input:  []byte(`{"provider":"google"}`),
			output: map[string]interface{}{"provider": "google"},
		},
		"empty_data": {
			output: map[string]interface{}{},
		},
		"invalid_json": {
			input:  []byte(`{`),
			output: map[string]interface{}{},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			repo := &implRepository{}

			output := repo.unmarshalMetadata(tc.input)

			assert.Equal(t, tc.output, output)
			assert.NoError(t, tc.err)
		})
	}
}

func TestBuildBatchInsertQuery(t *testing.T) {
	now := time.Now()
	type mock struct{}

	tcs := map[string]struct {
		input  []audit.AuditEvent
		mock   mock
		output int
		err    error
	}{
		"success_single_event": {
			input: []audit.AuditEvent{{
				UserID:       "user-1",
				Action:       audit.ActionLogin,
				ResourceType: "project",
				ResourceID:   "project-1",
				Metadata:     map[string]string{"provider": "google"},
				IPAddress:    "127.0.0.1",
				UserAgent:    "Mozilla/5.0",
				Timestamp:    now,
			}},
			output: 8,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			repo := &implRepository{}

			query, args := repo.buildBatchInsertQuery(tc.input)

			assert.Contains(t, query, "INSERT INTO audit_logs")
			assert.Contains(t, query, "$8::timestamp + interval '90 days'")
			assert.Equal(t, tc.output, len(args))
			assert.Equal(t, tc.input[0].UserID, args[0])
			assert.Equal(t, string(tc.input[0].Action), args[1])
			assert.NoError(t, tc.err)
		})
	}
}

func TestBuildQueryFilter(t *testing.T) {
	from := time.Now().Add(-time.Hour)
	to := time.Now()
	type mock struct{}

	tcs := map[string]struct {
		input  repository.QueryOptions
		mock   mock
		output struct {
			where    string
			args     []interface{}
			argIndex int
		}
		err error
	}{
		"empty_filter": {
			output: struct {
				where    string
				args     []interface{}
				argIndex int
			}{args: []interface{}{}, argIndex: 1},
		},
		"all_filters": {
			input: repository.QueryOptions{
				UserID: "user-1",
				Action: model.ActionLogin,
				From:   &from,
				To:     &to,
			},
			output: struct {
				where    string
				args     []interface{}
				argIndex int
			}{
				where:    "WHERE user_id = $1 AND action = $2 AND created_at >= $3 AND created_at <= $4",
				args:     []interface{}{"user-1", model.ActionLogin, from, to},
				argIndex: 5,
			},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			repo := &implRepository{}

			where, args, argIndex := repo.buildQueryFilter(tc.input)

			assert.Equal(t, tc.output.where, where)
			assert.Equal(t, tc.output.args, args)
			assert.Equal(t, tc.output.argIndex, argIndex)
			assert.NoError(t, tc.err)
		})
	}
}

func TestBuildCountQuery(t *testing.T) {
	type mock struct{}

	tcs := map[string]struct {
		input  string
		mock   mock
		output string
		err    error
	}{
		"success": {
			input:  "WHERE user_id = $1",
			output: "SELECT COUNT(*) FROM audit_logs WHERE user_id = $1",
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			repo := &implRepository{}

			output := repo.buildCountQuery(tc.input)

			assert.Equal(t, tc.output, output)
			assert.NoError(t, tc.err)
		})
	}
}

func TestBuildPaginatedQuery(t *testing.T) {
	type input struct {
		whereClause string
		argIndex    int
	}
	type mock struct{}

	tcs := map[string]struct {
		input  input
		mock   mock
		output []string
		err    error
	}{
		"success": {
			input: input{whereClause: "WHERE user_id = $1", argIndex: 2},
			output: []string{
				"FROM audit_logs",
				"WHERE user_id = $1",
				"LIMIT $2 OFFSET $3",
			},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			repo := &implRepository{}

			output := repo.buildPaginatedQuery(tc.input.whereClause, tc.input.argIndex)

			for _, expected := range tc.output {
				assert.Contains(t, strings.Join(strings.Fields(output), " "), expected)
			}
			assert.NoError(t, tc.err)
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
			output := New(nil, nil)

			assert.Equal(t, tc.output, output != nil)
			assert.NoError(t, tc.err)
		})
	}
}

func TestBatchInsert(t *testing.T) {
	now := time.Now()
	type mock struct {
		execErr error
	}

	tcs := map[string]struct {
		input  []audit.AuditEvent
		mock   mock
		output error
		err    error
	}{
		"empty_events": {},
		"success": {
			input: []audit.AuditEvent{{
				UserID:       "user-1",
				Action:       audit.ActionLogin,
				ResourceType: "project",
				ResourceID:   "project-1",
				Metadata:     map[string]string{"provider": "google"},
				IPAddress:    "127.0.0.1",
				UserAgent:    "Mozilla/5.0",
				Timestamp:    now,
			}},
		},
		"exec_error": {
			input: []audit.AuditEvent{{
				UserID:       "user-1",
				Action:       audit.ActionLogin,
				ResourceType: "project",
				ResourceID:   "project-1",
				Timestamp:    now,
			}},
			mock: mock{execErr: errors.New("exec error")},
			err:  errors.New("failed to batch insert audit logs"),
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			db, sqlMock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()
			repo := &implRepository{db: db}
			if len(tc.input) > 0 {
				expect := sqlMock.ExpectExec("INSERT INTO audit_logs").WithArgs(
					tc.input[0].UserID,
					string(tc.input[0].Action),
					tc.input[0].ResourceType,
					tc.input[0].ResourceID,
					sqlmock.AnyArg(),
					tc.input[0].IPAddress,
					tc.input[0].UserAgent,
					tc.input[0].Timestamp,
				)
				if tc.mock.execErr != nil {
					expect.WillReturnError(tc.mock.execErr)
				} else {
					expect.WillReturnResult(sqlmock.NewResult(0, 1))
				}
			}

			output := repo.BatchInsert(context.Background(), tc.input)

			if tc.err != nil {
				require.Error(t, output)
				assert.Contains(t, output.Error(), tc.err.Error())
			} else {
				assert.NoError(t, output)
			}
			assert.NoError(t, sqlMock.ExpectationsWereMet())
		})
	}
}

func TestDeleteExpired(t *testing.T) {
	type mock struct {
		execErr         error
		rowsAffectedErr error
	}

	tcs := map[string]struct {
		input  struct{}
		mock   mock
		output int64
		err    error
	}{
		"success": {
			output: 2,
		},
		"exec_error": {
			mock: mock{execErr: errors.New("exec error")},
			err:  errors.New("failed to delete expired audit logs"),
		},
		"rows_affected_error": {
			mock: mock{rowsAffectedErr: errors.New("rows affected error")},
			err:  errors.New("failed to get rows affected"),
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			db, sqlMock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()
			repo := &implRepository{db: db}
			expect := sqlMock.ExpectExec(regexp.QuoteMeta("DELETE FROM audit_logs WHERE expires_at < NOW()"))
			switch {
			case tc.mock.execErr != nil:
				expect.WillReturnError(tc.mock.execErr)
			case tc.mock.rowsAffectedErr != nil:
				expect.WillReturnResult(sqlmock.NewErrorResult(tc.mock.rowsAffectedErr))
			default:
				expect.WillReturnResult(sqlmock.NewResult(0, tc.output))
			}

			output, err := repo.DeleteExpired(context.Background())

			if tc.err != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.err.Error())
				assert.Zero(t, output)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.output, output)
			}
			assert.NoError(t, sqlMock.ExpectationsWereMet())
		})
	}
}

func TestQuery(t *testing.T) {
	now := time.Now()
	type mock struct {
		countErr error
		queryErr error
		rows     *sqlmock.Rows
	}

	tcs := map[string]struct {
		input  repository.QueryOptions
		mock   mock
		output []model.AuditLog
		err    error
	}{
		"success": {
			input: repository.QueryOptions{UserID: "user-1", Page: 2, Limit: 10},
			mock: mock{rows: sqlmock.NewRows([]string{
				"id", "user_id", "action", "resource_type", "resource_id", "metadata", "ip_address", "user_agent", "created_at", "expires_at",
			}).AddRow(
				"audit-1", "user-1", model.ActionLogin, "project", "project-1", []byte(`{"provider":"google"}`), "127.0.0.1", "Mozilla/5.0", now, now.AddDate(0, 0, 90),
			)},
			output: []model.AuditLog{{
				ID:           "audit-1",
				UserID:       strPtr("user-1"),
				Action:       model.ActionLogin,
				ResourceType: strPtr("project"),
				ResourceID:   strPtr("project-1"),
				Metadata:     map[string]interface{}{"provider": "google"},
				IPAddress:    strPtr("127.0.0.1"),
				UserAgent:    strPtr("Mozilla/5.0"),
				CreatedAt:    now,
				ExpiresAt:    now.AddDate(0, 0, 90),
			}},
		},
		"count_error": {
			input: repository.QueryOptions{UserID: "user-1", Page: 1, Limit: 10},
			mock:  mock{countErr: errors.New("count error")},
			err:   errors.New("failed to count audit logs"),
		},
		"query_error": {
			input: repository.QueryOptions{UserID: "user-1", Page: 1, Limit: 10},
			mock:  mock{queryErr: errors.New("query error")},
			err:   errors.New("failed to query audit logs"),
		},
		"scan_error": {
			input: repository.QueryOptions{UserID: "user-1", Page: 1, Limit: 10},
			mock: mock{rows: sqlmock.NewRows([]string{
				"id", "user_id", "action", "resource_type", "resource_id", "metadata", "ip_address", "user_agent", "created_at", "expires_at",
			}).AddRow(
				"audit-1", "user-1", model.ActionLogin, "project", "project-1", []byte(`{}`), "127.0.0.1", "Mozilla/5.0", "bad-time", now.AddDate(0, 0, 90),
			)},
			err: errors.New("failed to scan audit log"),
		},
		"rows_error": {
			input: repository.QueryOptions{UserID: "user-1", Page: 1, Limit: 10},
			mock: mock{rows: sqlmock.NewRows([]string{
				"id", "user_id", "action", "resource_type", "resource_id", "metadata", "ip_address", "user_agent", "created_at", "expires_at",
			}).AddRow(
				"audit-1", "user-1", model.ActionLogin, "project", "project-1", []byte(`{}`), "127.0.0.1", "Mozilla/5.0", now, now.AddDate(0, 0, 90),
			).RowError(0, errors.New("rows error"))},
			err: errors.New("error iterating audit logs"),
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			db, sqlMock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()
			repo := &implRepository{db: db}
			countQuery := regexp.QuoteMeta("SELECT COUNT(*) FROM audit_logs WHERE user_id = $1")
			countExpect := sqlMock.ExpectQuery(countQuery).WithArgs(tc.input.UserID)
			if tc.mock.countErr != nil {
				countExpect.WillReturnError(tc.mock.countErr)
			} else {
				countExpect.WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
				queryExpect := sqlMock.ExpectQuery("SELECT id, user_id, action, resource_type, resource_id").WithArgs(tc.input.UserID, tc.input.Limit, (tc.input.Page-1)*tc.input.Limit)
				if tc.mock.queryErr != nil {
					queryExpect.WillReturnError(tc.mock.queryErr)
				} else {
					queryExpect.WillReturnRows(tc.mock.rows)
				}
			}

			output, total, err := repo.Query(context.Background(), tc.input)

			if tc.err != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.err.Error())
				assert.Nil(t, output)
				assert.Zero(t, total)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.output, output)
				assert.Equal(t, 1, total)
			}
			assert.NoError(t, sqlMock.ExpectationsWereMet())
		})
	}
}

func strPtr(value string) *string {
	return &value
}
