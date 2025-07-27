import asyncio
import logging
from concurrent import futures
import os
import time
import hashlib
from typing import Dict, Optional, Tuple

import grpc
import openai  # 引入openai库
from dotenv import load_dotenv

# 导入我们生成的gRPC代码
import intelligence_pb2
import intelligence_pb2_grpc

# 导入对话状态管理器
from dialogue_state import DialogueStateManager, ChatMessage

# --- 从 .env 文件加载环境变量 ---
load_dotenv()

# --- OpenAI客户端优化配置 ---
# 创建一个OpenAI客户端，优化连接池和超时设置
client = openai.OpenAI(
    api_key=os.getenv("OPENAI_API_KEY"),
    base_url=os.getenv("OPENAI_API_BASE_URL"),
    max_retries=2,  # 减少重试次数以提高响应速度
    timeout=15.0,   # 设置较短的超时时间
)
# -------------------------

# AI响应缓存类
class AIResponseCache:
    def __init__(self, max_size: int = 1000, ttl_seconds: int = 3600):
        """
        初始化AI响应缓存
        :param max_size: 最大缓存条目数
        :param ttl_seconds: 缓存生存时间（秒）
        """
        self.cache: Dict[str, Tuple[intelligence_pb2.InterpretResponse, float]] = {}
        self.max_size = max_size
        self.ttl_seconds = ttl_seconds
    
    def _generate_cache_key(self, user_text: str) -> str:
        """生成缓存键"""
        return hashlib.md5(user_text.encode('utf-8')).hexdigest()
    
    def _is_expired(self, timestamp: float) -> bool:
        """检查缓存是否过期"""
        return time.time() - timestamp > self.ttl_seconds
    
    def _cleanup_expired(self):
        """清理过期的缓存条目"""
        current_time = time.time()
        expired_keys = [
            key for key, (_, timestamp) in self.cache.items()
            if current_time - timestamp > self.ttl_seconds
        ]
        for key in expired_keys:
            del self.cache[key]
    
    def get(self, user_text: str) -> Optional[intelligence_pb2.InterpretResponse]:
        """从缓存中获取响应"""
        cache_key = self._generate_cache_key(user_text)
        
        logging.info(f"检查缓存: {user_text[:30]}... (缓存键: {cache_key[:8]}...)")
        
        if cache_key in self.cache:
            response, timestamp = self.cache[cache_key]
            if not self._is_expired(timestamp):
                logging.info(f"缓存命中: {user_text[:30]}...")
                return response
            else:
                # 移除过期缓存
                logging.info(f"缓存过期，移除: {user_text[:30]}...")
                del self.cache[cache_key]
        else:
            logging.info(f"缓存未命中: {user_text[:30]}...")
        
        return None
    
    def put(self, user_text: str, response: intelligence_pb2.InterpretResponse):
        """将响应存入缓存"""
        cache_key = self._generate_cache_key(user_text)
        
        logging.info(f"准备存储缓存: {user_text[:30]}... (缓存键: {cache_key[:8]}...)")
        
        # 如果缓存已满，先清理过期项
        if len(self.cache) >= self.max_size:
            self._cleanup_expired()
            
            # 如果清理后仍然满了，移除最旧的条目
            if len(self.cache) >= self.max_size:
                oldest_key = min(self.cache.keys(), key=lambda k: self.cache[k][1])
                del self.cache[oldest_key]
        
        # 创建响应的深拷贝以避免引用问题
        cached_response = intelligence_pb2.InterpretResponse()
        cached_response.intent = response.intent
        cached_response.structured_data_json = response.structured_data_json
        
        self.cache[cache_key] = (cached_response, time.time())
        logging.info(f"缓存存储成功: {user_text[:30]}... (缓存大小: {len(self.cache)})")

# 全局缓存实例
ai_cache = AIResponseCache(max_size=500, ttl_seconds=1800)  # 30分钟TTL

# --- gRPC Service Implementation ---
# ... main.py 文件前面的部分保持不变 ...

