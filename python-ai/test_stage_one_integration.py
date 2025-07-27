"""
阶段一功能集成测试
测试Redis对话状态管理、结构化日志和Temporal工作流
"""

import asyncio
import pytest
import time
import uuid
import redis
import logging
from typing import Dict, Any

# 导入我们的模块
from dialogue_state import DialogueStateManager, ChatMessage
import intelligence_pb2
import intelligence_pb2_grpc
import grpc
from grpc import aio as grpc_aio


class TestRedisDialogueState:
    """Redis对话状态管理测试"""
    
    @pytest.fixture
    def redis_client(self):
        """Redis客户端fixture"""
        try:
            client = redis.Redis(host='localhost', port=6379, decode_responses=True)
            client.ping()
            return client
        except Exception:
            pytest.skip("Redis server not available")
    
    @pytest.fixture
    def dialogue_manager(self, redis_client):
        """对话状态管理器fixture"""
        return DialogueStateManager(
            redis_host='localhost',
            redis_port=6379,
            session_ttl=300,  # 5分钟测试TTL
            max_history_length=10
        )
    
    def test_dialogue_manager_initialization(self, dialogue_manager):
        """测试对话管理器初始化"""
        assert dialogue_manager is not None
        assert dialogue_manager.session_ttl == 300
        assert dialogue_manager.max_history_length == 10
    
    def test_create_session(self, dialogue_manager):
        """测试创建会话"""
        session_id = str(uuid.uuid4())
        user_id = str(uuid.uuid4())
        tenant_id = str(uuid.uuid4())
        
        result = dialogue_manager.create_session(session_id, user_id, tenant_id)
        assert result is True
        
        # 验证会话上下文
        context = dialogue_manager.get_conversation_context(session_id)
        assert context['context']['user_id'] == user_id
        assert context['context']['tenant_id'] == tenant_id
        assert context['context']['conversation_state'] == 'active'
    
    def test_save_and_retrieve_conversation(self, dialogue_manager):
        """测试保存和检索对话"""
        session_id = str(uuid.uuid4())
        dialogue_manager.create_session(session_id)
        
        # 创建测试消息
        user_message = ChatMessage(
            role="user",
            content="你好，我想查询员工信息",
            timestamp=time.time(),
            intent="list_employees"
        )
        
        assistant_message = ChatMessage(
            role="assistant",
            content="好的，我可以帮您查询员工信息",
            timestamp=time.time(),
            intent="list_employees"
        )
        
        # 保存对话
        result = dialogue_manager.save_conversation_turn(
            session_id, user_message, assistant_message, 
            {"last_intent": "list_employees"}
        )
        assert result is True
        
        # 检索对话历史
        history = dialogue_manager.get_conversation_history(session_id, limit=10)
        assert len(history) == 2
        assert history[0].role == "user"
        assert history[0].content == "你好，我想查询员工信息"
        assert history[1].role == "assistant"
        assert history[1].content == "好的，我可以帮您查询员工信息"
    
    def test_conversation_context_updates(self, dialogue_manager):
        """测试对话上下文更新"""
        session_id = str(uuid.uuid4())
        dialogue_manager.create_session(session_id)
        
        # 保存多轮对话
        for i in range(3):
            user_msg = ChatMessage(
                role="user",
                content=f"第{i+1}轮用户消息",
                timestamp=time.time(),
                intent=f"intent_{i}"
            )
            
            assistant_msg = ChatMessage(
                role="assistant", 
                content=f"第{i+1}轮助手回复",
                timestamp=time.time(),
                intent=f"intent_{i}"
            )
            
            dialogue_manager.save_conversation_turn(
                session_id, user_msg, assistant_msg,
                {"last_intent": f"intent_{i}", "turn_count": i+1}
            )
        
        # 检查上下文
        context = dialogue_manager.get_conversation_context(session_id)
        assert context['context']['last_intent'] == 'intent_2'
        assert context['context']['turn_count'] == '3'
    
    def test_session_cleanup(self, dialogue_manager, redis_client):
        """测试会话清理"""
        session_id = str(uuid.uuid4())
        dialogue_manager.create_session(session_id)
        
        # 模拟过期会话
        session_key = f"{dialogue_manager.SESSION_PREFIX}{session_id}"
        redis_client.hset(session_key, "last_activity", time.time() - 1000)  # 设置为1000秒前
        
        # 执行清理
        cleaned_count = dialogue_manager.cleanup_expired_sessions()
        assert cleaned_count >= 1
    
    def test_health_check(self, dialogue_manager):
        """测试健康检查"""
        health = dialogue_manager.health_check()
        assert health['status'] == 'healthy'
        assert 'redis_ping' in health
        assert health['redis_ping'] is True


