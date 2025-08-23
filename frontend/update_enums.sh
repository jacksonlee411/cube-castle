#!/bin/bash

# 批量更新前端文件中的unitType枚举
# 将 COST_CENTER 删除，COMPANY 改为 ORGANIZATION_UNIT

echo "开始批量更新前端unitType枚举..."

# 查找并替换TypeScript类型定义
find src -name "*.tsx" -o -name "*.ts" | grep -v node_modules | grep -v ".test." | while read file; do
  if grep -q "COST_CENTER\|COMPANY.*PROJECT_TEAM" "$file"; then
    echo "更新文件: $file"
    
    # 替换枚举定义
    sed -i "s/'DEPARTMENT' | 'COST_CENTER' | 'COMPANY' | 'PROJECT_TEAM'/'DEPARTMENT' | 'ORGANIZATION_UNIT' | 'PROJECT_TEAM'/g" "$file"
    sed -i "s/'COMPANY' | 'DEPARTMENT' | 'COST_CENTER' | 'PROJECT_TEAM'/'DEPARTMENT' | 'ORGANIZATION_UNIT' | 'PROJECT_TEAM'/g" "$file"
    sed -i "s/'COST_CENTER' | 'COMPANY' | 'DEPARTMENT' | 'PROJECT_TEAM'/'DEPARTMENT' | 'ORGANIZATION_UNIT' | 'PROJECT_TEAM'/g" "$file"
    
    # 替换选择器选项
    sed -i 's/<option value="COST_CENTER">成本中心<\/option>//g' "$file"
    sed -i 's/<option value="COMPANY">公司<\/option>/<option value="ORGANIZATION_UNIT">组织单位<\/option>/g' "$file"
    
    # 替换标签映射
    sed -i "s/'COST_CENTER': '成本中心'//g" "$file"
    sed -i "s/'COMPANY': '公司'/'ORGANIZATION_UNIT': '组织单位'/g" "$file"
    
    # 替换case语句
    sed -i "s/case 'COST_CENTER'://g" "$file"
    sed -i "s/case 'COMPANY':/case 'ORGANIZATION_UNIT':/g" "$file"
    
    # 清理空行
    sed -i '/^[[:space:]]*$/d' "$file" || true
  fi
done

echo "批量更新完成！"