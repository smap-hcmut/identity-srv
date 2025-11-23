# Social Media Scraper Services

A collection of high-performance social media data scraping services built with Clean Architecture principles. Each service is designed as an independent worker that consumes tasks from RabbitMQ, scrapes social media platforms, and persists data to MongoDB with media files stored in MinIO.

## Overview

This repository contains enterprise-grade scraper workers for:
- **TikTok**: Video metadata, creator profiles, comments, and audio/video media
- **YouTube**: Video information, channel data, comments, and media downloads
- **Instagram**: Legacy worker (currently in production, documentation in progress)

All services share a common architecture pattern and infrastructure stack for consistency and maintainability.

## Architecture

```
Producer System → RabbitMQ → Worker Services → MongoDB + MinIO
                    ↓
        ┌──────────┴──────────┬──────────────┐
        │                     │              │
   TikTok Worker        YouTube Worker   Instagram Worker
```

### Shared Infrastructure

- **RabbitMQ**: Task queue for distributed job processing
- **MongoDB**: Document storage for metadata and tracking
- **MinIO**: S3-compatible object storage for media files
- **Docker Compose**: Container orchestration for all services

### Storage Strategy

- **MinIO Buckets**:
  - `tiktok`: TikTok audio files
  - `youtube`: YouTube audio and video files (temporary storage during processing)
- **MongoDB Collections**: Per-service databases with dedicated collections for videos, creators/channels, comments, search results, and job tracking

## Services

### TikTok Scraper

A robust TikTok data scraper using Playwright for browser automation and HTTP APIs for efficient data collection.

**Key Features:**
- Browser automation with remote Playwright service
- Concurrent video crawling
- Creator profile extraction
- Comment harvesting
- Audio/video media download with FFmpeg

**[📖 TikTok Documentation →](tiktok/README.md)**

**Quick Start:**
```bash
cd tiktok
cp .env.example .env
# Edit .env with your configuration
python -m app.main
```

**Docker:**
```bash
docker compose up tiktok-worker
```

---

### YouTube Scraper

A high-performance YouTube scraper built with yt-dlp for comprehensive data extraction.

**Key Features:**
- yt-dlp integration for video metadata
- 100% comment coverage with youtube-comment-downloader
- Channel information scraping
- Search functionality with multiple sort options
- Audio/video download with FFmpeg support

**[📖 YouTube Documentation →](youtube/README.md)**

**Quick Start:**
```bash
cd youtube
cp .env.example .env
# Edit .env with your configuration
python -m app.worker_service
```

**Docker:**
```bash
docker compose up youtube-worker
```

---

### Instagram Scraper

Legacy worker currently serving production. Documentation and refactoring in progress.

**[📖 Production Deployment Guide →](insta/README_PRODUCTION.md)**

## Common Task Types

All scraper services support three standard task types:

### 1. Research Keyword
Search by keyword and save results.

**Use Case:** Discover trending content, find videos by topic

### 2. Crawl Links
Crawl specific URLs with full metadata extraction.

**Use Case:** Update existing data, crawl specific content

### 3. Research and Crawl
Search then crawl all found content in one job.

**Use Case:** Comprehensive data collection for specific topics

## Docker Compose

The repository includes a unified `docker-compose.yml` for running all services.

### Services Defined

```yaml
services:
  playwright-service:    # Remote browser automation for TikTok
  tiktok-worker:         # TikTok scraper worker
  youtube-worker:        # YouTube scraper worker
```

### Running Services

```bash
# Build all services
docker compose build

# Run specific service
docker compose up tiktok-worker
docker compose up youtube-worker

# Run all services
docker compose up

# Run in background
docker compose up -d

# View logs
docker compose logs -f tiktok-worker
docker compose logs -f youtube-worker

# Stop services
docker compose down
```

### Service Configuration

Each service:
- Uses its own `.env` file in its directory
- Mounts source code as volume for live development
- Connects to shared RabbitMQ, MongoDB, and MinIO instances
- Has isolated dependencies and runtime environment

## Prerequisites

### Required Software

- **Python 3.11** (TikTok) / **Python 3.10+** (YouTube)
- **Docker & Docker Compose** (for containerized deployment)
- **RabbitMQ 3.x** (message broker)
- **MongoDB 6.x** (data persistence)
- **MinIO** (object storage)
- **FFmpeg** (media processing)

