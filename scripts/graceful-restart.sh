#!/bin/bash

# Graceful Server Restart Script
# 更安全的服务器重启脚本

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
GO_APP_DIR="$PROJECT_ROOT/go-app"
SERVER_BIN="$GO_APP_DIR/bin/server"
PID_FILE="/tmp/cube-castle-server.pid"
LOG_FILE="/tmp/cube-castle-restart.log"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# Function to find server process
find_server_process() {
    pgrep -f "cube-castle.*server" || pgrep -f "./server" || echo ""
}

# Function to check if server is running
is_server_running() {
    local pid=$(find_server_process)
    if [ -n "$pid" ]; then
        return 0  # Running
    else
        return 1  # Not running
    fi
}

# Function to gracefully stop server
graceful_stop() {
    local pid=$(find_server_process)
    if [ -n "$pid" ]; then
        log "Found server process: $pid"
        log "Sending SIGTERM for graceful shutdown..."
        
        # Send SIGTERM signal for graceful shutdown
        kill -TERM "$pid" 2>/dev/null || {
            log "Failed to send SIGTERM to process $pid"
            return 1
        }
        
        # Wait for graceful shutdown (max 35 seconds, server has 30s timeout)
        local count=0
        while [ $count -lt 35 ]; do
            if ! kill -0 "$pid" 2>/dev/null; then
                log "Server stopped gracefully after ${count}s"
                return 0
            fi
            sleep 1
            count=$((count + 1))
        done
        
        # If still running after timeout, force kill
        log "Server didn't stop gracefully, forcing shutdown..."
        kill -KILL "$pid" 2>/dev/null || true
        sleep 2
        
        if kill -0 "$pid" 2>/dev/null; then
            log "ERROR: Failed to stop server process $pid"
            return 1
        else
            log "Server force-stopped successfully"
            return 0
        fi
    else
        log "No server process found"
        return 0
    fi
}

# Function to start server
start_server() {
    cd "$GO_APP_DIR"
    
    # Check if binary exists and is newer than source
    if [ ! -f "$SERVER_BIN" ]; then
        log "Server binary not found, building..."
        if ! go build -o bin/server cmd/server/main.go; then
            log "ERROR: Failed to build server"
            return 1
        fi
    fi
    
    log "Starting server..."
    nohup "$SERVER_BIN" > /tmp/cube-castle-server.log 2>&1 &
    local new_pid=$!
    
    # Save PID
    echo "$new_pid" > "$PID_FILE"
    log "Server started with PID: $new_pid"
    
    # Wait a moment and verify it's running
    sleep 3
    if kill -0 "$new_pid" 2>/dev/null; then
        log "Server startup verified successfully"
        return 0
    else
        log "ERROR: Server failed to start"
        return 1
    fi
}

# Function to rebuild and restart
rebuild_and_restart() {
    cd "$GO_APP_DIR"
    
    log "Building latest version..."
    if ! go build -o bin/server cmd/server/main.go; then
        log "ERROR: Build failed, aborting restart"
        return 1
    fi
    
    log "Build successful, proceeding with restart..."
    
    # Stop existing server
    if ! graceful_stop; then
        log "ERROR: Failed to stop existing server"
        return 1
    fi
    
    # Start new server
    if ! start_server; then
        log "ERROR: Failed to start new server"
        return 1
    fi
    
    log "✅ Server restart completed successfully"
    return 0
}

# Main script logic
case "${1:-restart}" in
    "stop")
        log "=== Stopping Server ==="
        graceful_stop
        ;;
    "start")
        log "=== Starting Server ==="
        if is_server_running; then
            log "Server is already running"
            exit 1
        fi
        start_server
        ;;
    "restart")
        log "=== Graceful Server Restart ==="
        rebuild_and_restart
        ;;
    "status")
        if is_server_running; then
            local pid=$(find_server_process)
            log "Server is running (PID: $pid)"
            # Check if server is responding
            if curl -sf http://localhost:8080/health > /dev/null 2>&1; then
                log "Server health check: ✅ HEALTHY"
            else
                log "Server health check: ❌ UNHEALTHY"
            fi
        else
            log "Server is not running"
        fi
        ;;
    "rebuild")
        log "=== Rebuilding Server ==="
        cd "$GO_APP_DIR"
        go build -o bin/server cmd/server/main.go
        log "Build completed"
        ;;
    *)
        echo "Usage: $0 {start|stop|restart|status|rebuild}"
        echo "  start   - Start the server"
        echo "  stop    - Gracefully stop the server"
        echo "  restart - Rebuild and gracefully restart (default)"
        echo "  status  - Check server status and health"
        echo "  rebuild - Rebuild binary without restart"
        exit 1
        ;;
esac