# 修复 Go 模块锁定问题
Write-Host "🔧 修复 Go 模块锁定问题" -ForegroundColor Cyan
Write-Host "========================" -ForegroundColor Cyan

# 检查是否在正确的目录
if (-not (Test-Path "go.mod")) {
    Write-Host "❌ 错误: 请在 go-app 目录下运行此脚本" -ForegroundColor Red
    exit 1
}

Write-Host "✅ 当前目录: $(Get-Location)" -ForegroundColor Green

# 清理 Go 模块缓存
Write-Host "🧹 清理 Go 模块缓存..." -ForegroundColor Yellow
go clean -modcache

# 删除可能损坏的文件
Write-Host "🗑️  删除可能损坏的文件..." -ForegroundColor Yellow
if (Test-Path "go.sum") {
    Remove-Item "go.sum" -Force
}
if (Test-Path "vendor") {
    Remove-Item "vendor" -Recurse -Force
}

# 重新初始化模块
Write-Host "🔄 重新初始化 Go 模块..." -ForegroundColor Yellow
go mod tidy

# 验证模块
Write-Host "✅ 验证模块..." -ForegroundColor Yellow
go mod verify

Write-Host ""
Write-Host "🎉 修复完成！" -ForegroundColor Green
Write-Host ""
Write-Host "现在可以尝试启动服务器：" -ForegroundColor Cyan
Write-Host "go run cmd/server/main.go" -ForegroundColor White 