### Infrastructure Setup

You can run infrastructure services via Docker:

```bash
# RabbitMQ
docker run -d --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management

# MongoDB
docker run -d --name mongodb -p 27017:27017 mongo:6

# MinIO
docker run -d --name minio -p 9000:9000 -p 9001:9001 \
  -e "MINIO_ROOT_USER=minioadmin" \
  -e "MINIO_ROOT_PASSWORD=minioadmin" \
  minio/minio server /data --console-address ":9001"
```

Or use docker-compose infrastructure file (if available).

## Local Development Workflow

### 1. Infrastructure Setup

Ensure RabbitMQ, MongoDB, and MinIO are running and accessible.

### 2. Service Configuration

Each service has an `.env.example` file. Copy it to `.env` and configure:

```bash
cd tiktok
cp .env.example .env
# Edit .env with your credentials

cd ../youtube
cp .env.example .env
# Edit .env with your credentials
```

### 3. Install Dependencies

```bash
# TikTok
cd tiktok
python -m venv .venv
source .venv/bin/activate  # or .venv\Scripts\activate on Windows
pip install -r requirements.txt
playwright install --with-deps chromium

# YouTube
cd ../youtube
python -m venv .venv
source .venv/bin/activate  # or .venv\Scripts\activate on Windows
pip install -r requirements.txt
```

### 4. Run Workers

**Manual Execution:**
```bash
# TikTok
cd tiktok
python -m app.main

# YouTube
cd youtube
python -m app.worker_service
```

**Docker Compose (Recommended):**
```bash
docker compose up
```

## Testing

Each service manages its own test suite in a `tests/` directory.

### Run Tests for a Service

```bash
# TikTok
cd tiktok
pytest tests/

# YouTube
cd youtube
pytest tests/
```

### Test Categories

- **Unit Tests**: `pytest tests/unit/` - Pure business logic
- **Integration Tests**: `pytest tests/integration/` - Database and queue interactions
- **E2E Tests**: `pytest tests/e2e/` - Complete workflow validation

**Best Practice:** Use separate test databases and queues to avoid affecting production data.

## Project Structure

```
scrapper/
├── docker-compose.yml           # Service orchestration
├── playwright.Dockerfile        # Shared Playwright service
├── README.MD                    # This file
│
├── tiktok/                      # TikTok scraper service
│   ├── README.md                # TikTok documentation
│   ├── Dockerfile
│   ├── requirements.txt
│   ├── .env.example
│   ├── app/                     # Entry points & DI
│   ├── application/             # Use cases
│   ├── domain/                  # Business entities
│   ├── internal/                # Infrastructure adapters
│   ├── config/                  # Configuration
│   ├── utils/                   # Utilities
│   ├── message_queue/           # RabbitMQ integration
│   ├── tests/                   # Test suite
│   └── docs/                    # Additional documentation
│
├── youtube/                     # YouTube scraper service
│   ├── README.md                # YouTube documentation
│   ├── Dockerfile
│   ├── requirements.txt
│   ├── .env.example
│   ├── app/                     # Entry points & DI
│   ├── application/             # Use cases
│   ├── domain/                  # Business entities
│   ├── internal/                # Infrastructure adapters
│   ├── config/                  # Configuration
│   ├── utils/                   # Utilities
│   └── tests/                   # Test suite
│
└── insta/                       # Instagram scraper (legacy)
    ├── README_PRODUCTION.md     # Production deployment guide
    └── ...
```

## Clean Architecture

All services follow **Clean Architecture** principles:

### Layer Structure

```
┌─────────────────────────────────────────┐
│              app/                       │  Entry points & DI container
└───────────────┬─────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────┐
│          application/                   │  Use cases & interfaces
└───────────────┬─────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────┐
│            domain/                      │  Business entities (NO dependencies)
└─────────────────────────────────────────┘
                ▲
                │
┌───────────────┴─────────────────────────┐
│          internal/                      │  Adapters & infrastructure
└─────────────────────────────────────────┘
```

### Benefits

- **Testability**: Each layer can be tested in isolation
- **Flexibility**: Easy to swap implementations (e.g., MongoDB → PostgreSQL)
- **Maintainability**: Clear separation of concerns
- **Independence**: Business logic independent of frameworks
- **Scalability**: Easy to add new features without breaking existing code

## Adding a New Service

