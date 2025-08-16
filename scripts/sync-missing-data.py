#!/usr/bin/env python3
"""
数据一致性修复脚本
从PostgreSQL同步缺失的组织记录到Neo4j
"""

import psycopg2
from neo4j import GraphDatabase
import json
from datetime import datetime
import uuid

# 数据库连接配置
PG_CONFIG = {
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

# 缺失的组织代码列表
MISSING_CODES = [
    '1001018', '1001019', '1001020', '1001021', '1001022',
    '1001023', '1001024', '1001025', '1001026', '1001027'
]

def get_missing_organizations_from_pg():
    """从PostgreSQL获取缺失的组织记录"""
    conn = psycopg2.connect(**PG_CONFIG)
    cursor = conn.cursor()
    
    # 获取缺失的组织记录
    query = """
    SELECT 
        tenant_id, code, parent_code, name, unit_type, status,
        level, path, sort_order, description, 
        created_at, updated_at, effective_date, end_date,
        is_temporal, change_reason, is_current
    FROM organization_units 
    WHERE code = ANY(%s)
    ORDER BY created_at;
    """
    
    cursor.execute(query, (MISSING_CODES,))
    results = cursor.fetchall()
    
    organizations = []
    for row in results:
        org = {
            'tenant_id': row[0],
            'code': row[1], 
            'parent_code': row[2],
            'name': row[3],
            'unit_type': row[4],
            'status': row[5],
            'level': row[6],
            'path': row[7],
            'sort_order': row[8],
            'description': row[9],
            'created_at': row[10].isoformat() if row[10] else None,
            'updated_at': row[11].isoformat() if row[11] else None,
            'effective_date': row[12].isoformat() if row[12] else None,
            'end_date': row[13].isoformat() if row[13] else None,
            'is_temporal': row[14],
            'change_reason': row[15],
            'is_current': row[16] if row[16] is not None else True
        }
        organizations.append(org)
    
    cursor.close()
    conn.close()
    return organizations

def sync_to_neo4j(organizations):
    """同步组织记录到Neo4j"""
    driver = GraphDatabase.driver(NEO4J_CONFIG['uri'], 
                                 auth=(NEO4J_CONFIG['user'], NEO4J_CONFIG['password']))
    
    with driver.session() as session:
        success_count = 0
        for org in organizations:
            try:
                # 生成确定性UUID (与同步服务逻辑保持一致)
                uuid_input = f"{org['tenant_id']}-{org['code']}"
                org_uuid = str(uuid.uuid5(uuid.NAMESPACE_DNS, uuid_input))
                
                # 创建Neo4j节点
                query = """
                MERGE (o:OrganizationUnit {tenant_id: $tenant_id, code: $code})
                SET o.uuid = $uuid,
                    o.parent_code = $parent_code,
                    o.name = $name,
                    o.unit_type = $unit_type,
                    o.status = $status,
                    o.level = $level,
                    o.path = $path,
                    o.sort_order = $sort_order,
                    o.description = $description,
                    o.created_at = $created_at,
                    o.updated_at = $updated_at,
                    o.effective_date = $effective_date,
                    o.end_date = $end_date,
                    o.is_temporal = $is_temporal,
                    o.change_reason = $change_reason,
                    o.is_current = $is_current,
                    o.last_synced = datetime()
                RETURN o.code, o.name
                """
                
                result = session.run(query, 
                    uuid=org_uuid,
                    tenant_id=org['tenant_id'],
                    code=org['code'],
                    parent_code=org['parent_code'],
                    name=org['name'],
                    unit_type=org['unit_type'],
                    status=org['status'],
                    level=org['level'],
                    path=org['path'],
                    sort_order=org['sort_order'],
                    description=org['description'],
                    created_at=org['created_at'],
                    updated_at=org['updated_at'],
                    effective_date=org['effective_date'],
                    end_date=org['end_date'],
                    is_temporal=org['is_temporal'],
                    change_reason=org['change_reason'],
                    is_current=org['is_current']
                )
                
                record = result.single()
                if record:
                    print(f"✅ 同步成功: {record['o.code']} - {record['o.name']}")
                    success_count += 1
                else:
                    print(f"❌ 同步失败: {org['code']} - {org['name']}")
                    
            except Exception as e:
                print(f"❌ 同步错误: {org['code']} - {e}")
    
    driver.close()
    return success_count

def main():
    print("🚀 开始数据一致性修复...")
    print(f"📋 目标修复记录数: {len(MISSING_CODES)}")
    
    # 从PostgreSQL获取缺失的记录
    print("\n📥 从PostgreSQL获取缺失记录...")
    organizations = get_missing_organizations_from_pg()
    print(f"📊 找到 {len(organizations)} 条缺失记录")
    
    if not organizations:
        print("⚠️  没有找到缺失的记录")
        return
    
    # 同步到Neo4j
    print("\n📤 同步记录到Neo4j...")
    success_count = sync_to_neo4j(organizations)
    
    print(f"\n🎯 修复完成:")
    print(f"   - 成功同步: {success_count} 条记录")
    print(f"   - 失败记录: {len(organizations) - success_count} 条")
    
    if success_count == len(organizations):
        print("✅ 数据一致性修复成功!")
    else:
        print("⚠️  部分记录同步失败，请检查错误日志")

if __name__ == "__main__":
    main()