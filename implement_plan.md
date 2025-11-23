# Implement Plan (Collector Service)

Tổng hợp từ `architech+code_standard.md`, `design.md`, `plan.md`, `rabbitmq.md`. Dùng làm tài liệu triển khai collector service và chuẩn kiến trúc.

## Mục tiêu
- Trung tâm nhận và điều phối job crawl qua RabbitMQ, tách producer ↔ crawler.
- Chuẩn hóa payload (`CrawlRequest` → `CollectorTask`), quản lý retry, quan sát được lag/lỗi.
- Giữ Clean Architecture (delivery → usecase → repository), dễ mở rộng thêm crawler/platform.

## Kiến trúc & Cấu trúc thư mục
- **Clean architecture template** (tham chiếu thư mục `example/`): `uc_types.go`, `uc_interface.go`, `uc_errors.go`, `usecase/`, `repository/`, `delivery/` (http/job/rabbitmq). Sentinel error, mapper/presenter tách riêng, không IO trong usecase.
- **Root layout hiện có**
  - `cmd/api`, `cmd/consumer`: entrypoint cho HTTP API và consumer worker.
  - `config/`: load config (HTTP, Mongo, JWT, RabbitMQ, Discord...).
  - `internal/`: business/service code.
    - `collector/`: logic collector (kế hoạch mới) gồm dispatcher + result/retry.
    - `consumer/`, `httpserver/`, `middleware/`, `models/`, `appconfig/`: server setup & domain models hiện có.
  - `pkg/`: shared lib (rabbitmq client, mongo, log, errors, paginator, discord, otp, util...).
  - `deployment/`, `docs/`, `docs2/`: tài liệu, cấu hình triển khai.
  - `example/`: mẫu module chuẩn (usecase/repo/delivery) dùng làm khuôn tạo module mới.
- **Collector module đề xuất** (từ `design.md`):
  - `internal/dispatcher/`: nhận `CrawlRequest`, map Strategy platform → queue, publish task.
    - `usecase/dispatch_uc.go`, `delivery/rabbitmq/inbound_consumer.go`, `task_producer.go`.
  - `internal/collector/`: nhận `CrawlerResult`, quyết định retry, republish.
    - `usecase/result_uc.go`, `retry_policy.go`, `delivery/rabbitmq/result_consumer.go`, `retry_producer.go`.
  - Repository chưa cần (worker tự ghi DB); có thể thêm `jobs` collection sau.

## Luồng hệ thống end-to-end
- **Ingress**: Upstream service publish `CrawlRequest` vào exchange `collector.inbound` (topic). Collector inbound consumer validate, set default `attempt=1`, `max_attempts` từ config.
- **Dispatch**: Strategy map `platform/task_type` → queue/routing; publish `CollectorTask` tới từng platform queue (`crawler.youtube.queue`, `crawler.tiktok.queue`, `crawler.facebook.queue`, ...). Headers chuẩn: `content_type=application/json`, `content_encoding=utf-8`, `delivery_mode=2`, `x-schema-version=1`, optional `x-trace-id`.
- **Worker xử lý**: Crawler worker nhận task, crawl, tự ghi DB (giai đoạn đầu) và luôn `ack` (không tự retry). Khi lỗi, vẫn `ack` và publish kết quả.
- **Fan-in kết quả**: Worker publish `CrawlerResult` vào exchange `collector.results` (topic/fanout), routing `crawler.<platform>.result`. Collector result consumer parse, kiểm tra idempotency (`job_id + cursor/bookmark`).
- **Retry**: Nếu `status=failed`, `error_code` thuộc retryable và `attempt < max_attempts` → republish task với `attempt++`, set header `retry=true` (có thể delay/backoff qua delay queue). Lỗi business hoặc hết lượt → đánh dấu failed/partial, optional DLQ/Audit.
- **Quan sát**: Log/metric tại boundary (consume/publish/usecase/repo), đếm retry/success/fail per platform, đo lag/latency. Không log PII.
- **Transition storage** (kế hoạch): Phase 0-3 từ crawler tự ghi DB → collector làm primary writer → Redis stream + data-writer service; giữ feature flag rollback.

