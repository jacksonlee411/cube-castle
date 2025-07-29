# Cube Castle 脚本开发规范

## 🎯 核心原则

**只使用Bash脚本，不创建PowerShell脚本**

## 📋 脚本开发规范

### 1. **脚本类型限制**
- ✅ **允许**: Bash脚本 (`.sh`后缀)
- ❌ **禁止**: PowerShell脚本 (`.ps1`后缀)
- ❌ **禁止**: Windows批处理文件 (`.bat`后缀)

### 2. **命名规范**
```
✅ 正确命名示例:
- test_api.sh
- start.sh
- build.sh
- deploy.sh
- verify_implementation.sh

❌ 错误命名示例:
- test_api.ps1
- start.bat
- build.ps1
```

### 3. **脚本头部规范**
```bash
#!/bin/bash
# 脚本描述：这个脚本的用途
# 作者：开发者姓名
# 创建时间：YYYY-MM-DD
# 版本：1.0

set -e  # 遇到错误立即退出
set -u  # 使用未定义变量时报错
set -o pipefail  # 管道中任何命令失败都会导致整个管道失败
```

### 4. **颜色输出规范**
```bash
# 定义颜色常量
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 使用示例
echo -e "${GREEN}✅ 操作成功${NC}"
echo -e "${RED}❌ 操作失败${NC}"
echo -e "${YELLOW}⚠️ 警告信息${NC}"
echo -e "${BLUE}ℹ️ 提示信息${NC}"
```

### 5. **错误处理规范**
```bash
#!/bin/bash
set -e

# 错误处理函数
handle_error() {
    echo -e "${RED}❌ 脚本执行失败: $1${NC}"
    exit 1
}

# 使用trap捕获错误
trap 'handle_error "未知错误"' ERR

# 检查依赖
check_dependencies() {
    if ! command -v curl &> /dev/null; then
        handle_error "curl 命令未找到"
    fi
}

# 主函数
main() {
    echo -e "${BLUE}🚀 开始执行脚本...${NC}"
    
    # 检查依赖
    check_dependencies
    
    # 执行主要逻辑
    echo -e "${GREEN}✅ 脚本执行完成${NC}"
}

# 执行主函数
main "$@"
```

### 6. **验证脚本模板**
```bash
#!/bin/bash
# 验证实现状态的脚本模板

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 统计变量
total_checks=0
passed_checks=0

# 检查函数
check_file() {
    local file=$1
    local description=$2
    
    ((total_checks++))
    if [ -f "$file" ]; then
        echo -e "${GREEN}✅ $description${NC}"
        ((passed_checks++))
        return 0
    else
        echo -e "${RED}❌ $description${NC}"
        return 1
    fi
}

check_content() {
    local file=$1
    local pattern=$2
    local description=$3
    
    ((total_checks++))
    if grep -q "$pattern" "$file"; then
        echo -e "${GREEN}✅ $description${NC}"
        ((passed_checks++))
        return 0
    else
        echo -e "${RED}❌ $description${NC}"
        return 1
    fi
}

# 主验证逻辑
main() {
    echo -e "${BLUE}🔍 开始验证实现状态...${NC}"
    echo ""
    
    # 检查核心文件
    check_file "internal/corehr/models.go" "数据模型文件"
    check_file "internal/corehr/repository.go" "Repository层文件"
    check_file "internal/corehr/service.go" "Service层文件"
    
    # 检查关键方法
    check_content "internal/corehr/repository.go" "CreateEmployee" "员工创建方法"
    check_content "internal/corehr/repository.go" "GetEmployeeByID" "员工查询方法"
    check_content "internal/corehr/repository.go" "ListEmployees" "员工列表方法"
    
    # 输出统计结果
    echo ""
    echo -e "${BLUE}📊 验证总结${NC}"
    echo "总检查项: $total_checks"
    echo -e "通过检查: ${GREEN}$passed_checks${NC}"
    echo -e "失败检查: ${RED}$((total_checks - passed_checks))${NC}"
    
    success_rate=$(echo "scale=1; $passed_checks * 100 / $total_checks" | bc)
    echo -e "成功率: ${BLUE}${success_rate}%${NC}"
    
    if [ "$success_rate" -ge 90 ]; then
        echo -e "${GREEN}🎉 验证通过！${NC}"
    else
        echo -e "${RED}❌ 验证失败，需要改进${NC}"
        exit 1
    fi
}

# 执行主函数
main "$@"
```

