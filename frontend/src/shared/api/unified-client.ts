/**
 * 统一的API客户端配置
 * 消除重复的GraphQL和REST客户端实现
 * 基于CQRS架构：查询使用GraphQL，命令使用REST API
 */
import { logger } from '@/shared/utils/logger';
import { authManager } from "./auth";
import { env } from "../config/environment";
import { authEvents } from "../auth/events";
import type { GraphQLResponse } from "../types";
import type { JsonValue } from "../types/json";
import { isJsonObject } from "../types/json";
// import { CQRS_ENDPOINTS } from '../config/ports'; // TODO: 将来可能用于直接端点配置

// 🔧 CQRS架构端点配置 - 使用代理避免CORS问题
const API_ENDPOINTS = {
  GRAPHQL_QUERY: "/graphql", // 查询服务 (PostgreSQL GraphQL) - 通过Vite代理
  REST_COMMAND: "/api/v1", // 命令服务 (REST API) - 通过Vite代理
} as const;

/**
 * 统一的GraphQL客户端 - 专用于查询操作
 * 遵循CQRS原则：所有查询统一使用GraphQL
 */
export class UnifiedGraphQLClient {
  private endpoint: string;

  constructor(endpoint: string = API_ENDPOINTS.GRAPHQL_QUERY) {
    this.endpoint = endpoint;
  }

  async request<T>(
    query: string,
    variables?: Record<string, JsonValue>,
  ): Promise<T> {
    const doRequest = async (): Promise<Response> => {
      // 🔧 开发和生产环境都需要JWT认证
      const headers: Record<string, string> = {
        "Content-Type": "application/json",
        // 租户头：优先使用会话返回的 tenantId，回退到环境默认
        "X-Tenant-ID": authManager.getTenantId() || env.defaultTenantId,
      };

      // 所有环境都需要JWT认证
      const accessToken = await authManager.getAccessToken();
      if (accessToken) {
        headers["Authorization"] = `Bearer ${accessToken}`;
      }

      return fetch(this.endpoint, {
        method: "POST",
        headers,
        body: JSON.stringify({
          query,
          variables,
        }),
      });
    };

    let retried = false;
    try {
      // 获取OAuth访问令牌
      let response = await doRequest();

      if (!response.ok) {
        // 401：强制刷新令牌并重试一次
        if (response.status === 401) {
          logger.warn(
            "[GraphQL Client] 401 未认证，尝试强制刷新令牌并重试一次",
          );
          if (!retried) {
            retried = true;
            await authManager.forceRefresh();
            response = await doRequest();
            if (!response.ok) {
              authEvents.emitUnauthorized();
              throw new Error("认证已过期，请刷新页面重新登录");
            }
          } else {
            authEvents.emitUnauthorized();
            throw new Error("认证已过期，请刷新页面重新登录");
          }
        }

        // 403：区分租户访问与权限不足
        if (response.status === 403) {
          try {
            const text = await response.text();
            const maybeJson = text ? JSON.parse(text) : undefined;
            const code = maybeJson?.error?.code as string | undefined;
            if (
              code === "TENANT_ACCESS_DENIED" ||
              code === "TENANT_MISMATCH" ||
              code === "TENANT_ID_MISMATCH"
            ) {
              throw new Error("无权访问所选租户，请切换到有权限的租户");
            }
            if (code === "INSUFFICIENT_PERMISSIONS") {
              throw new Error("权限不足，无法访问该资源，请联系管理员");
            }
            // 无法解析具体码时的兜底
            throw new Error("访问被禁止：请检查权限或租户设置");
          } catch (e) {
            if (e instanceof SyntaxError) {
              // 非JSON错误体
              throw new Error("访问被禁止：请检查权限或租户设置");
            }
            throw e;
          }
        }

        // 服务器内部错误时提供更友好的错误信息
        if (response.status === 500) {
          logger.error("[GraphQL Client] 服务器内部错误:", {
            query,
            variables,
            status: response.status,
          });
          throw new Error("服务器内部错误，请稍后重试或联系管理员");
        }

        throw new Error(
          `GraphQL Error: ${response.status} ${response.statusText}`,
        );
      }

      const responseBody = await response.json();

      // 检查是否为企业级API响应信封格式
      if (responseBody.success !== undefined) {
        // 企业级信封格式: {success: true, data: {...}, message: "...", timestamp: "..."}
        if (!responseBody.success) {
          const errorMsg =
            responseBody.error?.message ||
            responseBody.message ||
            "API调用失败";
          throw new Error(`API Error: ${errorMsg}`);
        }

        if (!responseBody.data) {
          throw new Error("API Error: No data returned");
        }

        return responseBody.data as T;
      } else {
        // 标准GraphQL格式: {data: {...}, errors: [...]}
        const result = responseBody as GraphQLResponse<T>;

        if (result.errors && result.errors.length > 0) {
          throw new Error(`GraphQL Error: ${result.errors[0].message}`);
        }

        if (!result.data) {
          throw new Error("GraphQL Error: No data returned");
        }

        return result.data;
      }
    } catch (error) {
      logger.error("GraphQL request failed:", { query, variables, error });
      throw error;
    }
  }
}

