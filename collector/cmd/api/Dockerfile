# syntax=docker/dockerfile:1

##
## STEP 1 - BUILD
##

# specify the base image to  be used for the application, alpine or ubuntu
FROM golang:1.23.8-alpine AS builder

# create a working directory inside the image
WORKDIR /app

# Copy go mod and sum files
COPY go.mod ./

# download all dependencies
RUN go mod tidy
RUN go mod vendor
RUN go install github.com/swaggo/swag/cmd/swag@v1.8.12

# Copy the source code into the container
COPY . .

# compile application
# /api: directory stores binaries file
RUN swag init -g cmd/api/main.go
RUN go build -o api ./cmd/api

##
## STEP 2 - DEPLOY
##
FROM alpine:latest

# Set environment variables
ENV UID=2025
ENV USER=sefthost
ENV GID=2025
ENV GROUP=sefthost

# Create user
RUN CONFIG="\
    --user=$USER \
    --group=$GROUP \
    " && addgroup -g $GID -S $USER \
    && adduser -D -S -h /app -s /bin/bash -G $GROUP -u $UID $USER

# Install ca-certificates and timezone
RUN apk add --no-cache ca-certificates tzdata && update-ca-certificates

# Set the timezone.
ENV TZ=Asia/Ho_Chi_Minh
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# Copy the binary from the builder stage
WORKDIR /app
COPY --from=builder /app/api .

# Expose port 8080 instead of 80 (non-root user can't bind to port < 1024)
EXPOSE 8080

# Run the application as sefthost user
USER sefthost

# Run the application
ENTRYPOINT ["./api"]