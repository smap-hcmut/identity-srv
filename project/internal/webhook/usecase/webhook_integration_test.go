package usecase

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"smap-project/internal/webhook"
	"smap-project/pkg/log"
)

// mockRedisClient implements a simple in-memory Redis client for testing
type mockRedisClient struct {
	data map[string][]byte
}

func newMockRedisClient() *mockRedisClient {
	return &mockRedisClient{
		data: make(map[string][]byte),
	}
}

func (m *mockRedisClient) Disconnect() error {
	return nil
}

func (m *mockRedisClient) Set(ctx context.Context, key string, value interface{}, expirationSeconds int) error {
	// Convert value to bytes
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return nil
	}
	m.data[key] = bytes
	return nil
}

func (m *mockRedisClient) Get(ctx context.Context, key string) ([]byte, error) {
	if val, ok := m.data[key]; ok {
		return val, nil
	}
	// Return error similar to redis.Nil
	return nil, context.DeadlineExceeded // Using a standard error instead
}

func (m *mockRedisClient) Lock(ctx context.Context, key string, expiration int) (bool, error) {
	return true, nil
}

func (m *mockRedisClient) Unlock(ctx context.Context, key string) error {
	return nil
}

func (m *mockRedisClient) Publish(ctx context.Context, channel string, message interface{}) error {
	// For testing, we just verify it doesn't error
	return nil
}

