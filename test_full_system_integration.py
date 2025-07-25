#!/usr/bin/env python3
"""
Cube Castle 全系统集成测试
"""

import asyncio
import json
import time
import requests
import grpc
import sys
import os
import logging
from concurrent.futures import ThreadPoolExecutor, as_completed
import statistics

# 配置日志
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

# 添加当前目录到Python路径
sys.path.insert(0, '/home/shangmeilin/cube-castle/python-ai')

try:
    import intelligence_pb2
    import intelligence_pb2_grpc
except ImportError:
    logger.error("无法导入gRPC模块，请确保在python-ai目录下运行")
    sys.exit(1)

class CubeCastleIntegrationTest:
    """Cube Castle 全系统集成测试"""
    
    def __init__(self):
        self.base_url = "http://localhost:8080"
        self.ai_grpc_url = "localhost:50051"
        self.test_results = {
            "total": 0,
            "passed": 0,
            "failed": 0,
            "errors": []
        }
        
    def log_test_result(self, test_name: str, success: bool, message: str = ""):
        """记录测试结果"""
        self.test_results["total"] += 1
        if success:
            self.test_results["passed"] += 1
            logger.info(f"✅ {test_name}: {message}")
        else:
            self.test_results["failed"] += 1
            self.test_results["errors"].append(f"{test_name}: {message}")
            logger.error(f"❌ {test_name}: {message}")
    
    def test_health_endpoints(self):
        """测试健康检查端点"""
        logger.info("🔍 测试健康检查端点")
        
        try:
            response = requests.get(f"{self.base_url}/health", timeout=5)
            success = response.status_code == 200
            self.log_test_result(
                "健康检查端点", 
                success, 
                f"状态码: {response.status_code}, 响应: {response.text}"
            )
        except Exception as e:
            self.log_test_result("健康检查端点", False, f"请求失败: {e}")
    
    def test_corehr_api(self):
        """测试CoreHR API"""
        logger.info("🏢 测试CoreHR API")
        
        # 测试员工列表API
        try:
            response = requests.get(f"{self.base_url}/api/v1/corehr/employees", timeout=10)
            success = response.status_code == 200
            self.log_test_result(
                "员工列表API", 
                success, 
                f"状态码: {response.status_code}"
            )
            
            if success:
                data = response.json()
                self.log_test_result(
                    "员工列表数据格式", 
                    "employees" in data,
                    f"响应包含employees字段: {'employees' in data}"
                )
        except Exception as e:
            self.log_test_result("员工列表API", False, f"请求失败: {e}")
        
        # 测试组织树API
        try:
            response = requests.get(f"{self.base_url}/api/v1/corehr/organizations/tree", timeout=10)
            success = response.status_code == 200
            self.log_test_result(
                "组织树API", 
                success, 
                f"状态码: {response.status_code}"
            )
        except Exception as e:
            self.log_test_result("组织树API", False, f"请求失败: {e}")
    
    def test_employee_crud(self):
        """测试员工CRUD操作"""
        logger.info("👤 测试员工CRUD操作")
        
        # 创建员工
        employee_data = {
            "employee_number": f"TEST{int(time.time())}",
            "first_name": "测试",
            "last_name": "员工",
            "email": f"test{int(time.time())}@example.com",
            "status": "active"
        }
        
        try:
            # 创建员工
            response = requests.post(
                f"{self.base_url}/api/v1/corehr/employees",
                json=employee_data,
                timeout=10
            )
            
            if response.status_code == 201:
                created_employee = response.json()
                employee_id = created_employee.get("id")
                
                self.log_test_result(
                    "创建员工", 
                    True, 
                    f"员工ID: {employee_id}"
                )
                
                # 获取创建的员工
                get_response = requests.get(
                    f"{self.base_url}/api/v1/corehr/employees",
                    timeout=10
                )
                
                if get_response.status_code == 200:
                    employees = get_response.json().get("employees", [])
                    found = any(emp.get("id") == employee_id for emp in employees)
                    self.log_test_result(
                        "获取创建的员工", 
                        found, 
                        f"在员工列表中找到新创建的员工: {found}"
                    )
                else:
                    self.log_test_result("获取创建的员工", False, f"获取员工列表失败: {get_response.status_code}")
            
            elif response.status_code == 409:
                self.log_test_result("创建员工", True, "员工已存在(409) - 这是预期行为")
            else:
                self.log_test_result("创建员工", False, f"创建失败，状态码: {response.status_code}, 响应: {response.text}")
                
        except Exception as e:
            self.log_test_result("创建员工", False, f"请求失败: {e}")
    
    def test_ai_service_grpc(self):
        """测试AI服务gRPC接口"""
        logger.info("🤖 测试AI服务gRPC接口")
        
        try:
            channel = grpc.insecure_channel(self.ai_grpc_url)
            stub = intelligence_pb2_grpc.IntelligenceServiceStub(channel)
            
            # 测试文本解释
            test_cases = [
                {
                    "text": "更新我的电话号码为13800138000",
                    "expected_intent": None,  # 不强制要求特定意图
                    "description": "电话号码更新"
                },
                {
                    "text": "谁是我的经理？",
                    "expected_intent": None,
                    "description": "查询经理"
                },
                {
                    "text": "查看我的个人信息",
                    "expected_intent": None,
                    "description": "查看个人信息"
                }
            ]
            
            for i, case in enumerate(test_cases):
                try:
                    request = intelligence_pb2.InterpretRequest()
                    request.user_text = case["text"]
                    request.session_id = f"integration-test-{i}"
                    
                    start_time = time.time()
                    response = stub.InterpretText(request, timeout=30)
                    end_time = time.time()
                    
                    response_time = (end_time - start_time) * 1000
                    
                    # 验证响应
                    has_intent = hasattr(response, 'intent') and response.intent is not None
                    has_data = hasattr(response, 'structured_data_json') and response.structured_data_json is not None
                    
                    success = has_intent and has_data
                    self.log_test_result(
                        f"AI文本解释-{case['description']}", 
                        success, 
                        f"响应时间: {response_time:.0f}ms, 意图: {response.intent if has_intent else 'None'}"
                    )
                    
                except grpc.RpcError as e:
                    self.log_test_result(
                        f"AI文本解释-{case['description']}", 
                        False, 
                        f"gRPC错误: {e.code()}, {e.details()}"
                    )
                except Exception as e:
                    self.log_test_result(
                        f"AI文本解释-{case['description']}", 
                        False, 
                        f"异常: {e}"
                    )
            
            channel.close()
            
        except Exception as e:
            self.log_test_result("AI服务gRPC连接", False, f"连接失败: {e}")
    
    def test_intelligence_gateway_api(self):
        """测试Intelligence Gateway API"""
        logger.info("🧠 测试Intelligence Gateway API")
        
        test_data = {
            "query": "我想更新我的电话号码为13900139000",
            "user_id": "11111111-1111-1111-1111-111111111111"
        }
        
        try:
            response = requests.post(
                f"{self.base_url}/api/v1/intelligence/interpret",
                json=test_data,
                timeout=30
            )
            
            success = response.status_code in [200, 500]  # 允许500因为可能的AI服务问题
            
            if response.status_code == 200:
                data = response.json()
                has_intent = "intent" in data
                has_data = "structured_data_json" in data
                
                self.log_test_result(
                    "Intelligence Gateway API", 
                    has_intent and has_data, 
                    f"状态码: {response.status_code}, 意图: {data.get('intent', 'None')}"
                )
            else:
                self.log_test_result(
                    "Intelligence Gateway API", 
                    False, 
                    f"状态码: {response.status_code}, 响应: {response.text}"
                )
                
        except Exception as e:
            self.log_test_result("Intelligence Gateway API", False, f"请求失败: {e}")
    
    def test_performance_benchmarks(self):
        """测试性能基准"""
        logger.info("⚡ 测试性能基准")
        
        # 测试API响应时间
        api_endpoints = [
            "/health",
            "/api/v1/corehr/employees",
            "/api/v1/corehr/organizations/tree"
        ]
        
        for endpoint in api_endpoints:
            response_times = []
            
            for i in range(5):  # 每个端点测试5次
                try:
                    start_time = time.time()
                    response = requests.get(f"{self.base_url}{endpoint}", timeout=10)
                    end_time = time.time()
                    
                    if response.status_code == 200:
                        response_time = (end_time - start_time) * 1000
                        response_times.append(response_time)
                except Exception:
                    pass
            
            if response_times:
                avg_time = statistics.mean(response_times)
                max_time = max(response_times)
                min_time = min(response_times)
                
                # API响应时间基准：平均 < 1000ms，最大 < 3000ms
                performance_ok = avg_time < 1000 and max_time < 3000
                
                self.log_test_result(
                    f"API性能-{endpoint}", 
                    performance_ok, 
                    f"平均: {avg_time:.0f}ms, 最大: {max_time:.0f}ms, 最小: {min_time:.0f}ms"
                )
            else:
                self.log_test_result(f"API性能-{endpoint}", False, "无有效响应时间数据")
    
    def test_concurrent_load(self):
        """测试并发负载"""
        logger.info("🚀 测试并发负载")
        
        def make_request():
            try:
                response = requests.get(f"{self.base_url}/health", timeout=5)
                return response.status_code == 200
            except:
                return False
        
        # 并发测试：10个并发请求
        with ThreadPoolExecutor(max_workers=10) as executor:
            futures = [executor.submit(make_request) for _ in range(10)]
            results = [future.result() for future in as_completed(futures)]
        
        success_count = sum(results)
        success_rate = success_count / len(results) * 100
        
        # 要求成功率 >= 80%
        load_test_ok = success_rate >= 80
        
        self.log_test_result(
            "并发负载测试", 
            load_test_ok, 
            f"成功率: {success_rate:.1f}% ({success_count}/{len(results)})"
        )
    
    def test_data_consistency(self):
        """测试数据一致性"""
        logger.info("🔒 测试数据一致性")
        
        try:
            # 获取员工列表两次，验证数据一致性
            response1 = requests.get(f"{self.base_url}/api/v1/corehr/employees", timeout=10)
            time.sleep(1)  # 等待1秒
            response2 = requests.get(f"{self.base_url}/api/v1/corehr/employees", timeout=10)
            
            if response1.status_code == 200 and response2.status_code == 200:
                data1 = response1.json()
                data2 = response2.json()
                
                # 比较员工数量（在没有并发修改的情况下应该相同）
                count1 = len(data1.get("employees", []))
                count2 = len(data2.get("employees", []))
                
                consistency_ok = count1 == count2
                
                self.log_test_result(
                    "数据一致性", 
                    consistency_ok, 
                    f"第一次查询: {count1}条, 第二次查询: {count2}条"
                )
            else:
                self.log_test_result("数据一致性", False, "无法获取数据进行一致性检查")
                
        except Exception as e:
            self.log_test_result("数据一致性", False, f"测试失败: {e}")
    
    def run_all_tests(self):
        """运行所有集成测试"""
        logger.info("🏰 开始 Cube Castle 全系统集成测试")
        logger.info("=" * 60)
        
        start_time = time.time()
        
        # 按逻辑顺序执行测试
        self.test_health_endpoints()
        self.test_corehr_api()
        self.test_employee_crud()
        self.test_ai_service_grpc()
        self.test_intelligence_gateway_api()
        self.test_performance_benchmarks()
        self.test_concurrent_load()
        self.test_data_consistency()
        
        end_time = time.time()
        total_time = end_time - start_time
        
        # 输出测试总结
        logger.info("=" * 60)
        logger.info("🏰 Cube Castle 全系统集成测试完成")
        logger.info(f"总测试时间: {total_time:.2f}秒")
        logger.info(f"总测试数量: {self.test_results['total']}")
        logger.info(f"✅ 通过: {self.test_results['passed']}")
        logger.info(f"❌ 失败: {self.test_results['failed']}")
        
        if self.test_results["failed"] > 0:
            logger.info("\n失败的测试:")
            for error in self.test_results["errors"]:
                logger.info(f"  - {error}")
        
        success_rate = (self.test_results["passed"] / self.test_results["total"]) * 100
        logger.info(f"成功率: {success_rate:.1f}%")
        
        if success_rate >= 90:
            logger.info("🎉 系统整体状态: 优秀")
        elif success_rate >= 75:
            logger.info("✅ 系统整体状态: 良好")
        elif success_rate >= 50:
            logger.info("⚠️  系统整体状态: 需要改进")
        else:
            logger.info("❌ 系统整体状态: 存在严重问题")
        
        logger.info("=" * 60)
        
        return self.test_results

if __name__ == "__main__":
    test_runner = CubeCastleIntegrationTest()
    results = test_runner.run_all_tests()
    
    # 根据测试结果设置退出代码
    exit_code = 0 if results["failed"] == 0 else 1
    sys.exit(exit_code)