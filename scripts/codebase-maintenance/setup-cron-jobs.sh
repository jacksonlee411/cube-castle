#!/bin/bash
# 设置代码库维护定时任务
# 遵循CLAUDE.md的持续维护原则

set -e

echo "⏰ 设置代码库维护定时任务..."

PROJECT_ROOT="/home/shangmeilin/cube-castle"

# 定义定时任务
CRON_JOBS="
# Cube Castle 代码库维护任务 - 自动生成于 $(date)
# 每日凌晨2点清理日志文件
0 2 * * * $PROJECT_ROOT/scripts/codebase-maintenance/cleanup-logs.sh >> $PROJECT_ROOT/logs/maintenance.log 2>&1

# 每周日凌晨3点清理测试结果（如果有新的）
0 3 * * 0 find $PROJECT_ROOT/frontend/test-results -type f -name '*.png' -o -name '*.webm' -mtime +7 -delete 2>/dev/null || true

# 每月1号凌晨4点清理超过90天的备份
0 4 1 * * find $PROJECT_ROOT/backup -name '*.tar.gz' -mtime +90 -delete 2>/dev/null || true
"

# 检查当前crontab
if crontab -l >/dev/null 2>&1; then
    echo "当前定时任务:"
    crontab -l | grep -v "Cube Castle" || true
    echo ""
fi

# 创建临时cron文件
TEMP_CRON=$(mktemp)

# 保留现有的非Cube Castle任务
if crontab -l >/dev/null 2>&1; then
    crontab -l | grep -v "Cube Castle" > "$TEMP_CRON" || true
fi

# 添加Cube Castle维护任务
echo "$CRON_JOBS" >> "$TEMP_CRON"

# 安装新的定时任务
crontab "$TEMP_CRON"
rm "$TEMP_CRON"

echo "✅ 定时任务设置完成！"
echo ""
echo "已设置的维护任务:"
echo "📅 每日 02:00 - 清理日志文件"
echo "📅 每周日 03:00 - 清理测试结果"  
echo "📅 每月1号 04:00 - 清理超期备份"
echo ""
echo "查看当前任务: crontab -l"
echo "查看维护日志: tail -f logs/maintenance.log"