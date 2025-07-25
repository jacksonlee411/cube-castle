#!/usr/bin/env python3
"""
Python AI服务全面单元测试
"""

import unittest
import time
import hashlib
from unittest.mock import Mock, patch, MagicMock
import sys
import os

# 添加当前目录到Python路径
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

# Mock dependencies before importing main module
sys.modules['grpc'] = Mock()
sys.modules['openai'] = Mock()
sys.modules['intelligence_pb2'] = Mock()
sys.modules['intelligence_pb2_grpc'] = Mock()

# Import after mocking dependencies
from main import AIResponseCache, IntelligenceServiceImpl

class TestAIResponseCache(unittest.TestCase):
    """测试AI响应缓存类"""
    
    def setUp(self):
        """测试前设置"""
        self.cache = AIResponseCache(max_size=5, ttl_seconds=1)  # 小缓存用于测试
        
        # Mock intelligence_pb2.InterpretResponse
        self.mock_response = Mock()
        self.mock_response.intent = "test_intent"
        self.mock_response.structured_data_json = "{\"test\": \"data\"}"
    
    def test_cache_initialization(self):
        """测试缓存初始化"""
        self.assertEqual(self.cache.max_size, 5)
        self.assertEqual(self.cache.ttl_seconds, 1)
        self.assertEqual(len(self.cache.cache), 0)
    
    def test_generate_cache_key(self):
        """测试缓存键生成"""
        text = "测试文本"
        key1 = self.cache._generate_cache_key(text)
        key2 = self.cache._generate_cache_key(text)
        
        # 相同文本应该生成相同的键
        self.assertEqual(key1, key2)
        self.assertEqual(len(key1), 32)  # MD5长度
        
        # 不同文本应该生成不同的键
        key3 = self.cache._generate_cache_key("不同文本")
        self.assertNotEqual(key1, key3)
    
    def test_cache_put_and_get(self):
        """测试缓存存储和获取"""
        text = "测试文本"
        
        # 首次获取应该返回None
        result = self.cache.get(text)
        self.assertIsNone(result)
        
        # 存储响应
        self.cache.put(text, self.mock_response)
        
        # 再次获取应该返回缓存的响应
        result = self.cache.get(text)
        self.assertIsNotNone(result)
        self.assertEqual(result.intent, "test_intent")
    
    def test_cache_expiration(self):
        """测试缓存过期"""
        text = "测试文本"
        
        # 存储响应
        self.cache.put(text, self.mock_response)
        
        # 立即获取应该成功
        result = self.cache.get(text)
        self.assertIsNotNone(result)
        
        # 等待过期
        time.sleep(1.1)
        
        # 过期后获取应该返回None
        result = self.cache.get(text)
        self.assertIsNone(result)
    
    def test_cache_max_size(self):
        """测试缓存最大大小限制"""
        # 填满缓存
        for i in range(6):  # 超过max_size=5
            text = f"测试文本{i}"
            mock_resp = Mock()
            mock_resp.intent = f"intent_{i}"
            self.cache.put(text, mock_resp)
        
        # 缓存大小不应该超过限制
        self.assertLessEqual(len(self.cache.cache), 5)
    
    def test_cache_cleanup_expired(self):
        """测试过期缓存清理"""
        # 添加一些缓存项
        for i in range(3):
            text = f"测试文本{i}"
            mock_resp = Mock()
            mock_resp.intent = f"intent_{i}"
            self.cache.put(text, mock_resp)
        
        self.assertEqual(len(self.cache.cache), 3)
        
        # 等待过期
        time.sleep(1.1)
        
        # 调用清理方法
        self.cache._cleanup_expired()
        
        # 所有过期项应该被清理
        self.assertEqual(len(self.cache.cache), 0)

