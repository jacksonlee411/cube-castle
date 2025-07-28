// src/pages/index.tsx - Simple homepage without antd
import React from 'react';
import Link from 'next/link';

const HomePage: React.FC = () => {
  return (
    <div style={{ 
      fontFamily: 'Arial, sans-serif', 
      maxWidth: '1200px', 
      margin: '0 auto', 
      padding: '20px',
      lineHeight: '1.6'
    }}>
      <header style={{ 
        textAlign: 'center', 
        marginBottom: '40px',
        borderBottom: '2px solid #1890ff',
        paddingBottom: '20px'
      }}>
        <h1 style={{ 
          color: '#1890ff', 
          fontSize: '2.5rem',
          margin: '0 0 10px 0'
        }}>
          Cube Castle
        </h1>
        <p style={{ 
          color: '#666', 
          fontSize: '1.2rem',
          margin: '0'
        }}>
          企业级HR管理平台
        </p>
      </header>

      <div style={{ 
        display: 'grid', 
        gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))', 
        gap: '20px',
        marginBottom: '40px'
      }}>
        <div style={{ 
          border: '1px solid #d9d9d9', 
          borderRadius: '8px', 
          padding: '20px',
          backgroundColor: '#fafafa',
          transition: 'box-shadow 0.3s ease'
        }}>
          <h3 style={{ color: '#1890ff', marginTop: '0' }}>👥 员工管理</h3>
          <p style={{ color: '#666', marginBottom: '15px' }}>
            全面的员工信息管理，包括个人档案、职位历史和绩效记录
          </p>
          <Link href="/employees" style={{ 
            color: '#1890ff', 
            textDecoration: 'none',
            fontWeight: 'bold'
          }}>
            进入员工管理 →
          </Link>
        </div>

        <div style={{ 
          border: '1px solid #d9d9d9', 
          borderRadius: '8px', 
          padding: '20px',
          backgroundColor: '#fafafa'
        }}>
          <h3 style={{ color: '#52c41a', marginTop: '0' }}>🏢 组织架构</h3>
          <p style={{ color: '#666', marginBottom: '15px' }}>
            可视化展示公司组织架构，支持部门筛选和数据同步
          </p>
          <Link href="/organization/chart" style={{ 
            color: '#52c41a', 
            textDecoration: 'none',
            fontWeight: 'bold'
          }}>
            查看组织架构 →
          </Link>
        </div>

        <div style={{ 
          border: '1px solid #d9d9d9', 
          borderRadius: '8px', 
          padding: '20px',
          backgroundColor: '#fafafa'
        }}>
          <h3 style={{ color: '#fa8c16', marginTop: '0' }}>📊 SAM仪表板</h3>
          <p style={{ color: '#666', marginBottom: '15px' }}>
            AI驱动的组织态势感知和决策支持系统
          </p>
          <Link href="/sam/dashboard" style={{ 
            color: '#fa8c16', 
            textDecoration: 'none',
            fontWeight: 'bold'
          }}>
            查看SAM仪表板 →
          </Link>
        </div>

        <div style={{ 
          border: '1px solid #d9d9d9', 
          borderRadius: '8px', 
          padding: '20px',
          backgroundColor: '#fafafa'
        }}>
          <h3 style={{ color: '#f5222d', marginTop: '0' }}>⚡ 工作流管理</h3>
          <p style={{ color: '#666', marginBottom: '15px' }}>
            Temporal.io驱动的业务流程自动化和审批管理
          </p>
          <Link href="/workflows/demo" style={{ 
            color: '#f5222d', 
            textDecoration: 'none',
            fontWeight: 'bold'
          }}>
            查看工作流演示 →
          </Link>
        </div>

        <div style={{ 
          border: '1px solid #d9d9d9', 
          borderRadius: '8px', 
          padding: '20px',
          backgroundColor: '#fafafa'
        }}>
          <h3 style={{ color: '#722ed1', marginTop: '0' }}>📝 Meta-Contract编辑器</h3>
          <p style={{ color: '#666', marginBottom: '15px' }}>
            智能化的元合约编辑器，支持YAML语法、实时编译和模板管理
          </p>
          <div style={{ marginBottom: '10px' }}>
            <Link href="/metacontract-editor/demo" style={{ 
              color: '#722ed1', 
              textDecoration: 'none',
              fontWeight: 'bold',
              marginRight: '15px'
            }}>
              开始体验 →
            </Link>
            <Link href="/metacontract-editor" style={{ 
              color: '#722ed1', 
              textDecoration: 'none',
              fontSize: '0.9rem',
              marginRight: '15px'
            }}>
              完整编辑器
            </Link>
            <Link href="/metacontract-editor/advanced" style={{ 
              color: '#722ed1', 
              textDecoration: 'none',
              fontSize: '0.9rem'
            }}>
              高级功能
            </Link>
          </div>
          <div style={{ 
            fontSize: '0.8rem', 
            color: '#999',
            display: 'flex',
            gap: '10px',
            flexWrap: 'wrap'
          }}>
            <span>✨ 语法高亮</span>
            <span>🔧 实时编译</span>
            <span>📋 模板库</span>
            <span>💾 项目管理</span>
          </div>
        </div>
      </div>

      <div style={{ 
        backgroundColor: '#f0f8ff', 
        border: '1px solid #1890ff', 
        borderRadius: '8px', 
        padding: '20px',
        textAlign: 'center'
      }}>
        <h3 style={{ color: '#1890ff', marginTop: '0' }}>系统状态</h3>
        <p style={{ color: '#666', margin: '10px 0' }}>
          ✅ 系统运行正常 | 🔧 开发服务器已启动 | 📦 所有服务可用
        </p>
        <div style={{ 
          display: 'flex', 
          justifyContent: 'center', 
          gap: '20px',
          marginTop: '15px',
          flexWrap: 'wrap'
        }}>
          <span style={{ 
            backgroundColor: '#52c41a', 
            color: 'white', 
            padding: '5px 10px', 
            borderRadius: '4px',
            fontSize: '0.9rem'
          }}>
            Next.js 14.2.30
          </span>
          <span style={{ 
            backgroundColor: '#1890ff', 
            color: 'white', 
            padding: '5px 10px', 
            borderRadius: '4px',
            fontSize: '0.9rem'
          }}>
            React 18.3.0
          </span>
          <span style={{ 
            backgroundColor: '#fa8c16', 
            color: 'white', 
            padding: '5px 10px', 
            borderRadius: '4px',
            fontSize: '0.9rem'
          }}>
            TypeScript
          </span>
        </div>
      </div>

      <footer style={{ 
        textAlign: 'center', 
        marginTop: '40px',
        padding: '20px 0',
        borderTop: '1px solid #d9d9d9',
        color: '#999',
        fontSize: '0.9rem'
      }}>
        <p>© 2025 Cube Castle - 现代化企业级HR管理平台</p>
        <p>基于城堡模型架构的SaaS解决方案</p>
      </footer>
    </div>
  );
};

export default HomePage;