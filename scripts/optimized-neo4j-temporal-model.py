#!/usr/bin/env python3

"""
优化的Neo4j时态数据模型重建脚本
基于干净的PostgreSQL数据，创建高性能的图数据库结构
"""

import psycopg2
from neo4j import GraphDatabase
import json
from datetime import datetime
import sys
import uuid

class OptimizedNeo4jTemporalModel:
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
    
    def create_optimized_schema(self):
        """创建优化的Neo4j Schema"""
        print("🏗️ 创建优化的Neo4j Schema...")
        
        with self.neo4j_driver.session() as session:
            # 1. 创建节点标签和属性约束
            schema_queries = [
                # 组织节点唯一约束
                """
                CREATE CONSTRAINT org_unique_current 
                FOR (n:Organization) 
                REQUIRE (n.tenant_id, n.code) IS UNIQUE
                """,
                
                # 时态节点约束
                """
                CREATE CONSTRAINT temporal_org_unique 
                FOR (n:TemporalOrganization) 
                REQUIRE (n.tenant_id, n.code, n.effective_date) IS UNIQUE
                """,
                
                # 性能索引
                """
                CREATE INDEX org_current_lookup 
                FOR (n:Organization) ON (n.tenant_id, n.is_current) 
                WHERE n.is_current = true
                """,
                
                """
                CREATE INDEX temporal_date_lookup 
                FOR (n:TemporalOrganization) ON (n.tenant_id, n.effective_date)
                """,
                
                """
                CREATE INDEX org_hierarchy_lookup 
                FOR (n:Organization) ON (n.tenant_id, n.parent_code)
                """,
                
                """
                CREATE INDEX org_level_lookup 
                FOR (n:Organization) ON (n.tenant_id, n.level)
                """
            ]
            
            for query in schema_queries:
                try:
                    session.run(query)
                    print(f"  ✅ Schema创建成功: {query.split()[1]}")
                except Exception as e:
                    if "already exists" in str(e):
                        print(f"  ⚠️ Schema已存在: {query.split()[1]}")
                    else:
                        print(f"  ❌ Schema创建失败: {e}")
    
    def get_clean_postgres_data(self):
        """获取清理后的PostgreSQL数据"""
        print("📥 获取干净的PostgreSQL数据...")
        
        cursor = self.pg_conn.cursor()
        
        # 只获取当前有效的组织数据
        query = """
        SELECT 
            code, parent_code, name, unit_type, status,
            level, path, sort_order, description,
            effective_date, end_date, is_current,
            tenant_id, created_at, updated_at
        FROM organization_units 
        WHERE is_current = true
        ORDER BY level, code
        """
        
        cursor.execute(query)
        results = cursor.fetchall()
        
        columns = [desc[0] for desc in cursor.description]
        organizations = [dict(zip(columns, row)) for row in results]
        
        print(f"  📊 获取到 {len(organizations)} 个干净的组织记录")
        return organizations
    
    def create_current_organizations(self, organizations):
        """创建当前组织节点"""
        print("🏗️ 创建当前组织节点...")
        
        with self.neo4j_driver.session() as session:
            # 批量创建当前组织节点
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
            for org in organizations:
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
            
            print(f"  ✅ 创建了 {created_count} 个组织节点")
            return created_count
    
    def create_hierarchy_relationships(self):
        """创建层级关系"""
        print("🔗 创建优化的层级关系...")
        
        with self.neo4j_driver.session() as session:
            # 创建直接父子关系
            parent_child_query = """
            MATCH (child:Organization), (parent:Organization)
            WHERE child.parent_code = parent.code
              AND child.tenant_id = parent.tenant_id
              AND child.parent_code IS NOT NULL
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
            
            print(f"  ✅ 创建了 {direct_count} 个直接层级关系")
            
            # 创建祖先后代关系（用于快速查询）
            ancestor_query = """
            MATCH (descendant:Organization)
            WHERE descendant.path IS NOT NULL
            WITH descendant, 
                 split(replace(descendant.path, '/', ''), '') as path_codes
            UNWIND path_codes as ancestor_code
            WITH descendant, ancestor_code
            WHERE ancestor_code <> descendant.code AND ancestor_code <> ''
            MATCH (ancestor:Organization {code: ancestor_code, tenant_id: descendant.tenant_id})
            MERGE (ancestor)-[:ANCESTOR_OF {
                created_at: datetime(),
                relationship_level: size(split(replace(descendant.path, '/', ''), '')) - 
                                   size(split(replace(ancestor.path, '/', ''), ''))
            }]->(descendant)
            RETURN count(*) as ancestor_relationships
            """
            
            result = session.run(ancestor_query)
            ancestor_count = list(result)[0]["ancestor_relationships"]
            
            print(f"  ✅ 创建了 {ancestor_count} 个祖先关系")
            
            return {"direct": direct_count, "ancestor": ancestor_count}
    
    def create_optimized_hierarchy_queries(self):
        """创建优化的层级查询函数"""
        print("⚡ 创建优化的查询函数...")
        
        with self.neo4j_driver.session() as session:
            # 1. 快速层级路径查询
            hierarchy_path_query = """
            CALL apoc.custom.asProcedure(
                'temporal.getOptimizedHierarchy',
                'WITH $tenant_id as tenant_id, $code as code
                 MATCH (org:Organization {tenant_id: tenant_id, code: code, is_current: true})
                 OPTIONAL MATCH path = (org)-[:PARENT_OF*]->(ancestors:Organization)
                 WHERE ancestors.is_current = true
                 WITH org, 
                      CASE WHEN ancestors IS NULL THEN [org] ELSE collect(DISTINCT ancestors) + [org] END as hierarchy_nodes
                 UNWIND hierarchy_nodes as node
                 RETURN node.code as code,
                        node.name as name,
                        node.level as level,
                        node.path as path,
                        node.unit_type as unit_type,
                        node.status as status
                 ORDER BY node.level',
                'READ',
                [['tenant_id','STRING'], ['code','STRING']]
            )
            """
            
            # 2. 快速子树查询
            subtree_query = """
            CALL apoc.custom.asProcedure(
                'temporal.getOptimizedSubtree',
                'WITH $tenant_id as tenant_id, $root_code as root_code, 
                      coalesce($max_depth, 10) as max_depth
                 MATCH (root:Organization {tenant_id: tenant_id, code: root_code, is_current: true})
                 MATCH path = (root)-[:HAS_CHILD*0..max_depth]->(descendants:Organization)
                 WHERE descendants.is_current = true
                 RETURN descendants.code as code,
                        descendants.name as name,
                        descendants.parent_code as parent_code,
                        descendants.level as level,
                        descendants.path as path,
                        descendants.unit_type as unit_type,
                        descendants.status as status,
                        length(path) as depth_from_root
                 ORDER BY descendants.level, descendants.path',
                'READ',
                [['tenant_id','STRING'], ['root_code','STRING'], ['max_depth','LONG']]
            )
            """
            
            try:
                session.run(hierarchy_path_query)
                print("  ✅ 层级路径查询函数创建成功")
            except Exception as e:
                print(f"  ⚠️ 层级路径函数创建失败: {e}")
            
            try:
                session.run(subtree_query)
                print("  ✅ 子树查询函数创建成功")
            except Exception as e:
                print(f"  ⚠️ 子树查询函数创建失败: {e}")
    
    def validate_optimized_model(self):
        """验证优化后的模型"""
        print("🔍 验证优化后的数据模型...")
        
        with self.neo4j_driver.session() as session:
            # 1. 检查节点数量
            nodes_result = session.run("MATCH (org:Organization) RETURN count(org) as count")
            node_count = list(nodes_result)[0]["count"]
            
            # 2. 检查关系数量
            rels_result = session.run("MATCH ()-[r]->() RETURN count(r) as count")
            rel_count = list(rels_result)[0]["count"]
            
            # 3. 检查层级完整性
            hierarchy_result = session.run("""
                MATCH (org:Organization)
                WHERE org.parent_code IS NOT NULL
                OPTIONAL MATCH (parent:Organization {code: org.parent_code, tenant_id: org.tenant_id})
                RETURN 
                  count(org) as orgs_with_parent,
                  count(parent) as valid_parents,
                  count(org) - count(parent) as orphan_count
            """)
            hierarchy_stats = list(hierarchy_result)[0]
            
            validation_report = {
                "total_nodes": node_count,
                "total_relationships": rel_count,
                "orgs_with_parent": hierarchy_stats["orgs_with_parent"],
                "valid_parents": hierarchy_stats["valid_parents"],
                "orphan_organizations": hierarchy_stats["orphan_count"],
                "data_integrity": hierarchy_stats["orphan_count"] == 0,
                "validated_at": datetime.now().isoformat()
            }
            
            print(f"  📊 总节点数: {node_count}")
            print(f"  📊 总关系数: {rel_count}")
            print(f"  📊 层级完整性: {'✅ 完整' if validation_report['data_integrity'] else '❌ 有孤儿节点'}")
            
            return validation_report
    
    def run_optimization(self):
        """执行完整的优化过程"""
        print("🚀 开始Neo4j时态数据模型优化...")
        
        try:
            # 1. 创建Schema
            self.create_optimized_schema()
            
            # 2. 获取干净数据
            organizations = self.get_clean_postgres_data()
            
            # 3. 创建组织节点
            created_count = self.create_current_organizations(organizations)
            
            # 4. 创建关系
            relationship_counts = self.create_hierarchy_relationships()
            
            # 5. 创建优化查询
            self.create_optimized_hierarchy_queries()
            
            # 6. 验证模型
            validation_report = self.validate_optimized_model()
            
            # 7. 生成报告
            optimization_report = {
                "optimization_completed": True,
                "organizations_created": created_count,
                "relationships_created": relationship_counts,
                "validation": validation_report,
                "completed_at": datetime.now().isoformat()
            }
            
            # 保存报告
            with open("/home/shangmeilin/cube-castle/neo4j-optimization-report.json", "w", encoding="utf-8") as f:
                json.dump(optimization_report, f, indent=2, ensure_ascii=False)
            
            print("\n✅ Neo4j时态数据模型优化完成！")
            print(f"📋 详细报告已保存到: neo4j-optimization-report.json")
            
            return optimization_report
            
        except Exception as e:
            print(f"❌ 优化过程失败: {e}")
            return None
    
    def close(self):
        """关闭连接"""
        self.pg_conn.close()
        self.neo4j_driver.close()

if __name__ == "__main__":
    optimizer = OptimizedNeo4jTemporalModel()
    try:
        result = optimizer.run_optimization()
        if result:
            sys.exit(0)
        else:
            sys.exit(1)
    except Exception as e:
        print(f"❌ 脚本执行失败: {e}")
        sys.exit(1)
    finally:
        optimizer.close()