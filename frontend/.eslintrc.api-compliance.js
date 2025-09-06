/**
 * ESLint自定义规则：API合规性检查 (ADR-008实施)
 * 防止在TypeScript代码中使用弃用的API模式
 */

module.exports = {
  rules: {
    // 禁用弃用的API方法调用
    'no-deprecated-org-api': {
      create(context) {
        return {
          // 检查方法调用
          CallExpression(node) {
            if (node.callee.property) {
              const methodName = node.callee.property.name;
              
              if (methodName === 'reactivate') {
                context.report({
                  node,
                  message: 'ADR-008: 使用 .activate() 替代 .reactivate() 方法',
                  fix: function(fixer) {
                    return fixer.replaceText(node.callee.property, 'activate');
                  }
                });
              }
            }
          },
          
          // 检查对象属性定义
          Property(node) {
            if (node.key.name === 'reactivate') {
              context.report({
                node,
                message: 'ADR-008: 移除 reactivate 属性定义，使用 activate',
                fix: function(fixer) {
                  return fixer.replaceText(node.key, 'activate');
                }
              });
            }
          },
          
          // 检查字符串字面量中的端点路径
          Literal(node) {
            if (typeof node.value === 'string') {
              if (node.value.includes('/reactivate')) {
                context.report({
                  node,
                  message: 'ADR-008: API端点路径使用 /activate 替代 /reactivate',
                  fix: function(fixer) {
                    const newValue = node.value.replace('/reactivate', '/activate');
                    return fixer.replaceText(node, `"${newValue}"`);
                  }
                });
              }
              
              if (node.value.includes('org:reactivate')) {
                context.report({
                  node,
                  message: 'ADR-008: 权限scope使用 org:activate 替代 org:reactivate',
                  fix: function(fixer) {
                    const newValue = node.value.replace('org:reactivate', 'org:activate');
                    return fixer.replaceText(node, `"${newValue}"`);
                  }
                });
              }
            }
          },
          
          // 检查模板字符串
          TemplateLiteral(node) {
            node.quasis.forEach(quasi => {
              if (quasi.value.raw.includes('/reactivate')) {
                context.report({
                  node: quasi,
                  message: 'ADR-008: 模板字符串中使用 /activate 替代 /reactivate'
                });
              }
            });
          }
        };
      }
    },

    // 禁止PATCH修改status字段
    'no-patch-status-modification': {
      create(context) {
        return {
          CallExpression(node) {
            // 检查fetch或axios的PATCH请求
            if (node.callee.name === 'fetch' || 
                (node.callee.object && node.callee.object.name === 'axios') ||
                (node.callee.property && node.callee.property.name === 'patch')) {
              
              // 检查请求体是否包含status字段
              node.arguments.forEach(arg => {
                if (arg.type === 'ObjectExpression') {
                  arg.properties.forEach(prop => {
                    if (prop.key && prop.key.name === 'body') {
                      // 这里需要更复杂的AST分析来检查JSON.stringify内容
                      context.report({
                        node,
                        message: 'ADR-008: 禁止通过PATCH修改status字段，使用POST /activate或/suspend',
                        data: {
                          suggestion: '使用 organizationAPI.activate() 或 organizationAPI.suspend() 方法'
                        }
                      });
                    }
                  });
                }
              });
            }
          }
        };
      }
    },

    // 强制使用标准错误处理
    'use-standard-error-handling': {
      create(context) {
        return {
          CallExpression(node) {
            // 检查是否使用了自定义错误处理而不是标准工具
            if (node.callee.property && node.callee.property.name === 'catch') {
              const sourceCode = context.getSourceCode();
              const catchHandler = node.arguments[0];
              
              if (catchHandler && catchHandler.type === 'ArrowFunctionExpression') {
                const bodyText = sourceCode.getText(catchHandler.body);
                
                // 检查是否使用了标准错误处理工具
                if (!bodyText.includes('formatErrorForUser') && 
                    !bodyText.includes('getErrorMessage')) {
                  context.report({
                    node,
                    message: 'ADR-008: 建议使用标准错误处理工具 formatErrorForUser() 或 getErrorMessage()',
                    data: {
                      suggestion: 'import { formatErrorForUser } from "@/shared/api/error-messages"'
                    }
                  });
                }
              }
            }
          }
        };
      }
    }
  }
};