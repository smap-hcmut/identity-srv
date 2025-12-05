# IMPLEMENTATION GUIDE: EVENT-DRIVEN CHOREOGRAPHY FOR SMAP

**Mục tiêu:** Thiết kế luồng dữ liệu tự hành (Autonomous Data Flow) từ Project → Collector → Analytics → Dashboard.

---

## 1. Thiết kế Hạ tầng Sự kiện (Event Infrastructure)

Để Choreography hoạt động, chúng ta cần một **Message Broker (RabbitMQ)** được cấu hình đúng chuẩn **Topic Exchange**.

### Cấu trúc Exchange & Routing Key

Chúng ta sẽ sử dụng 1 Exchange chính: `smap.events` (Type: `topic`).

| Routing Key | Ý nghĩa | Producer (Người gửi) | Consumers (Người nhận) |
| :--- | :--- | :--- | :--- |
| `project.created` | Có dự án mới cần chạy | Project Service | Collector Service |
| `data.collected` | Dữ liệu thô đã nằm trên MinIO | Collector Service | Analytics Service |
| `analysis.finished` | Phân tích xong 1 bài | Analytics Service | Insight Service / Notification |
| `job.completed` | Toàn bộ dự án đã xong | Analytics Service (Logic Redis) | Notification Service |

---

## 2. Quản lý Trạng thái Phân tán (Distributed State Management)

Vì không có ông Nhạc trưởng cầm sổ theo dõi, chúng ta dùng **Redis** làm "Bảng thông báo chung" - đóng vai trò "Trọng tài" (The Arbitrator) để đảm bảo các service rời rạc hiểu được bức tranh toàn cảnh.

### 2.1. Chiến lược chọn Database (Trong 16 DBs)

Để tránh việc cơ chế "dọn dẹp bộ nhớ" (Cache Eviction) của Redis vô tình xóa mất bộ đếm tiến độ, hãy tách biệt DB:

* **DB 0:** Dùng cho Cache (Session, API Response...).
* **DB 1:** Dùng riêng cho **SMAP State Management**.
  * *Lý do:* Dữ liệu này cần sống dai cho đến khi Project xong. Nếu dùng chung DB 0, khi RAM đầy, Redis có thể xóa nhầm key tracking → Hệ thống mất khả năng theo dõi.

### 2.2. Thiết kế Data Structure & Key

Thay vì dùng nhiều key rời rạc (`proj:{id}:status`, `proj:{id}:total`...), sử dụng **Redis HASH** để gom nhóm dữ liệu của 1 Project vào 1 key duy nhất.

**Tại sao dùng Hash?**
- Gọn gàng hơn khi đọc/ghi
- Dễ set TTL (hạn sử dụng) cho toàn bộ project
- Atomic operations trên nhiều fields

**Cấu trúc:**

| Key | Field | Kiểu | Mô tả | Ai Ghi? |
| :--- | :--- | :--- | :--- | :--- |
| `smap:proj:{id}` | `status` | String | `INITIALIZING`, `CRAWLING`, `PROCESSING`, `DONE` | Project / Collector / Analytics |
| | `total` | Int | Tổng số bài cần xử lý (VD: 1000) | Collector |
| | `done` | Int | Số bài đã xong (Atomic Counter) | Analytics |
| | `errors` | Int | Số bài bị lỗi | Analytics |

### 2.3. Cấu hình kết nối Redis (Python)

```python
import redis

# Kết nối vào DB 1 (Dành riêng cho State)
r = redis.Redis(host='smap-redis-master', port=6379, db=1, decode_responses=True)
```

---

## 3. Chi tiết Luồng Phối hợp (The Choreographed Flow)

Dưới đây là kịch bản chi tiết cho một vòng đời dữ liệu.

### BƯỚC 1: KÍCH HOẠT (The Trigger)

* **Tại:** `Project Service`
* **Hành động:** User bấm "Create Project".
* **Logic:**
  1. Lưu thông tin dự án vào PostgreSQL (`projects` table).
  2. Khởi tạo trạng thái trên Redis với TTL để tránh rác hệ thống.
  3. **Publish Event:** `project.created`.

**Implementation:**

