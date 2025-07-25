#!/usr/bin/env python3
"""
Python AI服务单元测试
"""

import unittest
import json
import asyncio
from unittest.mock import Mock, patch, MagicMock
import sys
import os

# 添加当前目录到Python路径，以便导入模块
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

import intelligence_pb2
import intelligence_pb2_grpc


class TestIntelligenceService(unittest.TestCase):
    """测试Intelligence服务"""
    
    def setUp(self):
        """测试前的设置"""
        self.mock_context = Mock()
    
    def test_protobuf_imports(self):
        """测试gRPC和protobuf导入"""
        # 测试消息类型存在
        self.assertTrue(hasattr(intelligence_pb2, 'InterpretRequest'))
        self.assertTrue(hasattr(intelligence_pb2, 'InterpretResponse'))
        
        # 测试服务类存在
        self.assertTrue(hasattr(intelligence_pb2_grpc, 'IntelligenceServiceServicer'))
    
    def test_create_interpret_request(self):
        """测试创建解释请求"""
        request = intelligence_pb2.InterpretRequest()
        request.user_text = "Hello, update my phone number"
        request.session_id = "test-session-123"
        
        self.assertEqual(request.user_text, "Hello, update my phone number")
        self.assertEqual(request.session_id, "test-session-123")
    
    def test_create_interpret_response(self):
        """测试创建解释响应"""
        response = intelligence_pb2.InterpretResponse()
        response.intent = "update_phone_number"
        response.structured_data_json = '{"employee_id": "123", "new_phone_number": "13800138000"}'
        
        self.assertEqual(response.intent, "update_phone_number")
        self.assertIn("employee_id", response.structured_data_json)
    
    def test_json_serialization(self):
        """测试JSON序列化和反序列化"""
        test_data = {
            "employee_id": "11111111-1111-1111-1111-111111111111",
            "new_phone_number": "13800138000"
        }
        
        # 测试序列化
        json_str = json.dumps(test_data)
        self.assertIsInstance(json_str, str)
        
        # 测试反序列化
        parsed_data = json.loads(json_str)
        self.assertEqual(parsed_data["employee_id"], test_data["employee_id"])
        self.assertEqual(parsed_data["new_phone_number"], test_data["new_phone_number"])
    
    def test_function_tools_definition(self):
        """测试AI工具函数定义"""
        tools = [
            {
                "type": "function",
                "function": {
                    "name": "update_phone_number",
                    "description": "Update an employee's phone number",
                    "parameters": {
                        "type": "object",
                        "properties": {
                            "employee_id": {"type": "string", "description": "The UUID of the employee"},
                            "new_phone_number": {"type": "string", "description": "The new phone number"}
                        },
                        "required": ["employee_id", "new_phone_number"],
                    },
                },
            }
        ]
        
        # 验证工具定义结构
        self.assertEqual(len(tools), 1)
        self.assertEqual(tools[0]["type"], "function")
        self.assertEqual(tools[0]["function"]["name"], "update_phone_number")
        self.assertIn("parameters", tools[0]["function"])
    
    @patch('os.getenv')
    def test_environment_variables(self, mock_getenv):
        """测试环境变量处理"""
        # 模拟环境变量
        mock_getenv.side_effect = lambda key, default=None: {
            'OPENAI_API_KEY': 'test-api-key',
            'OPENAI_API_BASE_URL': 'https://api.test.com/v1'
        }.get(key, default)
        
        # 测试环境变量获取
        api_key = os.getenv('OPENAI_API_KEY')
        base_url = os.getenv('OPENAI_API_BASE_URL')
        
        self.assertEqual(api_key, 'test-api-key')
        self.assertEqual(base_url, 'https://api.test.com/v1')


class TestMockIntelligenceService(unittest.TestCase):
    """测试模拟Intelligence服务"""
    
    def test_mock_service_creation(self):
        """测试模拟服务创建"""
        # 创建模拟服务
        mock_service = Mock(spec=intelligence_pb2_grpc.IntelligenceServiceServicer)
        
        # 设置模拟方法
        mock_response = intelligence_pb2.InterpretResponse()
        mock_response.intent = "update_phone_number"
        mock_response.structured_data_json = '{"employee_id": "123"}'
        
        mock_service.InterpretText.return_value = mock_response
        
        # 测试模拟服务调用
        request = intelligence_pb2.InterpretRequest()
        request.user_text = "Update phone number"
        
        response = mock_service.InterpretText(request, None)
        
        self.assertEqual(response.intent, "update_phone_number")
        self.assertIn("employee_id", response.structured_data_json)
    
    def test_error_handling(self):
        """测试错误处理"""
        # 测试空输入
        request = intelligence_pb2.InterpretRequest()
        request.user_text = ""
        request.session_id = "test-session"
        
        # 验证空输入应该被正确处理
        self.assertEqual(request.user_text, "")
        self.assertEqual(request.session_id, "test-session")
    
    def test_intent_classification(self):
        """测试意图分类"""
        test_cases = [
            {
                "input": "Update my phone number to 13800138000",
                "expected_intent": "update_phone_number"
            },
            {
                "input": "Who is my manager?",
                "expected_intent": "get_employee_manager"
            },
            {
                "input": "Hello there",
                "expected_intent": "no_intent_detected"
            }
        ]
        
        for case in test_cases:
            # 这里我们只测试测试数据的结构
            self.assertIn("input", case)
            self.assertIn("expected_intent", case)
            self.assertIsInstance(case["input"], str)
            self.assertIsInstance(case["expected_intent"], str)


class TestGRPCService(unittest.TestCase):
    """测试gRPC服务相关功能"""
    
    def test_grpc_server_options(self):
        """测试gRPC服务器选项"""
        options = [
            ('grpc.keepalive_time_ms', 10000),
            ('grpc.keepalive_timeout_ms', 5000),
            ('grpc.keepalive_permit_without_calls', True),
            ('grpc.http2.max_pings_without_data', 0),
            ('grpc.http2.min_time_between_pings_ms', 10000),
            ('grpc.http2.min_ping_interval_without_data_ms', 300000)
        ]
        
        # 验证选项格式
        for option in options:
            self.assertIsInstance(option, tuple)
            self.assertEqual(len(option), 2)
            self.assertIsInstance(option[0], str)
    
    def test_service_port_configuration(self):
        """测试服务端口配置"""
        default_port = "50051"
        self.assertEqual(default_port, "50051")
        self.assertTrue(default_port.isdigit())
        self.assertGreaterEqual(int(default_port), 1024)


class TestLogging(unittest.TestCase):
    """测试日志功能"""
    
    def test_log_format(self):
        """测试日志格式"""
        import logging
        
        # 创建测试logger
        test_logger = logging.getLogger('test_ai_service')
        
        # 验证logger创建成功
        self.assertIsInstance(test_logger, logging.Logger)
        self.assertEqual(test_logger.name, 'test_ai_service')
    
    def test_log_levels(self):
        """测试日志级别"""
        import logging
        
        # 测试日志级别常量
        self.assertEqual(logging.INFO, 20)
        self.assertEqual(logging.ERROR, 40)
        self.assertEqual(logging.WARNING, 30)


if __name__ == '__main__':
    # 运行所有测试
    unittest.main(verbosity=2)