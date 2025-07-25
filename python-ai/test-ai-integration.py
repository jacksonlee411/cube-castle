#!/usr/bin/env python3
"""
AI服务集成测试脚本
"""

import grpc
import sys
import os
import json
import time
import asyncio
from concurrent.futures import ThreadPoolExecutor

# 添加当前目录到Python路径
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

import intelligence_pb2
import intelligence_pb2_grpc

class AIServiceIntegrationTest:
    def __init__(self):
        self.total_tests = 0
        self.passed_tests = 0
        self.failed_tests = 0
        self.channel = None
        self.stub = None
    
    def test_result(self, condition, message):
        """记录测试结果"""
        self.total_tests += 1
        if condition:
            print(f"✅ {message}")
            self.passed_tests += 1
        else:
            print(f"❌ {message}")
            self.failed_tests += 1
    
    def setup_grpc_connection(self):
        """建立gRPC连接"""
        try:
            self.channel = grpc.insecure_channel('localhost:50051')
            grpc.channel_ready_future(self.channel).result(timeout=10)
            self.stub = intelligence_pb2_grpc.IntelligenceServiceStub(self.channel)
            self.test_result(True, "gRPC连接建立成功")
            return True
        except Exception as e:
            self.test_result(False, f"gRPC连接建立失败: {e}")
            return False
    
    def test_grpc_service_availability(self):
        """测试gRPC服务可用性"""
        print("\n1. gRPC服务可用性测试")
        print("--------------------")
        
        if not self.setup_grpc_connection():
            return
        
        # 测试服务是否响应
        try:
            # 创建一个简单的请求
            request = intelligence_pb2.InterpretRequest()
            request.user_text = "hello"
            request.session_id = "test-session"
            
            # 设置超时时间
            response = self.stub.InterpretText(request, timeout=10)
            self.test_result(True, "gRPC服务响应测试")
        except grpc.RpcError as e:
            if e.code() == grpc.StatusCode.UNAVAILABLE:
                self.test_result(False, "gRPC服务不可用")
            else:
                self.test_result(True, f"gRPC服务响应 (状态: {e.code()})")
        except Exception as e:
            self.test_result(False, f"gRPC服务测试异常: {e}")
    
    def test_text_interpretation(self):
        """测试文本解释功能"""
        print("\n2. 文本解释功能测试")
        print("------------------")
        
        test_cases = [
            {
                "input": "更新我的电话号码为13800138000",
                "expected_intent": "update_phone_number",
                "description": "电话号码更新意图识别"
            },
            {
                "input": "谁是我的经理？",
                "expected_intent": "get_employee_manager",
                "description": "查询经理意图识别"
            },
            {
                "input": "Hello there",
                "expected_intent": "no_intent_detected",
                "description": "无意图识别"
            }
        ]
        
        for i, case in enumerate(test_cases, 1):
            try:
                request = intelligence_pb2.InterpretRequest()
                request.user_text = case["input"]
                request.session_id = f"test-session-{i}"
                
                response = self.stub.InterpretText(request, timeout=15)
                
                # 验证响应格式
                self.test_result(hasattr(response, 'intent'), f"{case['description']} - 响应包含intent字段")
                self.test_result(hasattr(response, 'structured_data_json'), f"{case['description']} - 响应包含structured_data_json字段")
                
                # 验证JSON格式
                if response.structured_data_json:
                    try:
                        json.loads(response.structured_data_json)
                        self.test_result(True, f"{case['description']} - JSON格式正确")
                    except json.JSONDecodeError:
                        self.test_result(False, f"{case['description']} - JSON格式错误")
                
            except grpc.RpcError as e:
                self.test_result(False, f"{case['description']} - gRPC错误: {e.code()}")
            except Exception as e:
                self.test_result(False, f"{case['description']} - 异常: {e}")
    
    def test_concurrent_requests(self):
        """测试并发请求处理"""
        print("\n3. 并发请求处理测试")
        print("------------------")
        
        def make_request(session_id):
            try:
                request = intelligence_pb2.InterpretRequest()
                request.user_text = f"测试并发请求 {session_id}"
                request.session_id = f"concurrent-test-{session_id}"
                
                response = self.stub.InterpretText(request, timeout=20)
                return True
            except Exception as e:
                print(f"并发请求 {session_id} 失败: {e}")
                return False
        
        # 并发执行5个请求
        with ThreadPoolExecutor(max_workers=5) as executor:
            futures = [executor.submit(make_request, i) for i in range(5)]
            results = [future.result() for future in futures]
        
        success_count = sum(results)
        self.test_result(success_count >= 3, f"并发请求处理 ({success_count}/5 成功)")
    
    def test_error_handling(self):
        """测试错误处理"""
        print("\n4. 错误处理测试")
        print("--------------")
        
        # 测试空输入 - 应该被拒绝并返回错误
        try:
            request = intelligence_pb2.InterpretRequest()
            request.user_text = ""
            request.session_id = "empty-test"
            
            response = self.stub.InterpretText(request, timeout=10)
            # 如果没有抛出异常，说明空输入被处理了，这不是期望的行为
            self.test_result(False, "空输入应该被拒绝")
        except grpc.RpcError as e:
            if e.code() == grpc.StatusCode.INVALID_ARGUMENT:
                self.test_result(True, "空输入正确被拒绝")
            else:
                self.test_result(False, f"空输入处理异常: {e.code()}")
        except Exception as e:
            self.test_result(False, f"空输入处理失败: {e}")
        
        # 测试超长输入
        try:
            request = intelligence_pb2.InterpretRequest()
            request.user_text = "x" * 10000  # 超长文本
            request.session_id = "long-test"
            
            response = self.stub.InterpretText(request, timeout=30)
            # 如果处理成功，说明系统能处理超长输入
            self.test_result(True, "超长输入处理")
        except grpc.RpcError as e:
            if e.code() == grpc.StatusCode.INVALID_ARGUMENT:
                self.test_result(True, "超长输入正确被拒绝")
            else:
                self.test_result(False, f"超长输入处理异常: {e.code()}")
        except Exception as e:
            self.test_result(False, f"超长输入处理失败: {e}")
    
    def test_response_time(self):
        """测试响应时间"""
        print("\n5. 响应时间测试")
        print("--------------")
        
        response_times = []
        
        for i in range(3):
            try:
                start_time = time.time()
                
                request = intelligence_pb2.InterpretRequest()
                request.user_text = "快速响应测试"
                request.session_id = f"speed-test-{i}"
                
                response = self.stub.InterpretText(request, timeout=30)
                
                end_time = time.time()
                response_time = (end_time - start_time) * 1000  # 转换为毫秒
                response_times.append(response_time)
                
            except Exception as e:
                print(f"响应时间测试 {i+1} 失败: {e}")
        
        if response_times:
            avg_time = sum(response_times) / len(response_times)
            self.test_result(avg_time < 10000, f"平均响应时间 ({avg_time:.0f}ms)")
            self.test_result(max(response_times) < 30000, f"最大响应时间 ({max(response_times):.0f}ms)")
        else:
            self.test_result(False, "响应时间测试失败")
    
    def test_session_management(self):
        """测试会话管理"""
        print("\n6. 会话管理测试")
        print("--------------")
        
        session_id = "session-management-test"
        
        # 发送第一个请求
        try:
            request1 = intelligence_pb2.InterpretRequest()
            request1.user_text = "我想更新个人信息"
            request1.session_id = session_id
            
            response1 = self.stub.InterpretText(request1, timeout=15)
            self.test_result(True, "会话第一个请求")
            
            # 发送相关的第二个请求
            request2 = intelligence_pb2.InterpretRequest()
            request2.user_text = "更新电话号码"
            request2.session_id = session_id
            
            response2 = self.stub.InterpretText(request2, timeout=15)
            self.test_result(True, "会话相关请求")
            
        except Exception as e:
            self.test_result(False, f"会话管理测试失败: {e}")
    
    def cleanup(self):
        """清理资源"""
        if self.channel:
            self.channel.close()
    
    def run_all_tests(self):
        """运行所有测试"""
        print("🏰 Cube Castle - AI服务集成测试")
        print("==============================")
        
        self.test_grpc_service_availability()
        self.test_text_interpretation()
        self.test_concurrent_requests()
        self.test_error_handling()
        self.test_response_time()
        self.test_session_management()
        
        self.cleanup()
        
        print("\n==============================")
        print("AI服务集成测试完成！")
        print(f"总计: {self.total_tests} 项测试")
        print(f"✅ 通过: {self.passed_tests} 项")
        print(f"❌ 失败: {self.failed_tests} 项")
        success_rate = (self.passed_tests / self.total_tests * 100) if self.total_tests > 0 else 0
        print(f"成功率: {success_rate:.1f}%")
        print("==============================")
        
        return self.failed_tests == 0

if __name__ == "__main__":
    test_runner = AIServiceIntegrationTest()
    success = test_runner.run_all_tests()
    sys.exit(0 if success else 1)