#!/usr/bin/env python3
"""
组织架构数据同步脚本 - PostgreSQL to Neo4j
严格按照CQRS统一实施指南标准实施
"""

import psycopg2
from neo4j import GraphDatabase
import json
import uuid
from datetime import datetime
import logging

# 项目默认租户配置
DEFAULT_TENANT_ID = "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9"
DEFAULT_TENANT_NAME = "高谷集团"

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

class OrganizationDataSyncer:
    """城堡CQRS查询端数据同步器"""
    
    def __init__(self):
        # PostgreSQL连接 (命令端数据源)
        self.pg_conn = psycopg2.connect(**POSTGRES_CONFIG)
        self.pg_cursor = self.pg_conn.cursor()
        
        # Neo4j连接 (查询端数据存储)
        self.neo4j_driver = GraphDatabase.driver(
            NEO4J_CONFIG['uri'], 
            auth=(NEO4J_CONFIG['user'], NEO4J_CONFIG['password'])
        )
        
    def close(self):
        """关闭数据库连接"""
        if self.pg_cursor:
            self.pg_cursor.close()
        if self.pg_conn:
            self.pg_conn.close()
        if self.neo4j_driver:
            self.neo4j_driver.close()
    
    def fetch_organization_units(self):
        """从PostgreSQL获取组织单元数据"""
        query = """
        SELECT 
            code, parent_code, tenant_id, name, unit_type, status,
            level, path, sort_order, description, profile,
            created_at, updated_at
        FROM organization_units
        ORDER BY level, sort_order, code
        """
        
        self.pg_cursor.execute(query)
        rows = self.pg_cursor.fetchall()
        
        organizations = []
        for row in rows:
            org = {
                'code': row[0],
                'parent_code': row[1],
                'tenant_id': str(row[2]),
                'name': row[3],
                'unit_type': row[4],
                'status': row[5] or 'ACTIVE',
                'level': row[6] or 1,
                'path': row[7] or '',
                'sort_order': row[8] or 0,
                'description': row[9] or '',
                'profile': row[10] or {},
                'created_at': row[11].isoformat() if row[11] else datetime.now().isoformat(),
                'updated_at': row[12].isoformat() if row[12] else datetime.now().isoformat()
            }
            organizations.append(org)
        
        logger.info(f"从PostgreSQL获取到 {len(organizations)} 个组织单元")
        return organizations
    
    def create_neo4j_constraints(self, session):
        """创建Neo4j约束和索引"""
        constraints = [
            "CREATE CONSTRAINT org_code_unique IF NOT EXISTS FOR (o:OrganizationUnit) REQUIRE o.code IS UNIQUE",
            "CREATE INDEX org_tenant_index IF NOT EXISTS FOR (o:OrganizationUnit) ON (o.tenant_id)",
            "CREATE INDEX org_status_index IF NOT EXISTS FOR (o:OrganizationUnit) ON (o.status)",
            "CREATE INDEX org_type_index IF NOT EXISTS FOR (o:OrganizationUnit) ON (o.unit_type)",
            "CREATE INDEX org_name_index IF NOT EXISTS FOR (o:OrganizationUnit) ON (o.name)"
        ]
        
        for constraint in constraints:
            try:
                session.run(constraint)
                logger.info(f"创建约束/索引: {constraint.split()[1]}")
            except Exception as e:
                logger.warning(f"约束创建可能已存在: {e}")
    
    def clear_existing_data(self, session):
        """清理现有的组织数据"""
        result = session.run("MATCH (o:OrganizationUnit) DETACH DELETE o")
        summary = result.consume()
        logger.info(f"清理了现有的组织单元数据")
    
    def sync_organization_to_neo4j(self, organizations):
        """同步组织数据到Neo4j"""
        with self.neo4j_driver.session() as session:
            # 创建约束和索引
            self.create_neo4j_constraints(session)
            
            # 清理现有数据
            self.clear_existing_data(session)
            
            # 第一步：创建所有组织节点
            logger.info("开始创建组织单元节点...")
            for org in organizations:
                create_query = """
                CREATE (o:OrganizationUnit {
                    code: $code,
                    tenant_id: $tenant_id,
                    name: $name,
                    unit_type: $unit_type,
                    status: $status,
                    level: $level,
                    path: $path,
                    sort_order: $sort_order,
                    description: $description,
                    profile: $profile,
                    created_at: $created_at,
                    updated_at: $updated_at
                })
                """
                
                session.run(create_query, {
                    'code': org['code'],
                    'tenant_id': org['tenant_id'], 
                    'name': org['name'],
                    'unit_type': org['unit_type'],
                    'status': org['status'],
                    'level': org['level'],
                    'path': org['path'],
                    'sort_order': org['sort_order'],
                    'description': org['description'],
                    'profile': json.dumps(org['profile']) if org['profile'] else '{}',
                    'created_at': org['created_at'],
                    'updated_at': org['updated_at']
                })
            
            logger.info(f"创建了 {len(organizations)} 个组织单元节点")
            
            # 第二步：创建父子关系
            logger.info("开始创建父子关系...")
            relationship_count = 0
            for org in organizations:
                if org['parent_code']:
                    relationship_query = """
                    MATCH (parent:OrganizationUnit {code: $parent_code})
                    MATCH (child:OrganizationUnit {code: $child_code})
                    CREATE (parent)-[:PARENT_OF]->(child)
                    """
                    
                    session.run(relationship_query, {
                        'parent_code': org['parent_code'],
                        'child_code': org['code']
                    })
                    relationship_count += 1
            
            logger.info(f"创建了 {relationship_count} 个父子关系")
            
            # 验证同步结果
            result = session.run("MATCH (o:OrganizationUnit) RETURN count(o) as total")
            total_count = result.single()['total']
            logger.info(f"Neo4j中现有组织单元总数: {total_count}")
            
            return total_count
    
    def verify_sync_integrity(self):
        """验证数据同步完整性"""
        # PostgreSQL计数
        self.pg_cursor.execute("SELECT COUNT(*) FROM organization_units")
        pg_count = self.pg_cursor.fetchone()[0]
        
        # Neo4j计数
        with self.neo4j_driver.session() as session:
            result = session.run("MATCH (o:OrganizationUnit) RETURN count(o) as total")
            neo4j_count = result.single()['total']
        
        logger.info(f"数据完整性验证:")
        logger.info(f"  PostgreSQL: {pg_count} 条记录")
        logger.info(f"  Neo4j: {neo4j_count} 条记录")
        
        if pg_count == neo4j_count:
            logger.info("✅ 数据同步完整性验证通过")
            return True
        else:
            logger.error("❌ 数据同步完整性验证失败")
            return False

def main():
    """主执行函数"""
    syncer = None
    try:
        logger.info("🚀 开始组织架构数据同步 (PostgreSQL -> Neo4j)")
        logger.info("严格按照CQRS统一实施指南标准执行")
        
        syncer = OrganizationDataSyncer()
        
        # 获取PostgreSQL数据
        organizations = syncer.fetch_organization_units()
        
        if not organizations:
            logger.warning("PostgreSQL中没有找到组织单元数据")
            return
        
        # 同步到Neo4j
        total_synced = syncer.sync_organization_to_neo4j(organizations)
        
        # 验证完整性
        if syncer.verify_sync_integrity():
            logger.info(f"✅ 组织架构数据同步完成! 共同步 {total_synced} 个组织单元")
            logger.info("🎯 CQRS查询端数据层准备就绪")
        else:
            logger.error("❌ 数据同步验证失败，请检查日志")
            
    except Exception as e:
        logger.error(f"同步过程中发生错误: {e}")
        import traceback
        traceback.print_exc()
    finally:
        if syncer:
            syncer.close()
            logger.info("数据库连接已关闭")

if __name__ == "__main__":
    main()