#!/usr/bin/env python3
"""
时态历史数据同步脚本 (PostgreSQL → Neo4j)
将时态历史记录从命令端同步到查询端，确保CQRS架构完整性
"""
import psycopg2
from neo4j import GraphDatabase
import logging
import sys
from datetime import datetime

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# 数据库连接配置
POSTGRES_CONFIG = {
    'host': 'localhost',
    'port': '5432',
    'database': 'cubecastle',
    'user': 'user',
    'password': 'password'
}

NEO4J_CONFIG = {
    'uri': 'bolt://localhost:7687',
    'user': 'neo4j',
    'password': 'password'
}

def main():
    logger.info("🚀 开始时态历史数据同步 (PostgreSQL → Neo4j)")
    logger.info("🎯 目标：为CQRS架构修复 - 支持GraphQL时态查询")
    
    # 连接PostgreSQL
    try:
        pg_conn = psycopg2.connect(**POSTGRES_CONFIG)
        logger.info("✅ PostgreSQL连接成功")
    except Exception as e:
        logger.error(f"❌ PostgreSQL连接失败: {e}")
        return 1
    
    # 连接Neo4j
    try:
        neo4j_driver = GraphDatabase.driver(
            NEO4J_CONFIG['uri'],
            auth=(NEO4J_CONFIG['user'], NEO4J_CONFIG['password'])
        )
        logger.info("✅ Neo4j连接成功")
    except Exception as e:
        logger.error(f"❌ Neo4j连接失败: {e}")
        return 1
    
    try:
        # 获取PostgreSQL中的所有时态历史记录
        cursor = pg_conn.cursor()
        query = """
        SELECT record_id, tenant_id, code, parent_code, name, unit_type, status,
               level, path, sort_order, description, profile,
               created_at, updated_at, effective_date, end_date,
               change_reason, is_current
        FROM organization_units 
        ORDER BY code, effective_date
        """
        cursor.execute(query)
        temporal_records = cursor.fetchall()
        logger.info(f"📋 从PostgreSQL获取到 {len(temporal_records)} 条时态记录")
        
        # 清理Neo4j中的现有时态数据
        with neo4j_driver.session(database="neo4j") as session:
            logger.info("🧹 清理Neo4j中的现有时态数据...")
            session.run("MATCH (o:OrganizationUnit) DETACH DELETE o")
            logger.info("✅ Neo4j数据清理完成")
            
            # 创建约束和索引
            constraints_and_indexes = [
                "CREATE CONSTRAINT org_record_unique IF NOT EXISTS FOR (o:OrganizationUnit) REQUIRE o.record_id IS UNIQUE",
                "CREATE INDEX org_code_index IF NOT EXISTS FOR (o:OrganizationUnit) ON (o.code)",
                "CREATE INDEX org_effective_date_index IF NOT EXISTS FOR (o:OrganizationUnit) ON (o.effective_date)",
                "CREATE INDEX org_is_current_index IF NOT EXISTS FOR (o:OrganizationUnit) ON (o.is_current)"
            ]
            
            for constraint in constraints_and_indexes:
                try:
                    session.run(constraint)
                    logger.info(f"✅ 约束/索引创建: {constraint.split()[1]}")
                except Exception as e:
                    logger.warning(f"⚠️ 约束/索引已存在: {e}")
            
            # 批量插入时态历史记录
            logger.info("📝 开始插入时态历史记录...")
            insert_query = """
            CREATE (o:OrganizationUnit {
                record_id: $record_id,
                tenant_id: $tenant_id,
                code: $code,
                parent_code: $parent_code,
                name: $name,
                unit_type: $unit_type,
                status: $status,
                level: toInteger($level),
                path: $path,
                sort_order: toInteger($sort_order),
                description: $description,
                profile: $profile,
                created_at: datetime($created_at),
                updated_at: datetime($updated_at),
                effective_date: date($effective_date),
                end_date: CASE WHEN $end_date IS NOT NULL THEN date($end_date) ELSE null END,
                change_reason: $change_reason,
                is_current: $is_current
            })
            """
            
            inserted_count = 0
            for record in temporal_records:
                params = {
                    'record_id': str(record[0]),
                    'tenant_id': str(record[1]),
                    'code': record[2],
                    'parent_code': record[3],
                    'name': record[4],
                    'unit_type': record[5],
                    'status': record[6],
                    'level': record[7],
                    'path': record[8],
                    'sort_order': record[9],
                    'description': record[10] or '',
                    'profile': record[11] or '',
                    'created_at': record[12].isoformat() if record[12] else None,
                    'updated_at': record[13].isoformat() if record[13] else None,
                    'effective_date': record[14].isoformat() if record[14] else None,
                    'end_date': record[15].isoformat() if record[15] else None,
                    'change_reason': record[16],
                    'is_current': record[17] if record[17] is not None else False
                }
                
                session.run(insert_query, params)
                inserted_count += 1
                
                if inserted_count % 10 == 0:
                    logger.info(f"📈 已插入 {inserted_count} 条记录...")
            
            logger.info(f"✅ 成功插入 {inserted_count} 条时态历史记录")
            
            # 创建父子关系
            logger.info("🔗 创建组织层级关系...")
            relationship_query = """
            MATCH (child:OrganizationUnit), (parent:OrganizationUnit)
            WHERE child.parent_code = parent.code 
            AND child.is_current = true 
            AND parent.is_current = true
            CREATE (parent)-[:HAS_CHILD]->(child)
            """
            result = session.run(relationship_query)
            summary = result.consume()
            logger.info(f"🔗 创建了 {summary.counters.relationships_created} 个父子关系")
            
            # 验证时态数据
            logger.info("🔍 验证时态历史数据...")
            verification_queries = [
                ("总记录数", "MATCH (o:OrganizationUnit) RETURN count(o) as count"),
                ("当前记录数", "MATCH (o:OrganizationUnit {is_current: true}) RETURN count(o) as count"),
                ("历史记录数", "MATCH (o:OrganizationUnit {is_current: false}) RETURN count(o) as count"),
                ("1000004时态记录", "MATCH (o:OrganizationUnit {code: '1000004'}) RETURN count(o) as count")
            ]
            
            for desc, query in verification_queries:
                result = session.run(query)
                count = result.single()['count']
                logger.info(f"📊 {desc}: {count}")
        
        logger.info("✅ 时态历史数据同步完成!")
        logger.info("🎯 CQRS架构修复完成 - GraphQL时态查询现已支持")
        return 0
        
    except Exception as e:
        logger.error(f"❌ 同步过程中发生错误: {e}")
        return 1
    finally:
        # 关闭连接
        if 'cursor' in locals():
            cursor.close()
        if 'pg_conn' in locals():
            pg_conn.close()
        if 'neo4j_driver' in locals():
            neo4j_driver.close()
        logger.info("📋 数据库连接已关闭")

if __name__ == "__main__":
    sys.exit(main())