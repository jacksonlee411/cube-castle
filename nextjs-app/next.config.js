/** @type {import('next').NextConfig} */
const nextConfig = {
  typescript: {
    // 在生产构建时进行类型检查
    tsconfigPath: './tsconfig.json',
  },
  eslint: {
    // 在生产构建时进行 ESLint 检查
    dirs: ['src'],
  },
  experimental: {
    // 启用实验性功能
    typedRoutes: true,
  },
  images: {
    // 图片优化配置
    domains: ['localhost'],
    formats: ['image/webp', 'image/avif'],
  },
  // API 路由配置
  async rewrites() {
    return [
      {
        source: '/api/v1/:path*',
        destination: 'http://localhost:8080/api/v1/:path*',
      },
    ]
  },
  // 环境变量配置
  env: {
    CUBE_CASTLE_API_URL: process.env.CUBE_CASTLE_API_URL || 'http://localhost:8080',
    CUBE_CASTLE_WS_URL: process.env.CUBE_CASTLE_WS_URL || 'ws://localhost:8080',
  },
  // 输出配置
  output: 'standalone',
  // 压缩配置
  compress: true,
  // 性能配置
  poweredByHeader: false,
  reactStrictMode: true,
  swcMinify: true,
}

module.exports = nextConfig