```python
def init_project_state(project_id):
    key = f"smap:proj:{project_id}"
    
    # Dùng Pipeline để gom nhiều lệnh thành 1 transaction (Atomic)
    pipe = r.pipeline()
    pipe.hset(key, mapping={
        "status": "INITIALIZING",
        "total": 0,
        "done": 0,
        "errors": 0
    })
    pipe.expire(key, 604800)  # Tự hủy sau 7 ngày
    pipe.execute()
    
    print(f"Project {project_id} initialized in Redis.")
```

**Event Message:**

```json
// Topic: smap.events | Key: project.created
{
  "event_id": "evt_001",
  "timestamp": "2025-11-29T10:00:00Z",
  "payload": {
    "project_id": "proj_abc",
    "keywords": ["VinFast", "VF3"],
    "targets": [{"platform": "tiktok", "url": "..."}],
    "date_range": ["2025-01-01", "2025-02-01"]
  }
}
```

### BƯỚC 2: THU THẬP & SẢN XUẤT (The Producer)

* **Tại:** `Collector Service` (Collector Manager)
* **Hành động:**
  1. Lắng nghe `project.created`.
  2. Phân rã thành các Job con. Ví dụ: Tìm thấy 1000 bài viết cần crawl.
  3. Cập nhật Redis với tổng số items.
  4. Điều phối Worker đi crawl.
  5. Worker crawl xong 1 bài → Upload MinIO → **Publish Event:** `data.collected`.

**Implementation:**

```python
def set_project_total(project_id, total_items):
    key = f"smap:proj:{project_id}"
    
    r.hset(key, "total", total_items)
    r.hset(key, "status", "CRAWLING")
    
    print(f"Project {project_id} total items set to {total_items}")
```

**Event Message:**

```json
// Topic: smap.events | Key: data.collected
{
  "event_id": "evt_002",
  "payload": {
    "project_id": "proj_abc",
    "platform": "TIKTOK",
    "minio_path": "raw/tiktok/vid_888.json",
    "crawled_at": "..."
  }
}
```

### BƯỚC 3: XỬ LÝ & KIỂM TRA ĐÍCH (The Processor) - **QUAN TRỌNG NHẤT**

* **Tại:** `Analytics Service`
* **Hành động:**
  1. Lắng nghe `data.collected`. (Queue này nên set `prefetch_count` để không bị quá tải).
  2. Tải JSON từ MinIO dựa trên `minio_path`.
  3. Chạy Pipeline (5 Modules AI).
  4. Lưu kết quả vào PostgreSQL (`post_analytics`).
  5. Cập nhật Redis và kiểm tra hoàn thành.
  6. Nếu hoàn thành → **Publish Event:** `job.completed`.

**Implementation (Atomic Finish Check):**

Đây là nơi logic phức tạp nhất. Cần tăng biến đếm `done` và kiểm tra xem đã xong chưa. Thao tác này phải **Atomic** (bất khả phân chia) để tránh Race Condition khi có 100 workers chạy cùng lúc.

```python
def mark_item_done(project_id, is_error=False):
    key = f"smap:proj:{project_id}"
    
    # 1. Tăng biến đếm (Atomic Increment)
    # HINCRBY trả về giá trị MỚI sau khi cộng
    current_done = r.hincrby(key, "done", 1)
    
    if is_error:
        r.hincrby(key, "errors", 1)
        
    # 2. Lấy tổng số (Total) để so sánh
    total_str = r.hget(key, "total")
    total = int(total_str) if total_str else 0
    
    # 3. Kiểm tra vạch đích (The Finish Line Check)
    if total > 0 and current_done >= total:
        # Double check để đảm bảo chỉ bắn event 1 lần duy nhất
        current_status = r.hget(key, "status")
        if current_status != "DONE":
            r.hset(key, "status", "DONE")
            return "COMPLETED"  # Báo hiệu để code bên ngoài bắn RabbitMQ Event
            
    return "PROCESSING"


# --- Sử dụng trong code xử lý ---
# ... Xử lý AI xong ...
status = mark_item_done("proj_abc")

if status == "COMPLETED":
    rabbitmq.publish("smap.events", "job.completed", {"project_id": "proj_abc"})
    print("Job finished! Event sent.")
```

### BƯỚC 4: HIỂN THỊ & THÔNG BÁO (The View)

* **Tại:** `Insight Service` (hoặc Notification Service)
* **Hành động:**
  * **Real-time:** Frontend gọi API polling vào Redis để hiện Progress Bar.
  * **Hoàn thành:** Notification Service nghe `job.completed` → Gửi Email/Zalo cho User.

