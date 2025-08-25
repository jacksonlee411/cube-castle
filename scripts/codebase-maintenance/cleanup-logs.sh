#!/bin/bash
# 日志文件自动清理脚本
# 遵循CLAUDE.md的技术债务管控原则

set -e

echo "🧹 开始日志文件清理..."

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# 配置参数
LOG_DIR="/home/shangmeilin/cube-castle/logs"
ARCHIVE_DIR="/home/shangmeilin/cube-castle/archive/logs-backup-$(date +%Y%m%d)"
KEEP_DAYS=7        # 保留最近7天的日志
CRITICAL_KEEP_DAYS=30  # 关键错误日志保留30天

print_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
print_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
print_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# 检查日志目录
if [ ! -d "$LOG_DIR" ]; then
    print_error "日志目录不存在: $LOG_DIR"
    exit 1
fi

# 创建归档目录
mkdir -p "$ARCHIVE_DIR"
print_info "创建归档目录: $ARCHIVE_DIR"

# 统计当前日志文件
TOTAL_FILES=$(find "$LOG_DIR" -name "*.log" | wc -l)
TOTAL_SIZE=$(du -sh "$LOG_DIR" | cut -f1)
print_info "当前日志文件: $TOTAL_FILES 个，总大小: $TOTAL_SIZE"

# 查找需要清理的普通日志文件 (>7天)
OLD_LOGS=$(find "$LOG_DIR" -name "*.log" -mtime +$KEEP_DAYS -not -name "*error*" -not -name "*critical*")

if [ -n "$OLD_LOGS" ]; then
    print_info "找到需要归档的普通日志文件:"
    echo "$OLD_LOGS" | while read -r log_file; do
        if [ -f "$log_file" ]; then
            filename=$(basename "$log_file")
            print_info "  归档: $filename"
            mv "$log_file" "$ARCHIVE_DIR/"
        fi
    done
else
    print_info "没有需要归档的普通日志文件"
fi

# 查找需要清理的关键错误日志 (>30天)
CRITICAL_OLD_LOGS=$(find "$LOG_DIR" -name "*error*.log" -o -name "*critical*.log" -mtime +$CRITICAL_KEEP_DAYS)

if [ -n "$CRITICAL_OLD_LOGS" ]; then
    print_info "找到需要归档的关键日志文件:"
    echo "$CRITICAL_OLD_LOGS" | while read -r log_file; do
        if [ -f "$log_file" ]; then
            filename=$(basename "$log_file")
            print_warn "  归档关键日志: $filename"
            mv "$log_file" "$ARCHIVE_DIR/"
        fi
    done
else
    print_info "没有需要归档的关键日志文件"
fi

# 压缩归档目录
if [ "$(ls -A $ARCHIVE_DIR 2>/dev/null)" ]; then
    print_info "压缩归档日志..."
    tar -czf "${ARCHIVE_DIR}.tar.gz" -C "$(dirname $ARCHIVE_DIR)" "$(basename $ARCHIVE_DIR)"
    rm -rf "$ARCHIVE_DIR"
    print_info "归档完成: ${ARCHIVE_DIR}.tar.gz"
else
    print_info "没有文件需要归档，删除空目录"
    rmdir "$ARCHIVE_DIR"
fi

# 统计清理后状态
AFTER_FILES=$(find "$LOG_DIR" -name "*.log" | wc -l)
AFTER_SIZE=$(du -sh "$LOG_DIR" | cut -f1)
print_info "清理后日志文件: $AFTER_FILES 个，总大小: $AFTER_SIZE"

# 清理超过90天的归档文件
OLD_ARCHIVES=$(find "$(dirname $ARCHIVE_DIR)" -name "logs-backup-*.tar.gz" -mtime +90)
if [ -n "$OLD_ARCHIVES" ]; then
    print_info "删除90天以上的归档文件:"
    echo "$OLD_ARCHIVES" | while read -r archive; do
        print_warn "  删除旧归档: $(basename $archive)"
        rm -f "$archive"
    done
fi

print_info "日志清理完成！"