#!/usr/bin/env python3
"""
数据库性能专家分析工具 - PostgreSQL vs Neo4j 层级计算性能对比
==========================================================

深入分析和验证PostgreSQL和Neo4j在组织层级计算场景下的实际性能表现。
"""

import time
import statistics
import psycopg2
import requests
import json
from typing import List, Dict, Tuple
from dataclasses import dataclass
from neo4j import GraphDatabase
import concurrent.futures
import sys
import traceback

@dataclass
class PerformanceResult:
    """性能测试结果"""
    algorithm: str
    operation: str
    execution_times: List[float]
    avg_time: float
    median_time: float
    min_time: float
    max_time: float
    std_dev: float
    result_count: int
    memory_usage: str = "N/A"

class DatabasePerformanceAnalyzer:
    """数据库性能分析器"""
    
    def __init__(self):
        self.pg_conn = None
        self.neo4j_driver = None
        self.setup_connections()
    
    def setup_connections(self):
        """建立数据库连接"""
        try:
            # PostgreSQL连接
            self.pg_conn = psycopg2.connect(
                host="localhost",
                port=5432,
                database="cubecastle",
                user="user",
                password="password"
            )
            print("✅ PostgreSQL连接成功")
            
            # Neo4j连接
            self.neo4j_driver = GraphDatabase.driver(
                "bolt://localhost:7687",
                auth=("neo4j", "password")
            )
            print("✅ Neo4j连接成功")
            
        except Exception as e:
            print(f"❌ 数据库连接失败: {e}")
            sys.exit(1)
    
    def postgresql_hierarchy_query(self, org_code: str) -> Tuple[List[Dict], float]:
        """PostgreSQL递归层级查询 - 分析实际算法实现"""
        start_time = time.perf_counter()
        
        with self.pg_conn.cursor() as cursor:
            # 实际使用的递归CTE查询（向上查找父组织）
            query = """
            WITH RECURSIVE org_hierarchy AS (
              -- 基础查询：从目标组织开始
              SELECT 
                code,
                name,
                parent_code,
                level,
                1 as hierarchy_depth,
                code::text as path
              FROM organization_units 
              WHERE code = %s AND is_current = true
              
              UNION ALL
              
              -- 递归查询：向上查找父组织
              SELECT 
                p.code,
                p.name,
                p.parent_code,
                p.level,
                oh.hierarchy_depth + 1,
                p.code || ' -> ' || oh.path
              FROM organization_units p
              INNER JOIN org_hierarchy oh ON p.code = oh.parent_code
              WHERE p.is_current = true
            )
            SELECT 
              code,
              name,
              level,
              hierarchy_depth,
              path
            FROM org_hierarchy 
            ORDER BY hierarchy_depth DESC;
            """
            
            cursor.execute(query, (org_code,))
            results = cursor.fetchall()
            
        end_time = time.perf_counter()
        execution_time = end_time - start_time
        
        # 转换为字典格式
        formatted_results = []
        for row in results:
            formatted_results.append({
                'code': row[0],
                'name': row[1],
                'level': row[2],
                'hierarchy_depth': row[3],
                'path': row[4]
            })
        
        return formatted_results, execution_time
    
    def neo4j_hierarchy_query(self, org_code: str) -> Tuple[List[Dict], float]:
        """Neo4j图遍历层级查询 - 分析实际算法实现"""
        start_time = time.perf_counter()
        
        with self.neo4j_driver.session() as session:
            # Neo4j图遍历查询（一次性查找所有祖先）
            query = """
            MATCH (org:Organization {code: $org_code})
            OPTIONAL MATCH path = (org)-[:PARENT*0..]->(ancestor:Organization)
            WITH org, ancestor, length(path) as depth
            RETURN 
              ancestor.code as code,
              ancestor.name as name,
              ancestor.level as level,
              depth,
              [node in nodes(path) | node.code] as hierarchy_path
            ORDER BY depth
            """
            
            result = session.run(query, org_code=org_code)
            records = list(result)
            
        end_time = time.perf_counter()
        execution_time = end_time - start_time
        
        # 转换为字典格式
        formatted_results = []
        for record in records:
            formatted_results.append({
                'code': record['code'],
                'name': record['name'],
                'level': record['level'],
                'hierarchy_depth': record['depth'],
                'path': ' -> '.join(reversed(record['hierarchy_path'])) if record['hierarchy_path'] else record['code']
            })
        
        return formatted_results, execution_time
    
    def run_performance_test(self, test_name: str, test_func, iterations: int = 10) -> PerformanceResult:
        """执行性能测试"""
        print(f"\n🔍 执行测试: {test_name} (运行 {iterations} 次)")
        
        execution_times = []
        results_count = 0
        
        for i in range(iterations):
            try:
                results, exec_time = test_func()
                execution_times.append(exec_time)
                results_count = len(results)
                
                if i == 0:  # 第一次执行时显示详细结果
                    print(f"   📊 结果数量: {results_count}")
                    if results:
                        print(f"   📝 示例路径: {results[0].get('path', 'N/A')}")
                
                print(f"   ⏱️  第{i+1}次: {exec_time*1000:.3f}ms")
                
            except Exception as e:
                print(f"   ❌ 第{i+1}次执行失败: {e}")
                continue
        
        if not execution_times:
            print(f"   ❌ 测试 {test_name} 完全失败")
            return None
        
        # 计算统计指标
        avg_time = statistics.mean(execution_times)
        median_time = statistics.median(execution_times)
        min_time = min(execution_times)
        max_time = max(execution_times)
        std_dev = statistics.stdev(execution_times) if len(execution_times) > 1 else 0
        
        return PerformanceResult(
            algorithm=test_name.split(' - ')[0],
            operation=test_name.split(' - ')[1] if ' - ' in test_name else test_name,
            execution_times=execution_times,
            avg_time=avg_time,
            median_time=median_time,
            min_time=min_time,
            max_time=max_time,
            std_dev=std_dev,
            result_count=results_count
        )
    
    def analyze_algorithm_complexity(self):
        """分析算法复杂度"""
        print("\n" + "="*80)
        print("🧮 算法复杂度理论分析")
        print("="*80)
        
        print("\n📊 PostgreSQL递归CTE算法:")
        print("   🔄 算法类型: 递归查询 (Recursive Common Table Expression)")
        print("   📈 时间复杂度: O(h) - h为层级深度")
        print("   💾 空间复杂度: O(h) - 递归调用栈")
        print("   🎯 执行策略: 逐级向上查找父组织")
        print("   ⚡ 优势: 适合深层级结构，内存使用可控")
        print("   ⚠️  劣势: 需要多次表连接，IO操作较多")
        
        print("\n📊 Neo4j图遍历算法:")
        print("   🔄 算法类型: 图遍历 (Variable Length Path)")
        print("   📈 时间复杂度: O(h) - h为层级深度")
        print("   💾 空间复杂度: O(n) - n为遍历节点数")
        print("   🎯 执行策略: 一次性查找所有祖先路径")
        print("   ⚡ 优势: 专为图结构优化，单次查询获取完整路径")
        print("   ⚠️  劣势: 内存使用随层级数据量增长")
    
    def run_comprehensive_tests(self):
        """运行全面的性能测试"""
        print("\n" + "="*80)
        print("🚀 数据库层级计算性能对比测试")
        print("="*80)
        
        # 测试用例
        test_cases = [
            ("1000056", "测试组织 - 多父级结构"),
            ("1000000", "根组织 - 无父级"),
            ("1000002", "中层组织 - 标准层级"),
        ]
        
        results = []
        
        for org_code, description in test_cases:
            print(f"\n📋 测试用例: {description} (组织代码: {org_code})")
            print("-" * 60)
            
            # PostgreSQL测试
            pg_test = lambda: self.postgresql_hierarchy_query(org_code)
            pg_result = self.run_performance_test(f"PostgreSQL - 层级查询", pg_test)
            if pg_result:
                results.append(pg_result)
            
            # Neo4j测试
            neo4j_test = lambda: self.neo4j_hierarchy_query(org_code)
            neo4j_result = self.run_performance_test(f"Neo4j - 图遍历", neo4j_test)
            if neo4j_result:
                results.append(neo4j_result)
        
        return results
    
    def run_scalability_tests(self):
        """运行可扩展性测试"""
        print("\n" + "="*80)
        print("📊 可扩展性测试 - 批量操作性能")
        print("="*80)
        
        # 获取多个测试组织
        with self.pg_conn.cursor() as cursor:
            cursor.execute("""
                SELECT DISTINCT code 
                FROM organization_units 
                WHERE is_current = true AND parent_code IS NOT NULL
                LIMIT 20
            """)
            test_orgs = [row[0] for row in cursor.fetchall()]
        
        print(f"📝 测试组织数量: {len(test_orgs)}")
        
        # 批量测试PostgreSQL
        def batch_postgresql_test():
            start_time = time.perf_counter()
            total_results = 0
            for org_code in test_orgs:
                results, _ = self.postgresql_hierarchy_query(org_code)
                total_results += len(results)
            end_time = time.perf_counter()
            return [(total_results, "batch")], end_time - start_time
        
        # 批量测试Neo4j
        def batch_neo4j_test():
            start_time = time.perf_counter()
            total_results = 0
            for org_code in test_orgs:
                results, _ = self.neo4j_hierarchy_query(org_code)
                total_results += len(results)
            end_time = time.perf_counter()
            return [(total_results, "batch")], end_time - start_time
        
        batch_results = []
        
        # 执行批量测试
        pg_batch_result = self.run_performance_test("PostgreSQL - 批量层级查询", batch_postgresql_test, 5)
        if pg_batch_result:
            batch_results.append(pg_batch_result)
        
        neo4j_batch_result = self.run_performance_test("Neo4j - 批量图遍历", batch_neo4j_test, 5)
        if neo4j_batch_result:
            batch_results.append(neo4j_batch_result)
        
        return batch_results
    
    def generate_performance_report(self, results: List[PerformanceResult]):
        """生成性能分析报告"""
        print("\n" + "="*80)
        print("📊 性能测试结果报告")
        print("="*80)
        
        for result in results:
            print(f"\n🔍 {result.algorithm} - {result.operation}")
            print("-" * 50)
            print(f"   📊 平均执行时间: {result.avg_time*1000:.3f}ms")
            print(f"   📊 中位数时间: {result.median_time*1000:.3f}ms")
            print(f"   📊 最快执行: {result.min_time*1000:.3f}ms")
            print(f"   📊 最慢执行: {result.max_time*1000:.3f}ms")
            print(f"   📊 标准差: {result.std_dev*1000:.3f}ms")
            print(f"   📊 结果数量: {result.result_count}")
            
            # 性能等级评估
            avg_ms = result.avg_time * 1000
            if avg_ms < 1:
                performance_grade = "🟢 优秀 (<1ms)"
            elif avg_ms < 10:
                performance_grade = "🟡 良好 (<10ms)"
            elif avg_ms < 100:
                performance_grade = "🟠 一般 (<100ms)"
            else:
                performance_grade = "🔴 需优化 (>100ms)"
            
            print(f"   🎯 性能等级: {performance_grade}")
    
    def generate_comparison_analysis(self, results: List[PerformanceResult]):
        """生成对比分析"""
        print("\n" + "="*80)
        print("⚖️  PostgreSQL vs Neo4j 对比分析")
        print("="*80)
        
        # 按操作类型分组
        pg_results = [r for r in results if "PostgreSQL" in r.algorithm]
        neo4j_results = [r for r in results if "Neo4j" in r.algorithm]
        
        print(f"\n📊 总体性能对比:")
        print("-" * 40)
        
        if pg_results:
            pg_avg = statistics.mean([r.avg_time for r in pg_results]) * 1000
            print(f"   🐘 PostgreSQL平均响应: {pg_avg:.3f}ms")
        
        if neo4j_results:
            neo4j_avg = statistics.mean([r.avg_time for r in neo4j_results]) * 1000
            print(f"   🌐 Neo4j平均响应: {neo4j_avg:.3f}ms")
        
        if pg_results and neo4j_results:
            pg_total_avg = statistics.mean([r.avg_time for r in pg_results])
            neo4j_total_avg = statistics.mean([r.avg_time for r in neo4j_results])
            
            if pg_total_avg < neo4j_total_avg:
                speedup = neo4j_total_avg / pg_total_avg
                print(f"   🏆 PostgreSQL比Neo4j快 {speedup:.2f}倍")
            else:
                speedup = pg_total_avg / neo4j_total_avg
                print(f"   🏆 Neo4j比PostgreSQL快 {speedup:.2f}倍")
        
        print(f"\n🎯 专业建议:")
        print("-" * 40)
        print("   💡 单次查询场景: 选择平均响应时间更短的方案")
        print("   💡 批量查询场景: 考虑连接复用和缓存策略")
        print("   💡 深层级结构: PostgreSQL递归CTE内存使用更可控")
        print("   💡 复杂图关系: Neo4j原生图算法优势明显")
        print("   💡 CQRS架构: 建议查询端使用Neo4j，命令端使用PostgreSQL")
    
    def close_connections(self):
        """关闭数据库连接"""
        if self.pg_conn:
            self.pg_conn.close()
        if self.neo4j_driver:
            self.neo4j_driver.close()

def main():
    """主函数"""
    analyzer = DatabasePerformanceAnalyzer()
    
    try:
        # 算法复杂度分析
        analyzer.analyze_algorithm_complexity()
        
        # 基础性能测试
        basic_results = analyzer.run_comprehensive_tests()
        
        # 可扩展性测试
        scalability_results = analyzer.run_scalability_tests()
        
        # 合并所有结果
        all_results = basic_results + scalability_results
        
        # 生成报告
        analyzer.generate_performance_report(all_results)
        analyzer.generate_comparison_analysis(all_results)
        
        print("\n" + "="*80)
        print("✅ 性能分析完成")
        print("="*80)
        
    except Exception as e:
        print(f"❌ 分析过程中出现错误: {e}")
        traceback.print_exc()
    
    finally:
        analyzer.close_connections()

if __name__ == "__main__":
    main()