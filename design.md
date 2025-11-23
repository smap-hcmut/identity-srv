# Collector Service Design Notes

Tóm tắt các quyết định kiến trúc và contract cho collector service (worker tự
ghi DB, collector chỉ điều phối + retry).

## Message Topology
- Inbound exchange: `collector.inbound` (topic, durable); routing
  `crawler.<platform>.<task_type>` có thể dùng wildcard. Collector fan-out tới
  tất cả platform qua Strategy map.
- Crawler queues: `crawler.<platform>.queue` (durable, per platform).
- Result exchange: `collector.results` (topic/fanout, durable); routing
  `crawler.<platform>.result`. Collector result consumer nghe và quyết định
  retry/hoàn thành.
- Headers chung khi publish: `content_type=application/json`,
  `content_encoding=utf-8`, `delivery_mode=2`, `x-schema-version=1`,
  optional `x-trace-id`.

## CrawlRequest (inbound)
Collector nhận JSON tối giản, sau đó map payload theo từng worker.
```json
{
  "job_id": "uuid-or-string",
  "task_type": "research_keyword | crawl_links | research_and_crawl",
  "payload": {},           // dữ liệu gốc, sẽ map theo platform
  "time_range": 7,         // ngày, optional
  "attempt": 1,            // luôn 1 khi inbound
  "max_attempts": 3,       // collector dùng để retry
  "emitted_at": "2025-01-01T12:00:00Z"
}
```
- `job_id`: track/idempotency cross-platform.
- `task_type`: chỉ 3 loại chuẩn.
- `payload`: collector biến tấu sang payload worker-specific.
- `time_range`: giới hạn thời gian, pass-through nếu worker hỗ trợ.
- `attempt/max_attempts`: collector kiểm soát retry.
- `emitted_at`: timestamp gốc để đo trễ.

## Dispatcher Module (ingress)
- Validate `CrawlRequest`, set default attempt=1, max_attempts từ config nếu
  trống.
- Strategy map platform → queue/routing; publish task tới mọi platform (fan-out
  logic trong collector).
- Producer set header/version/content_type/delivery_mode persistent.
- Không dùng repo (worker tự lưu DB).

## Result Module (fan-in + retry)
- Consume `collector.results`, parse `CrawlerResult` (`job_id`, `task_type`,
  `status`, `cursor/bookmark`, `metrics`, `error_code`, `attempt`,
  `emitted_at`).
- Idempotency key: `job_id + cursor` để tránh nhân bản khi retry.
- Retry policy: nếu `status=failed` và `error_code` thuộc nhóm retryable và
  `attempt < max_attempts` → republish task với `attempt++` (header
  `retry=true`, optional delay/backoff). Worker luôn ack, không tự retry.
- Khi hết lượt hoặc lỗi business: mark failed/partial (in-memory/log/metrics),
  optional DLQ/Audit.

## Retry & Backoff
- Collector thực hiện retry, queue chính không auto requeue/TTL.
- Có thể dùng delay queue để backoff (TTL + DLX về queue chính); config
  backoff per platform.

## Worker Expectations
- Worker nhận task, xử lý, tự lưu DB/upsert, luôn `ack` kể cả lỗi.
- Khi lỗi, publish `CrawlerResult` với `status=failed`, `error_code`,
  `attempt` hiện tại để collector quyết định retry.
- Idempotent theo `job_id + cursor`.

## Config Needed
- RabbitMQ: URL, vhost, inbound exchange, per-platform queue/exchange,
  result exchange, schema_version, delivery_mode, content_type.
- Retry: max_attempts default, backoff, retryable_error_codes.

## Folder Layout (collector)
- `cmd/consumer/main.go` – entry khởi động consumers/producers.
- `internal/dispatcher/` (nhận từ src khác)
  - `uc_types.go`, `uc_errors.go`, `uc_interface.go`
  - `usecase/dispatch_uc.go` – fan-out task tới từng platform, định nghĩa `CrawlRequest`, `CollectorTask`.
  - `delivery/rabbitmq/inbound_consumer.go` – consume `collector.inbound`.
  - `delivery/rabbitmq/task_producer.go` – publish tới `crawler.<platform>.queue` (map payload theo platform).
- `internal/collector/` (nhận từ worker)
  - `uc_types.go`, `uc_errors.go`, `uc_interface.go`
  - `usecase/result_uc.go`, `retry_policy.go` – định nghĩa `CrawlerResult`, xử lý kết quả, quyết định retry.
  - `delivery/rabbitmq/result_consumer.go` – consume `collector.results`.
  - `delivery/rabbitmq/retry_producer.go` – republish task khi retry.
- (Hiện chưa cần repo vì worker tự lưu DB; bổ sung sau nếu collector ghi DB).
