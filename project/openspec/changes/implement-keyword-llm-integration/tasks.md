# Implementation Tasks

## Phase 1: Infrastructure Setup

### 1.1 LLM Service Package
- [ ] 1.1.1 Create `pkg/llm/` directory structure
- [ ] 1.1.2 Create `pkg/llm/interface.go` with `Provider` interface
  - [ ] Define `SuggestKeywords(ctx, brandName) ([]string, []string, error)`
  - [ ] Define `CheckAmbiguity(ctx, keyword) (bool, string, error)`
- [ ] 1.1.3 Create `pkg/llm/errors.go` with error types
  - [ ] `ErrLLMUnavailable`
  - [ ] `ErrLLMTimeout`
  - [ ] `ErrLLMInvalidResponse`
  - [ ] `ErrLLMInvalidAPIKey`
- [ ] 1.1.4 Create `pkg/llm/gemini.go` with Gemini implementation
  - [ ] Implement HTTP client for Gemini API
  - [ ] Implement `SuggestKeywords` with proper prompt
  - [ ] Implement `CheckAmbiguity` with proper prompt
  - [ ] Add retry logic with exponential backoff
  - [ ] Add timeout handling
  - [ ] Parse JSON responses safely
- [ ] 1.1.5 Create `pkg/llm/new.go` with constructor
  - [ ] Accept logger and config
  - [ ] Initialize HTTP client with timeout
  - [ ] Return appropriate provider based on config
- [ ] 1.1.6 Write unit tests for Gemini provider
  - [ ] Test successful suggestions
  - [ ] Test successful ambiguity check
  - [ ] Test error handling (timeout, invalid response, etc.)
  - [ ] Test retry logic

### 1.2 Collector Service Package
- [ ] 1.2.1 Create `pkg/collector/` directory structure
- [ ] 1.2.2 Create `pkg/collector/types.go` with data structures
  - [ ] `Post` struct with fields:
    - [ ] `ID` (string): Unique post identifier
    - [ ] `Content` (string): Post text content
    - [ ] `Source` (string): Platform name (facebook, twitter, reddit, etc.)
    - [ ] `SourceID` (string): Original post ID on platform
    - [ ] `Author` (string): Author name/username
    - [ ] `AuthorID` (string): Author identifier on platform
    - [ ] `Date` (time.Time): Post creation date
    - [ ] `URL` (string): Direct link to original post
    - [ ] `Engagement` (Engagement struct): Engagement metrics
    - [ ] `Metadata` (Metadata struct): Additional metadata
  - [ ] `Engagement` struct (likes, comments, shares)
  - [ ] `Metadata` struct (language, sentiment)
  - [ ] `DryRunRequest` struct (keywords, limit)
  - [ ] `DryRunResponse` struct (posts array, total_found, limit)
- [ ] 1.2.3 Create `pkg/collector/client.go` with `Client` interface
  - [ ] Define `DryRun(ctx, keywords, limit) ([]Post, error)`
- [ ] 1.2.4 Create `pkg/collector/errors.go` with error types
  - [ ] `ErrCollectorUnavailable`
  - [ ] `ErrCollectorTimeout`
  - [ ] `ErrCollectorInvalidResponse`
- [ ] 1.2.5 Create `pkg/collector/http_client.go` with HTTP implementation
  - [ ] Implement HTTP client with timeout
  - [ ] Implement `DryRun` method
  - [ ] Construct request to `{baseURL}/api/v1/collector/dry-run`
  - [ ] Parse JSON response safely with all Post fields
  - [ ] Handle HTTP errors (4xx, 5xx)
  - [ ] Support default mock URL `http://localhost:8081` when URL not configured
- [ ] 1.2.6 Create `pkg/collector/new.go` with constructor
  - [ ] Accept logger and config
  - [ ] Initialize HTTP client with timeout
  - [ ] Return HTTP client implementation
- [ ] 1.2.7 Write unit tests for Collector client
  - [ ] Test successful dry run
  - [ ] Test error handling (timeout, 4xx, 5xx)
  - [ ] Test response parsing