export interface RESTRequestOptions extends RequestInit {
  includeRawResponse?: boolean;
}

export interface RESTResponseMeta<T> {
  data: T;
  headers: Record<string, string>;
  response: Response;
}

const normalizeHeaders = (response: Response): Record<string, string> => {
  const headers: Record<string, string> = {};
  response.headers.forEach((value, key) => {
    headers[key.toLowerCase()] = value;
  });
  return headers;
};

/**
 * 统一的REST API客户端 - 专用于命令操作
 * 遵循CQRS原则：所有命令统一使用REST API
 */
export class UnifiedRESTClient {
  private baseURL: string;
  private defaultHeaders: Record<string, string>;

  constructor(baseURL: string = API_ENDPOINTS.REST_COMMAND) {
    this.baseURL = baseURL;
    this.defaultHeaders = {
      "Content-Type": "application/json",
      // 注意：实际请求时会覆盖为 authManager.getTenantId() || env.defaultTenantId
      "X-Tenant-ID": env.defaultTenantId,
    };
  }

  async request<T>(
    endpoint: string,
    options: RESTRequestOptions & { includeRawResponse: true },
  ): Promise<RESTResponseMeta<T>>;
  async request<T>(endpoint: string, options?: RESTRequestOptions): Promise<T>;
  async request<T>(
    endpoint: string,
    options: RESTRequestOptions = {},
  ): Promise<T | RESTResponseMeta<T>> {
    const { includeRawResponse, ...fetchOptions } = options;
    const url = `${this.baseURL}${endpoint}`;

    const buildHeaders = async (): Promise<Record<string, string>> => {
      const headers: Record<string, string> = {
        ...this.defaultHeaders,
        "X-Tenant-ID": authManager.getTenantId() || env.defaultTenantId,
      };

      const customHeaders = new Headers(
        fetchOptions.headers as HeadersInit | undefined,
      );
      const hasCustomAuthorization = customHeaders.has("Authorization");

      const accessToken = await authManager.getAccessToken();
      if (accessToken && !hasCustomAuthorization) {
        headers.Authorization = `Bearer ${accessToken}`;
      }

      customHeaders.forEach((value, key) => {
        if (value === undefined || value === null) {
          return;
        }
        headers[key] = value;
      });

      return headers;
    };

    const doRequest = async (): Promise<Response> => {
      const headers = await buildHeaders();
      return fetch(url, {
        ...fetchOptions,
        headers,
      });
    };

    const readBody = async (
      response: Response,
    ): Promise<JsonValue> => {
      const contentType = response.headers.get("content-type") || "";
      const text = await response.text();

      if (!text) {
        return {};
      }

      const looksLikeJson =
        contentType.includes("application/json") || /^(\s*[[{])/.test(text);
      if (!looksLikeJson) {
        if (!response.ok) {
          logger.error("[REST Client] 非JSON错误体，返回HTTP错误:", {
            endpoint,
            status: response.status,
            statusText: response.statusText,
          });
          throw new Error(
            `REST Error: ${response.status} ${response.statusText}`,
          );
        }
        logger.error("[REST Client] JSON解析失败: 响应非JSON", {
          endpoint,
          text,
        });
        throw new Error(
          `响应解析失败: ${text.substring(0, 100)}${text.length > 100 ? "..." : ""}`,
        );
      }

      try {
        return JSON.parse(text) as JsonValue;
      } catch (parseError) {
        if (!response.ok) {
          logger.error("[REST Client] JSON解析失败 (错误响应):", {
            endpoint,
            status: response.status,
            statusText: response.statusText,
            text,
          });
          throw new Error(
            `REST Error: ${response.status} ${response.statusText}`,
          );
        }
        logger.error("[REST Client] JSON解析失败:", {
          endpoint,
          text,
          parseError,
        });
        throw new Error(
          `响应解析失败: ${text.substring(0, 100)}${text.length > 100 ? "..." : ""}`,
        );
      }
    };

    let retried = false;
    try {
      let response = await doRequest();
      let result = await readBody(response);

      if (!response.ok && response.status === 401) {
        logger.warn("[REST Client] 401 未认证，尝试强制刷新令牌并重试一次");
        if (!retried) {
          retried = true;
          await authManager.forceRefresh();
          response = await doRequest();
          result = await readBody(response);
        } else {
          authEvents.emitUnauthorized();
          throw new Error("认证已过期，请刷新页面重新登录");
        }

        if (!response.ok) {
          authEvents.emitUnauthorized();
          throw new Error("认证已过期，请刷新页面重新登录");
        }
      }

      if (!response.ok) {
        if (response.status === 403) {
          const code =
            isJsonObject(result) && "error" in result
              ? (result.error as { code?: string })?.code
              : undefined;
          if (
            code === "TENANT_ACCESS_DENIED" ||
            code === "TENANT_MISMATCH" ||
            code === "TENANT_ID_MISMATCH"
          ) {
            throw new Error("无权访问所选租户，请切换到有权限的租户");
          }
          if (code === "INSUFFICIENT_PERMISSIONS") {
            throw new Error("权限不足，无法执行此操作，请联系管理员");
          }
          throw new Error("访问被禁止：请检查权限或租户设置");
        }

        if (response.status === 500) {
          logger.error("[REST Client] 服务器内部错误:", {
            endpoint,
            status: response.status,
            result,
          });
          throw new Error("服务器内部错误，请稍后重试或联系管理员");
        }

        if (isJsonObject(result) && "error" in result) {
          const errorInfo = result.error as { message?: string };
          if (errorInfo && errorInfo.message) {
            throw new Error(errorInfo.message);
          }
        }

        throw new Error(
          `REST Error: ${response.status} ${response.statusText}`,
        );
      }

      const payload = (result ?? {}) as T;

      if (includeRawResponse) {
        return {
          data: payload,
          headers: normalizeHeaders(response),
          response,
        };
      }

      return payload;
    } catch (error) {
      logger.error("REST request failed:", { endpoint, options, error });
      throw error;
    }
  }
}

/**
 * 未认证REST客户端（用于OAuth/会话端点）
 * - 不自动附加 Authorization 头
 * - 允许传入 credentials、headers 等原样透传
 */
export class UnauthenticatedRESTClient {
  private baseURL: string;

  constructor(baseURL: string = "") {
    this.baseURL = baseURL;
  }

  async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const url = `${this.baseURL}${endpoint}`;
    try {
      const response: Response = await fetch(url, options);
      // 兼容测试环境的最小 fetch mock：headers/文本体可能不存在
      const contentType = response?.headers?.get?.("content-type") || "";
      let text = "";
      if (typeof response?.text === "function") {
        text = await response.text();
      } else if (typeof response?.json === "function") {
        // 某些测试仅提供 json()，则直接读取
        try {
          const j = await response.json();
          return (j ?? {}) as T;
        } catch {
          // ignore, fallback to empty text
        }
      }
      let json: JsonValue | undefined;
      if (text) {
        const looksLikeJson =
          contentType.includes("application/json") || /^(\s*[[{])/.test(text);
        if (looksLikeJson) {
          try {
            json = JSON.parse(text) as JsonValue;
          } catch {
            /* ignore parse errors for non-JSON bodies */
          }
        }
      }
      if (!response?.ok) {
        let message = `${response.status} ${response.statusText}`;
        if (json && isJsonObject(json) && "error" in json) {
          const errVal = json.error;
          if (errVal && isJsonObject(errVal) && typeof errVal.message === "string") {
            const m = errVal.message;
            if (m.trim()) {
              message = m;
            }
          }
        }
        throw new Error(message);
      }
      return (json ?? ({} as JsonValue)) as T;
    } catch (error) {
      logger.error("[UnauthREST] request failed:", {
        endpoint,
        options,
        error,
      });
      throw error;
    }
  }
}
// 🔧 单例实例 - 全局使用统一客户端
export const unifiedGraphQLClient = new UnifiedGraphQLClient();
export const unifiedRESTClient = new UnifiedRESTClient();
export const unauthenticatedRESTClient = new UnauthenticatedRESTClient();

// 📋 客户端工厂方法 - 支持自定义配置
export const createGraphQLClient = (endpoint?: string) =>
  new UnifiedGraphQLClient(endpoint);
export const createRESTClient = (baseURL?: string) =>
  new UnifiedRESTClient(baseURL);

// 🔧 架构原则检查器 - 开发模式下验证正确使用
export const validateCQRSUsage = (
  operation: "query" | "command",
  method: string,
) => {
  if (process.env.NODE_ENV === "development") {
    if (operation === "query" && !method.includes("GraphQL")) {
      logger.warn("⚠️ CQRS违反: 查询操作应该使用GraphQL客户端");
    }
    if (operation === "command" && !method.includes("REST")) {
      logger.warn("⚠️ CQRS违反: 命令操作应该使用REST客户端");
    }
  }
};

export default {
  graphql: unifiedGraphQLClient,
  rest: unifiedRESTClient,
  unauth: unauthenticatedRESTClient,
  validateCQRSUsage,
};
