# Auth Service Deployment Guide

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Google OAuth Setup](#google-oauth-setup)
3. [Google Service Account Setup](#google-service-account-setup)
4. [Generate JWT Keys](#generate-jwt-keys)
5. [Environment Variables](#environment-variables)
6. [Database Setup](#database-setup)
7. [Redis Setup](#redis-setup)
8. [Kafka Setup](#kafka-setup)
9. [Service Keys Configuration](#service-keys-configuration)
10. [Docker Compose (Local Development)](#docker-compose-local-development)
11. [Kubernetes Deployment](#kubernetes-deployment)
12. [Health Checks](#health-checks)
13. [Troubleshooting](#troubleshooting)

---

## Prerequisites

- Go 1.25+
- PostgreSQL 14+
- Redis 7+
- Kafka 3.0+
- Google Workspace account with admin access
- Domain with HTTPS support

---

## Google OAuth Setup

### Step 1: Create Google Cloud Project

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Click **"Create Project"**
3. Enter project name: `smap-auth-service`
4. Click **"Create"**

### Step 2: Enable APIs

1. Navigate to **"APIs & Services" → "Library"**
2. Search and enable:
   - **Google+ API** (for user info)
   - **Admin SDK API** (for Directory API)

### Step 3: Configure OAuth Consent Screen

1. Navigate to **"APIs & Services" → "OAuth consent screen"**
2. Select **"Internal"** (for Google Workspace users only)
3. Fill in application information:
   - **App name**: `SMAP Identity Service`
   - **User support email**: `admin@yourdomain.com`
   - **Developer contact**: `admin@yourdomain.com`
4. Click **"Save and Continue"**

### Step 4: Add Scopes

1. Click **"Add or Remove Scopes"**
2. Add the following scopes:
   ```
   https://www.googleapis.com/auth/userinfo.email
   https://www.googleapis.com/auth/userinfo.profile
   https://www.googleapis.com/auth/admin.directory.group.readonly
   ```
3. Click **"Update"** and **"Save and Continue"**

### Step 5: Create OAuth 2.0 Credentials

1. Navigate to **"APIs & Services" → "Credentials"**
2. Click **"Create Credentials" → "OAuth 2.0 Client ID"**
3. Select **"Web application"**
4. Configure:
   - **Name**: `SMAP Identity Service`
   - **Authorized JavaScript origins**:
     ```
     https://yourdomain.com
     ```
   - **Authorized redirect URIs**:
     ```
     https://yourdomain.com/authentication/callback
     http://localhost:8080/authentication/callback  (for local dev)
     ```
5. Click **"Create"**
6. **Save the Client ID and Client Secret** (you'll need these)

**Screenshot Example**:

```
┌─────────────────────────────────────────────────────┐
│ OAuth 2.0 Client ID created                         │
├─────────────────────────────────────────────────────┤
│ Client ID:                                          │
│ 123456789-abc123def456.apps.googleusercontent.com   │
│                                                     │
│ Client Secret:                                      │
│ GOCSPX-abc123def456ghi789jkl012                     │
└─────────────────────────────────────────────────────┘
```

---

## Google Service Account Setup

### Step 1: Create Service Account

1. Navigate to **"IAM & Admin" → "Service Accounts"**
2. Click **"Create Service Account"**
3. Enter details:
   - **Name**: `smap-directory-api`
   - **Description**: `Service account for accessing Google Directory API`
4. Click **"Create and Continue"**
5. Skip role assignment (not needed for domain-wide delegation)
6. Click **"Done"**

### Step 2: Enable Domain-Wide Delegation

1. Click on the created service account
2. Navigate to **"Details"** tab
3. Scroll to **"Advanced settings"**
4. Click **"Enable Google Workspace Domain-wide Delegation"**
5. **Save the Client ID** (numeric, e.g., `123456789012345678901`)

### Step 3: Create Service Account Key

1. Navigate to **"Keys"** tab
2. Click **"Add Key" → "Create new key"**
3. Select **"JSON"**
4. Click **"Create"**
5. **Save the downloaded JSON file** (you'll need this)

**JSON File Example**:

```json
{
  "type": "service_account",
  "project_id": "smap-auth-service",
  "private_key_id": "abc123...",
  "private_key": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n",
  "client_email": "smap-directory-api@smap-auth-service.iam.gserviceaccount.com",
  "client_id": "123456789012345678901",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://oauth2.googleapis.com/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/..."
}
```

### Step 4: Configure Domain-Wide Delegation in Google Workspace

1. Go to [Google Admin Console](https://admin.google.com/)
2. Navigate to **"Security" → "Access and data control" → "API controls"**
3. Scroll to **"Domain-wide delegation"**
4. Click **"Manage Domain-Wide Delegation"**
5. Click **"Add new"**
6. Enter:
   - **Client ID**: `123456789012345678901` (from Step 2)
   - **OAuth Scopes**:
     ```
     https://www.googleapis.com/auth/admin.directory.group.readonly
     ```
7. Click **"Authorize"**

**Screenshot Example**:

```
┌─────────────────────────────────────────────────────┐
│ Add a new Client ID                                 │
├─────────────────────────────────────────────────────┤
│ Client ID:                                          │
│ [123456789012345678901                          ]   │
│                                                     │
│ OAuth Scopes (comma-delimited):                     │
│ [https://www.googleapis.com/auth/admin.directory...]│
│                                                     │
│                          [Cancel]  [Authorize]      │
└─────────────────────────────────────────────────────┘
```

---

## Generate JWT Keys

### Step 1: Generate RSA Key Pair

```bash
# Create keys directory
mkdir -p keys

# Generate private key (2048-bit RSA)
openssl genrsa -out keys/jwt-private.pem 2048

# Generate public key
openssl rsa -in keys/jwt-private.pem -pubout -out keys/jwt-public.pem

# Set proper permissions
chmod 600 keys/jwt-private.pem
chmod 644 keys/jwt-public.pem
```

### Step 2: Verify Keys

```bash
# Verify private key
openssl rsa -in keys/jwt-private.pem -check

# View public key
cat keys/jwt-public.pem
```

**Output Example**:

```
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAxGOr+H7A...
-----END PUBLIC KEY-----
```

### Step 3: Store Keys Securely

**For Kubernetes**:

```bash
# Create secret
kubectl create secret generic jwt-keys \
  --from-file=private-key=keys/jwt-private.pem \
  --from-file=public-key=keys/jwt-public.pem \
  -n identity-service
```

**For Docker**:

```bash
# Mount as volume
docker run -v $(pwd)/keys:/app/keys identity-service
```

---

## Environment Variables

### Complete Reference

Create `auth-config.yaml`:

```yaml
# Environment Configuration
environment:
  name: production # production, staging, development

# HTTP Server Configuration
http_server:
  host: 0.0.0.0
  port: 8080
  mode: release # release, debug

# Database Configuration
postgres:
  host: postgres.example.com
  port: 5432
  db_name: identity_service
  user: identity_user
  password: <POSTGRES_PASSWORD>
  ssl_mode: require
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 300 # seconds

# Redis Configuration
redis:
  host: redis.example.com
  port: 6379
  password: <REDIS_PASSWORD>
  db: 0 # DB 0 for sessions/groups cache
  # Note: DB 1 is automatically used for token blacklist

# Kafka Configuration
kafka:
  brokers:
    - kafka-1.example.com:9092
    - kafka-2.example.com:9092
    - kafka-3.example.com:9092
  topic: audit.events
  group_id: identity-service-audit-consumer

# OAuth2 Configuration
oauth2:
  client_id: 123456789-abc123def456.apps.googleusercontent.com
  client_secret: GOCSPX-abc123def456ghi789jkl012
  redirect_uri: https://yourdomain.com/authentication/callback
  scopes:
    - https://www.googleapis.com/auth/userinfo.email
    - https://www.googleapis.com/auth/userinfo.profile

# JWT Configuration
jwt:
  algorithm: RS256
  private_key_path: /app/keys/jwt-private.pem
  public_key_path: /app/keys/jwt-public.pem
  issuer: smap-auth-service
  audience:
    - smap-api
  ttl: 28800 # 8 hours in seconds

# Google Workspace Configuration
google_workspace:
  service_account_key: /app/keys/service-account.json
  admin_email: admin@yourdomain.com
  domain: yourdomain.com

# Access Control Configuration
access_control:
  allowed_domains:
    - yourdomain.com
    - partner.com
  blocked_emails:
    - blocked@yourdomain.com
  allowed_redirect_urls:
    - https://yourdomain.com/dashboard
    - https://yourdomain.com/
    - https://app.yourdomain.com
  role_mapping:
    admins@yourdomain.com:
      - ADMIN
    analysts@yourdomain.com:
      - ANALYST
    viewers@yourdomain.com:
      - VIEWER
  default_role: VIEWER

# Session Configuration
session:
  ttl: 28800 # 8 hours
  remember_me_ttl: 604800 # 7 days
  backend: redis

# Blacklist Configuration
blacklist:
  enabled: true
  backend: redis

# Cookie Configuration
cookie:
  name: smap_auth_token
  domain: .yourdomain.com # Note the leading dot for subdomains
  secure: true # HTTPS only
  same_site: Lax
  max_age: 28800 # 8 hours

# Encrypter Configuration (for service keys)
encrypter:
  key: <32_CHARACTER_ENCRYPTION_KEY> # Generate with: openssl rand -base64 32

# Service Keys (for internal API authentication)
service_keys:
  project-service: <ENCRYPTED_SERVICE_KEY>
  ingest-service: <ENCRYPTED_SERVICE_KEY>
  knowledge-service: <ENCRYPTED_SERVICE_KEY>
  notification-service: <ENCRYPTED_SERVICE_KEY>

# Logger Configuration
logger:
  level: info # debug, info, warn, error
  mode: production # production, development
  encoding: json # json, console
  color_enabled: false

# Discord Webhook (optional - for alerts)
discord:
  webhook_id: <DISCORD_WEBHOOK_ID>
  webhook_token: <DISCORD_WEBHOOK_TOKEN>
```

### Environment-Specific Configurations

**Development** (`auth-config.dev.yaml`):

```yaml
environment:
  name: development

http_server:
  mode: debug

postgres:
  host: localhost
  ssl_mode: disable

redis:
  host: localhost
  password: ""

kafka:
  brokers:
    - localhost:9092

oauth2:
  redirect_uri: http://localhost:8080/authentication/callback

cookie:
  domain: localhost
  secure: false

logger:
  level: debug
  mode: development
  encoding: console
  color_enabled: true
```

---

## Database Setup

### Step 1: Create Database

```sql
-- Connect to PostgreSQL
psql -h postgres.example.com -U postgres

-- Create database
CREATE DATABASE identity_service;

-- Create user
CREATE USER identity_user WITH ENCRYPTED PASSWORD '<POSTGRES_PASSWORD>';

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE identity_service TO identity_user;
```

### Step 2: Run Migrations

```bash
# Apply migration
psql -h postgres.example.com -U identity_user -d identity_service -f migration/01_auth_service_schema.sql
```

### Step 3: Verify Tables

```sql
-- Connect to database
psql -h postgres.example.com -U identity_user -d identity_service

-- List tables
\dt

-- Expected output:
--  Schema |    Name     | Type  |     Owner
-- --------+-------------+-------+---------------
--  public | audit_logs  | table | identity_user
--  public | jwt_keys    | table | identity_user
--  public | users       | table | identity_user
```

---

## Redis Setup

### Step 1: Install Redis

**Docker**:

```bash
docker run -d \
  --name redis \
  -p 6379:6379 \
  redis:7-alpine redis-server --requirepass <REDIS_PASSWORD>
```

**Kubernetes**:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis
spec:
  replicas: 1
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
        - name: redis
          image: redis:7-alpine
          args:
            - redis-server
            - --requirepass
            - $(REDIS_PASSWORD)
          env:
            - name: REDIS_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: redis-secret
                  key: password
          ports:
            - containerPort: 6379
```

### Step 2: Verify Connection

```bash
# Test connection
redis-cli -h redis.example.com -p 6379 -a <REDIS_PASSWORD> PING

# Expected output: PONG
```

### Step 3: Configure Persistence (Optional)

```bash
# Enable RDB snapshots
redis-cli -h redis.example.com -p 6379 -a <REDIS_PASSWORD> CONFIG SET save "900 1 300 10 60 10000"

# Enable AOF
redis-cli -h redis.example.com -p 6379 -a <REDIS_PASSWORD> CONFIG SET appendonly yes
```

---

## Kafka Setup

### Step 1: Create Topic

```bash
# Create audit.events topic
kafka-topics.sh --create \
  --bootstrap-server kafka-1.example.com:9092 \
  --topic audit.events \
  --partitions 3 \
  --replication-factor 2 \
  --config retention.ms=7776000000  # 90 days
```

### Step 2: Verify Topic

```bash
# List topics
kafka-topics.sh --list \
  --bootstrap-server kafka-1.example.com:9092

# Describe topic
kafka-topics.sh --describe \
  --bootstrap-server kafka-1.example.com:9092 \
  --topic audit.events
```

### Step 3: Test Producer/Consumer

```bash
# Test producer
echo '{"test": "message"}' | kafka-console-producer.sh \
  --bootstrap-server kafka-1.example.com:9092 \
  --topic audit.events

# Test consumer
kafka-console-consumer.sh \
  --bootstrap-server kafka-1.example.com:9092 \
  --topic audit.events \
  --from-beginning \
  --max-messages 1
```

---

## Service Keys Configuration

### Step 1: Generate Service Keys

```bash
# Generate random key for each service
openssl rand -base64 32

# Example output:
# Xk7mP9qR2sT4vW6yZ8aB1cD3eF5gH7iJ9kL0mN2oP4qR
```

### Step 2: Encrypt Service Keys

```go
// Use the encrypter package
package main

import (
    "fmt"
    "smap-api/pkg/encrypter"
)

func main() {
    // Use the same key from auth-config.yaml
    enc := encrypter.New("<32_CHARACTER_ENCRYPTION_KEY>")

    // Encrypt service key
    plainKey := "Xk7mP9qR2sT4vW6yZ8aB1cD3eF5gH7iJ9kL0mN2oP4qR"
    encryptedKey, err := enc.Encrypt(plainKey)
    if err != nil {
        panic(err)
    }

    fmt.Println("Encrypted key:", encryptedKey)
}
```

### Step 3: Add to Configuration

```yaml
service_keys:
  project-service: "encrypted_key_here"
  ingest-service: "encrypted_key_here"
  knowledge-service: "encrypted_key_here"
  notification-service: "encrypted_key_here"
```

---

## Docker Compose (Local Development)

Create `docker-compose.dev.yml`:

```yaml
version: "3.8"

services:
  postgres:
    image: postgres:14-alpine
    environment:
      POSTGRES_DB: identity_service
      POSTGRES_USER: identity_user
      POSTGRES_PASSWORD: dev_password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migration:/docker-entrypoint-initdb.d

  redis:
    image: redis:7-alpine
    command: redis-server --requirepass dev_password
    ports:
      - "6379:6379"

  kafka:
    image: confluentinc/cp-kafka:7.5.0
    environment:
      KAFKA_BROKER_ID: 1
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
    ports:
      - "9092:9092"
    depends_on:
      - zookeeper

  zookeeper:
    image: confluentinc/cp-zookeeper:7.5.0
    environment:
      ZOOKEEPER_CLIENT_PORT: 2181
      ZOOKEEPER_TICK_TIME: 2000
    ports:
      - "2181:2181"

  auth-service:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./auth-config.dev.yaml:/app/auth-config.yaml
      - ./keys:/app/keys
    depends_on:
      - postgres
      - redis
      - kafka
    environment:
      - CONFIG_FILE=/app/auth-config.yaml

volumes:
  postgres_data:
```

### Run

```bash
# Start all services
docker-compose -f docker-compose.dev.yml up -d

# View logs
docker-compose -f docker-compose.dev.yml logs -f auth-service

# Stop all services
docker-compose -f docker-compose.dev.yml down
```

---

## Kubernetes Deployment

### Step 1: Create Namespace

```bash
kubectl create namespace auth-service
```

### Step 2: Create Secrets

```bash
# JWT keys
kubectl create secret generic jwt-keys \
  --from-file=private-key=keys/jwt-private.pem \
  --from-file=public-key=keys/jwt-public.pem \
  -n auth-service

# Google service account
kubectl create secret generic google-service-account \
  --from-file=service-account.json=keys/service-account.json \
  -n auth-service

# Configuration
kubectl create secret generic auth-config \
  --from-file=auth-config.yaml=auth-config.yaml \
  -n auth-service
```

### Step 3: Create Deployment

`manifests/deployment.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: auth-service
  namespace: auth-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: auth-service
  template:
    metadata:
      labels:
        app: auth-service
    spec:
      containers:
        - name: auth-service
          image: your-registry/auth-service:latest
          ports:
            - containerPort: 8080
          env:
            - name: CONFIG_FILE
              value: /app/config/auth-config.yaml
          volumeMounts:
            - name: config
              mountPath: /app/config
              readOnly: true
            - name: jwt-keys
              mountPath: /app/keys
              readOnly: true
            - name: google-service-account
              mountPath: /app/google
              readOnly: true
          livenessProbe:
            httpGet:
              path: /live
              port: 8080
            initialDelaySeconds: 30
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /ready
              port: 8080
            initialDelaySeconds: 10
            periodSeconds: 5
          resources:
            requests:
              memory: "256Mi"
              cpu: "250m"
            limits:
              memory: "512Mi"
              cpu: "500m"
      volumes:
        - name: config
          secret:
            secretName: auth-config
        - name: jwt-keys
          secret:
            secretName: jwt-keys
        - name: google-service-account
          secret:
            secretName: google-service-account
```

### Step 4: Create Service

`manifests/service.yaml`:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: auth-service
  namespace: auth-service
spec:
  selector:
    app: auth-service
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8080
  type: ClusterIP
```

### Step 5: Create Ingress

`manifests/ingress.yaml`:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: auth-service
  namespace: auth-service
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
spec:
  ingressClassName: nginx
  tls:
    - hosts:
        - yourdomain.com
      secretName: auth-service-tls
  rules:
    - host: yourdomain.com
      http:
        paths:
          - path: /authentication
            pathType: Prefix
            backend:
              service:
                name: auth-service
                port:
                  number: 80
```

### Step 6: Deploy

```bash
# Apply all manifests
kubectl apply -f manifests/

# Check deployment status
kubectl get pods -n auth-service
kubectl get svc -n auth-service
kubectl get ingress -n auth-service

# View logs
kubectl logs -f deployment/auth-service -n auth-service
```

---

## Health Checks

### Endpoints

1. **Health Check**: `GET /health`
   - Returns overall service health
   - Checks: PostgreSQL, Redis, Kafka

2. **Readiness Check**: `GET /ready`
   - Returns if service is ready to accept traffic
   - Used by Kubernetes readiness probe

3. **Liveness Check**: `GET /live`
   - Returns if service is alive
   - Used by Kubernetes liveness probe

### Test Health

```bash
# Health check
curl https://yourdomain.com/health

# Expected response:
{
  "status": "healthy",
  "timestamp": "2026-02-09T10:30:00Z",
  "checks": {
    "postgres": "ok",
    "redis": "ok",
    "kafka": "ok"
  }
}
```

---

## Troubleshooting

### Common Issues

#### 1. OAuth Callback Fails

**Symptom**: `OAUTH_EXCHANGE_FAILED` error

**Solutions**:

- Verify `oauth2.client_id` and `oauth2.client_secret` in config
- Check `oauth2.redirect_uri` matches Google Console configuration
- Ensure redirect URI uses HTTPS in production

#### 2. Google Directory API Fails

**Symptom**: `Failed to fetch user groups`

**Solutions**:

- Verify service account has domain-wide delegation enabled
- Check `google_workspace.admin_email` has admin privileges
- Verify OAuth scopes in Google Admin Console
- Check service account JSON file path

#### 3. JWT Verification Fails

**Symptom**: `401 Unauthorized` on protected endpoints

**Solutions**:

- Verify JWT keys are correctly mounted
- Check `jwt.public_key_path` points to correct file
- Ensure public key is accessible by other services
- Verify JWKS endpoint is reachable

#### 4. Redis Connection Fails

**Symptom**: `Failed to connect to Redis`

**Solutions**:

- Check `redis.host` and `redis.port` in config
- Verify Redis password is correct
- Test Redis connectivity: `redis-cli -h <host> -p <port> -a <password> PING`
- Check firewall rules

#### 5. Kafka Connection Fails

**Symptom**: `Failed to initialize Kafka producer`

**Solutions**:

- Check `kafka.brokers` list in config
- Verify Kafka topic exists: `kafka-topics.sh --list`
- Test Kafka connectivity
- Check Kafka broker logs

### Debug Mode

Enable debug logging:

```yaml
logger:
  level: debug
  mode: development
  encoding: console
  color_enabled: true
```

### View Logs

```bash
# Docker
docker logs -f auth-service

# Kubernetes
kubectl logs -f deployment/auth-service -n auth-service

# Follow specific pod
kubectl logs -f <pod-name> -n auth-service
```

---

## Next Steps

1. **Configure monitoring**: Set up Prometheus metrics
2. **Configure alerting**: Set up Discord/Slack webhooks
3. **Configure backup**: Set up PostgreSQL backups
4. **Configure scaling**: Set up HPA for Kubernetes
5. **Review security**: Run security audit

For more information:

- **API Documentation**: `documents/api-reference.md`
- **Troubleshooting Guide**: `documents/identity-service-troubleshooting.md`
- **Service Integration**: `documents/auth-service-integration.md`