**Implementation:**

```python
def get_project_progress(project_id):
    key = f"smap:proj:{project_id}"
    
    # Lấy toàn bộ thông tin
    data = r.hgetall(key)
    
    if not data:
        return {"status": "NOT_FOUND", "percent": 0}
        
    total = int(data.get("total", 1))  # Tránh chia cho 0
    done = int(data.get("done", 0))
    
    percent = (done / total) * 100 if total > 0 else 0
    
    return {
        "status": data.get("status"),
        "total": total,
        "processed": done,
        "errors": int(data.get("errors", 0)),
        "percent": round(percent, 2)
    }
```

---

## 4. Giải quyết các Bài toán Hóc búa (Advanced Scenarios)

### 4.1. Vấn đề "Con gà quả trứng" (Collector chậm hơn Analytics)

**Tình huống:** Collector mới tìm được 10 bài, update `total=10`. Analytics chạy nhanh quá, xử lý xong 10 bài → Redis thấy `done=10, total=10` → Bắn event `job.completed`. Nhưng thực tế Collector vẫn đang tìm tiếp và sau đó update `total=100`.

**Giải pháp:**
* Collector chỉ update trạng thái là `CRAWLING` khi đang tìm.
* Khi Collector tìm xong HẾT, nó update trạng thái sang `PROCESSING_WAIT`.
* Analytics chỉ bắn event `job.completed` khi: `done == total` **VÀ** `status != CRAWLING`.

### 4.2. Vấn đề Dead Letter (Bài lỗi)

**Quan trọng:** Dù bài viết bị lỗi (crash, file hỏng), bạn **vẫn phải gọi `mark_item_done(..., is_error=True)`**.

*Lý do:* Nếu không tăng biến `done`, tổng số `done` sẽ mãi mãi nhỏ hơn `total` (ví dụ 999/1000) và hệ thống không bao giờ Finish được.

**Cơ chế xử lý:**
1. Analytics bắt lỗi (`try...except`).
2. **Ack** message đó (để xóa khỏi hàng đợi chính, tránh block các bài sau).
3. Gửi message đó vào **Dead Letter Queue (DLQ)**: `analytics.errors`.
4. Gọi `mark_item_done(project_id, is_error=True)` để Project vẫn có thể về đích 100%.

### 4.3. Xử lý trùng lặp (Idempotency)

**Tình huống:** RabbitMQ gửi 1 bài viết 2 lần.

**Cơ chế:**
* Analytics Service kiểm tra DB trước khi Insert.
* Sử dụng `INSERT ... ON CONFLICT DO UPDATE`.
* Redis `INCR` vẫn có thể bị tăng 2 lần. Chấp nhận sai số nhỏ này hoặc dùng `SET` trong Redis để lưu list ID đã làm (tốn RAM hơn).

### 4.4. Analytics Service bị sập (Crash)

**Hiện tượng:** Hàng nghìn message `data.collected` dồn ứ trong Queue RabbitMQ.

**Hậu quả:** Không sao cả. MinIO vẫn giữ file. Redis vẫn giữ số đếm cũ.

**Khắc phục:** Khi Analytics bật lại, nó tiếp tục consume message từ Queue và chạy tiếp. Không mất dữ liệu.

---

## 5. Tổng kết Kiến trúc

Với thiết kế này, hệ thống SMAP đạt được:

1. **DB 1 riêng biệt:** Đảm bảo dữ liệu tracking an toàn, không bị eviction.
2. **Redis Hash:** Cấu trúc dữ liệu gọn gàng, dễ quản lý TTL.
3. **Atomic Increment:** Đảm bảo chính xác 100% dù chạy đa luồng.
4. **Finish Check tại chỗ:** Logic kiểm tra đích ngay sau khi tăng biến đếm là cách hiệu quả nhất.
5. **High Throughput:** Collector cứ việc đẩy hàng nghìn bài vào kho mà không sợ Analytics bị nghẹn.
6. **Decoupling:** Team làm Crawler không cần quan tâm Team AI code gì, chỉ cần thống nhất format JSON.
7. **Visibility:** User vẫn thấy được thanh tiến độ chạy vù vù nhờ Redis, dù bên dưới là hàng chục service đang chạy tán loạn.

Đây chính là **Event-Driven Choreography** chuẩn mực cho hệ thống xử lý dữ liệu lớn.
