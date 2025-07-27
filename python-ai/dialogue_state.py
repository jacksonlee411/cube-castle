"""
Redis对话状态管理器
负责处理AI服务的对话历史、上下文状态的持久化存储
"""

import redis
import json
import logging
import time
from typing import Dict, List, Optional, Tuple
from dataclasses import dataclass
from datetime import datetime, timedelta


@dataclass
class ChatMessage:
    """聊天消息数据类"""
    role: str  # 'user' 或 'assistant'
    content: str
    timestamp: float
    intent: Optional[str] = None
    metadata: Optional[Dict] = None


@dataclass
class ConversationContext:
    """对话上下文数据类"""
    session_id: str
    user_id: Optional[str] = None
    tenant_id: Optional[str] = None
    last_intent: Optional[str] = None
    conversation_state: str = "active"  # active, paused, ended
    created_at: Optional[float] = None
    updated_at: Optional[float] = None


class DialogueStateManager:
    """对话状态管理器"""
    
    def __init__(
        self, 
        redis_host: str = 'localhost', 
        redis_port: int = 6379, 
        redis_db: int = 0,
        session_ttl: int = 1800,  # 30分钟
        max_history_length: int = 20  # 最多保留20轮对话
    ):
        """
        初始化对话状态管理器
        
        Args:
            redis_host: Redis服务器地址
            redis_port: Redis服务器端口
            redis_db: Redis数据库编号
            session_ttl: 会话存活时间（秒）
            max_history_length: 最大历史记录长度
        """
        try:
            self.redis_client = redis.Redis(
                host=redis_host,
                port=redis_port,
                db=redis_db,
                decode_responses=True,
                socket_connect_timeout=5,
                socket_timeout=5,
                retry_on_timeout=True,
                health_check_interval=30
            )
            # 测试连接
            self.redis_client.ping()
            logging.info(f"Successfully connected to Redis at {redis_host}:{redis_port}")
        except redis.ConnectionError as e:
            logging.error(f"Failed to connect to Redis: {e}")
            raise
        
        self.session_ttl = session_ttl
        self.max_history_length = max_history_length
        
        # 键前缀
        self.HISTORY_PREFIX = "chat:history:"
        self.CONTEXT_PREFIX = "chat:context:"
        self.SESSION_PREFIX = "chat:session:"
    
    def _generate_keys(self, session_id: str) -> Tuple[str, str, str]:
        """生成Redis键"""
        history_key = f"{self.HISTORY_PREFIX}{session_id}"
        context_key = f"{self.CONTEXT_PREFIX}{session_id}"
        session_key = f"{self.SESSION_PREFIX}{session_id}"
        return history_key, context_key, session_key
    
    def save_conversation_turn(
        self, 
        session_id: str, 
        user_message: ChatMessage, 
        assistant_message: ChatMessage,
        context_updates: Optional[Dict] = None
    ) -> bool:
        """
        保存一轮对话到Redis
        
        Args:
            session_id: 会话ID
            user_message: 用户消息
            assistant_message: AI助手消息
            context_updates: 上下文更新
            
        Returns:
            bool: 保存是否成功
        """
        try:
            history_key, context_key, session_key = self._generate_keys(session_id)
            
            # 使用Pipeline提高性能
            pipeline = self.redis_client.pipeline()
            
            # 保存对话历史
            user_data = {
                "role": user_message.role,
                "content": user_message.content,
                "timestamp": user_message.timestamp,
                "intent": user_message.intent,
                "metadata": json.dumps(user_message.metadata or {})
            }
            
            assistant_data = {
                "role": assistant_message.role,
                "content": assistant_message.content,
                "timestamp": assistant_message.timestamp,
                "intent": assistant_message.intent,
                "metadata": json.dumps(assistant_message.metadata or {})
            }
            
            # 将消息添加到历史列表（左侧推入，保持时间顺序）
            pipeline.lpush(history_key, json.dumps(user_data))
            pipeline.lpush(history_key, json.dumps(assistant_data))
            
            # 限制历史长度
            pipeline.ltrim(history_key, 0, self.max_history_length - 1)
            
            # 更新上下文
            if context_updates:
                pipeline.hset(context_key, mapping=context_updates)
            
            # 更新会话信息
            session_info = {
                "last_activity": time.time(),
                "message_count": 2,  # 用户消息 + AI消息
                "last_intent": assistant_message.intent or "unknown"
            }
            pipeline.hset(session_key, mapping=session_info)
            
            # 设置过期时间
            pipeline.expire(history_key, self.session_ttl)
            pipeline.expire(context_key, self.session_ttl)
            pipeline.expire(session_key, self.session_ttl)
            
            # 执行所有操作
            results = pipeline.execute()
            
            logging.info(
                f"Saved conversation turn for session {session_id[:8]}..., "
                f"user: {user_message.content[:50]}..., "
                f"assistant: {assistant_message.content[:50]}..."
            )
            
            return True
            
        except Exception as e:
            logging.error(f"Failed to save conversation turn: {e}")
            return False
    
    def get_conversation_history(self, session_id: str, limit: int = 10) -> List[ChatMessage]:
        """
        获取对话历史
        
        Args:
            session_id: 会话ID
            limit: 返回的消息数量限制
            
        Returns:
            List[ChatMessage]: 对话历史列表（按时间倒序）
        """
        try:
            history_key, _, _ = self._generate_keys(session_id)
            
            # 获取历史消息（最近的limit条）
            raw_messages = self.redis_client.lrange(history_key, 0, limit - 1)
            
            messages = []
            for raw_msg in reversed(raw_messages):  # 反转以获得正确的时间顺序
                try:
                    msg_data = json.loads(raw_msg)
                    message = ChatMessage(
                        role=msg_data["role"],
                        content=msg_data["content"],
                        timestamp=msg_data["timestamp"],
                        intent=msg_data.get("intent"),
                        metadata=json.loads(msg_data.get("metadata", "{}"))
                    )
                    messages.append(message)
                except (json.JSONDecodeError, KeyError) as e:
                    logging.warning(f"Failed to parse message: {e}")
                    continue
            
            logging.info(f"Retrieved {len(messages)} messages for session {session_id[:8]}...")
            return messages
            
        except Exception as e:
            logging.error(f"Failed to get conversation history: {e}")
            return []
    
    def get_conversation_context(self, session_id: str) -> Dict:
        """
        获取对话上下文
        
        Args:
            session_id: 会话ID
            
        Returns:
            Dict: 上下文数据
        """
        try:
            _, context_key, session_key = self._generate_keys(session_id)
            
            pipeline = self.redis_client.pipeline()
            pipeline.hgetall(context_key)
            pipeline.hgetall(session_key)
            results = pipeline.execute()
            
            context = results[0] or {}
            session_info = results[1] or {}
            
            return {
                "context": context,
                "session_info": session_info
            }
            
        except Exception as e:
            logging.error(f"Failed to get conversation context: {e}")
            return {"context": {}, "session_info": {}}
    
    def create_session(self, session_id: str, user_id: str = None, tenant_id: str = None) -> bool:
        """
        创建新的会话
        
        Args:
            session_id: 会话ID
            user_id: 用户ID
            tenant_id: 租户ID
            
        Returns:
            bool: 创建是否成功
        """
        try:
            _, context_key, session_key = self._generate_keys(session_id)
            
            current_time = time.time()
            
            # 初始化上下文
            initial_context = {
                "user_id": user_id or "",
                "tenant_id": tenant_id or "",
                "created_at": current_time,
                "conversation_state": "active"
            }
            
            # 初始化会话信息
            session_info = {
                "created_at": current_time,
                "last_activity": current_time,
                "message_count": 0,
                "status": "active"
            }
            
            pipeline = self.redis_client.pipeline()
            pipeline.hset(context_key, mapping=initial_context)
            pipeline.hset(session_key, mapping=session_info)
            pipeline.expire(context_key, self.session_ttl)
            pipeline.expire(session_key, self.session_ttl)
            pipeline.execute()
            
            logging.info(f"Created new session {session_id[:8]}... for user {user_id}")
            return True
            
        except Exception as e:
            logging.error(f"Failed to create session: {e}")
            return False
    
    def end_session(self, session_id: str) -> bool:
        """
        结束会话
        
        Args:
            session_id: 会话ID
            
        Returns:
            bool: 操作是否成功
        """
        try:
            _, context_key, session_key = self._generate_keys(session_id)
            
            # 更新会话状态
            updates = {
                "conversation_state": "ended",
                "ended_at": time.time()
            }
            
            pipeline = self.redis_client.pipeline()
            pipeline.hset(context_key, mapping=updates)
            pipeline.hset(session_key, "status", "ended")
            pipeline.execute()
            
            logging.info(f"Ended session {session_id[:8]}...")
            return True
            
        except Exception as e:
            logging.error(f"Failed to end session: {e}")
            return False
    
    def cleanup_expired_sessions(self) -> int:
        """
        清理过期的会话
        
        Returns:
            int: 清理的会话数量
        """
        try:
            # 获取所有会话键
            session_keys = self.redis_client.keys(f"{self.SESSION_PREFIX}*")
            cleanup_count = 0
            
            current_time = time.time()
            
            for session_key in session_keys:
                session_info = self.redis_client.hgetall(session_key)
                if session_info:
                    last_activity = float(session_info.get("last_activity", 0))
                    if current_time - last_activity > self.session_ttl:
                        # 会话已过期，删除相关数据
                        session_id = session_key.replace(self.SESSION_PREFIX, "")
                        history_key, context_key, _ = self._generate_keys(session_id)
                        
                        pipeline = self.redis_client.pipeline()
                        pipeline.delete(history_key)
                        pipeline.delete(context_key)
                        pipeline.delete(session_key)
                        pipeline.execute()
                        
                        cleanup_count += 1
            
            if cleanup_count > 0:
                logging.info(f"Cleaned up {cleanup_count} expired sessions")
            
            return cleanup_count
            
        except Exception as e:
            logging.error(f"Failed to cleanup expired sessions: {e}")
            return 0
    
    def get_session_stats(self) -> Dict:
        """
        获取会话统计信息
        
        Returns:
            Dict: 统计信息
        """
        try:
            # 获取所有活跃会话
            session_keys = self.redis_client.keys(f"{self.SESSION_PREFIX}*")
            active_sessions = 0
            total_messages = 0
            
            for session_key in session_keys:
                session_info = self.redis_client.hgetall(session_key)
                if session_info.get("status") == "active":
                    active_sessions += 1
                    total_messages += int(session_info.get("message_count", 0))
            
            return {
                "active_sessions": active_sessions,
                "total_sessions": len(session_keys),
                "total_messages": total_messages,
                "redis_memory_usage": self.redis_client.memory_usage(),
                "redis_info": self.redis_client.info("memory")
            }
            
        except Exception as e:
            logging.error(f"Failed to get session stats: {e}")
            return {}
    
    def health_check(self) -> Dict:
        """
        健康检查
        
        Returns:
            Dict: 健康状态信息
        """
        try:
            # 测试Redis连接
            ping_result = self.redis_client.ping()
            redis_info = self.redis_client.info()
            
            return {
                "status": "healthy" if ping_result else "unhealthy",
                "redis_ping": ping_result,
                "redis_version": redis_info.get("redis_version"),
                "redis_uptime": redis_info.get("uptime_in_seconds"),
                "redis_connected_clients": redis_info.get("connected_clients"),
                "redis_used_memory": redis_info.get("used_memory_human"),
                "timestamp": time.time()
            }
            
        except Exception as e:
            logging.error(f"Health check failed: {e}")
            return {
                "status": "unhealthy",
                "error": str(e),
                "timestamp": time.time()
            }