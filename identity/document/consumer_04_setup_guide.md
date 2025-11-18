# 📧 SMAP Consumer Service Setup Guide

## 🎯 Overview

Consumer Service xử lý các tác vụ bất đồng bộ thông qua RabbitMQ:
- **Email Sending**: Gửi email verification, notifications
- **Future**: Notifications, webhooks, data processing, etc.

---

## 🏗️ Architecture

```
┌─────────────────┐         ┌──────────────┐         ┌─────────────────┐
│   API Server    │ ───────>│   RabbitMQ   │ ───────>│    Consumer     │
│  (Producer)     │ Publish │   Exchange   │ Consume │    Service      │
└─────────────────┘         └──────────────┘         └─────────────────┘
                                   │                          │
                                   │                          │
                                   ▼                          ▼
                            ┌──────────────┐         ┌─────────────────┐
                            │ Email Queue  │         │  SMTP Server    │
                            └──────────────┘         └─────────────────┘
```

**Flow:**
1. API Server publish message vào RabbitMQ Exchange (`smtp_send_email_exc`)
2. Message được route vào Queue (`smtp_send_email`)
3. Consumer Service consume message từ queue
4. Consumer gửi email qua SMTP server
5. Acknowledge message khi gửi thành công

---

## 📦 Components

### 1. **Consumer Service** (`internal/consumer/`)
- Orchestrate tất cả consumers
- Graceful shutdown handling
- Health check support

### 2. **SMTP Consumer** (`internal/smtp/rabbitmq/consumer/`)
- Consume email messages từ RabbitMQ
- Deserialize message data
- Call SMTP UseCase để gửi email

### 3. **SMTP UseCase** (`internal/smtp/usecase/`)
- Business logic gửi email
- Integrate với go-mail library
- Support attachments, CC, HTML body

### 4. **Authentication Producer** (`internal/authentication/delivery/rabbitmq/producer/`)
- Publish email messages từ API
- Serialize message data
- Declare exchange và publish

---

## 🚀 Quick Start

### Development (Local)

```bash
# Option 1: Run directly with Go
make run-consumer

# Option 2: Run with Docker
make consumer-run

# Option 3: Use script directly
./build-consumer.sh run
```

### Production (Docker)

```bash
# Build for AMD64 server
make consumer-build-amd64

# Or use script
./build-consumer.sh amd64

# Run container
docker run -d \
  --name smap-consumer \
  --env-file .env \
  smap-consumer:amd64
```

---

## ⚙️ Configuration

Consumer cần các environment variables sau (trong `.env`):

### Required:

```bash
# PostgreSQL (for potential future features)
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=smap_identity
POSTGRES_USER=postgres
POSTGRES_PASSWORD=your_password

# RabbitMQ
RABBITMQ_URL=amqp://guest:guest@localhost:5672/

# SMTP Configuration
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM=noreply@smap.com

# Logger
LOGGER_LEVEL=info
LOGGER_MODE=production
LOGGER_ENCODING=json
```

### Optional:

```bash
# For Gmail SMTP
# 1. Enable 2FA in Google Account
# 2. Generate App Password: https://myaccount.google.com/apppasswords
# 3. Use App Password as SMTP_PASSWORD
```

---

## 🔌 RabbitMQ Setup

### Exchange & Queue Configuration

**Exchange:**
- Name: `smtp_send_email_exc`
- Type: `fanout`
- Durable: `true`
- Auto-delete: `false`

**Queue:**
- Name: `smtp_send_email`
- Durable: `true`
- Binding: Bound to `smtp_send_email_exc`

**Message Format:**
```json
{
  "subject": "Email Verification",
  "recipient": "user@example.com",
  "body": "<html>...</html>",
  "reply_to": "support@smap.com",
  "cc_addresses": ["admin@smap.com"],
  "attachments": [
    {
      "filename": "document.pdf",
      "content_type": "application/pdf",
      "data": "base64_encoded_data"
    }
  ]
}
```

### Test RabbitMQ Connection

```bash
# Check RabbitMQ is running
curl http://localhost:15672/api/overview

# Login to RabbitMQ Management
# http://localhost:15672
# Username: guest
# Password: guest
```

---

## 🐳 Docker Build

### Local Platform

```bash
make consumer-build
# Or
./build-consumer.sh local
```

### AMD64 (Production Server)

```bash
make consumer-build-amd64
# Or
./build-consumer.sh amd64
```

### Multi-Platform + Push to Registry

```bash
export REGISTRY=docker.io/yourname
make consumer-push
# Or
./build-consumer.sh push
```

---

## 📊 Monitoring & Logs

### View Logs

```bash
# If running with Docker
docker logs -f smap-consumer-dev

# If running locally
# Logs output to stdout
```

### Log Format

```json
{
  "level": "info",
  "ts": 1234567890.123,
  "caller": "consumer/consumer.go:25",
  "msg": "Queue smtp_send_email is being consumed"
}
```

### Key Log Messages

