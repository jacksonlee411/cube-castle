// OAuth 2.0 JWT认证中间件 - 企业级安全标准
const fs = require("fs");
const path = require("path");
const crypto = require("crypto");

// RS256 密钥配置（遵循 CLAUDE.md 原则）
const ISSUER = process.env.JWT_ISSUER || "cube-castle";
const DEFAULT_PRIVATE_KEY_PATH = path.resolve(
  __dirname,
  "../secrets/dev-jwt-private.pem",
);
const DEFAULT_PUBLIC_KEY_PATH = path.resolve(
  __dirname,
  "../secrets/dev-jwt-public.pem",
);
const JWT_KEY_ID = process.env.JWT_KEY_ID || "bff-key-1";

function resolveKeyPath(inputPath, fallbackPath) {
  if (inputPath && inputPath.trim()) {
    return path.isAbsolute(inputPath)
      ? inputPath
      : path.resolve(process.cwd(), inputPath.trim());
  }
  return fallbackPath;
}

const PRIVATE_KEY_PATH = resolveKeyPath(
  process.env.JWT_PRIVATE_KEY_PATH,
  DEFAULT_PRIVATE_KEY_PATH,
);
const PUBLIC_KEY_PATH = resolveKeyPath(
  process.env.JWT_PUBLIC_KEY_PATH,
  DEFAULT_PUBLIC_KEY_PATH,
);

let cachedPrivateKey = null;
let cachedPublicKey = null;

function readPrivateKey() {
  if (!cachedPrivateKey) {
    cachedPrivateKey = fs.readFileSync(PRIVATE_KEY_PATH, "utf8");
  }
  return cachedPrivateKey;
}

function readPublicKey() {
  if (!cachedPublicKey) {
    cachedPublicKey = fs.readFileSync(PUBLIC_KEY_PATH, "utf8");
  }
  return cachedPublicKey;
}

// 默认租户配置
const DEFAULT_TENANT_ID = "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9";
const DEFAULT_CLIENT_ID = "cube-castle-api-client";

/**
 * JWT Token验证中间件
 * 实现OAuth 2.0 Bearer Token认证
 */
const authenticateToken = async (req, res, next) => {
  try {
    // 1. 提取Bearer Token
    const authHeader = req.headers["authorization"];
    if (!authHeader || !authHeader.startsWith("Bearer ")) {
      return res.status(401).json({
        success: false,
        error: {
          code: "UNAUTHORIZED",
          message: "Missing or invalid Authorization header",
          details: "Expected format: Authorization: Bearer <token>",
        },
        timestamp: new Date().toISOString(),
        requestId: generateRequestId(),
      });
    }

    const token = authHeader.substring(7);

    // 2. 验证JWT (简化实现 - 生产环境需要完整JWT库)
    const payload = await verifyJWT(token);

    // 3. 权限解析和用户上下文设置
    req.auth = {
      clientId: payload.client_id || DEFAULT_CLIENT_ID,
      tenantId: payload.tenant_id || DEFAULT_TENANT_ID,
      userId: payload.user_id,
      permissions: payload.permissions || [],
      scopes: payload.scope ? payload.scope.split(" ") : [],
      issuedAt: payload.iat,
      expiresAt: payload.exp,
    };

    next();
  } catch (error) {
    if (error.message === "TOKEN_EXPIRED") {
      return res.status(401).json({
        success: false,
        error: {
          code: "TOKEN_EXPIRED",
          message: "Access token has expired",
          details: "Please obtain a new token from /oauth/token endpoint",
        },
        timestamp: new Date().toISOString(),
        requestId: generateRequestId(),
      });
    }

    return res.status(401).json({
      success: false,
      error: {
        code: "INVALID_TOKEN",
        message: "Invalid access token",
        details: error.message,
      },
      timestamp: new Date().toISOString(),
      requestId: generateRequestId(),
    });
  }
};

/**
 * 权限检查中间件工厂函数
 * @param {string} permission 需要的权限
 */
