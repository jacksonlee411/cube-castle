# 简化启动脚本
Write-Host "🚀 简化启动脚本" -ForegroundColor Cyan
Write-Host "================" -ForegroundColor Cyan

# 检查 Go 版本
Write-Host "📋 检查 Go 版本..." -ForegroundColor Yellow
go version

# 检查是否在正确的目录
if (-not (Test-Path "go.mod")) {
    Write-Host "❌ 错误: 请在 go-app 目录下运行此脚本" -ForegroundColor Red
    exit 1
}

Write-Host "✅ 当前目录: $(Get-Location)" -ForegroundColor Green

# 清理并重新初始化模块
Write-Host "🔄 初始化 Go 模块..." -ForegroundColor Yellow
go mod tidy

# 设置环境变量
$env:APP_PORT = "8080"
$env:INTELLIGENCE_SERVICE_GRPC_TARGET = "localhost:50051"

Write-Host "📝 环境变量设置:" -ForegroundColor Yellow
Write-Host "  APP_PORT=$env:APP_PORT" -ForegroundColor White
Write-Host "  INTELLIGENCE_SERVICE_GRPC_TARGET=$env:INTELLIGENCE_SERVICE_GRPC_TARGET" -ForegroundColor White

# 启动服务器
Write-Host ""
Write-Host "🚀 启动 CoreHR API 服务器..." -ForegroundColor Green
Write-Host "📍 服务地址: http://localhost:$env:APP_PORT" -ForegroundColor Cyan
Write-Host "📋 API 文档: http://localhost:$env:APP_PORT/test.html" -ForegroundColor Cyan
Write-Host "🏥 健康检查: http://localhost:$env:APP_PORT/health" -ForegroundColor Cyan
Write-Host ""
Write-Host "按 Ctrl+C 停止服务器" -ForegroundColor Yellow
Write-Host ""

# 启动服务器
go run cmd/server/main.go 