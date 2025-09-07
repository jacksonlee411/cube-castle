/**
 * ESLint Rule: no-rest-queries
 * 禁止前端使用REST API进行查询操作，强制使用GraphQL
 * 
 * 检查项：
 * - 禁止使用fetch/axios等HTTP客户端的GET方法
 * - 禁止导入非命令操作的REST API端点
 * - 强制查询操作使用GraphQL客户端
 * 
 * 符合CQRS架构原则：查询用GraphQL，命令用REST
 */

module.exports = {
  meta: {
    type: 'problem',
    docs: {
      description: '禁止前端使用REST API进行查询操作，强制使用GraphQL',
      category: 'Architecture',
      recommended: true,
    },
    fixable: null,
    schema: [
      {
        type: 'object',
        properties: {
          allowedRestMethods: {
            type: 'array',
            items: { type: 'string' },
            default: ['POST', 'PUT', 'DELETE', 'PATCH']
          },
          allowedQueryEndpoints: {
            type: 'array', 
            items: { type: 'string' },
            default: ['/auth', '/health', '/metrics']
          },
          graphqlClient: {
            type: 'string',
            default: 'graphql-client'
          }
        },
        additionalProperties: false
      }
    ],
    messages: {
      noRestGet: '禁止使用REST GET请求进行查询操作，请使用GraphQL客户端',
      noRestQueryEndpoint: '禁止访问查询类REST端点 "{{endpoint}}"，请使用GraphQL',
      useGraphQLForQuery: '查询操作必须使用GraphQL客户端，而不是REST API',
      restOnlyForCommands: 'REST API仅用于命令操作 (POST/PUT/DELETE/PATCH)，查询请使用GraphQL'
    }
  },

  create(context) {
    const options = context.options[0] || {};
    const allowedRestMethods = options.allowedRestMethods || ['POST', 'PUT', 'DELETE', 'PATCH'];
    const allowedQueryEndpoints = options.allowedQueryEndpoints || ['/auth', '/health', '/metrics'];
    const graphqlClient = options.graphqlClient || 'graphql-client';

    // 检查HTTP方法调用
    function checkHttpMethodCall(node, method) {
      const upperMethod = method.toUpperCase();
      
      // GET方法总是禁止的
      if (upperMethod === 'GET') {
        context.report({
          node,
          messageId: 'noRestGet',
          data: { method }
        });
        return;
      }

      // 检查其他HTTP方法
      if (!allowedRestMethods.includes(upperMethod)) {
        context.report({
          node,
          messageId: 'restOnlyForCommands',
          data: { method }
        });
      }
    }

    // 检查URL端点是否为查询类型
    function checkQueryEndpoint(node, url) {
      if (typeof url !== 'string') return;

      // 跳过允许的端点
      if (allowedQueryEndpoints.some(allowed => url.includes(allowed))) {
        return;
      }

      // 检查典型的查询端点模式
      const queryPatterns = [
        /\/api\/v\d+\/organization-units\/[^\/]+$/, // GET单个组织
        /\/api\/v\d+\/organization-units$/, // GET组织列表
        /\/api\/v\d+\/.*\?.*/, // 带查询参数的请求
        /\/search/, // 搜索端点
        /\/list/, // 列表端点
        /\/get/, // 获取端点
      ];

      if (queryPatterns.some(pattern => pattern.test(url))) {
        context.report({
          node,
          messageId: 'noRestQueryEndpoint',
          data: { endpoint: url }
        });
      }
    }

    return {
      // 检查 fetch() 调用
      CallExpression(node) {
        if (node.callee.name === 'fetch' && node.arguments.length > 0) {
          const urlArg = node.arguments[0];
          const optionsArg = node.arguments[1];

          // 检查URL
          if (urlArg.type === 'Literal') {
            checkQueryEndpoint(node, urlArg.value);
          }

          // 检查HTTP方法
          if (optionsArg && optionsArg.type === 'ObjectExpression') {
            const methodProp = optionsArg.properties.find(
              prop => prop.key && prop.key.name === 'method'
            );
            
            if (methodProp && methodProp.value.type === 'Literal') {
              checkHttpMethodCall(node, methodProp.value.value);
            } else {
              // 没有指定method的fetch默认是GET
              context.report({
                node,
                messageId: 'noRestGet'
              });
            }
          } else {
            // 没有options的fetch默认是GET
            context.report({
              node,
              messageId: 'noRestGet'
            });
          }
        }

        // 检查 axios 调用
        if (node.callee.type === 'MemberExpression') {
          const object = node.callee.object;
          const property = node.callee.property;

          if (object.name === 'axios' && property.name) {
            const method = property.name;
            checkHttpMethodCall(node, method);

            // 检查URL参数
            if (node.arguments.length > 0 && node.arguments[0].type === 'Literal') {
              checkQueryEndpoint(node, node.arguments[0].value);
            }
          }
        }

        // 检查其他HTTP客户端库
        const httpClients = ['request', 'superagent', 'got'];
        if (httpClients.includes(node.callee.name) && node.arguments.length > 0) {
          if (node.arguments[0].type === 'Literal') {
            checkQueryEndpoint(node, node.arguments[0].value);
          }
        }
      },

      // 检查导入语句
      ImportDeclaration(node) {
        const source = node.source.value;
        
        // 检查导入的API客户端
        const queryApiPatterns = [
          /.*\/api\/.*query.*/,
          /.*\/services\/.*query.*/,
          /.*rest.*client.*query.*/
        ];

        if (queryApiPatterns.some(pattern => pattern.test(source))) {
          context.report({
            node,
            messageId: 'useGraphQLForQuery',
            data: { source }
          });
        }
      },

      // 检查变量声明中的查询操作
      VariableDeclarator(node) {
        if (node.init && node.init.type === 'CallExpression') {
          const callee = node.init.callee;
          
          // 检查类似 const data = getOrganizations() 的调用
          if (callee.name && /^(get|fetch|query|list|search)/.test(callee.name)) {
            // 如果不是GraphQL相关的调用，报告错误
            const isGraphQLCall = node.init.arguments.some(arg => 
              arg.type === 'Literal' && 
              (arg.value.includes('graphql') || arg.value.includes('query'))
            );

            if (!isGraphQLCall) {
              context.report({
                node,
                messageId: 'useGraphQLForQuery'
              });
            }
          }
        }
      }
    };
  }
};