- ✅ `Starting SMAP Consumer Service...` - Service started
- ✅ `Starting SMTP Email Consumer...` - Email consumer started
- ✅ `Queue smtp_send_email is being consumed` - Listening for messages
- ✅ `smtp.delivery.rabbitmq.consumer.sendEmailWorker` - Processing message
- ❌ `Failed to connect to RabbitMQ` - RabbitMQ connection error
- ❌ `smtp.usecase.SendEmail.DialAndSend` - SMTP send error

---

## 🧪 Testing

### 1. Test Consumer Locally

```bash
# Terminal 1: Start consumer
make run-consumer

# Terminal 2: Trigger API endpoint that sends email
curl -X POST http://localhost:8080/api/v1/authentication/send-otp \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "yourpassword"
  }'

# Check Terminal 1 for consumer logs
```

### 2. Test with Docker

```bash
# Start consumer in Docker
make consumer-run

# View logs
docker logs -f smap-consumer-dev

# Trigger API
curl -X POST http://localhost:8080/api/v1/authentication/send-otp ...
```

### 3. Manual RabbitMQ Test

```bash
# Publish test message to RabbitMQ
# Use RabbitMQ Management UI: http://localhost:15672
# Go to Exchanges → smtp_send_email_exc → Publish message
```

---

## 🔧 Troubleshooting

### Issue 1: Consumer không connect được RabbitMQ

**Symptoms:**
```
Failed to connect to RabbitMQ: dial tcp: connection refused
```

**Solutions:**
- Check RabbitMQ đang chạy: `docker ps | grep rabbitmq`
- Check `RABBITMQ_URL` trong `.env`
- Verify port 5672 (AMQP) và 15672 (Management) có mở

### Issue 2: Email không được gửi

**Symptoms:**
```
smtp.usecase.SendEmail.DialAndSend: 535 Authentication failed
```

**Solutions:**
- Check SMTP credentials trong `.env`
- Nếu dùng Gmail:
  - Enable 2FA
  - Generate App Password
  - Use App Password thay vì Gmail password
- Check SMTP port (587 cho TLS, 465 cho SSL)

### Issue 3: Consumer không consume messages

**Symptoms:**
- Consumer started nhưng không log messages
- Messages stuck trong queue

**Solutions:**
- Check queue binding: `smtp_send_email` → `smtp_send_email_exc`
- Verify exchange type là `fanout`
- Check consumer có log: `Queue smtp_send_email is being consumed`
- Restart consumer service

### Issue 4: Message format không đúng

**Symptoms:**
```
json.Unmarshal: invalid character ...
```

**Solutions:**
- Check producer publish đúng format JSON
- Verify `ContentType` = `application/json`
- Check encoding của message body

---

## 📈 Performance Tips

### 1. Multiple Consumer Instances

Chạy nhiều consumer instances để tăng throughput:

```bash
# Instance 1
docker run -d --name smap-consumer-1 --env-file .env smap-consumer:latest

# Instance 2
docker run -d --name smap-consumer-2 --env-file .env smap-consumer:latest

# RabbitMQ sẽ round-robin distribute messages
```

### 2. Prefetch Count

Adjust prefetch để optimize performance (future enhancement):

```go
// In consumer/common.go
ch.Qos(
    10,    // prefetch count
    0,     // prefetch size
    false, // global
)
```

### 3. Connection Pooling

SMTP connection pooling (đã có sẵn trong go-mail):
- Reuse SMTP connections
- Reduce handshake overhead

---

## 🔐 Security Best Practices

1. **No Shell in Container**
   - Distroless image → No shell access
   - Minimal attack surface

2. **Non-Root User**
   - Container runs as UID 65532 (nonroot)
   - Cannot modify system files

3. **Secrets Management**
   - Never commit `.env` to git
   - Use Kubernetes Secrets in production
   - Rotate SMTP passwords regularly

4. **Network Isolation**
   - Consumer only needs access to:
     - RabbitMQ
     - SMTP server
     - PostgreSQL (optional)
   - Use Docker networks to isolate

---

## 📚 Additional Resources

- [RabbitMQ Best Practices](https://www.rabbitmq.com/best-practices.html)
- [Go-Mail Documentation](https://github.com/go-mail/mail)
- [Gmail App Passwords](https://support.google.com/accounts/answer/185833)
- [Docker Distroless](https://github.com/GoogleContainerTools/distroless)

---

## 🆘 Support

### Check Service Health

```bash
# Check if consumer is running
docker ps | grep smap-consumer

# Check logs for errors
docker logs smap-consumer-dev --tail 100

# Check RabbitMQ queue
# http://localhost:15672/#/queues
# Look for smtp_send_email queue
```

### Common Commands

```bash
# Start consumer
make consumer-run

# View logs
docker logs -f smap-consumer-dev

# Stop consumer
docker stop smap-consumer-dev

# Restart consumer
docker restart smap-consumer-dev

# Clean up
make consumer-clean
```

---

**Happy Consuming! 📧**

