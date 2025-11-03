#!/bin/bash

# 时态数据运维任务cron设置脚本
# 用途：设置cron任务来定期执行数据维护脚本
# 执行：sudo bash setup-cron.sh

# 检查是否以root权限运行
if [[ $EUID -ne 0 ]]; then
   echo "此脚本需要root权限运行，请使用: sudo bash setup-cron.sh" 
   exit 1
fi

# 获取脚本所在目录
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# 数据库连接配置
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-cubecastle}"
DB_USER="${DB_USER:-user}"
DB_PASSWORD="${DB_PASSWORD:-password}"

# 日志目录
LOG_DIR="/var/log/cubecastle"
mkdir -p "$LOG_DIR"

# 创建cron执行脚本
cat > /usr/local/bin/cubecastle-daily-cutover.sh << 'EOL'
#!/bin/bash

# 每日cutover任务执行脚本
# 自动生成 - 请勿手动修改

# 配置变量
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-cubecastle}"
DB_USER="${DB_USER:-user}"
DB_PASSWORD="${DB_PASSWORD:-password}"
LOG_DIR="/var/log/cubecastle"
SCRIPT_DIR="/opt/cubecastle/scripts"

# 创建日志文件名（包含日期）
LOG_FILE="$LOG_DIR/daily-cutover-$(date +%Y%m%d).log"
ERROR_LOG="$LOG_DIR/daily-cutover-error.log"

# 记录开始时间
echo "========================================" >> "$LOG_FILE"
echo "开始执行每日cutover任务: $(date '+%Y-%m-%d %H:%M:%S')" >> "$LOG_FILE"
echo "========================================" >> "$LOG_FILE"

# 执行cutover SQL脚本
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
    -f "$SCRIPT_DIR/daily-cutover.sql" \
    >> "$LOG_FILE" 2>> "$ERROR_LOG"

CUTOVER_EXIT_CODE=$?

# 等待一分钟，然后执行一致性检查
sleep 60

# 执行一致性检查
echo "========================================" >> "$LOG_FILE"
echo "开始执行数据一致性检查: $(date '+%Y-%m-%d %H:%M:%S')" >> "$LOG_FILE"
echo "========================================" >> "$LOG_FILE"

PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
    -f "$SCRIPT_DIR/data-consistency-check.sql" \
    >> "$LOG_FILE" 2>> "$ERROR_LOG"

CHECK_EXIT_CODE=$?

# 记录完成时间和结果
echo "========================================" >> "$LOG_FILE"
echo "任务完成时间: $(date '+%Y-%m-%d %H:%M:%S')" >> "$LOG_FILE"
echo "Cutover退出代码: $CUTOVER_EXIT_CODE" >> "$LOG_FILE"
echo "一致性检查退出代码: $CHECK_EXIT_CODE" >> "$LOG_FILE"

# 如果有任何错误，发送系统通知
if [ $CUTOVER_EXIT_CODE -ne 0 ] || [ $CHECK_EXIT_CODE -ne 0 ]; then
    echo "错误: 时态数据维护任务失败" >> "$ERROR_LOG"
    # 可以在这里添加邮件或其他通知机制
    logger "CubeCastle: 时态数据维护任务失败，请检查日志文件 $ERROR_LOG"
fi

echo "========================================" >> "$LOG_FILE"

# 日志清理：保留最近30天的日志
find "$LOG_DIR" -name "daily-cutover-*.log" -mtime +30 -delete

exit $((CUTOVER_EXIT_CODE + CHECK_EXIT_CODE))
EOL

# 创建一致性检查脚本
cat > /usr/local/bin/cubecastle-consistency-check.sh << 'EOL'
#!/bin/bash

# 数据一致性检查脚本
# 自动生成 - 请勿手动修改

# 配置变量
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-cubecastle}"
DB_USER="${DB_USER:-user}"
DB_PASSWORD="${DB_PASSWORD:-password}"
LOG_DIR="/var/log/cubecastle"
SCRIPT_DIR="/opt/cubecastle/scripts"

# 创建日志文件名
LOG_FILE="$LOG_DIR/consistency-check-$(date +%Y%m%d-%H%M).log"
ERROR_LOG="$LOG_DIR/consistency-check-error.log"

echo "开始执行一致性检查: $(date '+%Y-%m-%d %H:%M:%S')" >> "$LOG_FILE"

# 执行一致性检查
PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
    -f "$SCRIPT_DIR/data-consistency-check.sql" \
    >> "$LOG_FILE" 2>> "$ERROR_LOG"

EXIT_CODE=$?

echo "检查完成: $(date '+%Y-%m-%d %H:%M:%S'), 退出代码: $EXIT_CODE" >> "$LOG_FILE"

# 日志清理：保留最近7天的检查日志
find "$LOG_DIR" -name "consistency-check-*.log" -mtime +7 -delete

exit $EXIT_CODE
EOL

# 设置执行权限
chmod +x /usr/local/bin/cubecastle-daily-cutover.sh
chmod +x /usr/local/bin/cubecastle-consistency-check.sh

# 复制SQL脚本到系统目录
mkdir -p /opt/cubecastle/scripts
cp "$SCRIPT_DIR/daily-cutover.sql" /opt/cubecastle/scripts/
cp "$SCRIPT_DIR/data-consistency-check.sql" /opt/cubecastle/scripts/

# 创建环境变量配置文件
cat > /etc/default/cubecastle << EOF
# CubeCastle 数据库配置
DB_HOST=$DB_HOST
DB_PORT=$DB_PORT
DB_NAME=$DB_NAME
DB_USER=$DB_USER
DB_PASSWORD=$DB_PASSWORD
EOF

# 设置文件权限（保护密码）
chmod 600 /etc/default/cubecastle

# 创建cron任务
echo "设置cron任务..."

# 添加到系统crontab
cat > /etc/cron.d/cubecastle-temporal << 'EOF'
# CubeCastle 时态数据维护任务
# 每天凌晨2点执行daily cutover任务
0 2 * * * root /bin/bash -l -c '. /etc/default/cubecastle && /usr/local/bin/cubecastle-daily-cutover.sh'

# 每4小时执行一次数据一致性检查
0 */4 * * * root /bin/bash -l -c '. /etc/default/cubecastle && /usr/local/bin/cubecastle-consistency-check.sh'

EOF

# 重新加载cron服务
systemctl reload cron 2>/dev/null || service cron reload 2>/dev/null || echo "警告: 无法重新加载cron服务"

echo "✅ Cron任务设置完成！"
echo ""
echo "已创建以下任务："
echo "  - 每日凌晨2:00: 时态数据cutover维护"
echo "  - 每4小时: 数据一致性检查"
echo ""
echo "日志目录: $LOG_DIR"
echo "配置文件: /etc/default/cubecastle"
echo "执行脚本: /usr/local/bin/cubecastle-*.sh"
echo ""
echo "查看已设置的cron任务:"
echo "  sudo crontab -l"
echo "  sudo cat /etc/cron.d/cubecastle-temporal"
echo ""
echo "手动测试任务:"
echo "  sudo /usr/local/bin/cubecastle-daily-cutover.sh"
echo "  sudo /usr/local/bin/cubecastle-consistency-check.sh"