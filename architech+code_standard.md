# Example Module Coding Standard

Tài liệu này mô tả chuẩn triển khai cho một module theo cấu trúc tại
`example/`. Dùng làm khuôn mẫu khi tạo module mới trong dự án.

## Mục tiêu
- Giữ Clean Architecture: delivery → usecase → repository, không lệch tầng.
- Dễ mở rộng: thêm repository/delivery mới mà không đổi business logic.
- Testable: usecase thuần, mock được dependency.
- Minh bạch: log/metrics đầy đủ, lỗi có mã rõ ràng.

## Cấu trúc thư mục
```
example/
  uc_types.go           # Domain types/options dùng chung cho usecase
  uc_interface.go       # Interface public của module
  uc_errors.go          # Sentinel errors + helper mapping
  mock_UseCase.go       # Mock usecase cho test delivery

  usecase/              # Business logic (no IO)
    new.go              # NewUseCase + DI deps
    job_uc.go           # Ví dụ luồng job; đổi tên theo domain chính
    core_uc.go          # Luồng domain tổng quát (đổi từ event.go để tránh nhầm)
    consumer.go         # Luồng xử lý message đầu vào (gọi usecase)
    publisher.go        # Hàm publish event/message
    utils.go            # Helper chung (pure)
    core_utils.go       # Helper domain (đổi từ event_utils.go để tránh nhầm)
    consumer_utils.go   # Helper consumer
    util_types.go       # Types phụ trợ
    constants.go        # Hằng số domain
    uc_test.go          # Unit test usecase (table-driven)

  repository/           # Interface + option/pipeline + impl
    repo_interface.go   # Interface repo + Scope
    <domain>_options.go # Options truy vấn (ví dụ hiện tại: event_options.go)
    <entity>_options.go # Options khác (vd: recurring_instance_options.go)
    mock_Repository.go  # Mock repo cho usecase test
    mongo/              # Triển khai Mongo
      new.go            # Khởi tạo repo Mongo + DI
      build.go          # Build model/update doc chung
      query.go          # Build filter/pipeline chung
      <domain>_repo.go  # CRUD domain (vd: event.go)
      <entity>_repo.go  # Các entity khác (vd: recurring_instance.go, ...)
      <entity>_build.go # Build model/update cho entity (vd: recurring_instance_build.go)
      <entity>_query.go # Query/pipeline cho entity (vd: recurring_instance_query.go)
      repository_test.go

  delivery/             # Adapter vào/ra
    http/               # HTTP layer
      new.go            # Wire HTTP handler với usecase
      routes.go         # Đăng ký route
      handlers.go       # Handler HTTP chính
      process_request.go# Parse/validate request
      presenters.go     # Map domain → response DTO
      errors.go         # Mapping lỗi → HTTP code
      handler_test.go   # Test HTTP handler
    job/                # Cron/worker trigger
      new.go            # Khởi tạo job runner
      handlers.go       # Job logic (gọi usecase)
      register.go       # Đăng ký job/schedule
    rabbitmq/           # MQ adapters (consumer/producer)
      consumer.go       # Consumer tổng hoặc <entity>_consumer.go
      producer.go       # Helper publish
      presenters.go     # Map domain ↔ MQ message
      constants.go      # Hằng số routing/key
```

## Luồng xử lý chuẩn
1. **Delivery**: Validate cứng đầu vào (required/range/type) → map sang
   `Scope` + input usecase; không nhúng business rule.
2. **Usecase**: Điều phối repo/service khác, kiểm soát timeout, retry, song
   song (errgroup); trả lỗi sentinel cho business, wrap lỗi infra.
3. **Repository**: Thực hiện IO, không nhúng rule; phân loại lỗi rõ (not found,
   duplicate, temporary).
4. **Delivery**: Map lỗi usecase → HTTP/MQ status; log đủ ngữ cảnh; không lộ
   stacktrace ra client.

