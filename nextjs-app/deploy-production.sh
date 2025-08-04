#!/bin/bash

# Production deployment script for Cube Castle Frontend
# This script builds and deploys the Next.js application to production environment

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
IMAGE_NAME="cube-castle-frontend"
CONTAINER_NAME="cube-castle-frontend-prod"
PORT=${PORT:-3000}
BUILD_CONTEXT="."

echo -e "${BLUE}🏰 Cube Castle Frontend Production Deployment${NC}"
echo "================================================="

# Function to print colored status messages
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Check if Docker is running
if ! docker info >/dev/null 2>&1; then
    log_error "Docker is not running. Please start Docker and try again."
    exit 1
fi

log_success "Docker is running"

# Stop and remove existing container if it exists
if docker ps -a --format "table {{.Names}}" | grep -q "^${CONTAINER_NAME}$"; then
    log_info "Stopping existing container: ${CONTAINER_NAME}"
    docker stop ${CONTAINER_NAME} || true
    docker rm ${CONTAINER_NAME} || true
    log_success "Existing container stopped and removed"
fi

# Remove existing image if it exists
if docker images --format "table {{.Repository}}:{{.Tag}}" | grep -q "^${IMAGE_NAME}:latest$"; then
    log_info "Removing existing image: ${IMAGE_NAME}:latest"
    docker rmi ${IMAGE_NAME}:latest || true
    log_success "Existing image removed"
fi

# Build the Docker image
log_info "Building Docker image: ${IMAGE_NAME}:latest"
docker build -t ${IMAGE_NAME}:latest ${BUILD_CONTEXT}
log_success "Docker image built successfully"

# Run the container
log_info "Starting production container on port ${PORT}"
docker run -d \
    --name ${CONTAINER_NAME} \
    -p ${PORT}:3000 \
    --restart unless-stopped \
    -e NODE_ENV=production \
    -e NEXT_TELEMETRY_DISABLED=1 \
    -e CUBE_CASTLE_API_URL=${CUBE_CASTLE_API_URL:-http://localhost:8080} \
    -e CUBE_CASTLE_WS_URL=${CUBE_CASTLE_WS_URL:-ws://localhost:8080} \
    ${IMAGE_NAME}:latest

log_success "Container started successfully"

# Wait for container to be healthy
log_info "Waiting for application to be ready..."
sleep 10

# Check container status
if docker ps --format "table {{.Names}}\t{{.Status}}" | grep -q "${CONTAINER_NAME}.*Up"; then
    log_success "Container is running"
    
    # Test health endpoint
    if curl -f -s http://localhost:${PORT}/api/health >/dev/null 2>&1; then
        log_success "Health check passed"
        echo ""
        echo -e "${GREEN}🎉 Deployment completed successfully!${NC}"
        echo -e "${BLUE}📱 Application is available at: http://localhost:${PORT}${NC}"
        echo -e "${BLUE}🔍 Health check: http://localhost:${PORT}/api/health${NC}"
        echo ""
        echo "Container logs:"
        docker logs ${CONTAINER_NAME} --tail 10
    else
        log_warning "Health check failed, but container is running"
        echo "Check the logs for more information:"
        docker logs ${CONTAINER_NAME}
    fi
else
    log_error "Container failed to start"
    docker logs ${CONTAINER_NAME}
    exit 1
fi

echo ""
echo "🛠️  Useful commands:"
echo "  • View logs: docker logs ${CONTAINER_NAME} -f"
echo "  • Stop container: docker stop ${CONTAINER_NAME}"
echo "  • Remove container: docker rm ${CONTAINER_NAME}"
echo "  • Restart container: docker restart ${CONTAINER_NAME}"