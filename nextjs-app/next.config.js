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
    // 松散的 ES 模块外部化处理
    esmExternals: 'loose',
  },
  // 处理 ES 模块兼容性问题
  transpilePackages: [
    'antd',
    '@ant-design/icons',
    'rc-util',
    '@rc-component/util',
    'rc-pagination',
    'rc-picker',
    'rc-table',
    'rc-tree',
    'rc-select',
    'rc-upload',
    'rc-tooltip',
    'rc-dropdown',
    'rc-menu',
    'rc-input',
    'rc-textarea',
    'rc-checkbox',
    'rc-radio',
    'rc-switch',
    'rc-rate',
    'rc-slider',
    'rc-steps',
    'rc-tabs',
    'rc-collapse',
    'rc-calendar',
    'rc-date-picker',
    'rc-time-picker',
    'rc-cascader',
    'rc-tree-select',
    'rc-mentions',
    'rc-anchor',
    'rc-affix',
    'rc-back-top',
    'rc-drawer',
    'rc-notification',
    'rc-progress',
    'rc-spin',
    'rc-badge',
    'rc-avatar',
    'rc-card',
    'rc-divider',
    'rc-list',
    'rc-statistic',
    'rc-timeline',
    'rc-tag',
    'rc-alert',
    'rc-modal',
    'rc-popover',
    'rc-popconfirm',
    'rc-result',
    'rc-skeleton',
    'rc-breadcrumb',
    'rc-page-header'
  ],
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
  // Webpack 配置优化 - 针对版本降级后的优化
  webpack: (config, { dev, isServer }) => {
    // 优化模块解析
    config.resolve.fallback = {
      ...config.resolve.fallback,
      fs: false,
      path: false,
      os: false,
    };

    // 处理 ES 模块导入问题 - 完全强制使用 CommonJS
    config.resolve.alias = {
      ...config.resolve.alias,
      // 完全重定向所有ES模块路径到CommonJS版本
      'antd/es': 'antd/lib',
      '@ant-design/icons/es': '@ant-design/icons/lib',
      'rc-util/es': 'rc-util/lib',
      '@rc-component/util/es': '@rc-component/util/lib',
      // 针对hooks的特殊处理
      'rc-util/es/hooks': 'rc-util/lib/hooks',
      'rc-util/es/hooks/useMemo': 'rc-util/lib/hooks/useMemo',
      // 处理其他常见的ES模块路径
      'rc-util/es/Dom': 'rc-util/lib/Dom',
      'rc-util/es/warning': 'rc-util/lib/warning',
      'rc-util/es/pickAttrs': 'rc-util/lib/pickAttrs',
    };

    // 强制模块解析顺序，优先使用CommonJS
    config.resolve.mainFields = ['main', 'module'];
    
    // 添加模块扩展名支持
    config.resolve.extensions = ['.js', '.jsx', '.ts', '.tsx', '.json', '.mjs'];

    // 模块替换规则
    config.module.rules.push({
      test: /\.m?js$/,
      resolve: {
        fullySpecified: false,
      },
    });

    return config;
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