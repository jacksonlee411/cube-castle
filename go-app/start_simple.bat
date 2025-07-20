@echo off
echo 🏰 Cube Castle - 启动Go服务
echo ================================

echo 🔍 检查Go环境...
go version
if %errorlevel% neq 0 (
    echo ❌ Go未安装或不在PATH中
    pause
    exit /b 1
)

echo 📦 清理缓存...
go clean -cache

echo 🔧 整理依赖...
go mod tidy

echo 🚀 启动Go服务...
echo 💡 服务将在 http://localhost:8080 运行
echo 💡 按Ctrl+C停止服务
echo.

go run cmd/server/main.go

pause 