const requirePermission = (permission) => {
  return (req, res, next) => {
    if (!req.auth || !req.auth.permissions) {
      return res.status(401).json({
        success: false,
        error: {
          code: "MISSING_AUTH_CONTEXT",
          message: "Authentication context not found",
          details:
            "Ensure authenticateToken middleware runs before permission checks",
        },
        timestamp: new Date().toISOString(),
        requestId: generateRequestId(),
      });
    }

    if (!req.auth.permissions.includes(permission)) {
      return res.status(403).json({
        success: false,
        error: {
          code: "INSUFFICIENT_PERMISSIONS",
          message: "Insufficient permissions for this operation",
          details: {
            required_permission: permission,
            current_permissions: req.auth.permissions,
            client_id: req.auth.clientId,
          },
        },
        timestamp: new Date().toISOString(),
        requestId: generateRequestId(),
      });
    }

    next();
  };
};

/**
 * OAuth 2.0 Token端点 (简化实现)
 */
const tokenEndpoint = async (req, res) => {
  try {
    const { grant_type, client_id, client_secret } = req.body;

    // 验证grant_type
    if (grant_type !== "client_credentials") {
      return res.status(400).json({
        error: "unsupported_grant_type",
        error_description: "Only client_credentials grant type is supported",
      });
    }

    // 简化的客户端验证 (生产环境需要数据库验证)
    if (!client_id || !client_secret) {
      return res.status(400).json({
        error: "invalid_request",
        error_description: "Missing client_id or client_secret",
      });
    }

    // 生成访问令牌
    const payload = {
      client_id: client_id,
      tenant_id: DEFAULT_TENANT_ID,
      sub: "dev-user-id",
      roles: ["ADMIN", "HR_STAFF"], // 🔧 修复: 添加用户角色以支持PBAC权限检查
      permissions: [
        "org:read",
        "org:write",
        "org:delete",
        "org:suspend",
        "org:activate",
        "hr.organization.maintenance",
      ],
      scope: "org:read org:write org:delete",
      iat: Math.floor(Date.now() / 1000),
      exp: Math.floor(Date.now() / 1000) + 3600, // 1小时过期
      iss: ISSUER,
      aud: "cube-castle-api",
    };

    const accessToken = await generateJWT(payload);

    res.json({
      accessToken: accessToken,
      tokenType: "Bearer",
      expiresIn: 3600,
      scope: payload.scope,
    });
  } catch (error) {
    res.status(500).json({
      error: "server_error",
      error_description: "Internal server error during token generation",
    });
  }
};

// 辅助函数

/**
 * 简化的JWT生成 (生产环境需要使用专业JWT库)
 */
async function generateJWT(payload) {
  const header = {
    alg: "RS256",
    kid: JWT_KEY_ID,
    typ: "JWT",
  };

  const encodedHeader = Buffer.from(JSON.stringify(header)).toString(
    "base64url",
  );
  const encodedPayload = Buffer.from(JSON.stringify(payload)).toString(
    "base64url",
  );

  const signer = crypto.createSign("RSA-SHA256");
  signer.update(`${encodedHeader}.${encodedPayload}`);
  signer.end();

  const signature = signer.sign(readPrivateKey(), "base64url");

  return `${encodedHeader}.${encodedPayload}.${signature}`;
}

/**
 * 简化的JWT验证
 */
async function verifyJWT(token) {
  const parts = token.split(".");
  if (parts.length !== 3) {
    throw new Error("Invalid JWT format");
  }

  const [header, payload, signature] = parts;

  let decodedHeader;
  try {
    decodedHeader = JSON.parse(Buffer.from(header, "base64url").toString());
  } catch (error) {
    throw new Error("Invalid JWT header");
  }

  if (decodedHeader.alg !== "RS256") {
    throw new Error(
      `Unsupported signing algorithm: ${decodedHeader.alg || "unknown"}`,
    );
  }

  const verifier = crypto.createVerify("RSA-SHA256");
  verifier.update(`${header}.${payload}`);
  verifier.end();

  const isValid = verifier.verify(readPublicKey(), signature, "base64url");
  if (!isValid) {
    throw new Error("Invalid signature");
  }

  // 解码载荷
  let decodedPayload;
  try {
    decodedPayload = JSON.parse(Buffer.from(payload, "base64url").toString());
  } catch (error) {
    throw new Error("Invalid JWT payload");
  }

  // 检查过期时间
  if (
    decodedPayload.exp &&
    decodedPayload.exp < Math.floor(Date.now() / 1000)
  ) {
    throw new Error("TOKEN_EXPIRED");
  }

  return decodedPayload;
}

/**
 * 生成请求ID
 */
function generateRequestId() {
  return `req_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
}

module.exports = {
  authenticateToken,
  requirePermission,
  tokenEndpoint,
};
