// OAuth 2.0认证服务 - 企业级认证端点
const express = require('express');
const cors = require('cors');
const { tokenEndpoint } = require('../../middleware/auth');

const app = express();
const PORT = process.env.OAUTH_PORT || 8080;

// 中间件配置
app.use(express.json());
app.use(express.urlencoded({ extended: true })); // 🔧 修复: 添加URL编码解析支持
app.use(cors({
    origin: ['http://localhost:3000', 'http://localhost:3001'],
    credentials: true
}));

// 健康检查
app.get('/health', (req, res) => {
    res.json({
        status: 'healthy',
        service: 'cube-castle-oauth-service',
        version: 'v1.0.0',
        timestamp: new Date().toISOString()
    });
});

// OAuth 2.0 Token端点
app.post('/oauth/token', tokenEndpoint);

// OAuth 2.0服务信息
app.get('/oauth/.well-known', (req, res) => {
    res.json({
        issuer: process.env.JWT_ISSUER || 'cube-castle',
        token_endpoint: `http://localhost:${PORT}/oauth/token`,
        supported_grant_types: ['client_credentials'],
        supported_token_endpoint_auth_methods: ['client_secret_post'],
        supported_scopes: ['org:read', 'org:write', 'org:delete', 'hr.organization.maintenance']
    });
});

// 启动服务
app.listen(PORT, () => {
    console.log(`🔐 OAuth 2.0认证服务启动在端口 ${PORT}`);
    console.log(`🔗 Token端点: http://localhost:${PORT}/oauth/token`);
    console.log(`📋 服务发现: http://localhost:${PORT}/oauth/.well-known`);
    console.log(`🏥 健康检查: http://localhost:${PORT}/health`);
});

// 优雅关闭
process.on('SIGTERM', () => {
    console.log('🛑 OAuth服务收到SIGTERM，正在关闭...');
    process.exit(0);
});

process.on('SIGINT', () => {
    console.log('🛑 OAuth服务收到SIGINT，正在关闭...');
    process.exit(0);
});