### 7. **测试脚本模板**
```bash
#!/bin/bash
# API测试脚本模板

set -e

# 配置
API_BASE_URL="http://localhost:8080"
API_VERSION="v1"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 测试函数
test_api_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local description=$4
    
    echo -e "${BLUE}🧪 测试: $description${NC}"
    
    if [ "$method" = "GET" ]; then
        response=$(curl -s -w "%{http_code}" "$API_BASE_URL/api/$API_VERSION$endpoint")
    else
        response=$(curl -s -w "%{http_code}" -X "$method" \
            -H "Content-Type: application/json" \
            -d "$data" \
            "$API_BASE_URL/api/$API_VERSION$endpoint")
    fi
    
    http_code="${response: -3}"
    body="${response%???}"
    
    if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
        echo -e "${GREEN}✅ 成功 (HTTP $http_code)${NC}"
        echo "响应: $body" | head -c 100
        echo "..."
    else
        echo -e "${RED}❌ 失败 (HTTP $http_code)${NC}"
        echo "响应: $body"
    fi
    
    echo ""
}

# 主测试逻辑
main() {
    echo -e "${BLUE}🚀 开始API测试...${NC}"
    echo "API地址: $API_BASE_URL"
    echo ""
    
    # 测试员工列表
    test_api_endpoint "GET" "/corehr/employees" "" "获取员工列表"
    
    # 测试创建员工
    employee_data='{
        "employee_number": "EMP001",
        "first_name": "张三",
        "last_name": "李",
        "email": "zhangsan@example.com",
        "phone_number": "13800138001",
        "position": "软件工程师",
        "department": "技术部",
        "hire_date": "2024-01-15"
    }'
    test_api_endpoint "POST" "/corehr/employees" "$employee_data" "创建员工"
    
    echo -e "${GREEN}🎉 API测试完成！${NC}"
}

# 执行主函数
main "$@"
```

## 🚫 禁止事项

### 1. **不要创建PowerShell脚本**
```bash
❌ 错误示例:
- 验证实现.ps1
- 测试API.ps1
- 启动服务.ps1
```

### 2. **不要使用Windows特定命令**
```bash
❌ 错误示例:
- dir (使用 ls)
- copy (使用 cp)
- del (使用 rm)
- echo %PATH% (使用 echo $PATH)
```

### 3. **不要使用Windows路径格式**
```bash
❌ 错误示例:
- C:\Users\username\project
- \\server\share\path

✅ 正确示例:
- /home/username/project
- /mnt/c/Users/username/project
```

## ✅ 最佳实践

### 1. **脚本组织**
```
scripts/
├── test/           # 测试脚本
│   ├── test_api.sh
│   └── test_db.sh
├── deploy/         # 部署脚本
│   ├── deploy.sh
│   └── rollback.sh
├── verify/         # 验证脚本
│   ├── verify_1.1.1.sh
│   └── verify_1.1.2.sh
└── utils/          # 工具脚本
    ├── backup.sh
    └── cleanup.sh
```

### 2. **脚本权限**
```bash
# 设置脚本可执行权限
chmod +x scripts/*.sh

# 检查脚本权限
ls -la scripts/
```

### 3. **脚本测试**
```bash
# 使用shellcheck检查脚本语法
shellcheck scripts/*.sh

# 在测试环境中运行脚本
./scripts/test_api.sh
```

## 📝 总结

- **只使用Bash脚本**，确保跨平台兼容性
- **遵循命名规范**，使用`.sh`后缀
- **包含完整的错误处理**，提高脚本可靠性
- **使用颜色输出**，提高可读性
- **编写清晰的文档**，便于维护

---

**记住**: 在Cube Castle项目中，所有脚本都应该是Bash脚本，这样可以确保在WSL/Linux环境中正常运行，避免编码问题和跨平台兼容性问题。 