class TestIntelligenceServiceImpl(unittest.TestCase):
    """测试IntelligenceServiceImpl类"""
    
    def setUp(self):
        """测试前设置"""
        # Mock grpc context
        self.mock_context = Mock()
        
        # Mock request
        self.mock_request = Mock()
        self.mock_request.user_text = "测试文本"
        self.mock_request.session_id = "test-session"
        
        # Mock OpenAI client response
        self.mock_openai_response = Mock()
        self.mock_message = Mock()
        self.mock_message.tool_calls = None
        self.mock_openai_response.choices = [Mock()]
        self.mock_openai_response.choices[0].message = self.mock_message
        
        # Initialize service
        self.service = IntelligenceServiceImpl()
    
    def test_service_initialization(self):
        """测试服务初始化"""
        self.assertIsNotNone(self.service.executor)
        self.assertEqual(self.service.executor._max_workers, 20)
    
    def test_empty_input_validation(self):
        """测试空输入验证"""
        # 测试空字符串
        self.mock_request.user_text = ""
        
        with patch('intelligence_pb2.InterpretResponse') as mock_response:
            result = self.service.InterpretText(self.mock_request, self.mock_context)
            
            # 应该设置错误码和错误信息
            self.mock_context.set_code.assert_called()
            self.mock_context.set_details.assert_called_with("User text cannot be empty")
    
    def test_long_input_validation(self):
        """测试超长输入验证"""
        # 测试超长文本
        self.mock_request.user_text = "x" * 6000  # 超过5000字符限制
        
        with patch('intelligence_pb2.InterpretResponse') as mock_response:
            result = self.service.InterpretText(self.mock_request, self.mock_context)
            
            # 应该设置错误码和错误信息
            self.mock_context.set_code.assert_called()
            self.mock_context.set_details.assert_called_with("User text is too long (max 5000 characters)")
    
    @patch('main.ai_cache')
    def test_cache_hit(self, mock_cache):
        """测试缓存命中"""
        # 设置缓存命中
        mock_cached_response = Mock()
        mock_cached_response.intent = "cached_intent"
        mock_cache.get.return_value = mock_cached_response
        
        result = self.service.InterpretText(self.mock_request, self.mock_context)
        
        # 应该返回缓存的响应
        mock_cache.get.assert_called_with("测试文本")
        self.assertEqual(result, mock_cached_response)
    
    @patch('main.ai_cache')
    @patch('main.client')
    def test_no_intent_detected(self, mock_client, mock_cache):
        """测试未检测到意图的情况"""
        # 设置缓存未命中
        mock_cache.get.return_value = None
        
        # 设置OpenAI响应（无工具调用）
        mock_client.chat.completions.create.return_value = self.mock_openai_response
        
        # Mock InterpretResponse
        with patch('intelligence_pb2.InterpretResponse') as mock_response_class:
            mock_response_instance = Mock()
            mock_response_class.return_value = mock_response_instance
            
            result = self.service.InterpretText(self.mock_request, self.mock_context)
            
            # 应该调用OpenAI API
            mock_client.chat.completions.create.assert_called_once()
            
            # 应该创建并存储响应
            mock_cache.put.assert_called()
    
    @patch('main.ai_cache')
    @patch('main.client')
    def test_function_call_detected(self, mock_client, mock_cache):
        """测试检测到函数调用的情况"""
        # 设置缓存未命中
        mock_cache.get.return_value = None
        
        # 设置OpenAI响应（有工具调用）
        mock_tool_call = Mock()
        mock_tool_call.function.name = "update_phone_number"
        mock_tool_call.function.arguments = '{"employee_id": "123", "phone": "456"}'
        
        self.mock_message.tool_calls = [mock_tool_call]
        mock_client.chat.completions.create.return_value = self.mock_openai_response
        
        # Mock InterpretResponse
        with patch('intelligence_pb2.InterpretResponse') as mock_response_class:
            mock_response_instance = Mock()
            mock_response_class.return_value = mock_response_instance
            
            result = self.service.InterpretText(self.mock_request, self.mock_context)
            
            # 应该调用OpenAI API
            mock_client.chat.completions.create.assert_called_once()
            
            # 应该创建正确的响应
            mock_response_class.assert_called_with(
                intent="update_phone_number",
                structured_data_json='{"employee_id": "123", "phone": "456"}'
            )
            
            # 应该存储响应到缓存
            mock_cache.put.assert_called()
    
    @patch('main.ai_cache')
    @patch('main.client')
    def test_openai_error_handling(self, mock_client, mock_cache):
        """测试OpenAI API错误处理"""
        # 设置缓存未命中
        mock_cache.get.return_value = None
        
        # 设置OpenAI API异常
        mock_client.chat.completions.create.side_effect = Exception("API Error")
        
        with patch('intelligence_pb2.InterpretResponse') as mock_response_class:
            result = self.service.InterpretText(self.mock_request, self.mock_context)
            
            # 应该设置错误码和错误信息
            self.mock_context.set_code.assert_called()
            self.mock_context.set_details.assert_called()


class TestAIServiceIntegration(unittest.TestCase):
    """AI服务集成测试"""
    
    def test_cache_and_service_integration(self):
        """测试缓存与服务的集成"""
        cache = AIResponseCache(max_size=10, ttl_seconds=60)
        
        # 创建mock响应
        mock_response = Mock()
        mock_response.intent = "test_intent"
        mock_response.structured_data_json = "{\"test\": \"data\"}"
        
        text = "集成测试文本"
        
        # 首次存储
        cache.put(text, mock_response)
        
        # 验证可以获取
        result = cache.get(text)
        self.assertIsNotNone(result)
        self.assertEqual(result.intent, "test_intent")
        
        # 验证深拷贝工作正常
        self.assertNotEqual(id(result), id(mock_response))


if __name__ == '__main__':
    # 运行测试
    unittest.main(verbosity=2)