## Quy ước code
- Usecase public: `func (uc implUseCase) Method(ctx context.Context, in Input)`.
- Dùng struct input thay cho nhiều primitive; method không nhận context nil.
- Logger tiêm qua struct; không lấy từ context.
- Goroutine chỉ tạo trong usecase; delivery không spawn goroutine.
- Không dùng biến global/ singleton không cần thiết; inject deps qua `New`.
- Tên method/action rõ: `CreateX`, `ListX`, `CheckX`; tránh viết tắt khó hiểu.

## Xử lý lỗi
- Sentinel ở `uc_errors.go`: `ErrInvalidInput`, `ErrNotFound`, `ErrDuplicate`,
  `ErrTemporary`, `ErrPermission`.
- Wrap lỗi infra với ngữ cảnh: `fmt.Errorf("repo.GetUser: %w", err)`.
- Delivery chỉ surface lỗi business; lỗi infra trả mã 500/ retry và log nội bộ.
- MQ/cron: `ErrTemporary` → nack + requeue; lỗi business → ack + log cảnh báo.
- Không nuốt lỗi: luôn trả hoặc log ở edge; tránh panic (trừ khi truly fatal).

## Logging & Metrics
- Thêm `trace_id`/`request_id` từ context; log ở boundary (trước/sau repo,
  trước return lỗi).
- Level: Debug (payload nhỏ), Info (state change), Warn (retryable), Error
  (fail). Không log PII/mật khẩu/token.
- Metrics: counter success/fail, histogram latency cho usecase và call IO,
  gauge queue depth (nếu consumer).

## Repository
- Interface nhận `context.Context` đầu tiên; không dùng context.Background.
- Option struct riêng cho mỗi hành động; không reuse option sai mục đích.
- Không chứa rule business; chỉ mapping data và xử lý lỗi storage.
- Phân loại lỗi: not found, duplicate (unique), temporary (timeout/conn).
- Tên hàm repo: `Get`, `List`, `Create`, `Update`, `Delete`, `Upsert`.

## Testing
- Usecase: table-driven, mock repo/producer; assert call count/order khi cần.
- Repository: integration test với DB/queue thật hoặc test container; check
  unique/transaction cases.
- Delivery: handler test với mock usecase; kiểm tra validate + mã trả về +
  payload.
- Deterministic: inject clock/timeNow, tránh sleep; dùng context timeout nhỏ.

## Query/Build Files & Options
- Option: mỗi hành động một file (`*_options.go`), chỉ chứa filter cần thiết.
- Build files (`<entity>_build.go`, `build.go`): chuyển option → model/payload
  hoặc update doc; set default, convert ID/timezone; không chứa rule business.
- Query files (`<entity>_query.go`, `query.go`): build filter/pipeline (Mongo)
  hoặc SQL query; hàm nhỏ, test input → expected query map/SQL string.
- Pagination chuẩn: `Limit` (có max), `Offset/Skip`, `SortBy`, `SortDir`.
- Không embed rule business vào query/builder; chỉ mapping và lọc theo input.

### Template Option File (ví dụ)
```go
// repository/list_jobs_options.go
package repository

type ListJobsOptions struct {
	Scope     Scope
	Status    []JobStatus
	Keyword   string
	Limit     int64
	Offset    int64
	SortBy    string
	SortDir   int    // 1 asc, -1 desc
}
```

## Consumer / Producer Pattern (RabbitMQ)
- **Folder chuẩn**: `delivery/rabbitmq/consumer` (common.go, consumer.go,
  workers.go, new.go), `delivery/rabbitmq/producer` (common.go, producer.go,
  new.go, mock_Producer.go), `delivery/rabbitmq/presenters.go`, `constants.go`.
- **Không đặt adapter MQ trong usecase**; usecase chỉ gọi qua interface producer
  được inject.
