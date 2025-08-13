#!/usr/bin/env python3
"""
修复Neo4j层级关系脚本
将缺少父级组织的组织设置为高谷集团下级
"""

import psycopg2
from neo4j import GraphDatabase
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

# 高谷集团配置
GAOGU_GROUP = {
    'code': '1000000',
    'name': '高谷集团',
    'tenant_id': '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9'
}

def get_postgres_connection():
    """获取PostgreSQL连接"""
    return psycopg2.connect(**POSTGRES_CONFIG)

def get_neo4j_driver():
    """获取Neo4j驱动"""
    return GraphDatabase.driver(NEO4J_CONFIG['uri'], 
                               auth=(NEO4J_CONFIG['user'], NEO4J_CONFIG['password']))

def ensure_gaogu_group_exists():
    """确保高谷集团组织存在"""
    postgres_conn = get_postgres_connection()
    neo4j_driver = get_neo4j_driver()
    
    try:
        # 检查PostgreSQL中是否存在高谷集团
        with postgres_conn.cursor() as cursor:
            cursor.execute("""
                SELECT code, name FROM organization_units 
                WHERE code = %s AND tenant_id = %s AND is_current = true
            """, (GAOGU_GROUP['code'], GAOGU_GROUP['tenant_id']))
            
            result = cursor.fetchone()
            if result:
                logger.info(f"✅ PostgreSQL中已存在高谷集团: {result[1]}")
            else:
                logger.info("🔧 在PostgreSQL中创建高谷集团...")
                cursor.execute("""
                    INSERT INTO organization_units (
                        code, tenant_id, name, unit_type, status, level, path, 
                        sort_order, description, effective_date, is_current
                    ) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, CURRENT_DATE, true)
                """, (
                    GAOGU_GROUP['code'], 
                    GAOGU_GROUP['tenant_id'],
                    GAOGU_GROUP['name'],
                    'COMPANY',
                    'ACTIVE',
                    1,
                    f"/{GAOGU_GROUP['code']}",
                    0,
                    '集团总部'
                ))
                postgres_conn.commit()
                logger.info("✅ PostgreSQL中高谷集团创建成功")
        
        # 确保Neo4j中存在高谷集团
        with neo4j_driver.session() as session:
            result = session.run("""
                MERGE (org:OrganizationUnit {code: $code, tenant_id: $tenant_id, is_current: true})
                SET org.name = $name,
                    org.unit_type = 'COMPANY',
                    org.status = 'ACTIVE',
                    org.level = 1,
                    org.path = $path,
                    org.sort_order = 0,
                    org.description = '集团总部',
                    org.effective_date = date(),
                    org.is_temporal = false
                RETURN org.code as code
            """, {
                'code': GAOGU_GROUP['code'],
                'tenant_id': GAOGU_GROUP['tenant_id'],
                'name': GAOGU_GROUP['name'],
                'path': f"/{GAOGU_GROUP['code']}"
            })
            
            if result.single():
                logger.info("✅ Neo4j中高谷集团确认存在")
            
    finally:
        postgres_conn.close()
        neo4j_driver.close()

def find_orphaned_organizations():
    """查找缺少父级的组织"""
    postgres_conn = get_postgres_connection()
    
    try:
        with postgres_conn.cursor() as cursor:
            # 查找当前有效但没有父级的组织（排除高谷集团本身）
            cursor.execute("""
                SELECT code, name, level 
                FROM organization_units 
                WHERE is_current = true 
                  AND (parent_code IS NULL OR parent_code = '')
                  AND code != %s
                  AND tenant_id = %s
                ORDER BY code
            """, (GAOGU_GROUP['code'], GAOGU_GROUP['tenant_id']))
            
            orphaned_orgs = cursor.fetchall()
            logger.info(f"🔍 找到 {len(orphaned_orgs)} 个缺少父级的组织:")
            
            for code, name, level in orphaned_orgs:
                logger.info(f"  - {code}: {name} (级别: {level})")
            
            return [{'code': code, 'name': name, 'level': level} 
                   for code, name, level in orphaned_orgs]
    
    finally:
        postgres_conn.close()

