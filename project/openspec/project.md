# Project Context

## Purpose

SMAP Project Service manages project-related operations for the SMAP platform. It provides CRUD operations for projects including brand tracking, competitor analysis, and keyword management.

## Tech Stack

- **Language**: Go 1.25+
- **Web Framework**: Gin
- **Database**: PostgreSQL 15+
- **ORM**: SQLBoiler
- **Authentication**: JWT
- **Documentation**: Swagger
- **Infrastructure**: Docker, Docker Compose
- **Object Storage**: MinIO
- **Messaging**: RabbitMQ

## Project Conventions

### Code Style

- Standard Go conventions (gofmt)
- Project layout follows standard Go patterns (`cmd`, `internal`, `pkg`)

### Architecture Patterns

- **Layered Architecture**: Handlers -> Services -> Repositories
- **Dependency Injection**: Manual injection in `main.go`

### Testing Strategy

- Standard Go testing (`go test`)
- Unit tests for business logic

### Git Workflow

- Feature branching strategy

## Domain Context

- **Projects**: Core entity, belongs to a user.
- **Brand & Competitors**: Projects track a brand and its competitors.
- **Keywords**: Used for tracking brand and competitor mentions.

## Important Constraints

- **User Isolation**: Users can only access their own projects.
- **Soft Delete**: Projects are soft-deleted for audit purposes.

## External Dependencies

- **PostgreSQL**: Primary data store.
- **MinIO**: For file storage (if applicable).
- **RabbitMQ**: For asynchronous tasks.
