# syntax=docker/dockerfile:1

# =============================================================================
# STAGE 1: BUILDER
# =============================================================================
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

WORKDIR /app

ARG TARGETOS
ARG TARGETARCH

# Install build dependencies & timezone data
RUN apk add --no-cache git tzdata

# Cache dependencies (rarely changes → cached layer)
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download && \
    go install github.com/swaggo/swag/cmd/swag@v1.8.12

# Copy source & generate swagger docs
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    swag init -g cmd/api/main.go

# Build static binary
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o api ./cmd/api

# =============================================================================
# STAGE 2: RUNTIME (Distroless Static ~2MB)
# =============================================================================
FROM gcr.io/distroless/static-debian12:nonroot AS runtime

LABEL org.opencontainers.image.title="smap-identity-api" \
      org.opencontainers.image.source="https://github.com/smap-hcmut/identity-srv"

WORKDIR /app

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
ENV TZ=Asia/Ho_Chi_Minh

COPY --from=builder --chown=nonroot:nonroot /app/api .

USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["./api"]