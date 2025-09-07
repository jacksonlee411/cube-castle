#!/usr/bin/env python3
"""
[已废弃 - 2025-09-07]
本脚本用于 CQRS 双数据库一致性检查（PostgreSQL ↔ Neo4j）。
现行架构为 PostgreSQL 单一数据源，已取消 Neo4j/CDC；仅作历史参考。
"""

import psycopg2
from neo4j import GraphDatabase
import json

# 项目默认租户配置
DEFAULT_TENANT_ID = "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9"
DEFAULT_TENANT_NAME = "高谷集团"

# 数据库配置
POSTGRES_CONFIG = {
    'host': 'localhost',
    'port': 5432,
    'database': 'cubecastle', 
    'user': 'user',
    'password': 'password'
}

NEO4J_CONFIG = {
    'uri': 'bolt://localhost:7687',
    'user': 'neo4j',
    'password': 'password'
}

def get_postgres_data():
    """从PostgreSQL获取组织数据"""
    conn = psycopg2.connect(**POSTGRES_CONFIG)
    cursor = conn.cursor()
    
    cursor.execute("""
        SELECT code, name, unit_type, status, level, parent_code, tenant_id, 
               created_at, updated_at, path, sort_order, description, profile
        FROM organization_units 
        ORDER BY code
    """)
    
    rows = cursor.fetchall()
    data = {}
    for row in rows:
        data[row[0]] = {  # code作为key
            'name': row[1],
            'unit_type': row[2], 
            'status': row[3],
            'level': row[4],
            'parent_code': row[5],
            'tenant_id': str(row[6]),
            'created_at': row[7].isoformat() if row[7] else None,
            'updated_at': row[8].isoformat() if row[8] else None,
            'path': row[9],
            'sort_order': row[10],
            'description': row[11],
            'profile': row[12]
        }
    
    cursor.close()
    conn.close()
    return data

def get_neo4j_data():
    """从Neo4j获取组织数据"""
    driver = GraphDatabase.driver(NEO4J_CONFIG['uri'], 
                                 auth=(NEO4J_CONFIG['user'], NEO4J_CONFIG['password']))
    
    with driver.session() as session:
        result = session.run("""
            MATCH (o:OrganizationUnit)
            RETURN o.code, o.name, o.unit_type, o.status, o.level, 
                   o.tenant_id, o.created_at, o.updated_at, o.path,
                   o.sort_order, o.description, o.profile
            ORDER BY o.code
        """)
        
        data = {}
        for record in result:
            code = record['o.code']
            data[code] = {
                'name': record['o.name'],
                'unit_type': record['o.unit_type'],
                'status': record['o.status'], 
                'level': record['o.level'],
                'tenant_id': record['o.tenant_id'],
                'created_at': record['o.created_at'],
                'updated_at': record['o.updated_at'],
                'path': record['o.path'],
                'sort_order': record['o.sort_order'],
                'description': record['o.description'],
                'profile': record['o.profile']
            }
    
    driver.close()
    return data

