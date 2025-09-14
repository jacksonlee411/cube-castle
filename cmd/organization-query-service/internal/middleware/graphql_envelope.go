package middleware

import (
    "encoding/json"
    "net/http"

    "cube-castle-deployment-test/internal/types"
)

// GraphQLEnvelopeMiddleware 企业级GraphQL响应信封中间件
type GraphQLEnvelopeMiddleware struct{}

// NewGraphQLEnvelopeMiddleware 创建新的GraphQL企业级信封中间件
func NewGraphQLEnvelopeMiddleware() *GraphQLEnvelopeMiddleware {
	return &GraphQLEnvelopeMiddleware{}
}

// Middleware 包装GraphQL响应为企业级信封格式
func (m *GraphQLEnvelopeMiddleware) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 获取RequestID
			requestID := r.Context().Value("requestId")
			if requestID == nil {
				requestID = "unknown"
			}

			// 创建响应拦截器
			interceptor := &responseInterceptor{
				ResponseWriter: w,
				requestID:      requestID.(string),
			}

			// 调用下游处理器
			next.ServeHTTP(interceptor, r)
		})
	}
}

// responseInterceptor 响应拦截器，将GraphQL标准响应包装为企业级信封
type responseInterceptor struct {
    http.ResponseWriter
    requestID string
    written   bool
}

func (ri *responseInterceptor) Write(data []byte) (int, error) {
	if ri.written {
		return ri.ResponseWriter.Write(data)
	}
	ri.written = true

    // 解析GraphQL响应
    var graphqlResponse map[string]interface{}
    if err := json.Unmarshal(data, &graphqlResponse); err != nil {
        // 如果解析失败，返回原始响应
        return ri.ResponseWriter.Write(data)
    }

    // 检查是否为GraphQL查询响应（包含data字段）
    if _, hasData := graphqlResponse["data"]; hasData {
        // 检查是否有错误
        errorMessage := "Query executed successfully"
		
        if errorsVal, hasErr := graphqlResponse["errors"]; hasErr {
            errorMessage = "Query completed with errors"

            // 权限错误识别：若任一错误message包含"INSUFFICIENT_PERMISSIONS"，映射企业错误码
            code := "GRAPHQL_EXECUTION_ERROR"
            if arr, ok := errorsVal.([]interface{}); ok {
                for _, e := range arr {
                    if m, ok := e.(map[string]interface{}); ok {
                        if msg, ok := m["message"].(string); ok {
                            if msg == "INSUFFICIENT_PERMISSIONS" ||
                               containsIgnoreCase(msg, "INSUFFICIENT_PERMISSIONS") {
                                code = "INSUFFICIENT_PERMISSIONS"
                                errorMessage = "权限不足，无法执行该查询"
                                break
                            }
                        }
                    }
                }
            }

            errorResponse := types.WriteErrorResponse(
                code,
                errorMessage,
                ri.requestID,
                errorsVal,
            )

            ri.ResponseWriter.Header().Set("Content-Type", "application/json")
            responseData, _ := json.Marshal(errorResponse)
            return ri.ResponseWriter.Write(responseData)
        }

        // 构建企业级成功响应信封
        successResponse := types.WriteSuccessResponse(
            graphqlResponse["data"],
            errorMessage,
            ri.requestID,
        )

		// 设置响应头
		ri.ResponseWriter.Header().Set("Content-Type", "application/json")
		
		// 序列化并返回企业级信封响应
		responseData, err := json.Marshal(successResponse)
		if err != nil {
			// 如果序列化失败，返回原始响应
			return ri.ResponseWriter.Write(data)
		}
		
		return ri.ResponseWriter.Write(responseData)
	}

	// 非GraphQL查询响应（如Schema请求），返回原始响应
	return ri.ResponseWriter.Write(data)
}

func (ri *responseInterceptor) WriteHeader(statusCode int) {
    ri.ResponseWriter.WriteHeader(statusCode)
}

// containsIgnoreCase 判断子串（不区分大小写）
func containsIgnoreCase(s, substr string) bool {
    if len(substr) == 0 {
        return true
    }
    // 简单小写比较
    bs := []rune(s)
    bsub := []rune(substr)
    ls := make([]rune, len(bs))
    lsub := make([]rune, len(bsub))
    for i, r := range bs { if r >= 'A' && r <= 'Z' { ls[i] = r + 32 } else { ls[i] = r } }
    for i, r := range bsub { if r >= 'A' && r <= 'Z' { lsub[i] = r + 32 } else { lsub[i] = r } }
    return stringContains(string(ls), string(lsub))
}

// stringContains 使用标准库子串搜索
func stringContains(s, sub string) bool { return len(sub) == 0 || (len(s) >= len(sub) && (indexOf(s, sub) >= 0)) }

// indexOf 朴素查找（避免引入strings包以保持低依赖）：返回首次位置或-1
func indexOf(s, sub string) int {
    n, m := len(s), len(sub)
    if m == 0 { return 0 }
    if m > n { return -1 }
    for i := 0; i <= n-m; i++ {
        if s[i:i+m] == sub { return i }
    }
    return -1
}
