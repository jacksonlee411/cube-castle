#!/bin/bash
# 重复功能检测脚本
# 用途：自动检测潜在的重复功能实现

set -e

echo "🔍 Cube Castle 重复功能检测工具"
echo "================================="
echo

# 定义检测规则
declare -A PATTERNS=(
    ["sync_org"]="func.*Sync.*Organization|Sync.*Employee"
    ["cdc_impl"]="CREATE TRIGGER|pg_notify|LISTEN.*cdc|INSERT.*UPDATE.*DELETE"
    ["monitor_impl"]="Monitor.*Health|Health.*Check.*Monitor"
    ["recovery_impl"]="Recovery.*Auto|Auto.*Recovery"
    ["neo4j_direct"]="driver\.NewDriver|session\.Run.*neo4j"
)

declare -A EXISTING_SERVICES=(
    ["organization_sync"]="go-app/internal/service/organization_sync_service.go"
    ["cdc_service"]="go-app/internal/neo4j/cdc_sync_service.go" 
    ["monitoring"]="go-app/internal/monitoring/monitor.go"
    ["event_bus"]="go-app/internal/events/"
)

DUPLICATES_FOUND=false
WARNINGS_COUNT=0
ERRORS_COUNT=0

# 检查现有服务是否存在
check_existing_services() {
    echo "📋 验证现有企业级服务状态..."
    
    for service_name in "${!EXISTING_SERVICES[@]}"; do
        service_path="${EXISTING_SERVICES[$service_name]}"
        if [ -e "$service_path" ]; then
            echo "✅ $service_name: $service_path"
        else
            echo "⚠️ $service_name: $service_path (不存在)"
            ((WARNINGS_COUNT++))
        fi
    done
    echo
}

# 检测重复模式
detect_duplicates() {
    echo "🔍 检测重复功能模式..."
    
    for pattern_name in "${!PATTERNS[@]}"; do
        pattern="${PATTERNS[$pattern_name]}"
        echo "检查模式: $pattern_name"
        
        # 搜索匹配的文件
        matches=$(grep -r -l "$pattern" --include="*.go" . 2>/dev/null | grep -v backup/ | head -10)
        
        if [ ! -z "$matches" ]; then
            match_count=$(echo "$matches" | wc -l)
            echo "  📁 找到 $match_count 个匹配文件:"
            
            # 检查是否为已知的合法实现
            legitimate_found=false
            while IFS= read -r file; do
                # 检查是否是已知的企业级服务
                is_legitimate=false
                for service_path in "${EXISTING_SERVICES[@]}"; do
                    if [[ "$file" == *"$service_path"* ]]; then
                        echo "  ✅ $file (合法的企业级服务)"
                        is_legitimate=true
                        legitimate_found=true
                        break
                    fi
                done
                
                if [ "$is_legitimate" = false ]; then
                    echo "  ⚠️ $file (可能重复)"
                    DUPLICATES_FOUND=true
                    ((WARNINGS_COUNT++))
                fi
            done <<< "$matches"
            
            # 如果找到多个非合法实现，标记为错误
            duplicate_count=$((match_count - (legitimate_found ? 1 : 0)))
            if [ "$duplicate_count" -gt 1 ]; then
                echo "  ❌ 检测到 $duplicate_count 个可能重复的实现"
                ((ERRORS_COUNT++))
            fi
        else
            echo "  ✅ 未发现匹配文件"
        fi
        echo
    done
}

# 检查备份文件夹中的重复工具
check_backup_folder() {
    echo "📦 检查备份文件夹中的重复工具..."
    
    backup_dir="backup/redundant-tools-*"
    if ls $backup_dir 1> /dev/null 2>&1; then
        for backup in $backup_dir; do
            if [ -d "$backup" ]; then
                file_count=$(find "$backup" -name "*.go" | wc -l)
                echo "  📂 $backup: $file_count 个Go文件已备份"
                
                # 检查是否有活跃的重复
                for file in $(find "$backup" -name "*.go"); do
                    filename=$(basename "$file")
                    active_duplicates=$(find . -name "$filename" -not -path "./backup/*" 2>/dev/null)
                    if [ ! -z "$active_duplicates" ]; then
                        echo "  ⚠️ 检测到活跃重复: $filename"
                        ((WARNINGS_COUNT++))
                    fi
                done
            fi
        done
    else
        echo "  ✅ 未发现备份的重复工具"
    fi
    echo
}

# 分析函数名相似度  
analyze_function_similarity() {
    echo "🔬 分析函数名相似度..."
    
    # 提取所有函数名
    func_names=$(grep -r "^func " --include="*.go" . | grep -v backup/ | sed 's/.*func \([^(]*\).*/\1/' | sort)
    
    # 检查相似的函数名
    declare -A similar_funcs
    
    for func in $func_names; do
        # 提取函数名的关键词
        if [[ "$func" =~ (Sync|Monitor|Recovery|Health|CDC) ]]; then
            keyword="${BASH_REMATCH[1]}"
            if [ -z "${similar_funcs[$keyword]}" ]; then
                similar_funcs[$keyword]="$func"
            else
                similar_funcs[$keyword]="${similar_funcs[$keyword]}, $func"
            fi
        fi
    done
    
    for keyword in "${!similar_funcs[@]}"; do
        func_list="${similar_funcs[$keyword]}"
        func_count=$(echo "$func_list" | tr ',' '\n' | wc -l)
        
        if [ "$func_count" -gt 2 ]; then
            echo "  ⚠️ 关键词 '$keyword' 有 $func_count 个相似函数:"
            echo "     $func_list"
            ((WARNINGS_COUNT++))
        fi
    done
    echo
}

# 生成报告
generate_report() {
    echo "📊 检测结果报告"
    echo "=============="
    echo "错误数量: $ERRORS_COUNT"
    echo "警告数量: $WARNINGS_COUNT"
    echo
    
    if [ "$ERRORS_COUNT" -gt 0 ]; then
        echo "❌ 检测到严重的重复功能问题！"
        echo "建议立即审查并整合重复实现"
        return 1
    elif [ "$WARNINGS_COUNT" -gt 3 ]; then
        echo "⚠️ 检测到多个潜在问题"
        echo "建议进行代码审查"
        return 1
    elif [ "$DUPLICATES_FOUND" = true ]; then
        echo "⚠️ 发现一些需要关注的模式"
        echo "建议确认是否为合理的重复实现"
        return 0
    else
        echo "✅ 未检测到明显的重复功能"
        echo "架构一致性良好"
        return 0
    fi
}

# 主执行流程
main() {
    check_existing_services
    detect_duplicates
    check_backup_folder
    analyze_function_similarity
    generate_report
}

# 执行主程序
if [ "${BASH_SOURCE[0]}" == "${0}" ]; then
    main "$@"
fi