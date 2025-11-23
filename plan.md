# Collector Service Blueprint

This document outlines the plan for evolving the scraping platform into a
centralized collector architecture. The collector service becomes the broker
between domain services that request crawls and the specialized crawler
workers (Facebook, TikTok, YouTube, …). No code exists yet; this is the design
reference before implementation starts.

## Goals
- Accept crawl jobs from upstream business services through a single message
  entry point.
- Route jobs to the correct crawler workers without tight coupling.
- Aggregate crawler results and persist them through _one_ writer component
  (MongoDB now, Redis pub/sub + writer service later).
- Allow new crawler workers to be added via configuration, not code rewrites.
- Provide observability and backpressure controls across all crawling flows.

## Current Situation
- Each crawler service exposes its own queue/topic, and producers publish to
  them directly.
- Crawlers connect to MongoDB independently and perform inserts/upserts.
- There is no global throttling or lifecycle tracking for jobs that span
  multiple crawlers.

## Target Architecture

```
Producer Services
      │
      ▼
  collector.inbound (topic exchange)
      │
      ▼
+--------------------+
|  Collector Service |
+--------------------+
  │ route task via Strategy map (crawler_dispatcher)
  │
  ├─► crawler.youtube.queue
  ├─► crawler.tiktok.queue
  └─► crawler.facebook.queue

Crawler Workers ──► collector.results (fanout)
                       │
                       ▼
               Result Aggregator
                       │
               Mongo Writer / Redis Stream
```

### Message Topology
1. **Ingress**: Upstream service publishes a `CrawlRequest` to
   `collector.inbound` with routing key `crawler.<platform>`.
2. **Dispatch**: Collector normalizes the payload into a canonical
   `CollectorTask`, stores a `jobs` record, then publishes to the per-platform
   queue (`crawler.youtube.queue`, etc.).
3. **Result Fan-in**: Each crawler publishes `CrawlerResult` messages to
   `collector.results`. Collector correlates them with the originating job,
   runs validations/transforms, and writes via the storage adapter.
4. **Future Storage**: Once Redis pub/sub is introduced, collector publishes the
   consolidated payload to a Redis stream where a dedicated `data-writer`
   service persists to MongoDB/other stores.

### Core Components
- `CollectorService` (Mediator) – orchestrates the full job lifecycle.
- `TaskRouter` (Strategy) – maps `platform` + `task_type` to dispatcher configs.
- `JobRepository` – persists job metadata, attempts, status transitions.
- `CrawlerDispatcher` – abstract interface per crawler specifying queue names,
  payload templates, and timeout/retry policies.
- `ResultAggregator` – correlates crawler results, runs smart upsert logic, and
  forwards to the writer adapter.
- `WriterAdapter` – current Mongo implementation, future Redis publisher.
- `Monitoring` – metrics + logging hooks for queue depth, lag, failures.

### Interfaces & Contracts
- `CrawlRequest` – external payload (`job_id`, `platform`, `task_type`,
  `payload`, `time_range`, metadata).
- `CollectorTask` – normalized internal data structure with derived defaults and
  tracing context.
- `CrawlerResult` – result contract emitted by crawler workers containing
  status (`success`, `skipped`, `failed`), payload chunk, metrics, and errors.
- `JobStatus` – transitions: `pending → dispatched → processing → completed |
  failed | timed_out | partial`.

## Retry & Failure Handling (Collector-Driven)
- Workers **always `ack`** messages (kể cả lỗi); không tự retry/requeue để
  tránh nhân bản. Worker vẫn update `craw_job` nội bộ và publish
  `CrawlerResult` với `status=failed`, `error_code`, `attempt`.
- Collector giữ trạng thái `attempt`/`max_attempts` trong `CollectorTask` và
  quyết định retry: nếu lỗi tạm (`error_code` thuộc nhóm retryable) và chưa
  quá `max_attempts`, collector **republish** task với `attempt++` và header
  `retry=true`.
- Queue chính không bật TTL/requeue tự động; retry được thực hiện bởi collector
  (có thể dùng delay queue tùy config để backoff).
- Idempotency: payload mang `job_id` + `cursor/bookmark` để worker xử lý lại
  không ghi đè sai; storage upsert theo cặp này.