def fix_orphaned_organizations(orphaned_orgs):
    """修复缺少父级的组织"""
    if not orphaned_orgs:
        logger.info("✅ 没有需要修复的孤立组织")
        return
    
    postgres_conn = get_postgres_connection()
    neo4j_driver = get_neo4j_driver()
    
    try:
        logger.info(f"🔧 开始修复 {len(orphaned_orgs)} 个孤立组织...")
        
        # 在PostgreSQL中更新父级关系
        with postgres_conn.cursor() as cursor:
            for org in orphaned_orgs:
                cursor.execute("""
                    UPDATE organization_units 
                    SET parent_code = %s, 
                        level = 2,
                        path = %s,
                        updated_at = CURRENT_TIMESTAMP
                    WHERE code = %s AND tenant_id = %s AND is_current = true
                """, (
                    GAOGU_GROUP['code'],
                    f"/{GAOGU_GROUP['code']}/{org['code']}",
                    org['code'],
                    GAOGU_GROUP['tenant_id']
                ))
                
                logger.info(f"✅ PostgreSQL: {org['code']} -> {GAOGU_GROUP['code']}")
            
            postgres_conn.commit()
        
        # 在Neo4j中更新层级关系
        with neo4j_driver.session() as session:
            for org in orphaned_orgs:
                # 更新组织的父级信息
                session.run("""
                    MATCH (child:OrganizationUnit {code: $child_code, tenant_id: $tenant_id, is_current: true})
                    SET child.level = 2,
                        child.path = $new_path
                    RETURN child.code as code
                """, {
                    'child_code': org['code'],
                    'tenant_id': GAOGU_GROUP['tenant_id'],
                    'new_path': f"/{GAOGU_GROUP['code']}/{org['code']}"
                })
                
                # 创建层级关系
                result = session.run("""
                    MATCH (parent:OrganizationUnit {code: $parent_code, tenant_id: $tenant_id, is_current: true})
                    MATCH (child:OrganizationUnit {code: $child_code, tenant_id: $tenant_id, is_current: true})
                    MERGE (parent)-[r:HAS_CHILD {
                        effective_from: child.effective_date,
                        relationship_type: 'REPORTING'
                    }]->(child)
                    RETURN r
                """, {
                    'parent_code': GAOGU_GROUP['code'],
                    'child_code': org['code'],
                    'tenant_id': GAOGU_GROUP['tenant_id']
                })
                
                if result.single():
                    logger.info(f"✅ Neo4j关系: {GAOGU_GROUP['code']} -> {org['code']}")
        
        logger.info("🎉 所有孤立组织已成功设置为高谷集团下级")
    
    finally:
        postgres_conn.close()
        neo4j_driver.close()

def verify_hierarchy_fix():
    """验证层级关系修复结果"""
    neo4j_driver = get_neo4j_driver()
    
    try:
        with neo4j_driver.session() as session:
            # 统计层级关系数量
            result = session.run("""
                MATCH ()-[r:HAS_CHILD]->()
                RETURN count(r) as total_relationships
            """)
            
            total_relationships = result.single()['total_relationships']
            
            # 统计高谷集团的直接下级
            result = session.run("""
                MATCH (parent:OrganizationUnit {code: $code, tenant_id: $tenant_id, is_current: true})
                       -[r:HAS_CHILD]->(child:OrganizationUnit)
                RETURN count(r) as gaogu_children
            """, {
                'code': GAOGU_GROUP['code'],
                'tenant_id': GAOGU_GROUP['tenant_id']
            })
            
            gaogu_children = result.single()['gaogu_children']
            
            logger.info("📊 层级关系修复验证结果:")
            logger.info(f"  - 总层级关系数: {total_relationships}")
            logger.info(f"  - 高谷集团直接下级: {gaogu_children}")
    
    finally:
        neo4j_driver.close()

def main():
    """主修复流程"""
    logger.info("🚀 开始修复Neo4j层级关系...")
    
    # 1. 确保高谷集团存在
    ensure_gaogu_group_exists()
    
    # 2. 查找孤立组织
    orphaned_orgs = find_orphaned_organizations()
    
    # 3. 修复孤立组织
    fix_orphaned_organizations(orphaned_orgs)
    
    # 4. 验证修复结果
    verify_hierarchy_fix()
    
    logger.info("✅ Neo4j层级关系修复完成")

if __name__ == '__main__':
    main()