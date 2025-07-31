#!/bin/bash
echo "停止P2/P3验证服务..."
if [ -f "/tmp/cube_castle_pids.txt" ]; then
    while read pid; do
        if kill -0 $pid 2>/dev/null; then
            echo "停止进程: $pid"
            kill $pid
        fi
    done < /tmp/cube_castle_pids.txt
    rm -f /tmp/cube_castle_pids.txt
fi
echo "服务已停止"
