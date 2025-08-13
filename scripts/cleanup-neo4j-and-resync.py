#!/usr/bin/env python3
"""
Neo4j数据清理和重新同步脚本
将PostgreSQL中的组织数据完全同步到Neo4j
"""

import psycopg2
from neo4j import GraphDatabase
import json
import sys
from datetime import datetime
import logging

# 配置日志
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

# 数据库连接配置
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

def get_postgres_connection():
    """获取PostgreSQL连接"""
    return psycopg2.connect(**POSTGRES_CONFIG)

def get_neo4j_driver():
    """获取Neo4j驱动"""
    return GraphDatabase.driver(NEO4J_CONFIG['uri'], 
                               auth=(NEO4J_CONFIG['user'], NEO4J_CONFIG['password']))

def cleanup_neo4j(driver):
    """清理Neo4j中所有组织数据"""
    logger.info("开始清理Neo4j数据...")
    
    with driver.session() as session:
        # 删除所有OrganizationUnit节点和关系
        result = session.run("""
            MATCH (o:OrganizationUnit)
            DETACH DELETE o
        """)
        
        # 获取删除统计
        result = session.run("MATCH (o:OrganizationUnit) RETURN count(o) as remaining_count")
        record = result.single()
        remaining_count = record['remaining_count'] if record else 0
        logger.info(f"Neo4j清理完成，剩余节点数: {remaining_count}")

def fetch_organizations_from_postgres(conn):
    """从PostgreSQL获取所有组织数据"""
    logger.info("从PostgreSQL获取组织数据...")
    
    with conn.cursor() as cursor:
        cursor.execute("""
            SELECT tenant_id, code, parent_code, name, unit_type, status, 
                   level, path, sort_order, description, created_at, updated_at,
                   effective_date, end_date, is_current, change_reason
            FROM organization_units 
            ORDER BY code, effective_date DESC
        """)
        
        rows = cursor.fetchall()
        columns = [desc[0] for desc in cursor.description]
        
        organizations = []
        for row in rows:
            org_data = dict(zip(columns, row))
            # 转换日期格式
            if org_data['created_at']:
                org_data['created_at'] = org_data['created_at'].isoformat()
            if org_data['updated_at']:
                org_data['updated_at'] = org_data['updated_at'].isoformat()
            if org_data['effective_date']:
                org_data['effective_date'] = org_data['effective_date'].strftime('%Y-%m-%d')
            if org_data['end_date']:
                org_data['end_date'] = org_data['end_date'].strftime('%Y-%m-%d')
            
            organizations.append(org_data)
        
        logger.info(f"获取到 {len(organizations)} 条组织记录")
        return organizations

def sync_organization_to_neo4j(session, org):
    """将单个组织数据同步到Neo4j"""
    query = """
        MERGE (o:OrganizationUnit {code: $code, tenant_id: $tenant_id, effective_date: $effective_date})
        SET o.name = $name,
            o.unit_type = $unit_type,
            o.status = $status,
            o.level = $level,
            o.path = $path,
            o.sort_order = $sort_order,
            o.description = COALESCE($description, ''),
            o.created_at = datetime($created_at),
            o.updated_at = datetime($updated_at),
            o.effective_date = date($effective_date),
            o.end_date = CASE WHEN $end_date IS NULL THEN NULL ELSE date($end_date) END,
            o.is_current = $is_current,
            o.change_reason = COALESCE($change_reason, ''),
            o.is_temporal = true,
            o.valid_from = datetime($created_at),
            o.valid_to = datetime('9999-12-31T23:59:59Z')
        RETURN o.code as code
    """
    
    params = {
        'code': org['code'],
        'tenant_id': org['tenant_id'],
        'name': org['name'],
        'unit_type': org['unit_type'],
        'status': org['status'],
        'level': org['level'],
        'path': org['path'] or f"/{org['code']}",
        'sort_order': org['sort_order'] or 0,
        'description': org['description'] or '',
        'created_at': org['created_at'],
        'updated_at': org['updated_at'],
        'effective_date': org['effective_date'],
        'end_date': org['end_date'],
        'is_current': org['is_current'],
        'change_reason': org['change_reason'] or ''
    }
    
    result = session.run(query, params)
    return result.single()

def create_hierarchy_relationships(session, organizations):
    """创建层级关系"""
    logger.info("创建组织层级关系...")
    
    # 按当前有效记录创建关系
    current_orgs = [org for org in organizations if org['is_current']]
    
    relationship_count = 0
    for org in current_orgs:
        if org['parent_code']:
            query = """
                MATCH (parent:OrganizationUnit {code: $parent_code, tenant_id: $tenant_id, is_current: true})
                MATCH (child:OrganizationUnit {code: $child_code, tenant_id: $tenant_id, is_current: true})
                MERGE (parent)-[r:HAS_CHILD {
                    effective_from: child.effective_date,
                    relationship_type: 'REPORTING'
                }]->(child)
                RETURN r
            """
            
            result = session.run(query, {
                'parent_code': org['parent_code'],
                'child_code': org['code'],
                'tenant_id': org['tenant_id']
            })
            
            if result.single():
                relationship_count += 1
    
    logger.info(f"创建了 {relationship_count} 个层级关系")

def sync_all_data():
    """完整的数据同步流程"""
    logger.info("开始完整的Neo4j数据同步...")
    
    # 连接数据库
    postgres_conn = get_postgres_connection()
    neo4j_driver = get_neo4j_driver()
    
    try:
        # 1. 清理Neo4j数据
        cleanup_neo4j(neo4j_driver)
        
        # 2. 从PostgreSQL获取数据
        organizations = fetch_organizations_from_postgres(postgres_conn)
        
        # 3. 同步组织数据到Neo4j
        with neo4j_driver.session() as session:
            synced_count = 0
            for org in organizations:
                try:
                    result = sync_organization_to_neo4j(session, org)
                    if result:
                        synced_count += 1
                        if synced_count % 10 == 0:
                            logger.info(f"已同步 {synced_count}/{len(organizations)} 个组织")
                except Exception as e:
                    logger.error(f"同步组织 {org['code']} 失败: {e}")
            
            logger.info(f"总共同步了 {synced_count} 个组织记录")
            
            # 4. 创建层级关系
            create_hierarchy_relationships(session, organizations)
        
        logger.info("✅ Neo4j数据同步完成")
        
        # 5. 验证同步结果
        verify_sync_result(neo4j_driver)
        
    except Exception as e:
        logger.error(f"同步过程中发生错误: {e}")
        sys.exit(1)
    finally:
        postgres_conn.close()
        neo4j_driver.close()

def verify_sync_result(driver):
    """验证同步结果"""
    logger.info("验证同步结果...")
    
    with driver.session() as session:
        # 统计总记录数
        result = session.run("MATCH (o:OrganizationUnit) RETURN count(o) as total")
        total_count = result.single()['total']
        
        # 统计当前有效记录数
        result = session.run("MATCH (o:OrganizationUnit {is_current: true}) RETURN count(o) as current")
        current_count = result.single()['current']
        
        # 统计关系数
        result = session.run("MATCH ()-[r:HAS_CHILD]->() RETURN count(r) as relationships")
        relationship_count = result.single()['relationships']
        
        logger.info(f"Neo4j同步验证结果:")
        logger.info(f"  - 总组织记录数: {total_count}")
        logger.info(f"  - 当前有效记录数: {current_count}")
        logger.info(f"  - 层级关系数: {relationship_count}")

if __name__ == '__main__':
    sync_all_data()