# SMAP Identity API - Documentation Overview

> Complete documentation structure for SMAP Identity Service - Authentication, Authorization, Subscription Management & Background Processing

---

## Table of Contents

- [Service Overview](#service-overview)
- [Architecture](#architecture)
- [Documentation Structure](#documentation-structure)
- [Entity-Relationship Diagram](#entity-relationship-diagram)
- [Technology Stack](#technology-stack)
- [Document Maintenance](#document-maintenance)

---

## Service Overview

**SMAP Identity API** là hệ thống quản lý identity và subscription-based authentication được xây dựng theo Clean Architecture pattern, bao gồm hai services chính:

### 1. API Server (`cmd/api/`)

RESTful API service cung cấp các chức năng:

**Authentication & Authorization:**
- User registration với email verification
- OTP-based verification qua email
- JWT-based authentication
- Role-based access control (USER, ADMIN)
- Password management với encryption

**Subscription Management:**
- Plan CRUD operations
- Subscription lifecycle management
- Automatic free trial creation (14 ngày)
- Subscription status tracking (trialing, active, cancelled, expired)
- User subscription query và cancellation

**User Management:**
- User profile management
- Admin user listing với pagination
- User detail query
- Soft delete support

### 2. Consumer Service (`cmd/consumer/`)

Background job processing service:

**Async Task Processing:**
- Email sending via RabbitMQ + SMTP
- Graceful shutdown handling
- Extensible architecture cho future consumers
- Error handling và retry logic

**Features:**
- Consume messages từ RabbitMQ queues
- Process email sending tasks
- Monitor queue depth và performance
- Horizontal scaling support

---

## Architecture

### System Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Client Applications                          │
│                    (Web, Mobile, Third-party)                        │
└────────────────────────┬────────────────────────────────────────────┘
                         │ HTTPS/REST API
                         ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      API Server (:8080)                              │
├─────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                  HTTP Delivery Layer                        │   │
│  │  ├─ Authentication Handlers  (/identity/auth/*)            │   │
│  │  ├─ Plan Handlers           (/identity/plans/*)            │   │
│  │  ├─ Subscription Handlers   (/identity/subscriptions/*)    │   │
│  │  └─ User Handlers           (/identity/users/*)            │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                            │                                         │
│  ┌─────────────────────────▼───────────────────────────────────┐   │
│  │                   UseCase Layer                             │   │
│  │             (Business Logic & Orchestration)                │   │
│  │  ├─ Authentication UseCase                                  │   │
│  │  ├─ Plan UseCase                                            │   │
│  │  ├─ Subscription UseCase                                    │   │
│  │  ├─ User UseCase                                            │   │
│  │  └─ Email Producer (RabbitMQ)                               │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                            │                                         │
│  ┌─────────────────────────▼───────────────────────────────────┐   │
│  │                  Repository Layer                           │   │
│  │                (Data Access & Queries)                      │   │
│  │  ├─ User Repository                                         │   │
│  │  ├─ Plan Repository                                         │   │
│  │  └─ Subscription Repository                                 │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                       │
└────────┬────────────────────────────────┬───────────────────────────┘
         │                                │
         │ SQL Queries                    │ Publish Messages (Async)
         ▼                                ▼
┌──────────────────────┐      ┌─────────────────────────────────────┐
│  PostgreSQL (:5432)  │      │       RabbitMQ (:5672)              │
│                      │      │  ┌────────────────────────────────┐ │
│  Tables:             │      │  │ Exchange: smtp_send_email_exc  │ │
│  ├─ users            │      │  │ Type: fanout                   │ │
│  ├─ plans            │      │  └────────────────────────────────┘ │
│  └─ subscriptions    │      │  ┌────────────────────────────────┐ │
│                      │      │  │ Queue: smtp_send_email         │ │
└──────────────────────┘      │  │ Durable: true                  │ │
                              │  └────────────────────────────────┘ │
                              └──────────┬──────────────────────────┘
                                         │ Consume Messages
                                         ▼
                              ┌─────────────────────────────────────┐
                              │    Consumer Service                 │
                              ├─────────────────────────────────────┤
                              │  ┌───────────────────────────────┐ │
                              │  │  Consumer Orchestrator        │ │
                              │  │  └─ Graceful Shutdown         │ │
                              │  └───────────────────────────────┘ │
                              │  ┌───────────────────────────────┐ │
                              │  │  SMTP Consumer                │ │
                              │  │  ├─ Consume Queue             │ │
                              │  │  ├─ Deserialize Messages      │ │
                              │  │  └─ Call SMTP UseCase         │ │
                              │  └───────────────────────────────┘ │
                              │  ┌───────────────────────────────┐ │
                              │  │  SMTP UseCase                 │ │
                              │  │  ├─ Build Email Content       │ │
                              │  │  ├─ Attach Files (if any)     │ │
                              │  │  └─ Send via SMTP             │ │
                              │  └───────────────────────────────┘ │
                              └──────────────┬──────────────────────┘
                                             │ Send Email
                                             ▼
                              ┌─────────────────────────────────────┐
                              │     SMTP Server (:587)              │
                              │  (Gmail, SendGrid, Mailgun, etc.)   │
                              └──────────────┬──────────────────────┘
                                             │ Deliver
                                             ▼
                                       User's Email Inbox
```

### Clean Architecture Layers

**4-Layer Architecture Pattern:**

```
┌──────────────────────────────────────────────────────────────────┐
│                      1. Delivery Layer                           │
│                                                                   │
│  - HTTP Handlers (Gin framework)                                │
│  - Request/Response DTOs                                         │
│  - Input validation                                              │
│  - RabbitMQ Producers                                            │
│  - Middleware (Auth, CORS, Error handling)                       │
│                                                                   │
└─────────────────────────────┬────────────────────────────────────┘
                              │ Calls
                              ▼
┌──────────────────────────────────────────────────────────────────┐
│                       2. UseCase Layer                           │
│                                                                   │
│  - Business Logic                                                │
│  - Orchestration between repositories                            │
│  - Validation rules                                              │
│  - Error handling                                                │
│  - Transaction management                                        │
│                                                                   │
└─────────────────────────────┬────────────────────────────────────┘
                              │ Calls
                              ▼
┌──────────────────────────────────────────────────────────────────┐
│                     3. Repository Layer                          │
│                                                                   │
│  - Data access abstraction                                       │
│  - Query builders                                                │
│  - Database operations (CRUD)                                    │
│  - Pagination logic                                              │
│  - Filtering & sorting                                           │
│                                                                   │
└─────────────────────────────┬────────────────────────────────────┘
                              │ Queries
                              ▼
┌──────────────────────────────────────────────────────────────────┐
│                  4. External Services Layer                      │
│                                                                   │
│  - PostgreSQL Database                                           │
│  - RabbitMQ Message Broker                                       │
│  - SMTP Email Server                                             │
│  - MinIO Object Storage                                          │
│                                                                   │
└──────────────────────────────────────────────────────────────────┘
```

### Key Architectural Decisions

**1. Separation of Concerns:**
- API Server: Synchronous request/response
- Consumer Service: Asynchronous task processing
- Decoupled via RabbitMQ message queue

**2. Database Design:**
- Users table: Authentication data với encrypted role
- Plans table: Subscription plan definitions
- Subscriptions table: User-plan relationships với status tracking
- Soft delete pattern: deleted_at timestamp

**3. Message Flow:**
```
API UseCase → Producer → RabbitMQ Exchange → Queue → Consumer → SMTP → Email
```

**4. Scalability:**
- Horizontal scaling: Multiple API instances với load balancer
- Consumer scaling: Multiple consumer instances consuming từ same queue
- Database: Connection pooling, prepared statements


## Entity-Relationship Diagram

```mermaid
erDiagram
    USERS ||--o{ SUBSCRIPTIONS : "has"
    PLANS ||--o{ SUBSCRIPTIONS : "defines"
    
    USERS {
        uuid id PK
        varchar username UK "Email address"
        varchar full_name "Full name"
        text password_hash "Encrypted password"
        text role_hash "Encrypted role (USER/ADMIN)"
        text avatar_url "Avatar image URL"
        boolean is_active "Account verified"
        varchar otp "6-digit OTP"
        timestamp otp_expired_at "OTP expiry time"
        timestamp created_at "Record creation"
        timestamp updated_at "Last update"
        timestamp deleted_at "Soft delete"
    }
    
    PLANS {
        uuid id PK
        varchar name "Plan display name"
        varchar code UK "Unique identifier"
        text description "Plan description"
        int max_usage "API call limit per day"
        timestamp created_at "Record creation"
        timestamp updated_at "Last update"
        timestamp deleted_at "Soft delete"
    }
    
    SUBSCRIPTIONS {
        uuid id PK
        uuid user_id FK "References users.id"
        uuid plan_id FK "References plans.id"
        varchar status "trialing, active, cancelled, expired, past_due"
        timestamp trial_ends_at "Trial expiry date"
        timestamp starts_at "Subscription start"
        timestamp ends_at "Subscription end"
        timestamp cancelled_at "Cancellation date"
        timestamp created_at "Record creation"
        timestamp updated_at "Last update"
        timestamp deleted_at "Soft delete"
    }
```

---

## Technology Stack

### Backend Core
| Technology | Version | Purpose |
|------------|---------|---------|
| **Go** | 1.25 | Primary programming language |
| **Gin** | Latest | HTTP web framework, routing |
| **PostgreSQL** | 15 | Primary database, relational data |
| **SQLBoiler** | Latest | Type-safe ORM, code generation |

### Messaging & Async
| Technology | Version | Purpose |
|------------|---------|---------|
| **RabbitMQ** | 3.x | Message broker, task queue |
| **amqp091-go** | Latest | RabbitMQ Go client |

### Security & Auth
| Technology | Version | Purpose |
|------------|---------|---------|
| **golang-jwt** | Latest | JWT token generation & validation |
| **bcrypt** | Built-in | Password hashing |
| **SHA256** | Built-in | Role encryption |

### Email & Communication
| Technology | Version | Purpose |
|------------|---------|---------|
| **go-mail** | v2 | SMTP email sending |
| **SMTP** | - | Email delivery protocol |

### Infrastructure & DevOps
| Technology | Version | Purpose |
|------------|---------|---------|
| **Docker** | >= 20.10 | Containerization |
| **BuildKit** | Latest | Docker build optimization |
| **Distroless** | debian12 | Minimal runtime image |

### Development Tools
| Technology | Version | Purpose |
|------------|---------|---------|
| **Swagger** | v1.8.12 | API documentation |
| **Zap** | Latest | Structured logging |
| **Make** | Built-in | Build automation |

---

## External Services

### Required Services

#### 1. PostgreSQL Database
**Purpose**: Primary data storage
- **Port**: 5432
- **Tables**: users, plans, subscriptions
- **Features Used**: 
  - Foreign keys & constraints
  - Indexes (username, deleted_at)
  - Soft delete (deleted_at timestamp)
  - UUID primary keys

**Connection:**
```
postgresql://user:password@host:5432/database?sslmode=disable
```

---

#### 2. RabbitMQ Message Broker
**Purpose**: Async task queue, message routing
- **Port**: 5672 (AMQP), 15672 (Management UI)
- **Exchanges**: 
  - `smtp_send_email_exc` (fanout, durable)
- **Queues**: 
  - `smtp_send_email` (durable, bound to exchange)
- **Features Used**:
  - Fanout exchange routing
  - Message persistence (durable)
  - Manual acknowledgment
  - Prefetch (for load balancing)

**Connection:**
```
amqp://guest:guest@localhost:5672/
```

---

#### 3. SMTP Email Server
**Purpose**: Email delivery
- **Port**: 587 (TLS), 465 (SSL)
- **Supported Providers**:
  - Gmail (smtp.gmail.com)
  - SendGrid (smtp.sendgrid.net)
  - Mailgun (smtp.mailgun.org)
  - Custom SMTP servers

**Gmail Configuration:**
1. Enable 2-Factor Authentication
2. Generate App Password: https://myaccount.google.com/apppasswords
3. Use App Password as SMTP_PASSWORD

**Connection:**
```
Host: smtp.gmail.com
Port: 587
Username: your-email@gmail.com
Password: app-password (16 characters)
From: noreply@smap.com
```

---

#### 4. MinIO Object Storage (Optional)
**Purpose**: File storage, avatar uploads
- **Port**: 9000 (API), 9001 (Console)
- **S3-Compatible**: Yes
- **Features Used**:
  - Bucket creation & management
  - Object upload/download
  - Presigned URLs

**Connection:**
```
Endpoint: localhost:9000
AccessKey: minioadmin
SecretKey: minioadmin
```

---

### Service Dependencies Map

```
API Server Dependencies:
├─ PostgreSQL (Required)
│  └─ Data persistence
├─ RabbitMQ (Required)
│  └─ Async task publishing
└─ MinIO (Optional)
   └─ File storage

Consumer Service Dependencies:
├─ RabbitMQ (Required)
│  └─ Task consumption
├─ SMTP Server (Required)
│  └─ Email delivery
└─ PostgreSQL (Optional)
   └─ Future features
```

## Document Maintenance

### Version History
- **v1.0.0** (November 20, 2025): Initial documentation structure
  - 3 core documentation files (overview, api, consumer)
  - 6 supporting documentation files
  - 17 sequence diagrams
  - Complete architecture diagrams
  - Setup & deployment guides

**Last Updated**: November 20, 2025  
**Document Version**: 1.0.0  
**Maintained By**: SMAP Development Team

---

*Built with love using Clean Architecture principles*