// TestCheckpoint_Phase1And2 verifies the checkpoint requirements:
// 1. Job mappings are stored when jobs are created
// 2. Callbacks work with both lookup and fallback
// 3. Old collector callbacks (with UserID) still work
func TestCheckpoint_Phase1And2(t *testing.T) {
	ctx := context.Background()
	mockRedis := newMockRedisClient()
	logger := log.NewNopLogger()

	uc := &usecase{
		l:           logger,
		redisClient: mockRedis,
	}

	t.Run("1. Verify job mappings are stored when jobs are created", func(t *testing.T) {
		jobID := "test-job-123"
		userID := "user-456"
		projectID := "project-789"

		// Store job mapping (simulating job creation)
		err := uc.StoreJobMapping(ctx, jobID, userID, projectID)
		if err != nil {
			t.Fatalf("Failed to store job mapping: %v", err)
		}

		// Verify the mapping was stored
		retrievedUserID, retrievedProjectID, err := uc.getJobMapping(ctx, jobID)
		if err != nil {
			t.Fatalf("Failed to retrieve job mapping: %v", err)
		}

		if retrievedUserID != userID {
			t.Errorf("Expected userID %s, got %s", userID, retrievedUserID)
		}

		if retrievedProjectID != projectID {
			t.Errorf("Expected projectID %s, got %s", projectID, retrievedProjectID)
		}

		t.Logf("✓ Job mapping stored and retrieved successfully: jobID=%s, userID=%s, projectID=%s", jobID, userID, projectID)
	})

	t.Run("2. Verify callbacks work with lookup (no UserID in request)", func(t *testing.T) {
		jobID := "test-job-lookup"
		userID := "user-lookup"
		projectID := "project-lookup"

		// Store job mapping first
		err := uc.StoreJobMapping(ctx, jobID, userID, projectID)
		if err != nil {
			t.Fatalf("Failed to store job mapping: %v", err)
		}

		// Create callback request WITHOUT UserID (new format)
		req := webhook.CallbackRequest{
			JobID:    jobID,
			Status:   "success",
			Platform: "youtube",
			Payload:  webhook.CallbackPayload{},
			UserID:   "", // Empty UserID - should use lookup
		}

		// Process callback
		err = uc.HandleDryRunCallback(ctx, req)
		if err != nil {
			t.Fatalf("Failed to handle callback with lookup: %v", err)
		}

		t.Logf("✓ Callback processed successfully using lookup: jobID=%s", jobID)
	})

	t.Run("3. Verify callbacks work with fallback (UserID provided)", func(t *testing.T) {
		jobID := "test-job-fallback"
		userID := "user-fallback"

		// Create callback request WITH UserID (old format)
		// Note: We're NOT storing a job mapping to test the fallback
		req := webhook.CallbackRequest{
			JobID:    jobID,
			Status:   "success",
			Platform: "tiktok",
			Payload:  webhook.CallbackPayload{},
			UserID:   userID, // Provided UserID - should use fallback
		}

		// Process callback
		err := uc.HandleDryRunCallback(ctx, req)
		if err != nil {
			t.Fatalf("Failed to handle callback with fallback: %v", err)
		}

		t.Logf("✓ Callback processed successfully using fallback: jobID=%s, userID=%s", jobID, userID)
	})

	t.Run("4. Verify old collector callbacks (with UserID) still work", func(t *testing.T) {
		jobID := "test-job-old-format"
		userID := "user-old-format"
		projectID := "project-old-format"

		// Store job mapping (simulating a job that was created)
		err := uc.StoreJobMapping(ctx, jobID, userID, projectID)
		if err != nil {
			t.Fatalf("Failed to store job mapping: %v", err)
		}

		// Create callback request WITH UserID (old collector format)
		req := webhook.CallbackRequest{
			JobID:    jobID,
			Status:   "success",
			Platform: "youtube",
			Payload:  webhook.CallbackPayload{},
			UserID:   userID, // Old format includes UserID
		}

		// Process callback - should prefer lookup but fallback works too
		err = uc.HandleDryRunCallback(ctx, req)
		if err != nil {
			t.Fatalf("Failed to handle old format callback: %v", err)
		}

		t.Logf("✓ Old collector callback processed successfully: jobID=%s, userID=%s", jobID, userID)
	})

	t.Run("5. Verify missing JobID is handled gracefully", func(t *testing.T) {
		jobID := "test-job-missing"

		// Create callback request WITHOUT UserID and WITHOUT stored mapping
		req := webhook.CallbackRequest{
			JobID:    jobID,
			Status:   "failed",
			Platform: "youtube",
			Payload:  webhook.CallbackPayload{},
			UserID:   "", // No UserID provided
		}

		// Process callback - should log error but return nil (acknowledge callback)
		// This prevents retries for missing job mappings
		err := uc.HandleDryRunCallback(ctx, req)
		if err != nil {
			t.Errorf("Expected nil error for missing JobID (should acknowledge callback), got: %v", err)
		}

		t.Logf("✓ Missing JobID handled gracefully (callback acknowledged, notification skipped)")
	})

	t.Run("6. Verify dry-run jobs with empty projectID work", func(t *testing.T) {
		jobID := "test-job-dryrun"
		userID := "user-dryrun"
		projectID := "" // Empty for dry-run jobs

		// Store job mapping with empty projectID (simulating dry-run job creation)
		err := uc.StoreJobMapping(ctx, jobID, userID, projectID)
		if err != nil {
			t.Fatalf("Failed to store job mapping with empty projectID: %v", err)
		}

		// Verify the mapping was stored
		retrievedUserID, retrievedProjectID, err := uc.getJobMapping(ctx, jobID)
		if err != nil {
			t.Fatalf("Failed to retrieve job mapping: %v", err)
		}

		if retrievedUserID != userID {
			t.Errorf("Expected userID %s, got %s", userID, retrievedUserID)
		}

		if retrievedProjectID != projectID {
			t.Errorf("Expected empty projectID, got %s", retrievedProjectID)
		}

		t.Logf("✓ Dry-run job mapping (empty projectID) stored and retrieved successfully")
	})

	t.Run("7. Verify job mapping data includes timestamp", func(t *testing.T) {
		jobID := "test-job-timestamp"
		userID := "user-timestamp"
		projectID := "project-timestamp"

		// Store job mapping
		err := uc.StoreJobMapping(ctx, jobID, userID, projectID)
		if err != nil {
			t.Fatalf("Failed to store job mapping: %v", err)
		}

		// Retrieve raw data from Redis to check timestamp
		key := "job:mapping:" + jobID
		jsonData, err := mockRedis.Get(ctx, key)
		if err != nil {
			t.Fatalf("Failed to get raw data from Redis: %v", err)
		}

		var data webhook.JobMappingData
		if err := json.Unmarshal(jsonData, &data); err != nil {
			t.Fatalf("Failed to unmarshal job mapping data: %v", err)
		}

		if data.CreatedAt.IsZero() {
			t.Error("Expected non-zero CreatedAt timestamp")
		}

		if time.Since(data.CreatedAt) > 5*time.Second {
			t.Errorf("CreatedAt timestamp seems too old: %v", data.CreatedAt)
		}

		t.Logf("✓ Job mapping includes timestamp: createdAt=%v", data.CreatedAt)
	})
}