# --- gRPC Service Implementation ---
class IntelligenceServiceImpl(intelligence_pb2_grpc.IntelligenceServiceServicer):
    def __init__(self):
        """初始化服务，设置工作线程池和对话状态管理器"""
        self.executor = futures.ThreadPoolExecutor(max_workers=20)
        
        # 初始化对话状态管理器
        try:
            redis_host = os.getenv("REDIS_HOST", "localhost")
            redis_port = int(os.getenv("REDIS_PORT", "6379"))
            self.dialogue_manager = DialogueStateManager(
                redis_host=redis_host,
                redis_port=redis_port,
                session_ttl=1800,  # 30分钟
                max_history_length=20
            )
            logging.info("✅ DialogueStateManager initialized successfully")
        except Exception as e:
            logging.warning(f"⚠️ Failed to initialize DialogueStateManager: {e}")
            logging.warning("⚠️ Running without persistent dialogue state")
            self.dialogue_manager = None
    
    def InterpretText(self, request: intelligence_pb2.InterpretRequest, context):
        start_time = time.time()
        logging.info(f"Received text: '{request.user_text}' from session '{request.session_id}'")
        
        # 验证输入 - 拒绝空输入
        if not request.user_text or request.user_text.strip() == "":
            logging.warning("Empty input rejected")
            context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
            context.set_details("User text cannot be empty")
            return intelligence_pb2.InterpretResponse()
        
        # 验证输入长度 - 拒绝过长输入  
        if len(request.user_text) > 5000:
            logging.warning(f"Input too long: {len(request.user_text)} characters")
            context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
            context.set_details("User text is too long (max 5000 characters)")
            return intelligence_pb2.InterpretResponse()

        # 获取对话历史（如果可用）
        conversation_history = []
        if self.dialogue_manager:
            try:
                # 确保会话存在
                self.dialogue_manager.create_session(request.session_id)
                
                # 获取对话历史
                history_messages = self.dialogue_manager.get_conversation_history(
                    request.session_id, limit=10
                )
                
                # 转换为OpenAI消息格式
                for msg in history_messages:
                    conversation_history.append({
                        "role": msg.role,
                        "content": msg.content
                    })
                
                logging.info(f"Retrieved {len(conversation_history)} messages from history")
                
            except Exception as e:
                logging.warning(f"Failed to get conversation history: {e}")

        # 检查简单缓存
        cached_response = ai_cache.get(request.user_text)
        if cached_response is not None and len(conversation_history) == 0:
            logging.info(f"返回缓存结果: {request.user_text[:30]}...")
            return cached_response

        # AI工具定义
        tools = [
            {
                "type": "function",
                "function": {
                    "name": "get_employee_manager",
                    "description": "Get the manager of a specified employee",
                    "parameters": {
                        "type": "object",
                        "properties": {
                            "employee_id": { "type": "string", "description": "The UUID of the employee" }
                        },
                        "required": ["employee_id"],
                    },
                },
            },
            {
                "type": "function",
                "function": {
                    "name": "update_phone_number",
                    "description": "Update an employee's phone number",
                    "parameters": {
                        "type": "object",
                        "properties": {
                            "employee_id": {
                                "type": "string",
                                "description": "The UUID of the employee whose phone number is to be updated",
                            },
                            "new_phone_number": {
                                "type": "string",
                                "description": "The new phone number",
                            }
                        },
                        "required": ["employee_id", "new_phone_number"],
                    },
                },
            },
            {
                "type": "function",
                "function": {
                    "name": "list_employees",
                    "description": "List employees with optional search criteria",
                    "parameters": {
                        "type": "object",
                        "properties": {
                            "search": {
                                "type": "string",
                                "description": "Search term for employee name or number"
                            },
                            "page": {
                                "type": "integer",
                                "description": "Page number (default: 1)"
                            }
                        }
                    },
                },
            }
        ]

        try:
            # 构建包含历史的消息列表
            system_prompt = """你是一个专业的HR系统智能助手，专门识别用户的HR相关意图。

你有记忆能力，可以记住之前的对话内容，请结合上下文进行回复。

核心意图类型和对应的函数：
1. list_employees - 查询员工列表 (关键词: 查询、查看、员工、列表、搜索)
2. update_phone_number - 更新电话号码 (关键词: 更新、修改、电话、手机、号码)
3. get_employee_manager - 查看经理信息 (关键词: 经理、上级、主管、领导)

识别规则：
- 仔细分析用户输入的关键词和上下文
- 结合之前的对话历史理解用户意图
- 如果用户询问或搜索员工，选择list_employees
- 如果用户提到更新、修改电话号码，选择update_phone_number
- 如果用户询问经理、上级信息，选择get_employee_manager
- 提取相关的结构化数据参数

请根据用户输入和对话历史识别意图并调用对应函数。"""

            messages = [{"role": "system", "content": system_prompt}]
            
            # 添加历史对话（最近5轮）
            messages.extend(conversation_history[-10:])
            
            # 添加当前用户消息
            messages.append({"role": "user", "content": request.user_text})

            response = client.chat.completions.create(
                model="deepseek-chat",
                messages=messages,
                tools=tools,
                tool_choice="auto",
                temperature=0.1,
                max_tokens=512,
                stream=False,
            )

            response_message = response.choices[0].message
            tool_calls = response_message.tool_calls
            
            # 处理AI响应
            if tool_calls:
                function_name = tool_calls[0].function.name
                function_args_json = tool_calls[0].function.arguments
                assistant_content = f"我识别到您的意图是：{function_name}，正在为您处理..."

                logging.info(f"LLM wants to call function: {function_name} with args: {function_args_json}")
            else:
                function_name = "no_intent_detected"
                function_args_json = "{}"
                assistant_content = response_message.content or "抱歉，我没有理解您的意图，请尝试重新表达。"

            # 保存对话到Redis（如果可用）
            if self.dialogue_manager:
                try:
                    user_message = ChatMessage(
                        role="user",
                        content=request.user_text,
                        timestamp=start_time,
                        intent=function_name
                    )
                    
                    assistant_message = ChatMessage(
                        role="assistant",
                        content=assistant_content,
                        timestamp=time.time(),
                        intent=function_name
                    )
                    
                    context_updates = {
                        "last_intent": function_name,
                        "last_activity": time.time(),
                        "processing_time": time.time() - start_time
                    }
                    
                    self.dialogue_manager.save_conversation_turn(
                        request.session_id,
                        user_message,
                        assistant_message,
                        context_updates
                    )
                    
                except Exception as e:
                    logging.warning(f"Failed to save conversation to Redis: {e}")

            # 创建响应对象
            result_response = intelligence_pb2.InterpretResponse(
                intent=function_name,
                structured_data_json=function_args_json
            )
            
            # 简单缓存仅用于无历史的单次查询
            if len(conversation_history) == 0:
                ai_cache.put(request.user_text, result_response)
            
            processing_time = time.time() - start_time
            logging.info(f"Request processed in {processing_time:.3f}s: {function_name}")
            
            return result_response

        except Exception as e:
            logging.error(f"Error calling OpenAI: {e}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Error communicating with LLM: {e}")
            return intelligence_pb2.InterpretResponse()

# ... 文件后面的 serve() 和 main() 函数保持不变 ...
async def serve() -> None:
    port = "50051"
    # 增加gRPC服务器的并发处理能力
    options = [
        ('grpc.keepalive_time_ms', 30000),
        ('grpc.keepalive_timeout_ms', 5000),
        ('grpc.keepalive_permit_without_calls', True),
        ('grpc.http2.max_pings_without_data', 0),  
        ('grpc.http2.min_time_between_pings_ms', 10000),
        ('grpc.http2.min_ping_interval_without_data_ms', 300000),
        ('grpc.max_connection_idle_ms', 60000),
    ]
    
    server = grpc.aio.server(
        futures.ThreadPoolExecutor(max_workers=50),
        options=options
    )
    intelligence_pb2_grpc.add_IntelligenceServiceServicer_to_server(
        IntelligenceServiceImpl(), server
    )
    
    server.add_insecure_port(f"[::]:{port}")
    logging.info(f"🧙 Python AI Service 'The Wizard Tower' is listening on gRPC port {port}")
    await server.start()
    await server.wait_for_termination()

if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO)
    asyncio.run(serve())