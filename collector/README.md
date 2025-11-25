# SMAP API - Dispatcher Service

This repository houses the **Dispatcher Service** for the SMAP data collection system. It serves as the central coordinator that receives high-level crawl requests, validates them, and distributes granular tasks to platform-specific workers via RabbitMQ.

## 🏗 System Architecture

The service follows a **Producer-Consumer** and **Fan-out** architecture pattern. It acts as an intermediary between the request initiator (e.g., a frontend or scheduler) and the actual scraping workers.

```mermaid
graph LR
    A[BE Service] -- "CrawlRequest" --> B(RabbitMQ: Dispatch Queue)
    B --> C{SMAP Dispatcher}
    C -- "Validate & Map" --> C
    C -- "CollectorTask (YouTube)" --> D[RabbitMQ: YouTube Queue]
    C -- "CollectorTask (TikTok)" --> E[RabbitMQ: TikTok Queue]
    D --> F[YouTube Worker]
    E --> G[TikTok Worker]
```

### Core Components

1.  **API Server (`cmd/api`)**:
    -   Currently provides health/readiness checks (`/health`, `/ready`, `/live`).
    -   Serves as the entry point for HTTP-based interactions (future expansion).
    -   Built with **Gin Framework**.

2.  **Dispatcher Consumer (`cmd/consumer`)**:
    -   The core worker process.
    -   Consumes `CrawlRequest` messages from the input queue.
    -   Executes the **Dispatch UseCase** to route tasks.

## Business Logic & Rules

The core logic resides in `internal/dispatcher`.

### 1. Dispatch Flow
The dispatch process (`internal/dispatcher/usecase/dispatch_uc.go`) follows these steps:
1.  **Receive Request**: Accepts a `CrawlRequest` containing `JobID`, `TaskType`, and a generic `Payload`.
2.  **Platform Selection**: Determines which platforms to target based on configuration (default: all enabled platforms).
3.  **Task Generation**: For each target platform:
    -   Creates a `CollectorTask` with a new trace ID and metadata.
    -   **Payload Mapping**: Converts the generic input payload into a strict, platform-specific struct (e.g., `YouTubeResearchKeywordPayload`) using the `Mapper` logic.
    -   **Routing**: Determines the correct RabbitMQ routing key for the platform.
4.  **Publish**: Sends the `CollectorTask` to the platform's queue.

### 2. Supported Task Types
The system supports three primary task types (`internal/models/task.go`):
-   `research_keyword`: Search for content based on keywords.
-   `crawl_links`: Scrape specific video/profile URLs.
-   `research_and_crawl`: A composite task to search and then immediately scrape results.

### 3. Supported Platforms
-   **YouTube** (`PlatformYouTube`)
-   **TikTok** (`PlatformTikTok`)

## 📐 Design Patterns

The project strictly follows **Clean Architecture** and **SOLID** principles:

-   **Clean Architecture**:
    -   `delivery/`: Transport layer (RabbitMQ consumers, HTTP handlers).
    -   `usecase/`: Pure business logic (Dispatching, Mapping).
    -   `models/`: Domain entities and DTOs.
    -   **Benefit**: The business logic is decoupled from the transport mechanism. We could easily switch from RabbitMQ to Kafka without changing the core dispatch logic.

-   **Dependency Injection (DI)**:
    -   All components (UseCases, Repositories, Consumers) are initialized with their dependencies passed via constructors (`New...`).
    -   **Benefit**: Makes testing easier (mocking dependencies) and dependencies explicit.

-   **Strategy/Factory Pattern**:
    -   The `mapPayload` function (`internal/dispatcher/usecase/mapper.go`) acts as a factory, selecting the correct payload structure and validation strategy based on the `Platform` and `TaskType`.

## 📂 Project Structure

```
smap-api/
├── cmd/
│   ├── api/          # HTTP Server entry point
│   └── consumer/     # RabbitMQ Worker entry point
├── config/           # Configuration loading (Viper/Env)
├── docs/             # Documentation & Design specs
├── internal/
│   ├── dispatcher/   # CORE DOMAIN: Dispatch logic
│   │   ├── delivery/ # RabbitMQ consumers/producers
│   │   └── usecase/  # Business logic (Dispatch, Map)
│   ├── httpserver/   # Gin server setup
│   ├── models/       # Shared data structures (CrawlRequest, CollectorTask)
│   └── middleware/   # Gin middlewares (Auth, Logging)
├── pkg/              # Shared utilities (Logger, RabbitMQ, Mongo, etc.)
└── ...
```

## 🚀 Getting Started

### Prerequisites
-   Go 1.23+
-   RabbitMQ
-   MongoDB (for future persistence)
-   Redis (optional, for caching)

### Configuration
Copy `env.template` to `.env` and configure:
```ini
# RabbitMQ
AMQP_URL=amqp://guest:guest@localhost:5672/

# Service
PORT=8080
MODE=debug
```

### Running the Service

**1. Start the Dispatcher Worker:**
This process listens for incoming requests and dispatches them.
```bash
go run cmd/consumer/main.go
```

**2. Start the API Server:**
(Optional, for health checks)
```bash
go run cmd/api/main.go
```

## 🔌 Integration Guide

### RabbitMQ Connection (Inbound)

External services (e.g., API Gateway, Scheduler) should publish `CrawlRequest` messages to the **Inbound** exchange.

-   **Exchange**: `collector.inbound` (Type: `topic`)
-   **Routing Key**: `crawler.#` (e.g., `crawler.request`)
-   **Queue**: `collector.inbound.queue`

> [!NOTE]
> The `collector.tiktok` and `collector.youtube` exchanges are **internal** and managed by the dispatcher. Do not publish to them directly.

#### Payload Example (`CrawlRequest`)

```json
{
  "job_id": "job_12345",
  "task_type": "research_keyword",
  "payload": {
    "keyword": "golang tutorial",
    "limit": 10
  },
  "time_range": 7,
  "attempt": 1,
  "max_attempts": 3,
  "emitted_at": "2023-10-27T10:00:00Z"
}
```