- **Consumer**:
  - `consumer.go`: đăng ký consume các queue/exchange và gắn worker function.
  - `common.go`: khởi tạo channel, declare/bind exchange/queue, consume message,
    chạy worker loop; luôn có `catchPanic` để không chết goroutine.
  - `workers.go`: decode JSON → build `Scope` → gọi usecase; `Ack` sau khi xử
    lý; với lỗi parse thì `Ack` bỏ qua (hoặc `Nack` tuỳ chính sách).
  - Dùng `context.Background()` hiện có; khi thêm timeout nên bọc
    `context.WithTimeout`.
  - Logging: log khi start consume, khi parse lỗi, khi usecase trả lỗi; thêm
    metric nếu cần.
- **Producer**:
  - `producer/common.go`: Run mở channel, declare exchange (và queue nếu cần),
    giữ writer; Close đóng channel.
  - `producer/producer.go`: Hàm publish cụ thể (PushNoti, UpdateRequestEventID,
    UpdateTaskEventID); marshal JSON, gửi qua writer với content-type rõ.
  - `producer/new.go`: Interface `Producer`, hàm `New` khởi tạo với logger +
    connection.
  - Retry/backoff tùy broker; có thể mở publish confirm nếu yêu cầu tin cậy hơn.
- **Presenters & constants**: `delivery/rabbitmq/presenters.go` map domain ↔
  message struct; `constants.go` chứa exchange/queue/routing key.

### Template Consumer (bám theo structure hiện tại)
```go
// delivery/rabbitmq/consumer/workers.go
func (c Consumer) createSystemEventWorker(d amqp.Delivery) {
	ctx := context.Background() // optional: WithTimeout

	var msg CreateEventMsg
	if err := json.Unmarshal(d.Body, &msg); err != nil {
		c.l.Warnf(ctx, "consumer.Unmarshal: %v", err)
		d.Ack(false) // drop bad message
		return
	}

	if err := c.uc.CreateSystemEvent(ctx, msg.ToInput()); err != nil {
		c.l.Errorf(ctx, "consumer.CreateSystemEvent: %v", err)
		d.Ack(false) // or Nack(true) if retryable policy
		return
	}

	d.Ack(false)
}
```

### Template Producer (bám theo structure hiện tại)
```go
// delivery/rabbitmq/producer/producer.go
func (p implProducer) PublishUpdateTaskEventIDMsg(ctx context.Context, msg UpdateTaskEventIDMsg) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal UpdateTaskEventIDMsg: %w", err)
	}

	return p.updateTaskEventIDWriter.Publish(ctx, rabbitmq.PublishArgs{
		Exchange: UpdateTaskEventIDExchange.Name,
		Msg: rabbitmq.Publishing{
			Body:        body,
			ContentType: rabbitmq.ContentTypePlainText,
		},
	})
}
```

## File Naming Gợi ý
- Options: `repository/<action>_<entity>_options.go`
- Build (model/update): `repository/<entity>_build.go`
- Query/pipeline: `repository/<entity>_query.go` (Mongo/SQL)
- Consumer/producer: `delivery/rabbitmq/<entity>_consumer.go`,
  `delivery/rabbitmq/producer.go`
- Usecase: `<feature>_uc.go`, test: `<feature>_uc_test.go`

## Checklist theo tầng & file
### Root contracts
- `uc_types.go`: Khai báo domain struct/input/output; không import hạ tầng. Đặt
  default/validate nhẹ nếu cần (ví dụ method `Validate()`).
- `uc_errors.go`: Sentinel errors (`ErrInvalidInput`, `ErrNotFound`,
  `ErrTemporary`...), helper map lỗi → mã HTTP/MQ. Không wrap lỗi ở đây.
- `uc_interface.go`: Interface public của module; gom các phương thức usecase.
- `mock_UseCase.go`: Mock sinh sẵn/viết tay cho delivery test; giữ cập nhật khi
  đổi interface.

### usecase/
- `new.go`: Định nghĩa `implUseCase` với deps (repo, logger, clock, producers).
  `NewUseCase(deps ...) UseCase` phải validate deps != nil.
- `<feature>_uc.go`: Business flow chính (vd: `job_uc.go`); orchestration repo,
  publish event; không chứa DTO delivery/IO.
