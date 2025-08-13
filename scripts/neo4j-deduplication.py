#!/usr/bin/env python3

"""
Neo4j数据去重和清理机制
解决重复节点和关系问题，优化数据质量
"""

import psycopg2
from neo4j import GraphDatabase
import json
from datetime import datetime

class Neo4jDataDeduplication:
    def __init__(self):
        # PostgreSQL连接
        self.pg_conn = psycopg2.connect(
            host="localhost",
            port=5432,
            database="cubecastle",
            user="user",
            password="password"
        )
        
        # Neo4j连接
        self.neo4j_driver = GraphDatabase.driver(
            "bolt://localhost:7687",
            auth=("neo4j", "password")
        )
    
    def detect_duplicate_nodes(self):
        """检测重复节点"""
        print("🔍 检测重复节点...")
        
        with self.neo4j_driver.session() as session:
            # 检查重复组织节点
            duplicate_query = """
            MATCH (org:Organization)
            WITH org.code as code, org.tenant_id as tenant_id, collect(org) as org_list
            WHERE size(org_list) > 1
            RETURN code, tenant_id, size(org_list) as duplicate_count
            ORDER BY duplicate_count DESC
            """
            
            result = session.run(duplicate_query)
            duplicates = list(result)
            
            if duplicates:
                print(f"  ⚠️ 发现 {len(duplicates)} 组重复节点:")
                for dup in duplicates:
                    print(f"    - 代码: {dup['code']}, 重复数量: {dup['duplicate_count']}")
                return duplicates
            else:
                print("  ✅ 没有发现重复节点")
                return []
    
    def remove_duplicate_nodes(self):
        """移除重复节点，保留最新的"""
        print("🧹 移除重复节点...")
        
        with self.neo4j_driver.session() as session:
            # 移除重复节点，保留updated_at最新的
            dedup_query = """
            MATCH (org:Organization)
            WITH org.code as code, org.tenant_id as tenant_id, collect(org) as org_list
            WHERE size(org_list) > 1
            WITH code, tenant_id, org_list,
                 [org IN org_list | org.updated_at] as update_times
            WITH code, tenant_id, org_list, max(update_times) as latest_update
            UNWIND org_list as org
            WITH code, tenant_id, org, latest_update
            WHERE org.updated_at < latest_update
            DETACH DELETE org
            RETURN count(*) as removed_duplicates
            """
            
            result = session.run(dedup_query)
            removed_count = list(result)[0]["removed_duplicates"]
            
            print(f"  ✅ 移除了 {removed_count} 个重复节点")
            return removed_count
    
    def clean_orphaned_relationships(self):
        """清理孤立的关系"""
        print("🧹 清理孤立的关系...")
        
        with self.neo4j_driver.session() as session:
            # 移除指向不存在节点的关系
            clean_query = """
            MATCH ()-[r]->()
            WHERE NOT (startNode(r):Organization) OR NOT (endNode(r):Organization)
            DELETE r
            RETURN count(r) as removed_relationships
            """
            
            result = session.run(clean_query)
            removed_count = list(result)[0]["removed_relationships"]
            
            print(f"  ✅ 清理了 {removed_count} 个孤立关系")
            return removed_count
    
    def synchronize_with_postgres_truth(self):
        """与PostgreSQL真实数据同步"""
        print("🔄 与PostgreSQL真实数据同步...")
        
        # 获取PostgreSQL的标准数据
        cursor = self.pg_conn.cursor()
        query = """
        SELECT code, parent_code, name, unit_type, status,
               level, path, sort_order, description,
               effective_date, end_date, is_current,
               tenant_id, created_at, updated_at
        FROM organization_units 
        WHERE is_current = true
        ORDER BY code
        """
        
        cursor.execute(query)
        results = cursor.fetchall()
        columns = [desc[0] for desc in cursor.description]
        postgres_orgs = [dict(zip(columns, row)) for row in results]
        
        print(f"  📊 PostgreSQL标准数据: {len(postgres_orgs)} 个组织")
        
        with self.neo4j_driver.session() as session:
            # 完全重建Neo4j数据以保证一致性
            print("  🧹 清空Neo4j现有数据...")
            session.run("MATCH (n) DETACH DELETE n")
            
            print("  🏗️ 重建标准数据...")
            
            # 批量创建标准组织节点
            create_query = """
            UNWIND $organizations as org
            CREATE (o:Organization {
                tenant_id: org.tenant_id,
                code: org.code,
                parent_code: org.parent_code,
                name: org.name,
                unit_type: org.unit_type,
                status: org.status,
                level: org.level,
                path: org.path,
                sort_order: org.sort_order,
                description: org.description,
                effective_date: date(org.effective_date),
                end_date: CASE WHEN org.end_date IS NOT NULL THEN date(org.end_date) ELSE null END,
                is_current: org.is_current,
                created_at: datetime(org.created_at),
                updated_at: datetime(org.updated_at),
                synced_at: datetime()
            })
            RETURN count(o) as created_count
            """
            
            # 转换数据格式
            org_data = []
            for org in postgres_orgs:
                org_data.append({
                    "tenant_id": str(org["tenant_id"]),
                    "code": org["code"],
                    "parent_code": org["parent_code"],
                    "name": org["name"],
                    "unit_type": org["unit_type"],
                    "status": org["status"],
                    "level": org["level"],
                    "path": org["path"],
                    "sort_order": org["sort_order"],
                    "description": org["description"],
                    "effective_date": org["effective_date"].strftime("%Y-%m-%d"),
                    "end_date": org["end_date"].strftime("%Y-%m-%d") if org["end_date"] else None,
                    "is_current": org["is_current"],
                    "created_at": org["created_at"].isoformat(),
                    "updated_at": org["updated_at"].isoformat()
                })
            
            result = session.run(create_query, {"organizations": org_data})
            created_count = list(result)[0]["created_count"]
            
            print(f"  ✅ 重建了 {created_count} 个标准组织节点")
            
            # 重建层级关系
            relationship_count = self.rebuild_clean_relationships()
            
            return {"organizations": created_count, "relationships": relationship_count}
    
    def rebuild_clean_relationships(self):
        """重建干净的层级关系"""
        print("  🔗 重建干净的层级关系...")
        
        with self.neo4j_driver.session() as session:
            # 创建直接父子关系
            parent_child_query = """
            MATCH (child:Organization), (parent:Organization)
            WHERE child.parent_code = parent.code
              AND child.tenant_id = parent.tenant_id
              AND child.parent_code IS NOT NULL
              AND child.is_current = true
              AND parent.is_current = true
            MERGE (parent)-[:HAS_CHILD {
                created_at: datetime(),
                relationship_level: 1
            }]->(child)
            MERGE (child)-[:PARENT_OF {
                created_at: datetime(),
                relationship_level: 1
            }]->(parent)
            RETURN count(*) as direct_relationships
            """
            
            result = session.run(parent_child_query)
            direct_count = list(result)[0]["direct_relationships"]
            
            print(f"    ✅ 创建了 {direct_count} 个直接层级关系")
            
            # 创建祖先后代关系（基于path字段）
            ancestor_query = """
            MATCH (descendant:Organization)
            WHERE descendant.path IS NOT NULL AND descendant.is_current = true
            WITH descendant, 
                 [segment IN split(descendant.path, '/') WHERE segment <> ''] as path_segments
            UNWIND path_segments as ancestor_code
            WITH descendant, ancestor_code
            WHERE ancestor_code <> descendant.code
            MATCH (ancestor:Organization {code: ancestor_code, tenant_id: descendant.tenant_id})
            WHERE ancestor.is_current = true
            MERGE (ancestor)-[:ANCESTOR_OF {
                created_at: datetime(),
                relationship_level: descendant.level - ancestor.level
            }]->(descendant)
            RETURN count(*) as ancestor_relationships
            """
            
            result = session.run(ancestor_query)
            ancestor_count = list(result)[0]["ancestor_relationships"]
            
            print(f"    ✅ 创建了 {ancestor_count} 个祖先关系")
            
            return {"direct": direct_count, "ancestor": ancestor_count}
    
    def validate_clean_data(self):
        """验证清理后的数据质量"""
        print("🔍 验证清理后的数据质量...")
        
        with self.neo4j_driver.session() as session:
            # 数据完整性检查
            integrity_query = """
            MATCH (org:Organization)
            WHERE org.is_current = true
            OPTIONAL MATCH (parent:Organization {code: org.parent_code, tenant_id: org.tenant_id})
            WHERE parent.is_current = true AND org.parent_code IS NOT NULL
            RETURN 
              count(org) as total_organizations,
              count(CASE WHEN org.parent_code IS NULL THEN 1 END) as root_organizations,
              count(CASE WHEN org.parent_code IS NOT NULL THEN 1 END) as child_organizations,
              count(parent) as valid_parent_references,
              count(CASE WHEN org.parent_code IS NOT NULL AND parent IS NULL THEN 1 END) as orphan_count
            """
            
            result = session.run(integrity_query)
            stats = list(result)[0]
            
            # 重复检查
            duplicate_check = """
            MATCH (org:Organization)
            WITH org.code as code, org.tenant_id as tenant_id, count(org) as node_count
            WHERE node_count > 1
            RETURN count(*) as duplicate_groups
            """
            
            result = session.run(duplicate_check)
            duplicate_groups = list(result)[0]["duplicate_groups"]
            
            # 关系统计
            relationship_stats = """
            MATCH ()-[r]->()
            RETURN 
              count(CASE WHEN type(r) = 'HAS_CHILD' THEN 1 END) as has_child_relations,
              count(CASE WHEN type(r) = 'PARENT_OF' THEN 1 END) as parent_of_relations,
              count(CASE WHEN type(r) = 'ANCESTOR_OF' THEN 1 END) as ancestor_relations,
              count(r) as total_relationships
            """
            
            result = session.run(relationship_stats)
            rel_stats = list(result)[0]
            
            validation_report = {
                "data_integrity": {
                    "total_organizations": stats["total_organizations"],
                    "root_organizations": stats["root_organizations"],
                    "child_organizations": stats["child_organizations"],
                    "valid_parent_references": stats["valid_parent_references"],
                    "orphan_organizations": stats["orphan_count"],
                    "data_is_clean": stats["orphan_count"] == 0 and duplicate_groups == 0
                },
                "duplicate_check": {
                    "duplicate_groups": duplicate_groups,
                    "no_duplicates": duplicate_groups == 0
                },
                "relationship_stats": {
                    "has_child_relations": rel_stats["has_child_relations"],
                    "parent_of_relations": rel_stats["parent_of_relations"],
                    "ancestor_relations": rel_stats["ancestor_relations"],
                    "total_relationships": rel_stats["total_relationships"]
                },
                "validated_at": datetime.now().isoformat()
            }
            
            print(f"  📊 总组织数: {validation_report['data_integrity']['total_organizations']}")
            print(f"  📊 根组织数: {validation_report['data_integrity']['root_organizations']}")
            print(f"  📊 孤儿组织: {validation_report['data_integrity']['orphan_organizations']}")
            print(f"  📊 重复组织组: {validation_report['duplicate_check']['duplicate_groups']}")
            print(f"  📊 总关系数: {validation_report['relationship_stats']['total_relationships']}")
            print(f"  📊 数据质量: {'✅ 优秀' if validation_report['data_integrity']['data_is_clean'] else '❌ 需要改进'}")
            
            return validation_report
    
    def run_deduplication(self):
        """运行完整的去重和清理过程"""
        print("🚀 开始Neo4j数据去重和清理...")
        
        try:
            # 1. 检测重复
            duplicates = self.detect_duplicate_nodes()
            
            # 2. 移除重复
            removed_duplicates = self.remove_duplicate_nodes()
            
            # 3. 清理孤立关系
            cleaned_relationships = self.clean_orphaned_relationships()
            
            # 4. 与PostgreSQL同步
            sync_results = self.synchronize_with_postgres_truth()
            
            # 5. 验证清理结果
            validation_report = self.validate_clean_data()
            
            # 6. 生成最终报告
            deduplication_report = {
                "deduplication_completed": True,
                "initial_duplicates": len(duplicates),
                "removed_duplicates": removed_duplicates,
                "cleaned_relationships": cleaned_relationships,
                "sync_results": sync_results,
                "validation": validation_report,
                "completed_at": datetime.now().isoformat()
            }
            
            # 保存报告
            with open("/home/shangmeilin/cube-castle/neo4j-deduplication-report.json", "w", encoding="utf-8") as f:
                json.dump(deduplication_report, f, indent=2, ensure_ascii=False)
            
            print("\n✅ Neo4j数据去重和清理完成！")
            print(f"📋 详细报告已保存到: neo4j-deduplication-report.json")
            
            return deduplication_report
            
        except Exception as e:
            print(f"❌ 去重清理失败: {e}")
            import traceback
            traceback.print_exc()
            return None
    
    def close(self):
        """关闭连接"""
        self.pg_conn.close()
        self.neo4j_driver.close()

if __name__ == "__main__":
    deduplicator = Neo4jDataDeduplication()
    try:
        result = deduplicator.run_deduplication()
        if result and result.get("validation", {}).get("data_integrity", {}).get("data_is_clean"):
            print("✅ 数据去重和清理成功完成，数据质量优秀")
        else:
            print("⚠️ 数据去重完成，但仍有质量问题需要解决")
    finally:
        deduplicator.close()