# 测试Repository实现的PowerShell脚本
Write-Host "🧪 测试 CoreHR Repository 实现" -ForegroundColor Green

# 检查数据库连接
Write-Host "📊 检查数据库连接..." -ForegroundColor Yellow
try {
    $dbTest = Invoke-WebRequest -Uri "http://localhost:8080/health" -UseBasicParsing -TimeoutSec 5
    Write-Host "✅ 数据库连接正常" -ForegroundColor Green
} catch {
    Write-Host "❌ 数据库未连接，请先启动数据库" -ForegroundColor Red
    exit 1
}

# 编译项目
Write-Host "🔨 编译项目..." -ForegroundColor Yellow
Set-Location go-app
go build -o server.exe cmd/server/main.go

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ 编译失败" -ForegroundColor Red
    exit 1
}

Write-Host "✅ 编译成功" -ForegroundColor Green

# 启动服务器（后台运行）
Write-Host "🚀 启动服务器..." -ForegroundColor Yellow
Start-Process -FilePath ".\server.exe" -WindowStyle Hidden
$serverProcess = Get-Process -Name "server" -ErrorAction SilentlyContinue

# 等待服务器启动
Start-Sleep -Seconds 3

# 测试API端点
Write-Host "🌐 测试API端点..." -ForegroundColor Yellow

# 测试健康检查
Write-Host "📋 测试健康检查..." -ForegroundColor Cyan
try {
    $healthResponse = Invoke-RestMethod -Uri "http://localhost:8080/health" -Method Get
    $healthResponse | ConvertTo-Json -Depth 3
} catch {
    Write-Host "❌ 健康检查失败: $($_.Exception.Message)" -ForegroundColor Red
}

# 测试员工列表
Write-Host "👥 测试员工列表..." -ForegroundColor Cyan
try {
    $employeesResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/employees?page=1&pageSize=10" -Method Get
    $employeesResponse | ConvertTo-Json -Depth 3
} catch {
    Write-Host "❌ 员工列表测试失败: $($_.Exception.Message)" -ForegroundColor Red
}

# 测试创建员工
Write-Host "➕ 测试创建员工..." -ForegroundColor Cyan
$createEmployeeBody = @{
    employee_number = "TEST001"
    first_name = "测试"
    last_name = "员工"
    email = "test@example.com"
    phone_number = "13800138000"
    position = "软件工程师"
    department = "技术部"
    hire_date = "2024-01-01"
} | ConvertTo-Json

try {
    $createResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/employees" -Method Post -Body $createEmployeeBody -ContentType "application/json"
    $createResponse | ConvertTo-Json -Depth 3
    
    $employeeId = $createResponse.id
    if ($employeeId) {
        Write-Host "✅ 员工创建成功，ID: $employeeId" -ForegroundColor Green
        
        # 测试获取员工详情
        Write-Host "👤 测试获取员工详情..." -ForegroundColor Cyan
        try {
            $employeeResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/employees/$employeeId" -Method Get
            $employeeResponse | ConvertTo-Json -Depth 3
        } catch {
            Write-Host "❌ 获取员工详情失败: $($_.Exception.Message)" -ForegroundColor Red
        }
        
        # 测试更新员工
        Write-Host "✏️ 测试更新员工..." -ForegroundColor Cyan
        $updateEmployeeBody = @{
            first_name = "更新后的名字"
            phone_number = "13900139000"
        } | ConvertTo-Json
        
        try {
            $updateResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/employees/$employeeId" -Method Put -Body $updateEmployeeBody -ContentType "application/json"
            $updateResponse | ConvertTo-Json -Depth 3
        } catch {
            Write-Host "❌ 更新员工失败: $($_.Exception.Message)" -ForegroundColor Red
        }
        
        # 测试删除员工
        Write-Host "🗑️ 测试删除员工..." -ForegroundColor Cyan
        try {
            Invoke-RestMethod -Uri "http://localhost:8080/api/v1/employees/$employeeId" -Method Delete
            Write-Host "✅ 员工删除成功" -ForegroundColor Green
        } catch {
            Write-Host "❌ 删除员工失败: $($_.Exception.Message)" -ForegroundColor Red
        }
    } else {
        Write-Host "❌ 员工创建失败" -ForegroundColor Red
    }
} catch {
    Write-Host "❌ 创建员工失败: $($_.Exception.Message)" -ForegroundColor Red
}

# 测试组织列表
Write-Host "🏢 测试组织列表..." -ForegroundColor Cyan
try {
    $organizationsResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/organizations" -Method Get
    $organizationsResponse | ConvertTo-Json -Depth 3
} catch {
    Write-Host "❌ 组织列表测试失败: $($_.Exception.Message)" -ForegroundColor Red
}

# 测试组织树
Write-Host "🌳 测试组织树..." -ForegroundColor Cyan
try {
    $treeResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/organizations/tree" -Method Get
    $treeResponse | ConvertTo-Json -Depth 3
} catch {
    Write-Host "❌ 组织树测试失败: $($_.Exception.Message)" -ForegroundColor Red
}

# 停止服务器
Write-Host "🛑 停止服务器..." -ForegroundColor Yellow
if ($serverProcess) {
    Stop-Process -Id $serverProcess.Id -Force
}

Write-Host "✅ 测试完成" -ForegroundColor Green 