### 1.3 Configuration Updates
- [ ] 1.3.1 Update `config/config.go`
  - [ ] Add `LLMConfig` struct with fields:
    - [ ] `Provider` (string, default: "gemini")
    - [ ] `APIKey` (string, required)
    - [ ] `Model` (string, default: "gemini-1.5-flash")
    - [ ] `Timeout` (int, default: 30)
    - [ ] `MaxRetries` (int, default: 3)
  - [ ] Add `CollectorConfig` struct with fields:
    - [ ] `BaseURL` (string, default: "http://localhost:8081" for development)
    - [ ] `Timeout` (int, default: 30)
  - [ ] Add `LLM` field to `Config` struct
  - [ ] Add `Collector` field to `Config` struct
- [ ] 1.3.2 Update `template.env` with new environment variables
  - [ ] `LLM_PROVIDER=gemini`
  - [ ] `LLM_API_KEY=`
  - [ ] `LLM_MODEL=gemini-1.5-flash`
  - [ ] `LLM_TIMEOUT=30`
  - [ ] `LLM_MAX_RETRIES=3`
  - [ ] `COLLECTOR_SERVICE_URL=http://localhost:8081` (default mock URL for development)
  - [ ] `COLLECTOR_TIMEOUT=30`
  - [ ] Add comment: "For production, set COLLECTOR_SERVICE_URL to actual Collector Service URL"
- [ ] 1.3.3 Update `k8s/secret.yaml.template` with new secret keys (if applicable)

## Phase 2: Keyword Usecase Integration

### 2.1 Update Keyword Usecase Structure
- [ ] 2.1.1 Update `internal/keyword/usecase/new.go`
  - [ ] Add `llmProvider` field to `usecase` struct
  - [ ] Add `collectorClient` field to `usecase` struct
  - [ ] Update constructor to accept LLM provider and Collector client
  - [ ] Store dependencies in struct
- [ ] 2.1.2 Update `internal/keyword/usecase/keyword.go` (if needed for interface changes)

### 2.2 Implement Real AI Suggestion
- [ ] 2.2.1 Update `internal/keyword/usecase/suggester.go`
  - [ ] Replace hardcoded logic in `suggestProcessing` with LLM call
  - [ ] Implement fallback to basic suggestions if LLM fails
  - [ ] Post-validate suggestions using existing `validate` method
  - [ ] Add proper error handling and logging
  - [ ] Ensure suggestions are deduplicated
- [ ] 2.2.2 Add `fallbackSuggestions` helper method
  - [ ] Generate basic suggestions (current hardcoded logic)
  - [ ] Return niche and negative keywords
- [ ] 2.2.3 Write unit tests for suggester
  - [ ] Test successful LLM suggestion
  - [ ] Test fallback when LLM fails
  - [ ] Test post-validation of suggestions
  - [ ] Test with mock LLM provider

