/**
 * ESLint Rule: no-hardcoded-ports
 * 禁止硬编码端口号，强制使用统一配置管理
 * 
 * 检查项：
 * - 禁止在代码中直接使用端口号常量
 * - 禁止硬编码localhost:port格式的URL
 * - 强制使用SERVICE_PORTS配置
 * 
 * 符合统一配置原则：所有端口配置集中管理
 */

module.exports = {
  meta: {
    type: 'problem',
    docs: {
      description: '禁止硬编码端口号，强制使用统一配置管理',
      category: 'Architecture',
      recommended: true,
    },
    fixable: 'code',
    schema: [
      {
        type: 'object',
        properties: {
          allowedPorts: {
            type: 'array',
            items: { type: 'number' },
            default: [80, 443] // HTTP/HTTPS标准端口
          },
          configModule: {
            type: 'string',
            default: '@shared/config/ports'
          },
          allowedPatterns: {
            type: 'array',
            items: { type: 'string' },
            default: [
              'SERVICE_PORTS\\.',
              'CQRS_ENDPOINTS\\.',
              'TEST_ENDPOINTS\\.'
            ]
          }
        },
        additionalProperties: false
      }
    ],
    messages: {
      hardcodedPort: '硬编码端口号 {{port}}，请使用 {{configModule}} 中的配置',
      hardcodedUrl: '硬编码URL "{{url}}"，请使用统一端点配置',
      usePortConfig: '请导入并使用 {{configModule}} 中的端口配置',
      replaceWithConfig: '将 {{hardcoded}} 替换为 {{suggested}}'
    }
  },

  create(context) {
    const options = context.options[0] || {};
    const allowedPorts = options.allowedPorts || [80, 443];
    const configModule = options.configModule || '@shared/config/ports';
    const allowedPatterns = options.allowedPatterns || [
      'SERVICE_PORTS\\.',
      'CQRS_ENDPOINTS\\.',
      'TEST_ENDPOINTS\\.'
    ];

    // 编译允许的模式为正则表达式
    const allowedRegexes = allowedPatterns.map(pattern => new RegExp(pattern));

    // 检查是否使用了允许的配置模式
    function isUsingAllowedConfig(text) {
      return allowedRegexes.some(regex => regex.test(text));
    }

    // 检查端口号
    function checkPort(node, port) {
      const portNum = parseInt(port);
      
      // 跳过标准HTTP/HTTPS端口
      if (allowedPorts.includes(portNum)) {
        return;
      }

      // 检查上下文是否使用了配置
      const sourceCode = context.getSourceCode();
      const text = sourceCode.getText(node);
      
      if (isUsingAllowedConfig(text)) {
        return;
      }

      // 建议使用的配置名称
      const portConfigMap = {
        3000: 'SERVICE_PORTS.FRONTEND_DEV',
        3001: 'SERVICE_PORTS.FRONTEND_PREVIEW', 
        8090: 'SERVICE_PORTS.GRAPHQL_QUERY_SERVICE',
        9090: 'SERVICE_PORTS.REST_COMMAND_SERVICE',
        5432: 'SERVICE_PORTS.POSTGRESQL',
        6379: 'SERVICE_PORTS.REDIS'
      };

      const suggested = portConfigMap[portNum] || 'SERVICE_PORTS.APPROPRIATE_PORT';

      context.report({
        node,
        messageId: 'hardcodedPort',
        data: { 
          port: portNum, 
          configModule,
          suggested
        },
        fix: portConfigMap[portNum] ? function(fixer) {
          return fixer.replaceText(node, suggested);
        } : null
      });
    }

    // 检查URL字符串
    function checkUrl(node, url) {
      if (typeof url !== 'string') return;

      // 检查localhost:port模式
      const localhostMatch = url.match(/localhost:(\d+)/);
      if (localhostMatch) {
        const port = parseInt(localhostMatch[1]);
        if (!allowedPorts.includes(port)) {
          const endpointConfigMap = {
            3000: 'CQRS_ENDPOINTS.FRONTEND_URL',
            8090: 'CQRS_ENDPOINTS.GRAPHQL_ENDPOINT', 
            9090: 'CQRS_ENDPOINTS.COMMAND_API'
          };

          const suggested = endpointConfigMap[port] || 'CQRS_ENDPOINTS.APPROPRIATE_ENDPOINT';

          context.report({
            node,
            messageId: 'hardcodedUrl',
            data: { 
              url, 
              suggested 
            },
            fix: endpointConfigMap[port] ? function(fixer) {
              return fixer.replaceText(node, `\`\${${suggested}}\``);
            } : null
          });
        }
      }

      // 检查IP:port模式
      const ipMatch = url.match(/(http:\/\/|https:\/\/)?(\d+\.\d+\.\d+\.\d+):(\d+)/);
      if (ipMatch) {
        const port = parseInt(ipMatch[3]);
        if (!allowedPorts.includes(port)) {
          context.report({
            node,
            messageId: 'hardcodedUrl',
            data: { url }
          });
        }
      }
    }

    return {
      // 检查数字字面量
      Literal(node) {
        if (typeof node.value === 'number') {
          // 检查是否为端口号范围 (1024-65535)
          if (node.value >= 1024 && node.value <= 65535) {
            checkPort(node, node.value);
          }
        }

        if (typeof node.value === 'string') {
          checkUrl(node, node.value);
        }
      },

      // 检查模板字符串
      TemplateLiteral(node) {
        const sourceCode = context.getSourceCode();
        const templateText = sourceCode.getText(node);
        
        // 检查模板字符串中的端口模式
        const portMatches = templateText.match(/:\$\{(\d+)\}/g);
        if (portMatches) {
          portMatches.forEach(match => {
            const port = match.match(/\d+/)[0];
            checkPort(node, port);
          });
        }

        // 检查模板字符串中的localhost模式
        if (templateText.includes('localhost:') && !isUsingAllowedConfig(templateText)) {
          context.report({
            node,
            messageId: 'hardcodedUrl',
            data: { url: templateText }
          });
        }
      },

      // 检查变量声明
      VariableDeclarator(node) {
        if (node.id.name && /port|PORT/.test(node.id.name)) {
          if (node.init && node.init.type === 'Literal' && typeof node.init.value === 'number') {
            const port = node.init.value;
            if (port >= 1024 && port <= 65535 && !allowedPorts.includes(port)) {
              context.report({
                node: node.init,
                messageId: 'hardcodedPort',
                data: { 
                  port, 
                  configModule 
                }
              });
            }
          }
        }
      },

      // 检查对象属性
      Property(node) {
        if (node.key.name && /port|PORT/.test(node.key.name)) {
          if (node.value.type === 'Literal' && typeof node.value.value === 'number') {
            const port = node.value.value;
            if (port >= 1024 && port <= 65535 && !allowedPorts.includes(port)) {
              checkPort(node.value, port);
            }
          }
        }
      },

      // 检查程序结束时是否导入了配置
      'Program:exit'(node) {
        const sourceCode = context.getSourceCode();
        const text = sourceCode.getText(node);
        
        // 检查是否有硬编码端口但未导入配置
        const hasHardcodedPorts = /:\s*(\d{4,5})/g.test(text);
        const hasConfigImport = text.includes(configModule) || 
                               text.includes('SERVICE_PORTS') ||
                               text.includes('CQRS_ENDPOINTS');

        if (hasHardcodedPorts && !hasConfigImport) {
          context.report({
            node,
            messageId: 'usePortConfig',
            data: { configModule },
            fix(fixer) {
              // 在文件开头添加导入
              const firstNode = node.body[0];
              const importStatement = `import { SERVICE_PORTS, CQRS_ENDPOINTS } from '${configModule}';\n`;
              return fixer.insertTextBefore(firstNode, importStatement);
            }
          });
        }
      }
    };
  }
};