# RabbitMQ Producer Guide (TikTok & YouTube)

Tài liệu này dành cho các service muốn đẩy job vào hai worker scraper qua RabbitMQ. Mặc định các worker tự khai báo exchange/queue, chỉ cần publish đúng exchange và routing key.

## Kết nối mặc định
- URL: `amqp://guest:guest@localhost:5672/`
- VHost: `/`
- User/Pass: `guest` / `guest`
- Management UI (nếu bật image `rabbitmq:3-management`): `http://localhost:15672`
- Có thể override bằng biến môi trường trong `.env` của từng worker (`RABBITMQ_HOST`, `RABBITMQ_PORT`, `RABBITMQ_USER`, `RABBITMQ_PASSWORD`, `RABBITMQ_VHOST`, `RABBITMQ_EXCHANGE`, `RABBITMQ_ROUTING_KEY`, `RABBITMQ_QUEUE_NAME`, `RABBITMQ_PREFETCH_COUNT`).

## Exchange / Queue / Routing key
- **YouTube Worker**
  - Exchange: `youtube_exchange` (direct, durable)
  - Routing key: `youtube.crawl`
  - Queue: `youtube_crawl_queue` (durable)
  - Override: giá trị tương ứng bằng biến môi trường `RABBITMQ_*` trong `scrapper/youtube/.env`.
- **TikTok Worker**
  - Exchange: `tiktok_exchange` (direct, durable)
  - Routing key: `tiktok.crawl`
  - Queue: `tiktok_crawl_queue` (durable)
  - Override: giá trị tương ứng bằng biến môi trường `RABBITMQ_*` trong `scrapper/tiktok/.env`.

## Định dạng message chung
Gửi JSON UTF-8, nên đặt `delivery_mode=2` để message persistent.

```json
{
  "task_type": "research_keyword | crawl_links | research_and_crawl",
  "payload": { ... },
  "job_id": "tùy chọn, UUID hoặc string do producer sinh"
}
```

- `task_type` bắt buộc. Sai hoặc thiếu sẽ bị ACK và bỏ qua.
- `job_id` tùy chọn; worker sẽ tự sinh nếu trống và gắn vào record DB.
- `payload.time_range` (int, ngày) là option chung: chỉ lưu video/comment trong `[now - time_range, now]`; video ngoài range nhưng có comment hợp lệ vẫn giữ, nếu không sẽ đánh dấu `skipped`.

## Payload chi tiết theo worker

### YouTube
- `research_keyword`
  ```json
  {
    "task_type": "research_keyword",
    "payload": {
      "keyword": "python tutorial",
      "limit": 50,
      "sort_by": "relevance",
      "time_range": 7
    },
    "job_id": "job-yt-001"
  }
  ```
- `crawl_links`
  ```json
  {
    "task_type": "crawl_links",
    "payload": {
      "video_urls": [
        "https://www.youtube.com/watch?v=abc123"
      ],
      "include_channel": true,
      "include_comments": true,
      "max_comments": 100,
      "download_media": true,
      "media_type": "audio",
      "time_range": 14
    },
    "job_id": "job-yt-002"
  }
  ```
- `research_and_crawl`
  ```json
  {
    "task_type": "research_and_crawl",
    "payload": {
      "keywords": ["python programming"],
      "limit_per_keyword": 30,
      "include_comments": true,
      "include_channel": true,
      "max_comments": 50,
      "download_media": false,
      "time_range": 30
    },
    "job_id": "job-yt-003"
  }
  ```

### TikTok
- `research_keyword`
  ```json
  {
    "task_type": "research_keyword",
    "payload": {
      "keyword": "fitness",
      "limit": 20,
      "sort_by": "most_viewed",
      "time_range": 7
    },
    "job_id": "job-tt-001"
  }
  ```
- `crawl_links`
  ```json
  {
    "task_type": "crawl_links",
    "payload": {
      "video_urls": [
        "https://www.tiktok.com/@user/video/123"
      ],
      "include_comments": true,
      "include_creator": true,
      "max_comments": 50,
      "download_media": true,
      "media_type": "audio",
      "media_save_dir": "./downloads",
      "time_range": 14
    },
    "job_id": "job-tt-002"
  }
  ```
- `research_and_crawl`
  ```json
  {
    "task_type": "research_and_crawl",
    "payload": {
      "keywords": ["food", "travel"],
      "limit_per_keyword": 15,
      "sort_by": "liked",
      "include_comments": false,
      "include_creator": true,
      "max_comments": 0,
      "download_media": true,
      "media_type": "video",
      "media_save_dir": "./downloads",
      "time_range": 30
    },
    "job_id": "job-tt-003"
  }
  ```

## Ví dụ publish bằng Python (pika)
```python
import json
import pika

body = {
    "task_type": "crawl_links",
    "payload": {"video_urls": ["https://www.youtube.com/watch?v=abc123"]},
    "job_id": "job-yt-demo"
}

connection = pika.BlockingConnection(pika.URLParameters("amqp://guest:guest@localhost:5672/"))
channel = connection.channel()
channel.basic_publish(
    exchange="youtube_exchange",
    routing_key="youtube.crawl",
    body=json.dumps(body),
    properties=pika.BasicProperties(
        content_type="application/json",
        delivery_mode=2  # persistent
    )
)
connection.close()
```

Chỉ cần thay `exchange` và `routing_key` theo worker ở bảng trên. Nếu môi trường chạy khác host/port/credential, đổi URL kết nối tương ứng.