def compare_datasets(pg_data, neo4j_data):
    """对比两个数据集"""
    print("🔍 CQRS双数据库一致性验证报告")
    print("=" * 60)
    
    # 基础统计
    print(f"📊 数据量对比:")
    print(f"  PostgreSQL: {len(pg_data)} 条记录")
    print(f"  Neo4j:      {len(neo4j_data)} 条记录")
    print(f"  一致性:     {'✅ 一致' if len(pg_data) == len(neo4j_data) else '❌ 不一致'}")
    print()
    
    # 记录级别对比
    print("📋 记录级别一致性检查:")
    all_codes = set(pg_data.keys()) | set(neo4j_data.keys())
    consistent_count = 0
    
    for code in sorted(all_codes):
        pg_record = pg_data.get(code)
        neo4j_record = neo4j_data.get(code)
        
        if not pg_record:
            print(f"  ❌ {code}: 仅存在于Neo4j")
            continue
        if not neo4j_record:
            print(f"  ❌ {code}: 仅存在于PostgreSQL")
            continue
            
        # 字段级别对比
        field_consistent = True
        differences = []
        
        # 核心字段对比
        core_fields = ['name', 'unit_type', 'status', 'level', 'tenant_id']
        for field in core_fields:
            pg_val = pg_record.get(field)
            neo4j_val = neo4j_record.get(field)
            
            if pg_val != neo4j_val:
                field_consistent = False
                differences.append(f"{field}: PG='{pg_val}' vs Neo4j='{neo4j_val}'")
        
        if field_consistent:
            print(f"  ✅ {code}: {pg_record['name']} - 完全一致")
            consistent_count += 1
        else:
            print(f"  ❌ {code}: {pg_record['name']} - 存在差异")
            for diff in differences:
                print(f"      {diff}")
    
    print()
    print(f"📈 一致性统计:")
    consistency_rate = (consistent_count / len(all_codes)) * 100 if all_codes else 0
    print(f"  一致记录: {consistent_count}/{len(all_codes)}")
    print(f"  一致性率: {consistency_rate:.2f}%")
    
    if consistency_rate >= 99:
        print("  🎯 CQRS数据同步: ✅ 优秀")
    elif consistency_rate >= 95:
        print("  🎯 CQRS数据同步: ⚠️  良好")
    else:
        print("  🎯 CQRS数据同步: ❌ 需要修复")
    
    return consistency_rate

def analyze_relationships(pg_data, neo4j_data):
    """分析组织关系结构"""
    print("\n🏗️  组织架构关系分析:")
    print("-" * 40)
    
    # 层级分布
    level_distribution = {}
    for code, data in pg_data.items():
        level = data['level']
        if level not in level_distribution:
            level_distribution[level] = []
        level_distribution[level].append((code, data['name']))
    
    print("📊 层级分布:")
    for level in sorted(level_distribution.keys()):
        orgs = level_distribution[level]
        print(f"  级别 {level}: {len(orgs)} 个组织")
        for code, name in orgs:
            print(f"    - {code}: {name}")
    
    # 父子关系
    print("\n🌳 父子关系:")
    for code, data in sorted(pg_data.items()):
        if data['parent_code']:
            parent = pg_data.get(data['parent_code'])
            parent_name = parent['name'] if parent else '未知'
            print(f"  {parent_name} ({data['parent_code']}) → {data['name']} ({code})")
    
    # 类型分布
    type_distribution = {}
    for data in pg_data.values():
        unit_type = data['unit_type']
        type_distribution[unit_type] = type_distribution.get(unit_type, 0) + 1
    
    print("\n📋 类型分布:")
    for unit_type, count in sorted(type_distribution.items()):
        print(f"  {unit_type}: {count} 个")

def main():
    try:
        print("🚀 开始CQRS数据一致性验证...")
        
        # 获取数据
        print("📥 从PostgreSQL获取数据...")
        pg_data = get_postgres_data()
        
        print("📥 从Neo4j获取数据...")  
        neo4j_data = get_neo4j_data()
        
        # 对比数据
        consistency_rate = compare_datasets(pg_data, neo4j_data)
        
        # 分析关系
        analyze_relationships(pg_data, neo4j_data)
        
        # 总结
        print(f"\n🎯 验证总结:")
        print(f"CQRS数据同步一致性: {consistency_rate:.2f}%")
        if consistency_rate == 100:
            print("✅ 完美！CQRS查询端数据完全一致")
        elif consistency_rate >= 99:
            print("🎉 优秀！CQRS实施成功") 
        else:
            print("⚠️  需要检查数据同步机制")
            
    except Exception as e:
        print(f"❌ 验证过程中发生错误: {e}")
        import traceback
        traceback.print_exc()

if __name__ == "__main__":
    main()
