# 🐳 Docker Build Guide - SMAP Identity API

## 📋 Yêu Cầu

- **Docker**: >= 20.10 (hỗ trợ BuildKit)
- **Docker Buildx**: Đã được kích hoạt (mặc định từ Docker Desktop >= 19.03)

Kiểm tra:
```bash
docker buildx version
# Kết quả: github.com/docker/buildx v0.x.x
```

---

## 🚀 Các Cách Build

### 1. Build Cho Máy Local (Apple Silicon M4 hoặc AMD64)

```bash
# Build cho kiến trúc hiện tại của máy bạn
docker build -t smap-identity:latest -f cmd/api/Dockerfile .

# Hoặc chỉ định rõ platform
docker build --platform linux/arm64 -t smap-identity:latest -f cmd/api/Dockerfile .
```

### 2. Build Cross-Platform (AMD64 cho Server Production)

```bash
# Build cho server Linux AMD64 (từ máy M4)
docker buildx build \
  --platform linux/amd64 \
  -t smap-identity:amd64 \
  -f cmd/api/Dockerfile \
  --load \
  .
```

**Giải thích flags:**
- `--platform linux/amd64`: Target platform (server production thường dùng AMD64)
- `--load`: Load image vào Docker local (để test)
- Nếu muốn push lên registry, thay `--load` bằng `--push`

### 3. Build Multi-Platform (ARM64 + AMD64) - Cho Registry

```bash
# Build cả 2 kiến trúc và push lên Docker Hub/Registry
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t yourusername/smap-identity:latest \
  -t yourusername/smap-identity:v1.0.0 \
  -f cmd/api/Dockerfile \
  --push \
  .
```

**Lưu ý:** Multi-platform build **không thể** dùng `--load` (chỉ dùng được `--push`)

---

## 🏃 Chạy Container

### Development (với logs chi tiết)

```bash
docker run -it --rm \
  -p 8080:8080 \
  -e LOG_LEVEL=debug \
  --name smap-identity-dev \
  smap-identity:latest
```

### Production (với restart policy)

```bash
docker run -d \
  -p 8080:8080 \
  -e LOG_LEVEL=info \
  --restart unless-stopped \
  --name smap-identity-prod \
  smap-identity:latest
```

### Với Environment Variables

```bash
docker run -d \
  -p 8080:8080 \
  --env-file .env \
  --name smap-identity \
  smap-identity:latest
```

---

## 🎯 Best Practices

### 1. Sử Dụng BuildKit Cache

BuildKit cache được enable mặc định trong Dockerfile này qua:
- `--mount=type=cache,target=/go/pkg/mod`
- `--mount=type=cache,target=/root/.cache/go-build`

**Lợi ích:**
- Lần build đầu: ~3-5 phút
- Lần build tiếp theo (chỉ thay đổi code): ~30-60 giây ⚡

### 2. Build Tags Với Git Commit

```bash
# Tự động tag bằng git commit SHA
COMMIT_SHA=$(git rev-parse --short HEAD)
docker build \
  -t smap-identity:${COMMIT_SHA} \
  -t smap-identity:latest \
  -f cmd/api/Dockerfile \
  .
```

### 3. Inspect Image Size

```bash
# Kiểm tra image size sau khi build
docker images smap-identity:latest

# Distroless static thường ~10-15MB (bao gồm binary + timezone)
# So sánh Alpine ~50-70MB
```

### 4. Security Scan

```bash
# Scan image với Docker Scout (nếu có)
docker scout cves smap-identity:latest

# Hoặc dùng Trivy
trivy image smap-identity:latest
```

---

## 🔧 Troubleshooting

### Lỗi: "failed to solve with frontend dockerfile.v0"

**Nguyên nhân:** BuildKit chưa được enable

**Fix:**
```bash
# Enable BuildKit
export DOCKER_BUILDKIT=1

# Hoặc thêm vào ~/.bashrc hoặc ~/.zshrc
echo 'export DOCKER_BUILDKIT=1' >> ~/.zshrc
```

### Lỗi: "multiple platforms feature is currently not supported"

**Nguyên nhân:** Dùng `--load` với multi-platform

**Fix:** Thay `--load` bằng `--push` hoặc build từng platform riêng lẻ

### Lỗi: Cannot connect to database

**Nguyên nhân:** Container không thể kết nối PostgreSQL/RabbitMQ

**Fix:**
```bash
# Chạy trong Docker network
docker network create smap-network

docker run -d \
  --network smap-network \
  --name smap-identity \
  smap-identity:latest
```

---

## 📊 So Sánh Performance

| Metrics | Alpine (Old) | Distroless (New) | Cải thiện |
|---------|--------------|------------------|-----------|
| Image Size | ~65MB | ~12MB | **81% nhỏ hơn** |
| Build Time (1st) | ~5 phút | ~4 phút | 20% nhanh hơn |
| Build Time (cached) | ~2 phút | ~45 giây | **63% nhanh hơn** |
| Attack Surface | Medium | Minimal | **Cao hơn nhiều** |
| Shell Access | ✅ | ❌ | Security trade-off |

---

## 🏗️ CI/CD Integration

### GitHub Actions Example

```yaml
name: Build and Push Docker Image

on:
  push:
    branches: [ main ]
    tags: [ 'v*' ]

jobs:
  docker:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3
      
      - name: Login to Docker Hub
        uses: docker/login-action@v3
        with:
          username: ${{ secrets.DOCKERHUB_USERNAME }}
          password: ${{ secrets.DOCKERHUB_TOKEN }}
      
      - name: Build and push
        uses: docker/build-push-action@v5
        with:
          context: .
          file: ./cmd/api/Dockerfile
          platforms: linux/amd64,linux/arm64
          push: true
          tags: |
            yourusername/smap-identity:latest
            yourusername/smap-identity:${{ github.sha }}
          cache-from: type=registry,ref=yourusername/smap-identity:buildcache
          cache-to: type=registry,ref=yourusername/smap-identity:buildcache,mode=max
```

---

## 🎓 Các Tính Năng Đặc Biệt Của Dockerfile Này

### 1. **Multi-Platform Build Native**
- Build trên M4 (ARM64) native → Cực nhanh
- Cross-compile sang AMD64 server → Không cần QEMU emulation

### 2. **BuildKit Cache Mounts**
- Go modules cache được preserve giữa các builds
- Build cache được reuse → Giảm thời gian build >60%

### 3. **Distroless Static Runtime**
- Không có shell, package manager
- Attack surface tối thiểu
- Image size siêu nhỏ (~2MB base + ~10MB binary)

### 4. **Non-Root User**
- Chạy với UID 65532 (nonroot user của Distroless)
- Security best practice

### 5. **Swagger Auto-Generation**
- Swagger docs được generate trong build time
- Không cần run manual trên local

---

## 📚 Tài Liệu Tham Khảo

- [Docker Buildx](https://docs.docker.com/buildx/working-with-buildx/)
- [Distroless Images](https://github.com/GoogleContainerTools/distroless)
- [Docker BuildKit](https://docs.docker.com/build/buildkit/)
- [Multi-platform builds](https://docs.docker.com/build/building/multi-platform/)

---

**Happy Building! 🚀**

