#!/usr/bin/env python3
"""
历史时态数据同步脚本
从PostgreSQL同步所有历史时态数据到Neo4j
"""

import psycopg2
from neo4j import GraphDatabase
import logging
import sys
from datetime import datetime

# 配置日志
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(message)s')
logger = logging.getLogger(__name__)

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

TENANT_ID = '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9'

def sync_historical_data():
    """同步历史时态数据从PostgreSQL到Neo4j"""
    
    # 连接PostgreSQL
    try:
        pg_conn = psycopg2.connect(**PG_CONFIG)
        logger.info("✅ PostgreSQL连接成功")
    except Exception as e:
        logger.error(f"❌ PostgreSQL连接失败: {e}")
        return False
    
    # 连接Neo4j
    try:
        neo4j_driver = GraphDatabase.driver(NEO4J_CONFIG['uri'], 
                                          auth=(NEO4J_CONFIG['user'], NEO4J_CONFIG['password']))
        logger.info("✅ Neo4j连接成功")
    except Exception as e:
        logger.error(f"❌ Neo4j连接失败: {e}")
        return False
    
    try:
        # 查询PostgreSQL中的所有时态数据
        query = """
        SELECT code, parent_code, tenant_id, name, unit_type, status, level, path, 
               sort_order, description, profile, created_at, updated_at, 
               effective_date, end_date, change_reason, is_current, is_temporal
        FROM organization_units 
        ORDER BY code, effective_date DESC
        """
        
        with pg_conn.cursor() as cursor:
            cursor.execute(query)
            records = cursor.fetchall()
            logger.info(f"📊 从PostgreSQL获取到 {len(records)} 条时态记录")
        
        # 清空Neo4j现有数据（仅组织单元）
        with neo4j_driver.session() as session:
            result = session.run("MATCH (n:OrganizationUnit) RETURN count(n) as count")
            old_count = result.single()['count']
            logger.info(f"📊 Neo4j中现有 {old_count} 条记录，准备清空重新同步")
            
            # 先删除所有约束，避免冲突
            try:
                session.run("DROP CONSTRAINT organization_unit_code IF EXISTS")
                logger.info("🔓 已删除组织代码唯一约束")
            except Exception as e:
                logger.info(f"约束删除结果: {e}")
                pass
            
            session.run("MATCH (n:OrganizationUnit) DETACH DELETE n")
            logger.info("🗑️ 已清空Neo4j中的组织单元数据")
        
        # 批量插入数据到Neo4j
        batch_size = 50
        successful = 0
        failed = 0
        
        with neo4j_driver.session() as session:
            for i in range(0, len(records), batch_size):
                batch = records[i:i + batch_size]
                
                # 构建批量插入语句
                cypher = """
                UNWIND $batch as row
                CREATE (org:OrganizationUnit)
                SET org.tenant_id = row.tenant_id,
                    org.code = row.code,
                    org.parent_code = row.parent_code,
                    org.name = row.name,
                    org.unit_type = row.unit_type,
                    org.status = row.status,
                    org.level = row.level,
                    org.path = row.path,
                    org.sort_order = row.sort_order,
                    org.description = row.description,
                    org.profile = row.profile,
                    org.created_at = row.created_at,
                    org.updated_at = row.updated_at,
                    org.effective_date = toString(row.effective_date),
                    org.end_date = toString(row.end_date),
                    org.change_reason = row.change_reason,
                    org.is_current = row.is_current,
                    org.is_temporal = row.is_temporal,
                    org.version = 1
                """
                
                # 准备批量数据
                batch_data = []
                for record in batch:
                    data = {
                        'tenant_id': str(record[2]) if record[2] else '',
                        'code': record[0] or '',
                        'parent_code': record[1] or '',
                        'name': record[3] or '',
                        'unit_type': record[4] or '',
                        'status': record[5] or '',
                        'level': record[6] or 1,
                        'path': record[7] or '',
                        'sort_order': record[8] or 0,
                        'description': record[9] or '',
                        'profile': str(record[10]) if record[10] else '{}',  # 确保JSON序列化
                        'created_at': record[11].isoformat() if record[11] else '',
                        'updated_at': record[12].isoformat() if record[12] else '',
                        'effective_date': record[13] if record[13] else None,
                        'end_date': record[14] if record[14] else None,
                        'change_reason': record[15] or '',
                        'is_current': record[16] if record[16] is not None else True,
                        'is_temporal': record[17] if record[17] is not None else False
                    }
                    batch_data.append(data)
                
                try:
                    session.run(cypher, batch=batch_data)
                    successful += len(batch)
                    logger.info(f"✅ 批量同步成功: {successful}/{len(records)} 条记录")
                except Exception as e:
                    failed += len(batch)
                    logger.error(f"❌ 批量同步失败: {e}")
        
        # 验证同步结果
        with neo4j_driver.session() as session:
            result = session.run("MATCH (n:OrganizationUnit) RETURN count(n) as count")
            final_count = result.single()['count']
            
            # 检查代码1000056的记录数
            result = session.run("MATCH (n:OrganizationUnit {code: '1000056'}) RETURN count(n) as count")
            test_count = result.single()['count']
            
        logger.info("🎉 历史数据同步完成!")
        logger.info(f"📊 同步统计:")
        logger.info(f"   - PostgreSQL源数据: {len(records)} 条")
        logger.info(f"   - 成功同步: {successful} 条")
        logger.info(f"   - 失败: {failed} 条")
        logger.info(f"   - Neo4j最终数量: {final_count} 条")
        logger.info(f"   - 测试代码1000056: {test_count} 条时态记录")
        
        return successful > 0
        
    except Exception as e:
        logger.error(f"❌ 同步过程失败: {e}")
        return False
    finally:
        pg_conn.close()
        neo4j_driver.close()

if __name__ == '__main__':
    logger.info("🚀 开始历史时态数据同步...")
    success = sync_historical_data()
    if success:
        logger.info("✅ 历史数据同步成功完成!")
        sys.exit(0)
    else:
        logger.error("❌ 历史数据同步失败!")
        sys.exit(1)