## Hợp đồng dữ liệu / schema
- `CrawlRequest` (inbound): `job_id`, `task_type (research_keyword|crawl_links|research_and_crawl)`, `platform`, `payload`, `time_range`, `attempt`, `max_attempts`, `emitted_at`. Collector chuẩn hóa thành `CollectorTask` (thêm default, idempotency token, writer target).
- `CollectorTask` (class cha): field chung `job_id`, `platform`, `task_type`, `payload` (map), `time_range`, `attempt`, `max_attempts`, `trace_ctx`, `emitted_at`, `retry` flag, `schema_version`, routing info. Collector giữ loại platform và map payload sang struct con theo worker.
  - **YouTube payload** (từ `rabbitmq.md`):  
    - `research_keyword`: `keyword`, `limit`, `sort_by`, `time_range`.  
    - `crawl_links`: `video_urls[]`, `include_channel`, `include_comments`, `max_comments`, `download_media`, `media_type`, `time_range`.  
    - `research_and_crawl`: `keywords[]`, `limit_per_keyword`, `include_comments`, `include_channel`, `max_comments`, `download_media`, `time_range`.
  - **TikTok payload** (từ `rabbitmq.md`):  
    - `research_keyword`: `keyword`, `limit`, `sort_by`, `time_range`.  
    - `crawl_links`: `video_urls[]`, `include_comments`, `include_creator`, `max_comments`, `download_media`, `media_type`, `media_save_dir`, `time_range`.  
    - `research_and_crawl`: `keywords[]`, `limit_per_keyword`, `sort_by`, `include_comments`, `include_creator`, `max_comments`, `download_media`, `media_type`, `media_save_dir`, `time_range`.
- `CrawlerResult` (worker → collector): `job_id`, `platform`, `task_type`, `status (success|skipped|failed)`, `payload_chunk/data`, `cursor/bookmark`, `metrics` (docs/bytes/duration), `errors` (machine codes), `attempt`, `emitted_at`. Idempotent và có thứ tự theo job.

## RabbitMQ topology & config
- **Collector exchanges/queues**:
  - Ingress exchange `collector.inbound` (topic, durable); routing `crawler.<platform>.<task_type>` hỗ trợ wildcard.
  - Per-platform queues: `crawler.<platform>.queue` (durable) nhận task đã map payload theo platform.
  - Result exchange `collector.results` (topic/fanout, durable); routing `crawler.<platform>.result`; consumer fan-in ở collector.
- **Legacy/worker-specific (từ `rabbitmq.md`)**:
  - YouTube: exchange `youtube_exchange` (direct), routing `youtube.crawl`, queue `youtube_crawl_queue`.
  - TikTok: exchange `tiktok_exchange` (direct), routing `tiktok.crawl`, queue `tiktok_crawl_queue`.
  - Mặc định URL: `amqp://guest:guest@localhost:5672/`, vhost `/`; override qua env `RABBITMQ_*` từng worker (`host/port/user/password/vhost/exchange/routing_key/queue/prefetch`).
- **Message format khi publish (producer guide)**:
  - Body JSON UTF-8, `delivery_mode=2` (persistent).
  - Ví dụ: `{"task_type":"crawl_links","payload":{...},"job_id":"job-yt-001"}`; `task_type` bắt buộc, `job_id` tùy chọn (worker sinh nếu trống).
  - `payload.time_range` (ngày) áp dụng chung để lọc data theo khoảng thời gian.
- **Ack/Nack policy**:
  - Worker: luôn `ack` cả khi lỗi; không tự `nack`/requeue.
  - Collector consumer: parse lỗi → `ack` bỏ (hoặc `nack` không requeue tùy chính sách); lỗi tạm từ usecase → `nack` + requeue/delay theo retry policy; lỗi business → `ack`, log warn.

## Quy ước code & triển khai
- Usecase public: `func (uc implUseCase) Method(ctx context.Context, in Input)`.
- Không goroutine ở delivery; goroutine ở usecase với errgroup, có timeout.
- Dependency inject qua `New`, không dùng global/singleton.
- Option struct per hành động repo, không nhúng business rule vào query/build.
- Sentinel errors: `ErrInvalidInput`, `ErrNotFound`, `ErrDuplicate`, `ErrTemporary`, `ErrPermission`.
- Testing: table-driven cho usecase, mock repo/producer; handler test validate mapping & code trả về; repo integration test với DB/queue thật khi cần.

## Việc cần làm tiếp theo (đề xuất)
- Tạo module `internal/dispatcher` & `internal/collector` theo template `example/`.
- Khai báo schema `CrawlRequest`, `CollectorTask`, `CrawlerResult` + strategy map config (platform → queue/routing/backoff).
- Implement consumer/producer RabbitMQ adapter theo pattern `delivery/rabbitmq` (common.go, consumer.go, workers.go, presenter/constants).
- Thêm metrics/logging hook ở ingress/egress; optional delay queue cho backoff.