- Khi hết lượt hoặc lỗi business: collector đánh dấu job `failed` (hoặc
  `partial`), optional publish vào DLQ/Audit stream để quan sát.
- Collector metrics: đếm retry/success/fail per platform, theo dõi lag và
  tỷ lệ lỗi để điều chỉnh `max_attempts`/backoff.

## Design Patterns Used
- **Strategy** – per-platform dispatcher configuration and result handlers.
- **Mediator** – collector coordinates producers, crawlers, and writers.
- **Template Method** – result validation pipeline (normalize, dedupe, enrich).
- **CQRS** – collector handles write path; future read models can be derived
  separately.
- **Event-Driven** – all communication through message brokers (RabbitMQ now,
  Redis later).

## Implementation Phases
1. **Foundations**
   - Define domain models, interfaces, and schemas.
   - Create collector service skeleton with message consumers/producers.
   - Introduce `jobs` collection for lifecycle tracking.
2. **Crawler Integration**
   - Update Facebook/TikTok/YouTube workers to consume from the new queues and
     publish standardized `CrawlerResult` messages.
   - Disable direct Mongo writes inside crawlers.
3. **Observability**
   - Add metrics (queue lag, job duration, failure rate) and dashboards.
   - Add dead-letter queues with alerting hooks.
4. **Redis Transition**
   - Introduce Redis stream for results and a standalone `data-writer`
     microservice that performs Mongo upserts.
   - Gradually move Mongo connectivity out of the collector.
5. **Extensibility**
   - Document the steps to register a new crawler via configuration (platform
     identifier, queue bindings, payload schema, result handler plugin).

## Transitional Plan: Crawler DB Writes → Collector-Managed Storage
- **Phase 0: Current** – Crawlers keep writing directly to Mongo to avoid
  production disruption; start emitting `CrawlerResult` events (idempotent) to
  `collector.results` so the collector can observe/validate without persisting.
- **Phase 1: Dual-Publish** – Crawlers write to Mongo and also send full
  payloads via `CrawlerResult`. Collector replays into a shadow writer to prove
  parity (checksum/count dashboards).
- **Phase 2: Collector Primary** – Feature flag in each crawler disables direct
  Mongo writes; collector acks only after writer succeeds. Keep a rollback flag
  to re-enable crawler writes if needed.
- **Phase 3: Redis/Data-Writer** – Collector switches to publishing to Redis
  stream; dedicated `data-writer` owns persistence. Crawlers no longer have DB
  drivers bundled.

## Extensibility Playbook (Add a New Crawler)
- Register `platform` in `TaskRouter` config: routing key, queue name,
  retry/timeout, concurrency limits.
- Define payload schema (per task type) and validation in `CrawlerDispatcher`.
- Implement result handler plugin if the crawler emits platform-specific fields.
- Add dashboards/alerts: queue depth, success/fail rate, lag per platform.
- Update IaC/broker bindings for the new queue + DLQ; run e2e canary job.

## Data/Message Contracts (Worker ↔ Collector)
- `CrawlRequest` includes `job_id`, `platform`, `task_type`, `payload`,
  `time_range`, `trace_ctx`, `priority`, `attempt`.
- `CollectorTask` adds derived defaults: max attempts, SLA deadline, idempotency
  token, and writer target (`mongo` now, `redis` later).
- `CrawlerResult` must carry `job_id`, `platform`, `task_type`, `status`,
  `payload_chunk` (data), `cursor/bookmark`, `metrics` (docs, bytes, duration),
  and `errors` (machine-readable codes). Results are idempotent and ordered per
  job.
- Storage adapter requirements: upsert semantics, dedupe on `job_id` + `cursor`,
  dead-letter on parse/validation failures with alert hooks.

## Open Questions
- Auth & security model between services (mTLS, token-based?).
-.Partial result handling when multiple crawlers contribute to one job.
- Rollback/compensation strategy if the writer fails after collectors ack the
  crawler result.
- How to version schemas so legacy crawlers can coexist during migration.

This plan should be reviewed and refined with stakeholders before coding. Once
approved, tasks for each implementation phase can be created in the backlog.