### 2.3 Enhance Semantic Validation
- [ ] 2.3.1 Update `internal/keyword/usecase/validator.go`
  - [ ] Add `isSingleWord` helper function
  - [ ] Update `validateOne` to call LLM for ambiguity check on single words
  - [ ] Implement warning system (log warning, don't reject)
  - [ ] Handle LLM errors gracefully (continue without LLM check)
  - [ ] Add context to ambiguity warnings
- [ ] 2.3.2 Write unit tests for validator
  - [ ] Test ambiguity detection for single words
  - [ ] Test that multi-word keywords skip LLM check
  - [ ] Test graceful degradation when LLM fails
  - [ ] Test with mock LLM provider

### 2.4 Implement Real Dry Run
- [ ] 2.4.1 Update `internal/keyword/usecase/tester.go`
  - [ ] Replace mock data with Collector Service call
  - [ ] Validate keywords before calling Collector
  - [ ] Call Collector Service with validated keywords and limit (10)
  - [ ] Convert Collector `Post` structs to `interface{}` for response
  - [ ] Add proper error handling and logging
- [ ] 2.4.2 Write unit tests for tester
  - [ ] Test successful dry run with Collector
  - [ ] Test error handling when Collector fails
  - [ ] Test keyword validation before dry run
  - [ ] Test with mock Collector client

## Phase 3: Dependency Injection & Initialization

### 3.1 Update HTTP Server Handler
- [ ] 3.1.1 Update `internal/httpserver/handler.go`
  - [ ] Initialize LLM provider using config
  - [ ] Initialize Collector client using config
  - [ ] Pass LLM provider to keyword usecase constructor
  - [ ] Pass Collector client to keyword usecase constructor
  - [ ] Handle initialization errors gracefully
- [ ] 3.1.2 Update `internal/httpserver/new.go` (if needed)
  - [ ] Add LLM config to `Config` struct
  - [ ] Add Collector config to `Config` struct
  - [ ] Validate new configs in `validate()` method

### 3.2 Update Main Application
- [ ] 3.2.1 Update `cmd/api/main.go`
  - [ ] Load LLM config from environment
  - [ ] Load Collector config from environment
  - [ ] Pass configs to HTTP server initialization
  - [ ] Handle missing required configs (log error, fail fast)

## Phase 4: Error Handling & Resilience

### 4.1 Error Type Definitions
- [ ] 4.1.1 Ensure all error types are properly defined
  - [ ] LLM errors in `pkg/llm/errors.go`
  - [ ] Collector errors in `pkg/collector/errors.go`
  - [ ] Keyword errors in `internal/keyword/error.go` (if needed)

### 4.2 Error Mapping
- [ ] 4.2.1 Update `internal/project/delivery/http/error.go` (if exists)
  - [ ] Map LLM errors to HTTP errors
  - [ ] Map Collector errors to HTTP errors
  - [ ] Ensure user-friendly error messages

### 4.3 Logging
- [ ] 4.3.1 Add structured logging for all external service calls
  - [ ] Log LLM API calls (request, response, latency)
  - [ ] Log Collector API calls (request, response, latency)
  - [ ] Log errors with full context
  - [ ] Log fallback activations

## Phase 5: Testing

### 5.1 Unit Tests
- [ ] 5.1.1 Test LLM provider with mock HTTP client
- [ ] 5.1.2 Test Collector client with mock HTTP client
- [ ] 5.1.3 Test keyword usecase with mock LLM and Collector
- [ ] 5.1.4 Test error handling paths
- [ ] 5.1.5 Test fallback mechanisms

### 5.2 Integration Tests (Optional)
- [ ] 5.2.1 Test with real Gemini API (staging environment)
- [ ] 5.2.2 Test with real Collector Service (staging environment)
- [ ] 5.2.3 Test full flow from API endpoint to response

### 5.3 Manual Testing
- [ ] 5.3.1 Test suggestion endpoint with various brand names
- [ ] 5.3.2 Test validation with ambiguous keywords
- [ ] 5.3.3 Test dry run with various keyword combinations
- [ ] 5.3.4 Test error scenarios (invalid API key, service down, etc.)

## Phase 6: Documentation & Cleanup

### 6.1 Code Documentation
- [ ] 6.1.1 Add Go doc comments to all exported functions
- [ ] 6.1.2 Add inline comments for complex logic
- [ ] 6.1.3 Document LLM prompt design decisions

### 6.2 Configuration Documentation
- [ ] 6.2.1 Update `README.md` with new environment variables
- [ ] 6.2.2 Document LLM provider setup instructions
- [ ] 6.2.3 Document Collector Service URL configuration

### 6.3 API Documentation
- [ ] 6.3.1 Update Swagger docs if API contracts changed (shouldn't)
- [ ] 6.3.2 Verify API examples in documentation

### 6.4 Code Review Checklist
- [ ] 6.4.1 All tests passing
- [ ] 6.4.2 No linter errors
- [ ] 6.4.3 Error handling comprehensive
- [ ] 6.4.4 Logging adequate
- [ ] 6.4.5 Code follows project conventions
- [ ] 6.4.6 No hardcoded values (use config)
- [ ] 6.4.7 All TODOs addressed or documented

## Validation Checklist

Before considering this change complete:
- [ ] All tasks in Phase 1-6 are checked off
- [ ] Unit tests have >80% coverage for new code
- [ ] Integration tests pass (if implemented)
- [ ] Manual testing completed successfully
- [ ] Code review approved
- [ ] Documentation updated
- [ ] Configuration templates updated
- [ ] No breaking changes to existing APIs
- [ ] Error handling tested for all failure scenarios
- [ ] Fallback mechanisms verified
