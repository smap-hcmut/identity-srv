#!/bin/bash

# ==============================================================================
# SMAP Consumer Service - Docker Build Helper Script
# ==============================================================================

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
IMAGE_NAME="smap-consumer"
DOCKERFILE="cmd/consumer/Dockerfile"
REGISTRY="" # Set this if pushing to registry (e.g., "docker.io/username")

# ==============================================================================
# Helper Functions
# ==============================================================================

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_usage() {
    cat << EOF
${GREEN}SMAP Consumer Service - Docker Build Script${NC}

Usage: ./build-consumer.sh [OPTIONS]

Options:
    local           Build for local development (current platform)
    amd64           Build for AMD64 (Linux servers)
    arm64           Build for ARM64 (Apple Silicon, AWS Graviton)
    multi           Build for both AMD64 and ARM64 (requires push)
    push            Build and push to registry (multi-platform)
    run             Build and run container locally
    clean           Remove all smap-consumer images
    help            Show this help message

Examples:
    ./build-consumer.sh local                    # Build for current platform
    ./build-consumer.sh amd64                    # Build for AMD64 server
    ./build-consumer.sh multi                    # Build multi-platform (requires registry)
    ./build-consumer.sh run                      # Build and run

Environment Variables:
    REGISTRY        Docker registry (default: empty for local)
    IMAGE_NAME      Image name (default: smap-consumer)
    TAG             Image tag (default: latest)

EOF
}

check_docker() {
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed"
        exit 1
    fi
    
    if ! docker buildx version &> /dev/null; then
        print_error "Docker Buildx is not available"
        exit 1
    fi
    
    print_success "Docker and Buildx are available"
}

get_git_info() {
    if git rev-parse --git-dir > /dev/null 2>&1; then
        GIT_COMMIT=$(git rev-parse --short HEAD)
        GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
        print_info "Git: Branch=${GIT_BRANCH}, Commit=${GIT_COMMIT}"
    else
        GIT_COMMIT="unknown"
        GIT_BRANCH="unknown"
        print_warning "Not a git repository"
    fi
}

# ==============================================================================
# Build Functions
# ==============================================================================

build_local() {
    print_info "Building consumer for local platform..."
    
    TAG="${TAG:-latest}"
    
    docker build \
        -t ${IMAGE_NAME}:${TAG} \
        -t ${IMAGE_NAME}:${GIT_COMMIT} \
        -f ${DOCKERFILE} \
        .
    
    print_success "Build completed: ${IMAGE_NAME}:${TAG}"
    docker images ${IMAGE_NAME}:${TAG}
}

build_amd64() {
    print_info "Building consumer for AMD64 platform..."
    
    TAG="${TAG:-amd64}"
    
    docker buildx build \
        --platform linux/amd64 \
        -t ${IMAGE_NAME}:${TAG} \
        -t ${IMAGE_NAME}:${GIT_COMMIT}-amd64 \
        -f ${DOCKERFILE} \
        --load \
        .
    
    print_success "Build completed: ${IMAGE_NAME}:${TAG}"
    docker images ${IMAGE_NAME}:${TAG}
}

build_arm64() {
    print_info "Building consumer for ARM64 platform..."
    
    TAG="${TAG:-arm64}"
    
    docker buildx build \
        --platform linux/arm64 \
        -t ${IMAGE_NAME}:${TAG} \
        -t ${IMAGE_NAME}:${GIT_COMMIT}-arm64 \
        -f ${DOCKERFILE} \
        --load \
        .
    
    print_success "Build completed: ${IMAGE_NAME}:${TAG}"
    docker images ${IMAGE_NAME}:${TAG}
}

build_multi() {
    if [ -z "$REGISTRY" ]; then
        print_error "REGISTRY is required for multi-platform build"
        print_info "Set REGISTRY environment variable (e.g., export REGISTRY=docker.io/username)"
        exit 1
    fi
    
    print_info "Building multi-platform consumer images..."
    
    TAG="${TAG:-latest}"
    FULL_IMAGE_NAME="${REGISTRY}/${IMAGE_NAME}"
    
    docker buildx build \
        --platform linux/amd64,linux/arm64 \
        -t ${FULL_IMAGE_NAME}:${TAG} \
        -t ${FULL_IMAGE_NAME}:${GIT_COMMIT} \
        -f ${DOCKERFILE} \
        --push \
        .
    
    print_success "Multi-platform build completed and pushed: ${FULL_IMAGE_NAME}:${TAG}"
}

build_and_push() {
    if [ -z "$REGISTRY" ]; then
        print_error "REGISTRY is required for push"
        print_info "Set REGISTRY environment variable (e.g., export REGISTRY=docker.io/username)"
        exit 1
    fi
    
    print_info "Building and pushing consumer to registry..."
    
    TAG="${TAG:-latest}"
    FULL_IMAGE_NAME="${REGISTRY}/${IMAGE_NAME}"
    
    docker buildx build \
        --platform linux/amd64,linux/arm64 \
        -t ${FULL_IMAGE_NAME}:${TAG} \
        -t ${FULL_IMAGE_NAME}:${GIT_COMMIT} \
        -t ${FULL_IMAGE_NAME}:latest \
        -f ${DOCKERFILE} \
        --push \
        .
    
    print_success "Build and push completed: ${FULL_IMAGE_NAME}:${TAG}"
}

build_and_run() {
    print_info "Building consumer for local and running..."
    
    build_local
    
    print_info "Starting consumer container..."
    
    # Stop existing container if running
    docker stop smap-consumer-dev 2>/dev/null || true
    docker rm smap-consumer-dev 2>/dev/null || true
    
    # Note: Consumer needs .env file with RabbitMQ and SMTP config
    if [ ! -f ".env" ]; then
        print_warning ".env file not found. Consumer needs RabbitMQ and SMTP configuration."
        print_info "Create .env file with required variables before running."
        exit 1
    fi
    
    docker run -d \
        --name smap-consumer-dev \
        --env-file .env \
        ${IMAGE_NAME}:latest
    
    print_success "Consumer container started: smap-consumer-dev"
    
    echo ""
    print_info "View logs: docker logs -f smap-consumer-dev"
    print_info "Stop: docker stop smap-consumer-dev"
}

clean_images() {
    print_warning "Cleaning all ${IMAGE_NAME} images..."
    
    docker images | grep ${IMAGE_NAME} | awk '{print $3}' | xargs docker rmi -f 2>/dev/null || true
    
    print_success "Cleanup completed"
    docker images | grep ${IMAGE_NAME} || print_info "No ${IMAGE_NAME} images found"
}

# ==============================================================================
# Main
# ==============================================================================

main() {
    check_docker
    get_git_info
    
    case "${1:-help}" in
        local)
            build_local
            ;;
        amd64)
            build_amd64
            ;;
        arm64)
            build_arm64
            ;;
        multi)
            build_multi
            ;;
        push)
            build_and_push
            ;;
        run)
            build_and_run
            ;;
        clean)
            clean_images
            ;;
        help|--help|-h)
            print_usage
            ;;
        *)
            print_error "Unknown option: $1"
            print_usage
            exit 1
            ;;
    esac
}

main "$@"