class TestIntelligenceServiceIntegration:
    """AI服务集成测试"""
    
    @pytest.fixture
    async def grpc_server(self):
        """启动测试用的gRPC服务器"""
        # 这里应该启动实际的AI服务
        # 简化测试，假设服务已经运行在localhost:50051
        return "localhost:50051"
    
    @pytest.fixture
    async def grpc_client(self, grpc_server):
        """gRPC客户端fixture"""
        channel = grpc_aio.insecure_channel(grpc_server)
        return intelligence_pb2_grpc.IntelligenceServiceStub(channel)
    
    @pytest.mark.asyncio
    async def test_ai_service_basic_intent_recognition(self, grpc_client):
        """测试基础意图识别"""
        request = intelligence_pb2.InterpretRequest(
            user_text="我想查询员工信息",
            session_id=str(uuid.uuid4())
        )
        
        try:
            response = await grpc_client.InterpretText(request)
            assert response.intent in ["list_employees", "no_intent_detected"]
            assert response.structured_data_json is not None
        except grpc.RpcError as e:
            if e.code() == grpc.StatusCode.UNAVAILABLE:
                pytest.skip("AI service not available for testing")
    
    @pytest.mark.asyncio 
    async def test_ai_service_conversation_memory(self, grpc_client):
        """测试对话记忆功能"""
        session_id = str(uuid.uuid4())
        
        # 第一轮对话
        request1 = intelligence_pb2.InterpretRequest(
            user_text="我想查询员工张三的信息",
            session_id=session_id
        )
        
        try:
            response1 = await grpc_client.InterpretText(request1)
            
            # 第二轮对话（基于上下文）
            request2 = intelligence_pb2.InterpretRequest(
                user_text="他的经理是谁？",
                session_id=session_id
            )
            
            response2 = await grpc_client.InterpretText(request2)
            
            # 验证第二轮对话能够理解上下文
            assert response2.intent in ["get_employee_manager", "list_employees"]
            
        except grpc.RpcError as e:
            if e.code() == grpc.StatusCode.UNAVAILABLE:
                pytest.skip("AI service not available for testing")
    
    @pytest.mark.asyncio
    async def test_ai_service_error_handling(self, grpc_client):
        """测试AI服务错误处理"""
        # 测试空输入
        request = intelligence_pb2.InterpretRequest(
            user_text="",
            session_id=str(uuid.uuid4())
        )
        
        try:
            with pytest.raises(grpc.RpcError) as exc_info:
                await grpc_client.InterpretText(request)
            assert exc_info.value.code() == grpc.StatusCode.INVALID_ARGUMENT
        except grpc.RpcError as e:
            if e.code() == grpc.StatusCode.UNAVAILABLE:
                pytest.skip("AI service not available for testing")


class TestLoggingAndMetrics:
    """结构化日志和监控测试"""
    
    def test_structured_logging_format(self, caplog):
        """测试结构化日志格式"""
        # 这个测试需要Go服务运行，暂时跳过
        pytest.skip("Go service integration test - requires running service")
    
    def test_prometheus_metrics_collection(self):
        """测试Prometheus指标收集"""
        # 这个测试需要Go服务运行，暂时跳过  
        pytest.skip("Go service integration test - requires running service")


class TestE2EIntegration:
    """端到端集成测试"""
    
    @pytest.mark.asyncio
    async def test_complete_conversation_flow(self):
        """测试完整对话流程"""
        # 创建Redis连接
        try:
            dialogue_manager = DialogueStateManager()
        except Exception:
            pytest.skip("Redis not available")
        
        session_id = str(uuid.uuid4())
        
        # 1. 创建会话
        dialogue_manager.create_session(session_id)
        
        # 2. 模拟AI对话
        conversation_turns = [
            ("你好", "您好！我是HR智能助手，有什么可以帮助您的吗？"),
            ("我想查询员工信息", "好的，请告诉我您要查询哪位员工的信息？"),
            ("张三", "我正在为您查询张三的员工信息..."),
            ("他的经理是谁？", "张三的直属经理是李四。")
        ]
        
        for user_text, assistant_text in conversation_turns:
            user_message = ChatMessage(
                role="user",
                content=user_text,
                timestamp=time.time()
            )
            
            assistant_message = ChatMessage(
                role="assistant",
                content=assistant_text,
                timestamp=time.time()
            )
            
            dialogue_manager.save_conversation_turn(
                session_id, user_message, assistant_message
            )
        
        # 3. 验证对话历史
        history = dialogue_manager.get_conversation_history(session_id)
        assert len(history) == 8  # 4轮对话 = 8条消息
        
        # 4. 验证上下文保持
        context = dialogue_manager.get_conversation_context(session_id)
        assert context['session_info']['message_count'] == '2'  # 最后一轮的消息数
    
    def test_workflow_health_checks(self):
        """测试工作流健康检查"""
        # 这需要Temporal服务运行，暂时跳过
        pytest.skip("Temporal service integration test - requires running service")


# 性能测试
class TestPerformance:
    """性能测试"""
    
    def test_redis_performance(self):
        """测试Redis操作性能"""
        try:
            dialogue_manager = DialogueStateManager()
        except Exception:
            pytest.skip("Redis not available")
        
        session_id = str(uuid.uuid4())
        dialogue_manager.create_session(session_id)
        
        # 测试批量保存对话的性能
        start_time = time.time()
        
        for i in range(100):
            user_message = ChatMessage(
                role="user",
                content=f"测试消息 {i}",
                timestamp=time.time()
            )
            
            assistant_message = ChatMessage(
                role="assistant",
                content=f"回复消息 {i}",
                timestamp=time.time()
            )
            
            dialogue_manager.save_conversation_turn(
                session_id, user_message, assistant_message
            )
        
        end_time = time.time()
        duration = end_time - start_time
        
        # 验证性能：100轮对话应该在5秒内完成
        assert duration < 5.0, f"Performance test failed: {duration}s > 5.0s"
        
        # 验证数据正确性
        history = dialogue_manager.get_conversation_history(session_id, limit=200)
        assert len(history) <= dialogue_manager.max_history_length


if __name__ == "__main__":
    # 运行测试
    pytest.main([
        __file__,
        "-v",
        "--tb=short",
        "--show-capture=no"
    ])