When adding a new social media scraper:

1. **Create Service Directory**
   ```bash
   mkdir scrapper/new-service
   cd scrapper/new-service
   ```

2. **Follow Clean Architecture Structure**
   - Copy structure from TikTok or YouTube as template
   - Adapt domain entities for the new platform
   - Implement platform-specific scrapers in `internal/adapters/`

3. **Add Docker Service**
   Edit `docker-compose.yml`:
   ```yaml
   new-service-worker:
     container_name: new-service-scraper
     build:
       context: ./new-service
       dockerfile: Dockerfile
     env_file:
       - ./new-service/.env
     command: ["python", "-m", "app.worker_service"]
   ```

4. **Create Documentation**
   - Write comprehensive `README.md` in service directory
   - Document task types and payload formats
   - Include configuration examples

5. **Update This README**
   - Add service to overview section
   - Add link to service documentation
   - Update architecture diagram if needed

## Configuration Management

Each service uses environment variables for configuration:

### Common Settings

All services share these configuration patterns:

- **RabbitMQ**: Connection, queue names, prefetch count
- **MongoDB**: Connection string, database name, authentication
- **MinIO**: Endpoint, credentials, bucket names
- **Crawler**: Concurrency limits, timeouts, retry logic
- **Media**: Download settings, FFmpeg options, storage paths
- **Logging**: Log level, output paths, worker naming

### Environment Variables

See individual service `.env.example` files for complete variable lists:
- [TikTok .env.example](tiktok/.env.example)
- [YouTube .env.example](youtube/.env.example)

## Monitoring and Observability

### Job Tracking

All services track jobs in MongoDB `jobs` collection:
- **Status**: `pending`, `processing`, `completed`, `failed`
- **Timestamps**: `created_at`, `started_at`, `completed_at`
- **Metadata**: Error messages, retry count, associated resource IDs

### Logging

Each service implements structured logging:
- **Output**: stdout + `logs/` directory
- **Format**: Timestamp, level, service name, message, context
- **Levels**: DEBUG, INFO, WARNING, ERROR

### Health Checks

Monitor service health by checking:
- RabbitMQ connection status
- MongoDB connection status
- Recent job completion rates
- Error rates in logs

## Troubleshooting

### Common Issues

#### RabbitMQ Connection Failed
- Verify RabbitMQ is running: `docker ps | grep rabbitmq`
- Check credentials in service `.env` files
- Verify port 5672 is accessible
- Check virtual host configuration

#### MongoDB Connection Failed
- Verify MongoDB is running: `docker ps | grep mongo`
- Check connection string format
- Verify authentication credentials if enabled
- Check network connectivity

#### MinIO Connection Failed
- Verify MinIO is running: `docker ps | grep minio`
- Check endpoint URL (include port)
- Verify access key and secret key
- Ensure buckets are created

#### FFmpeg Not Found
- Install FFmpeg: See service-specific installation guides
- Verify PATH environment variable
- Test: `ffmpeg -version`

#### Service Won't Start
- Check logs: `docker compose logs service-name`
- Verify all required environment variables are set
- Check dependency service health
- Ensure ports are not already in use

### Getting Help

For service-specific issues, refer to individual service documentation:
- [TikTok Troubleshooting](tiktok/README.md#troubleshooting)
- [YouTube Troubleshooting](youtube/README.md#troubleshooting)

## Contributing

When contributing to this project:

1. **Follow Clean Architecture principles** in all services
2. **Write comprehensive tests** (unit, integration, e2e)
3. **Document changes** in service READMEs
4. **Update environment examples** when adding new config options
5. **Follow Python PEP 8** style guidelines
6. **Use type hints** for better code clarity
7. **Keep domain layer pure** with no external dependencies

## License

*License information to be added*

## Changelog

### Version 2.0.0 (2025-11-06)
- Complete rewrite of TikTok scraper with Clean Architecture
- Complete rewrite of YouTube scraper with Clean Architecture
- Added remote Playwright service for TikTok
- Unified Docker Compose configuration
- Comprehensive documentation updates
- Smart upsert logic for efficient data updates

### Version 1.0.0
- Initial TikTok scraper implementation
- Initial YouTube scraper implementation
- Instagram production worker
- Basic Docker support

---

**Last Updated:** 2025-11-06
**Version:** 2.0.0
**Maintainer:** SMAP AI Team
