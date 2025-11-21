# SMAP Project Service

> Project management service for the SMAP platform

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-336791?style=flat&logo=postgresql)](https://www.postgresql.org/)
[![Docker](https://img.shields.io/badge/Docker-Optimized-2496ED?style=flat&logo=docker)](https://www.docker.com/)

---

## Overview

**SMAP Project Service** manages project-related operations for the SMAP platform. It provides CRUD operations for projects including brand tracking, competitor analysis, and keyword management.

### Key Features

- **Project Management**: Create, read, update, and delete projects
- **Brand Tracking**: Track brand names and keywords
- **Competitor Analysis**: Monitor competitor names and their associated keywords
- **Date Range Management**: Project timeline management with validation
- **Status Tracking**: Draft, Active, Completed, Archived, Cancelled
- **User Isolation**: Users can only access their own projects
- **Soft Delete**: Data retention for audit purposes

---

## API Endpoints

### Base URL
```
http://localhost:8080/project
```

### Project Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/projects` | List all user's projects | Yes |
| GET | `/projects/page` | Get projects with pagination | Yes |
| GET | `/projects/:id` | Get project details | Yes |
| POST | `/projects` | Create new project | Yes |
| PUT | `/projects/:id` | Update project | Yes |
| DELETE | `/projects/:id` | Delete project (soft delete) | Yes |

---

## Getting Started

### Prerequisites

- Go 1.23+
- PostgreSQL 15+
- Make

### Quick Start

```bash
# Install dependencies
go mod download

# Run migrations
make migrate-up

# Generate SQLBoiler models
make sqlboiler

# Run the service
make run-api
```

### API Examples

**Create Project:**
```bash
curl -X POST http://localhost:8080/project/projects \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Q1 2025 Campaign",
    "status": "draft",
    "from_date": "2025-01-01T00:00:00Z",
    "to_date": "2025-03-31T23:59:59Z",
    "brand_name": "MyBrand",
    "brand_keywords": ["mybrand", "my brand"],
    "competitor_names": ["Competitor A"],
    "competitor_keywords_map": {
      "Competitor A": ["competitor-a", "comp-a"]
    }
  }'
```

---

**Built for SMAP Graduation Project**
