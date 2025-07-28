// src/pages/sam/dashboard.tsx - SAM态势感知仪表板
import React, { useState, useEffect } from 'react';
import dayjs from 'dayjs';

const SAMDashboardPage: React.FC = () => {
  const [loading, setLoading] = useState(true);
  const [refreshTime, setRefreshTime] = useState<string>(dayjs().format('YYYY-MM-DD HH:mm:ss'));

  // 模拟数据加载
  useEffect(() => {
    const timer = setTimeout(() => {
      setLoading(false);
    }, 1000);
    return () => clearTimeout(timer);
  }, []);

  // 模拟SAM数据
  const mockData = {
    timestamp: dayjs().format('YYYY-MM-DD HH:mm:ss'),
    alertLevel: 'LOW',
    organizationHealth: {
      overallScore: 85,
      turnoverRate: 12.5,
      engagementLevel: 78,
      productivityIndex: 88,
      departmentHealth: [
        { department: '研发部', healthScore: 88, turnoverRate: 8.5, managerEffectiveness: 85 },
        { department: '产品部', healthScore: 82, turnoverRate: 15.2, managerEffectiveness: 78 },
        { department: '市场部', healthScore: 79, turnoverRate: 18.5, managerEffectiveness: 75 },
      ]
    },
    talentMetrics: {
      talentPipelineHealth: 85,
      successionReadiness: 72,
      skillGaps: [
        { skillArea: 'AI/ML技术', gapSize: 20 },
        { skillArea: '云原生技术', gapSize: 15 },
      ]
    },
    riskAssessment: {
      overallRiskScore: 35,
      keyRisks: [
        { 
          riskType: '核心人员流失', 
          employeeName: '张三', 
          riskScore: 85, 
          riskFactors: ['薪资偏低', '工作量过大', '缺乏晋升机会'] 
        }
      ]
    },
    recommendations: [
      {
        id: 'REC001',
        priority: 'HIGH',
        category: '人才保留',
        title: '核心人员保留计划',
        description: '针对高风险核心人员制定个性化保留策略',
        confidence: 85
      }
    ]
  };

  if (loading) {
    return (
      <div style={{ 
        display: 'flex', 
        justifyContent: 'center', 
        alignItems: 'center', 
        height: '400px',
        fontFamily: 'Arial, sans-serif'
      }}>
        <div style={{ textAlign: 'center' }}>
          <div style={{ 
            width: '40px', 
            height: '40px', 
            border: '4px solid #f3f3f3',
            borderTop: '4px solid #1890ff',
            borderRadius: '50%',
            animation: 'spin 2s linear infinite',
            margin: '0 auto 20px'
          }}></div>
          <p>正在加载SAM分析数据...</p>
          <style jsx>{`
            @keyframes spin {
              0% { transform: rotate(0deg); }
              100% { transform: rotate(360deg); }
            }
          `}</style>
        </div>
      </div>
    );
  }

  const getAlertLevelColor = (level: string) => {
    switch (level) {
      case 'HIGH': return '#f5222d';
      case 'MEDIUM': return '#fa8c16';
      case 'LOW': return '#52c41a';
      default: return '#d9d9d9';
    }
  };

  const getAlertLevelText = (level: string) => {
    switch (level) {
      case 'HIGH': return '高风险';
      case 'MEDIUM': return '中等风险';
      case 'LOW': return '低风险';
      default: return '未知';
    }
  };

  return (
    <div style={{ 
      fontFamily: 'Arial, sans-serif', 
      maxWidth: '1400px', 
      margin: '0 auto', 
      padding: '20px',
      backgroundColor: '#f5f5f5',
      minHeight: '100vh'
    }}>
      <header style={{ 
        textAlign: 'center', 
        marginBottom: '30px',
        backgroundColor: 'white',
        padding: '20px',
        borderRadius: '8px',
        boxShadow: '0 2px 8px rgba(0,0,0,0.1)'
      }}>
        <h1 style={{ 
          color: '#1890ff', 
          fontSize: '2.2rem',
          margin: '0 0 10px 0'
        }}>
          SAM 态势感知仪表板
        </h1>
        <p style={{ 
          color: '#666', 
          fontSize: '1.1rem',
          margin: '0 0 15px 0'
        }}>
          AI驱动的组织态势感知和决策支持系统
        </p>
        <div style={{ 
          display: 'flex', 
          justifyContent: 'center', 
          alignItems: 'center',
          gap: '20px',
          flexWrap: 'wrap'
        }}>
          <span style={{ 
            backgroundColor: getAlertLevelColor(mockData.alertLevel),
            color: 'white',
            padding: '8px 16px',
            borderRadius: '20px',
            fontWeight: 'bold'
          }}>
            系统状态: {getAlertLevelText(mockData.alertLevel)}
          </span>
          <span style={{ color: '#666', fontSize: '0.9rem' }}>
            最后更新: {refreshTime}
          </span>
        </div>
      </header>

      {/* 组织健康度概览 */}
      <div style={{ 
        backgroundColor: 'white', 
        padding: '20px', 
        borderRadius: '8px', 
        marginBottom: '20px',
        boxShadow: '0 2px 8px rgba(0,0,0,0.1)'
      }}>
        <h2 style={{ color: '#1890ff', borderBottom: '2px solid #1890ff', paddingBottom: '10px' }}>
          📊 组织健康度
        </h2>
        <div style={{ 
          display: 'grid', 
          gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', 
          gap: '20px',
          marginTop: '20px'
        }}>
          <div style={{ textAlign: 'center', padding: '15px', backgroundColor: '#f0f8ff', borderRadius: '8px' }}>
            <h3 style={{ color: '#1890ff', fontSize: '2rem', margin: '0' }}>
              {mockData.organizationHealth.overallScore}
            </h3>
            <p style={{ color: '#666', margin: '5px 0 0 0' }}>总体健康度</p>
          </div>
          <div style={{ textAlign: 'center', padding: '15px', backgroundColor: '#f6ffed', borderRadius: '8px' }}>
            <h3 style={{ color: '#52c41a', fontSize: '2rem', margin: '0' }}>
              {mockData.organizationHealth.turnoverRate}%
            </h3>
            <p style={{ color: '#666', margin: '5px 0 0 0' }}>离职率</p>
          </div>
          <div style={{ textAlign: 'center', padding: '15px', backgroundColor: '#fff7e6', borderRadius: '8px' }}>
            <h3 style={{ color: '#fa8c16', fontSize: '2rem', margin: '0' }}>
              {mockData.organizationHealth.engagementLevel}
            </h3>
            <p style={{ color: '#666', margin: '5px 0 0 0' }}>员工参与度</p>
          </div>
          <div style={{ textAlign: 'center', padding: '15px', backgroundColor: '#f0f5ff', borderRadius: '8px' }}>
            <h3 style={{ color: '#722ed1', fontSize: '2rem', margin: '0' }}>
              {mockData.organizationHealth.productivityIndex}
            </h3>
            <p style={{ color: '#666', margin: '5px 0 0 0' }}>生产力指数</p>
          </div>
        </div>
      </div>

      {/* 部门健康分析 */}
      <div style={{ 
        backgroundColor: 'white', 
        padding: '20px', 
        borderRadius: '8px', 
        marginBottom: '20px',
        boxShadow: '0 2px 8px rgba(0,0,0,0.1)'
      }}>
        <h2 style={{ color: '#52c41a', borderBottom: '2px solid #52c41a', paddingBottom: '10px' }}>
          🏢 部门健康分析
        </h2>
        <div style={{ marginTop: '20px' }}>
          {mockData.organizationHealth.departmentHealth.map((dept, index) => (
            <div key={index} style={{ 
              marginBottom: '15px', 
              padding: '15px', 
              backgroundColor: '#fafafa', 
              borderRadius: '8px',
              border: '1px solid #d9d9d9'
            }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '10px' }}>
                <h3 style={{ color: '#1890ff', margin: '0' }}>{dept.department}</h3>
                <span style={{ 
                  backgroundColor: dept.healthScore >= 85 ? '#52c41a' : dept.healthScore >= 75 ? '#fa8c16' : '#f5222d',
                  color: 'white',
                  padding: '4px 12px',
                  borderRadius: '12px',
                  fontSize: '0.9rem'
                }}>
                  {dept.healthScore}分
                </span>
              </div>
              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(150px, 1fr))', gap: '10px' }}>
                <div>
                  <span style={{ color: '#666', fontSize: '0.9rem' }}>离职率: </span>
                  <strong style={{ color: dept.turnoverRate <= 10 ? '#52c41a' : '#fa8c16' }}>
                    {dept.turnoverRate}%
                  </strong>
                </div>
                <div>
                  <span style={{ color: '#666', fontSize: '0.9rem' }}>管理效能: </span>
                  <strong style={{ color: '#1890ff' }}>{dept.managerEffectiveness}分</strong>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* 人才分析和风险评估 */}
      <div style={{ 
        display: 'grid', 
        gridTemplateColumns: 'repeat(auto-fit, minmax(400px, 1fr))', 
        gap: '20px', 
        marginBottom: '20px' 
      }}>
        <div style={{ 
          backgroundColor: 'white', 
          padding: '20px', 
          borderRadius: '8px',
          boxShadow: '0 2px 8px rgba(0,0,0,0.1)'
        }}>
          <h2 style={{ color: '#fa8c16', borderBottom: '2px solid #fa8c16', paddingBottom: '10px' }}>
            👥 人才分析
          </h2>
          <div style={{ marginTop: '20px' }}>
            <div style={{ marginBottom: '15px' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '5px' }}>
                <span>人才管道健康度</span>
                <strong style={{ color: '#52c41a' }}>{mockData.talentMetrics.talentPipelineHealth}%</strong>
              </div>
              <div style={{ backgroundColor: '#f5f5f5', height: '8px', borderRadius: '4px' }}>
                <div style={{ 
                  backgroundColor: '#52c41a', 
                  height: '100%', 
                  width: `${mockData.talentMetrics.talentPipelineHealth}%`, 
                  borderRadius: '4px'
                }}></div>
              </div>
            </div>
            <div style={{ marginBottom: '15px' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '5px' }}>
                <span>继任准备度</span>
                <strong style={{ color: '#fa8c16' }}>{mockData.talentMetrics.successionReadiness}%</strong>
              </div>
              <div style={{ backgroundColor: '#f5f5f5', height: '8px', borderRadius: '4px' }}>
                <div style={{ 
                  backgroundColor: '#fa8c16', 
                  height: '100%', 
                  width: `${mockData.talentMetrics.successionReadiness}%`, 
                  borderRadius: '4px'
                }}></div>
              </div>
            </div>
            <h4 style={{ color: '#1890ff', marginTop: '20px' }}>技能缺口分析</h4>
            {mockData.talentMetrics.skillGaps.map((gap, index) => (
              <div key={index} style={{ marginBottom: '10px' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <span>{gap.skillArea}</span>
                  <span style={{ color: '#f5222d', fontWeight: 'bold' }}>缺口: {gap.gapSize}分</span>
                </div>
              </div>
            ))}
          </div>
        </div>

        <div style={{ 
          backgroundColor: 'white', 
          padding: '20px', 
          borderRadius: '8px',
          boxShadow: '0 2px 8px rgba(0,0,0,0.1)'
        }}>
          <h2 style={{ color: '#f5222d', borderBottom: '2px solid #f5222d', paddingBottom: '10px' }}>
            ⚠️ 风险评估
          </h2>
          <div style={{ marginTop: '20px' }}>
            <div style={{ textAlign: 'center', marginBottom: '20px' }}>
              <div style={{ 
                fontSize: '3rem', 
                fontWeight: 'bold',
                color: mockData.riskAssessment.overallRiskScore <= 40 ? '#52c41a' : 
                      mockData.riskAssessment.overallRiskScore <= 70 ? '#fa8c16' : '#f5222d'
              }}>
                {mockData.riskAssessment.overallRiskScore}
              </div>
              <p style={{ color: '#666', margin: '0' }}>总体风险评分</p>
            </div>
            
            <h4 style={{ color: '#f5222d' }}>关键风险</h4>
            {mockData.riskAssessment.keyRisks.map((risk, index) => (
              <div key={index} style={{ 
                marginBottom: '15px', 
                padding: '12px', 
                backgroundColor: '#fff1f0', 
                borderLeft: '4px solid #f5222d',
                borderRadius: '4px'
              }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '8px' }}>
                  <strong>{risk.riskType}</strong>
                  <span style={{ 
                    backgroundColor: '#f5222d', 
                    color: 'white', 
                    padding: '2px 8px', 
                    borderRadius: '10px',
                    fontSize: '0.8rem'
                  }}>
                    风险: {risk.riskScore}%
                  </span>
                </div>
                <p style={{ margin: '0', color: '#666' }}>涉及员工: {risk.employeeName}</p>
                <div style={{ marginTop: '8px' }}>
                  {risk.riskFactors.map((factor, idx) => (
                    <span key={idx} style={{ 
                      display: 'inline-block',
                      backgroundColor: '#fff',
                      border: '1px solid #ffa39e',
                      color: '#f5222d',
                      padding: '2px 6px',
                      borderRadius: '10px',
                      fontSize: '0.8rem',
                      marginRight: '5px',
                      marginBottom: '5px'
                    }}>
                      {factor}
                    </span>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* 战略建议 */}
      <div style={{ 
        backgroundColor: 'white', 
        padding: '20px', 
        borderRadius: '8px', 
        marginBottom: '20px',
        boxShadow: '0 2px 8px rgba(0,0,0,0.1)'
      }}>
        <h2 style={{ color: '#722ed1', borderBottom: '2px solid #722ed1', paddingBottom: '10px' }}>
          💡 战略建议
        </h2>
        <div style={{ marginTop: '20px' }}>
          {mockData.recommendations.map((rec, index) => (
            <div key={index} style={{ 
              marginBottom: '20px', 
              padding: '20px', 
              backgroundColor: '#f9f0ff', 
              borderRadius: '8px',
              border: '1px solid #d3adf7'
            }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: '10px' }}>
                <div>
                  <h3 style={{ color: '#722ed1', margin: '0 0 5px 0' }}>{rec.title}</h3>
                  <span style={{ 
                    backgroundColor: rec.priority === 'HIGH' ? '#f5222d' : '#fa8c16',
                    color: 'white',
                    padding: '4px 8px',
                    borderRadius: '4px',
                    fontSize: '0.8rem'
                  }}>
                    {rec.priority} 优先级
                  </span>
                  <span style={{ 
                    backgroundColor: '#722ed1',
                    color: 'white',
                    padding: '4px 8px',
                    borderRadius: '4px',
                    fontSize: '0.8rem',
                    marginLeft: '8px'
                  }}>
                    {rec.category}
                  </span>
                </div>
                <div style={{ textAlign: 'right' }}>
                  <div style={{ color: '#52c41a', fontWeight: 'bold', fontSize: '1.2rem' }}>
                    {rec.confidence}%
                  </div>
                  <div style={{ color: '#666', fontSize: '0.8rem' }}>置信度</div>
                </div>
              </div>
              <p style={{ color: '#595959', margin: '10px 0 0 0', lineHeight: '1.6' }}>
                {rec.description}
              </p>
            </div>
          ))}
        </div>
      </div>

      {/* 操作按钮 */}
      <div style={{ 
        textAlign: 'center', 
        padding: '20px',
        backgroundColor: 'white',
        borderRadius: '8px',
        boxShadow: '0 2px 8px rgba(0,0,0,0.1)'
      }}>
        <button 
          onClick={() => {
            setRefreshTime(dayjs().format('YYYY-MM-DD HH:mm:ss'));
            setLoading(true);
            setTimeout(() => setLoading(false), 1000);
          }}
          style={{ 
            backgroundColor: '#1890ff', 
            color: 'white', 
            border: 'none', 
            padding: '12px 24px', 
            borderRadius: '6px', 
            fontSize: '1rem',
            cursor: 'pointer',
            marginRight: '10px'
          }}
        >
          🔄 刷新数据
        </button>
        <button 
          style={{ 
            backgroundColor: '#52c41a', 
            color: 'white', 
            border: 'none', 
            padding: '12px 24px', 
            borderRadius: '6px', 
            fontSize: '1rem',
            cursor: 'pointer',
            marginRight: '10px'
          }}
        >
          📊 生成报告
        </button>
        <button 
          style={{ 
            backgroundColor: '#fa8c16', 
            color: 'white', 
            border: 'none', 
            padding: '12px 24px', 
            borderRadius: '6px', 
            fontSize: '1rem',
            cursor: 'pointer'
          }}
        >
          📤 导出数据
        </button>
      </div>
    </div>
  );
};

export default SAMDashboardPage;