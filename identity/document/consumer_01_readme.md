# SMAP Consumer Service

> Async task processing service for SMAP Identity API

[![Go Version](https://img.shields.io/badge/Go-1.25-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Docker](https://img.shields.io/badge/Docker-Optimized-2496ED?style=flat&logo=docker)](https://www.docker.com/)
[![RabbitMQ](https://img.shields.io/badge/RabbitMQ-3.x-FF6600?style=flat&logo=rabbitmq)](https://www.rabbitmq.com/)

---

## Overview

Consumer Service xử lý các tác vụ bất đồng bộ qua RabbitMQ:
- **Email Sending**: Verification emails, notifications
- **Push Notifications**: Mobile & web push
- **Webhooks**: Event callbacks
- **Data Processing**: Background jobs

---

## Quick Start

### 1. Prerequisites

```bash
# Check Docker
docker --version

# Check RabbitMQ is running
docker run -d -p 5672:5672 -p 15672:15672 --name rabbitmq rabbitmq:3-management-alpine
```

### 2. Configuration

Create `.env` file:

```bash
# RabbitMQ
RABBITMQ_URL=amqp://guest:guest@localhost:5672/

# SMTP (Gmail example)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM=noreply@smap.com

# PostgreSQL
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=smap_identity
POSTGRES_USER=postgres
POSTGRES_PASSWORD=password

# Logger
LOGGER_LEVEL=info
LOGGER_MODE=production
LOGGER_ENCODING=json
```

### 3. Run

```bash
# Option 1: Local development
make run-consumer

# Option 2: Docker
make consumer-run

# Option 3: Build script
./build-consumer.sh run
```

---

## Architecture

```
API Server ──► RabbitMQ ──► Consumer Service ──► SMTP Server
   (Async)      (Queue)       (Process)           (Deliver)
```

**Flow:**
1. API publishes message to RabbitMQ
2. Consumer listens on queue
3. Consumer processes message (e.g., send email)
4. Consumer acknowledges message
5. Done

---

## Components

| Component | Description | Location |
|-----------|-------------|----------|
| **Consumer Service** | Orchestration layer | `internal/consumer/` |
| **SMTP Consumer** | Email processing | `internal/smtp/rabbitmq/consumer/` |
| **SMTP UseCase** | Send email logic | `internal/smtp/usecase/` |
| **Email Producer** | Publish messages (in API) | `internal/authentication/delivery/rabbitmq/producer/` |

---

## Commands

### Development

```bash
make run-consumer          # Run locally with Go
make run-api              # Run API server (in another terminal)
```

### Docker - Consumer

```bash
make consumer-build        # Build for local platform
make consumer-build-amd64  # Build for AMD64 servers
make consumer-run          # Build and run with Docker
make consumer-clean        # Remove Docker images
make consumer-push         # Push to registry (requires REGISTRY env)
```

### Docker - API

```bash
make docker-build          # Build API for local
make docker-run            # Build and run API
```

---

## Docker Optimization

Consumer uses **optimized Dockerfile**:

| Feature | Benefit |
|---------|---------|
| Multi-platform build | M4 → AMD64 cross-compile |
| BuildKit cache | 60% faster rebuilds |
| Distroless runtime | ~12MB image (vs 50MB Alpine) |
| Non-root user | Security best practice |
| No shell | Minimal attack surface |

**Performance:**

```bash
# First build: ~4 minutes
make consumer-build

# Cached rebuild: ~45 seconds
# (Only code changed, dependencies cached)
```

---

## Documentation

| Document | Description |
|----------|-------------|
| [CONSUMER_SETUP_GUIDE.md](CONSUMER_SETUP_GUIDE.md) | Complete setup guide |
| [CONSUMER_IMPLEMENTATION_SUMMARY.md](CONSUMER_IMPLEMENTATION_SUMMARY.md) | Implementation details |
| [CONSUMER_FLOW_DIAGRAM.md](CONSUMER_FLOW_DIAGRAM.md) | Sequence diagrams & architecture |

---

## Testing

### Test Email Flow

```bash
# Terminal 1: Start consumer
make run-consumer

# Terminal 2: Trigger email via API
curl -X POST http://localhost:8080/api/v1/authentication/send-otp \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "yourpassword"
  }'

# Check Terminal 1 for consumer logs
# Check email inbox for OTP email
```

### View Logs

```bash
# Docker logs
docker logs -f smap-consumer-dev

# Local logs (stdout)
# Logs appear in terminal where consumer is running
```

---

## Troubleshooting

### Issue: Consumer can't connect to RabbitMQ

```bash
# Check RabbitMQ is running
docker ps | grep rabbitmq

# Check RabbitMQ management UI
open http://localhost:15672
# Login: guest/guest

# Check RABBITMQ_URL in .env
```

### Issue: Email not sent

```bash
# Check SMTP credentials in .env
# For Gmail:
# 1. Enable 2FA: https://myaccount.google.com/security
# 2. Generate App Password: https://myaccount.google.com/apppasswords
# 3. Use App Password as SMTP_PASSWORD
```

### Issue: Messages stuck in queue

```bash
# Check consumer is running
docker ps | grep smap-consumer

# Check consumer logs for errors
docker logs smap-consumer-dev

# Check queue in RabbitMQ UI
open http://localhost:15672/#/queues
```

---

## Monitoring

### Key Metrics

- **Queue Depth**: Should be ~0 (messages processed immediately)
- **Error Rate**: Should be < 5%
- **Email Latency**: Should be < 5 seconds
- **Consumer Memory**: Should be < 100MB

### Check Queue Status

```bash
# RabbitMQ Management UI
open http://localhost:15672/#/queues

# Look for: smtp_send_email
# - Ready: Number of messages waiting
# - Unacked: Number of messages being processed
```

---

## Production Deployment

### 1. Build & Push

```bash
# Set your registry
export REGISTRY=docker.io/yourname

# Build and push
make consumer-push
```

### 2. Deploy

```bash
# Docker
docker run -d \
  --name smap-consumer \
  --env-file .env \
  --restart unless-stopped \
  yourname/smap-consumer:latest

# Docker Compose
docker-compose up -d consumer

# Kubernetes
kubectl apply -f k8s/consumer-deployment.yaml
```

### 3. Verify

```bash
# Check logs
docker logs -f smap-consumer

# Should see:
# "Starting SMAP Consumer Service..."
# "Starting SMTP Email Consumer..."
# "Queue smtp_send_email is being consumed"
```

---

## Security

- **No shell in container** (Distroless)
- **Non-root user** (UID 65532)
- **Secrets via env vars** (not hardcoded)
- **Minimal dependencies** (attack surface)
- **TLS for SMTP** (future)
- **Message encryption** (future)

---

## Scaling

### Horizontal Scaling

```bash
# Run multiple instances
docker run -d --name smap-consumer-1 --env-file .env smap-consumer:latest
docker run -d --name smap-consumer-2 --env-file .env smap-consumer:latest
docker run -d --name smap-consumer-3 --env-file .env smap-consumer:latest

# RabbitMQ will distribute messages round-robin
```

### When to Scale

| Queue Depth | Recommendation |
|-------------|----------------|
| < 100 | 1-2 consumers OK |
| 100-1000 | Add 2-3 consumers |
| 1000+ | Add 5+ consumers or investigate bottleneck |

---

## Development

### Add New Consumer

1. **Create consumer package**:
   ```go
   internal/notification/rabbitmq/consumer/
   ```

2. **Implement Consumer interface**:
   ```go
   func (c Consumer) Consume() {
       go c.consume(notifExc, notifQueue, c.worker)
   }
   ```

3. **Register in service**:
   ```go
   // internal/consumer/new.go
   notifCons := notifConsumer.New(...)
   return &Consumer{
       smtpConsumer: smtpCons,
       notifConsumer: notifCons,  // NEW
   }
   ```

4. **Start consumer**:
   ```go
   // internal/consumer/consumer.go
   c.notifConsumer.Consume()
   ```

---

## Example: Send Email via API

```bash
# 1. User registers
curl -X POST http://localhost:8080/api/v1/authentication/register \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "password123"}'

# 2. Request OTP (triggers email)
curl -X POST http://localhost:8080/api/v1/authentication/send-otp \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "password123"}'

# 3. Check consumer logs
docker logs -f smap-consumer-dev
# Output:
# smtp.delivery.rabbitmq.consumer.sendEmailWorker: {"recipient":"user@example.com",...}
# Email sent successfully

# 4. Check email inbox
# Subject: Email Verification
# Body: Your OTP is: 123456
```

---

## Related Projects

- **API Server**: `cmd/api/` - HTTP API with authentication, plans, subscriptions
- **Producer**: `internal/authentication/delivery/rabbitmq/producer/` - Publishes messages
- **SMTP UseCase**: `internal/smtp/usecase/` - Email sending logic

---

## Tech Stack

- **Go 1.25**: Programming language
- **RabbitMQ**: Message broker
- **SMTP**: Email delivery
- **PostgreSQL**: Database (optional for consumer)
- **Docker**: Containerization
- **Distroless**: Secure runtime
- **go-mail**: SMTP library

---

## Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

---

## License

This project is licensed under the MIT License.

---

## Support

### Get Help

- Read [CONSUMER_SETUP_GUIDE.md](CONSUMER_SETUP_GUIDE.md)
- Check [Troubleshooting](#troubleshooting) section
- Contact: support@smap.com

### Quick Commands

```bash
# Show all available commands
make help

# Consumer commands
make run-consumer
make consumer-build
make consumer-run
make consumer-clean

# API commands  
make run-api
make docker-build
make docker-run
```

---

Happy Consuming!

*Built with love for async task processing*