- `<feature>_utils.go`: Hàm util thuần (convert/merge) để test dễ; tránh dùng
  chữ helper mơ hồ.
- `consumer.go`/`publisher.go`: Luồng consumer/producer nội bộ usecase (gọi repo,
  phát sự kiện domain), không chứa code MQ cụ thể (adapter nằm ở delivery).
- `consumer_utils.go`/`util_types.go`/`constants.go`: Tiện ích cho consumer
  nội bộ (normalize input, loại trùng, hằng số domain).
- `uc_test.go` (hoặc `<feature>_uc_test.go`): Table-driven test, mock repo +
  producer, cover happy path + lỗi; dùng context timeout; kiểm tra call order
  nếu cần.

### repository/
- `repo_interface.go`: Khai báo interface repo + struct Scope. Mỗi method nhận
  `context.Context`, dùng option struct riêng.
- `<action>_<entity>_options.go`: Option/filter cho từng API repo. Chỉ field
  cần thiết, kèm default guard (limit max) trong repo implementation.
- `<entity>_build.go`: Build model hoặc update doc từ option; convert ID/time,
  set default; không chứa rule business.
- `<entity>_query.go` (Mongo/SQL): Build filter/pipeline/query. Unit test input
  → expected query map/SQL string.
- `<entity>_repo.go`: Triển khai repo, mapping model ↔ storage. Không nhúng
  rule business; wrap lỗi thêm ngữ cảnh.
- `mock_Repository.go`: Mock cho usecase test; update khi đổi interface.

### delivery/
- HTTP (`delivery/http`):
  - `dto.go`: Request/response struct; giữ field tags (json) rõ ràng.
  - `validator.go` (tuỳ chọn): Validate input (range/required), trả `ErrInvalidInput`.
  - `handlers.go`: Parse request → DTO → map sang usecase input; translate lỗi → HTTP code; không chạy goroutine.
  - `process_request.go`/`mapper.go`: Map DTO ↔ usecase; tách để test dễ.
  - `presenters.go`: Map domain → response DTO; không chứa rule business.
- MQ (`delivery/rabbitmq`):
  - `consumer.go` hoặc `<entity>_consumer.go`: Decode/validate message, call usecase, ack/nack theo guideline (temporary → requeue, business → ack).
  - `producer.go`: Helper publish message/event; wrap headers (trace_id), retry/backoff.
  - `presenters.go`: Map domain ↔ MQ payload; giữ schema chuẩn.
  - `constants.go`: Routing key/queue/exchange constants.
- Job/Cron (`delivery/job`):
  - `handlers.go`: Gọi usecase theo lịch; không nhúng business khác.
  - `register.go`/`new.go`: Wire scheduler, inject deps, set interval/cron spec.

## Quy trình tạo module mới (áp dụng cho `example/`)
1. **Khai báo contract**: định nghĩa input/output struct ở `uc_types.go`,
   lỗi business ở `uc_errors.go`, interface tại `uc_interface.go`.
2. **Repository**: thêm method cần thiết vào `repo_interface.go` + option struct.
   Nếu chưa có triển khai, tạo mock để test usecase trước.
3. **Usecase**: tạo file trong `usecase/`, implement interface, gom dependency
   qua `New…` (logger, repo, external clients). Không dùng biến global.
4. **Delivery**: nếu cần HTTP/gRPC/consumer, tạo adapter trong `delivery/`,
   validate input, map lỗi → HTTP status hoặc MQ ack/nack.
5. **Test**: viết unit test cho usecase; nếu có handler mới, test mapping đầu
   cuối; thêm integration test nếu chạm DB/queue.
6. **Tài liệu**: cập nhật README của module (file này) với API, payload, lỗi
   chuẩn và ví dụ lệnh chạy test.

## Lệnh tham khảo
- Chạy test module: `go test ./example/...`
- Kiểm tra format: `gofmt -w example`
- (Tuỳ chọn) lint: `golangci-lint run ./example/...` nếu có cấu hình chung.
