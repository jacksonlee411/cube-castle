// tests/unit/pages/sam/dashboard.test.tsx
import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import SAMDashboard from '../../../../src/pages/sam/dashboard';
import '@testing-library/jest-dom';

// Mock Chart.js and react-chartjs-2
jest.mock('react-chartjs-2', () => ({
  Line: ({ data, options }: any) => (
    <div data-testid="line-chart">
      <div>{data.datasets[0].label}</div>
      <div>{JSON.stringify(options)}</div>
    </div>
  ),
  Bar: ({ data, options }: any) => (
    <div data-testid="bar-chart">
      <div>{data.datasets[0].label}</div>
      <div>{JSON.stringify(options)}</div>
    </div>
  ),
  Doughnut: ({ data, options }: any) => (
    <div data-testid="doughnut-chart">
      <div>{data.datasets[0].label}</div>
      <div>{JSON.stringify(options)}</div>
    </div>
  ),
}));

// Mock chart.js
jest.mock('chart.js', () => ({
  Chart: {
    register: jest.fn(),
  },
  CategoryScale: jest.fn(),
  LinearScale: jest.fn(),
  PointElement: jest.fn(),
  LineElement: jest.fn(),
  BarElement: jest.fn(),
  ArcElement: jest.fn(),
  Title: jest.fn(),
  Tooltip: jest.fn(),
  Legend: jest.fn(),
}));

// Mock antd message
const mockMessage = {
  success: jest.fn(),
  error: jest.fn(),
  warning: jest.fn(),
};

// Mock fetch
global.fetch = jest.fn();

const mockSAMData = {
  timestamp: '2025-01-27T10:30:00Z',
  alertLevel: 'LOW',
  organizationHealth: {
    overallScore: 85,
    turnoverRate: 12.5,
    engagementLevel: 78,
    productivityIndex: 88,
    spanOfControlHealth: 82,
    departmentHealth: [
      {
        department: '研发部',
        healthScore: 88,
        turnoverRate: 8.5,
        averageTenure: 2.3,
        managerEffectiveness: 85,
        teamCohesion: 90,
        workloadBalance: 82,
        lastAssessment: '2025-01-26T09:00:00Z'
      },
      {
        department: '产品部',
        healthScore: 82,
        turnoverRate: 15.2,
        averageTenure: 1.8,
        managerEffectiveness: 78,
        teamCohesion: 85,
        workloadBalance: 75,
        lastAssessment: '2025-01-26T09:00:00Z'
      },
      {
        department: '市场部',
        healthScore: 79,
        turnoverRate: 18.5,
        averageTenure: 1.5,
        managerEffectiveness: 75,
        teamCohesion: 80,
        workloadBalance: 78,
        lastAssessment: '2025-01-26T09:00:00Z'
      }
    ],
    trendAnalysis: {
      trend: 'IMPROVING',
      trendStrength: 0.65,
      keyDrivers: ['管理效能提升', '团队协作改善', '工作负载优化'],
      predictedHealth: 87,
      confidence: 0.78
    }
  },
  talentMetrics: {
    talentPipelineHealth: 85,
    successionReadiness: 72,
    learningDevelopmentROI: 145,
    internalMobilityRate: 18.5,
    skillGapAnalysis: [
      {
        skillArea: 'AI/ML技术',
        currentLevel: 65,
        requiredLevel: 85,
        gapSize: 20,
        priority: 'HIGH',
        affectedRoles: ['数据科学家', '算法工程师'],
        closureStrategy: '内部培训 + 外部招聘'
      },
      {
        skillArea: '云原生技术',
        currentLevel: 70,
        requiredLevel: 90,
        gapSize: 20,
        priority: 'HIGH',
        affectedRoles: ['DevOps工程师', '架构师'],
        closureStrategy: '认证培训 + 实践项目'
      }
    ],
    performanceDistribution: {
      highPerformers: 25,
      solidPerformers: 65,
      lowPerformers: 10,
      performanceGaps: ['技术技能', '沟通能力']
    }
  },
  riskAssessment: {
    overallRiskScore: 35,
    keyPersonRisks: [
      {
        employeeId: 'EMP001',
        employeeName: '张三',
        position: '技术架构师',
        department: '研发部',
        riskScore: 85,
        riskFactors: ['关键技能依赖', '薪资偏低', '工作量过大'],
        businessImpact: 'HIGH',
        mitigationSteps: ['薪资调整', '团队扩充', '知识传承'],
        lastAssessment: '2025-01-25T14:30:00Z'
      }
    ],
    complianceRisks: [
      {
        riskType: '劳动法合规',
        severity: 'MEDIUM',
        description: '部分员工超时工作',
        affectedAreas: ['研发部', '产品部'],
        complianceGaps: ['工时管理', '加班审批'],
        remediationPlan: '完善工时管理制度',
        deadline: '2025-02-15'
      }
    ],
    operationalRisks: [
      {
        riskCategory: '人才流失',
        description: '核心人员离职风险',
        probability: 0.3,
        impact: 0.8,
        riskScore: 75,
        affectedTeams: ['核心技术团队'],
        contingencyPlan: '加速人才储备计划'
      }
    ],
    talentFlightRisks: [
      {
        employeeId: 'EMP002',
        employeeName: '李四',
        flightRisk: 0.72,
        riskIndicators: ['薪资不满', '晋升机会少', '工作压力大'],
        retentionActions: ['职业发展规划', '薪资调整', '工作分配优化'],
        timeFrame: '3个月内'
      }
    ]
  },
  opportunities: {
    talentOptimization: [
      {
        opportunityType: '内部人才流动',
        description: '跨部门人才配置优化',
        affectedRoles: ['产品经理', '项目经理'],
        expectedBenefit: '提升15%工作效率',
        implementationSteps: ['需求分析', '人才匹配', '培训过渡']
      }
    ],
    processImprovements: [
      {
        processArea: '招聘流程',
        currentState: '平均招聘周期45天',
        proposedState: '平均招聘周期30天',
        efficiencyGain: '33%时间节省',
        implementationComplexity: 'MEDIUM'
      }
    ]
  },
  recommendations: [
    {
      id: 'REC001',
      type: 'TALENT_RETENTION',
      priority: 'HIGH',
      category: '人才保留',
      title: '核心人员保留计划',
      description: '针对高风险核心人员制定个性化保留策略',
      businessImpact: '减少核心人才流失50%',
      implementation: {
        timeline: '6个月',
        phases: [
          {
            phaseNumber: 1,
            phaseName: '风险评估',
            duration: '2周',
            activities: ['人员风险评估', '影响分析'],
            dependencies: [],
            deliverables: ['风险评估报告']
          }
        ]
      },
      confidence: 0.85
    }
  ]
};

describe('SAMDashboard', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    
    // Mock successful API response
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(mockSAMData)
    });
  });

  it('should render the SAM dashboard title and description', async () => {
    render(<SAMDashboard />);
    
    expect(screen.getByText('SAM 态势感知仪表板')).toBeInTheDocument();
    expect(screen.getByText('AI驱动的组织态势感知和决策支持系统')).toBeInTheDocument();
  });

  it('should display loading state initially', () => {
    render(<SAMDashboard />);
    
    // Should show loading indicator
    expect(screen.getByRole('img', { name: 'loading' })).toBeInTheDocument();
  });

  it('should load and display SAM analysis data', async () => {
    render(<SAMDashboard />);
    
    // Wait for data to load
    await waitFor(() => {
      expect(screen.getByText('组织健康度')).toBeInTheDocument();
      expect(screen.getByText('风险评估')).toBeInTheDocument();
      expect(screen.getByText('人才分析')).toBeInTheDocument();
    });
  });

  it('should display alert level indicator', async () => {
    render(<SAMDashboard />);
    
    await waitFor(() => {
      expect(screen.getByText('系统状态')).toBeInTheDocument();
      expect(screen.getByText('低风险')).toBeInTheDocument(); // LOW alert level
    });
  });

  it('should show organization health metrics', async () => {
    render(<SAMDashboard />);
    
    await waitFor(() => {
      expect(screen.getByText('85')).toBeInTheDocument(); // Overall score
      expect(screen.getByText('12.5%')).toBeInTheDocument(); // Turnover rate
      expect(screen.getByText('78')).toBeInTheDocument(); // Engagement level
      expect(screen.getByText('88')).toBeInTheDocument(); // Productivity index
    });
  });

  it('should display department health analysis', async () => {
    render(<SAMDashboard />);
    
    await waitFor(() => {
      expect(screen.getByText('部门健康分析')).toBeInTheDocument();
      expect(screen.getByText('研发部')).toBeInTheDocument();
      expect(screen.getByText('产品部')).toBeInTheDocument();
      expect(screen.getByText('市场部')).toBeInTheDocument();
    });

    // Check health scores
    expect(screen.getByText('88分')).toBeInTheDocument(); // 研发部
    expect(screen.getByText('82分')).toBeInTheDocument(); // 产品部
    expect(screen.getByText('79分')).toBeInTheDocument(); // 市场部
  });

  it('should show talent metrics and skill gap analysis', async () => {
    render(<SAMDashboard />);
    
    await waitFor(() => {
      expect(screen.getByText('人才管道健康度')).toBeInTheDocument();
      expect(screen.getByText('继任准备度')).toBeInTheDocument();
      expect(screen.getByText('内部流动率')).toBeInTheDocument();
    });

    // Check skill gaps
    expect(screen.getByText('技能缺口分析')).toBeInTheDocument();
    expect(screen.getByText('AI/ML技术')).toBeInTheDocument();
    expect(screen.getByText('云原生技术')).toBeInTheDocument();
  });

  it('should display risk assessment section', async () => {
    render(<SAMDashboard />);
    
    await waitFor(() => {
      expect(screen.getByText('关键人员风险')).toBeInTheDocument();
      expect(screen.getByText('合规风险')).toBeInTheDocument();
      expect(screen.getByText('运营风险')).toBeInTheDocument();
      expect(screen.getByText('人才流失风险')).toBeInTheDocument();
    });

    // Check key person risks
    expect(screen.getByText('张三')).toBeInTheDocument();
    expect(screen.getByText('技术架构师')).toBeInTheDocument();
    expect(screen.getByText('85分')).toBeInTheDocument(); // Risk score
  });

  it('should show strategic recommendations', async () => {
    render(<SAMDashboard />);
    
    await waitFor(() => {
      expect(screen.getByText('战略建议')).toBeInTheDocument();
      expect(screen.getByText('核心人员保留计划')).toBeInTheDocument();
      expect(screen.getByText('人才保留')).toBeInTheDocument();
    });

    // Check recommendation details
    expect(screen.getByText('减少核心人才流失50%')).toBeInTheDocument();
    expect(screen.getByText('85%')).toBeInTheDocument(); // Confidence level
  });

  it('should handle different alert levels correctly', async () => {
    // Test HIGH alert level
    const highAlertData = { ...mockSAMData, alertLevel: 'HIGH' };
    (global.fetch as jest.Mock).mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve(highAlertData)
    });

    render(<SAMDashboard />);
    
    await waitFor(() => {
      expect(screen.getByText('高风险')).toBeInTheDocument();
    });
  });

  it('should refresh data when clicking refresh button', async () => {
    render(<SAMDashboard />);
    
    await waitFor(() => {
      expect(screen.getByText('组织健康度')).toBeInTheDocument();
    });

    const refreshButton = screen.getByText('刷新数据');
    fireEvent.click(refreshButton);

    // Should make another API call
    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledTimes(2); // Initial load + refresh
    });
  });

  it('should generate new analysis when clicking generate button', async () => {
    render(<SAMDashboard />);
    
    await waitFor(() => {
      expect(screen.getByText('组织健康度')).toBeInTheDocument();
    });

    const generateButton = screen.getByText('生成新分析');
    fireEvent.click(generateButton);

    // Should show loading state
    expect(generateButton.closest('button')).toHaveClass('ant-btn-loading');

    // Should make API call for new analysis
    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('generate'),
        expect.any(Object)
      );
    });
  });

  it('should export analysis report', async () => {
    render(<SAMDashboard />);
    
    await waitFor(() => {
      expect(screen.getByText('组织健康度')).toBeInTheDocument();
    });

    const exportButton = screen.getByText('导出报告');
    fireEvent.click(exportButton);

    // Should call export API
    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('export'),
        expect.any(Object)
      );
    });
  });

  it('should handle API errors gracefully', async () => {
    // Mock API error
    (global.fetch as jest.Mock).mockRejectedValueOnce(new Error('Network error'));
    
    render(<SAMDashboard />);
    
    await waitFor(() => {
      expect(mockMessage.error).toHaveBeenCalledWith('获取SAM分析数据失败');
    });
  });

  it('should display trend analysis information', async () => {
    render(<SAMDashboard />);
    
    await waitFor(() => {
      expect(screen.getByText('趋势分析')).toBeInTheDocument();
      expect(screen.getByText('改善中')).toBeInTheDocument(); // IMPROVING trend
      expect(screen.getByText('预测健康度: 87分')).toBeInTheDocument();
    });

    // Check key drivers
    expect(screen.getByText('管理效能提升')).toBeInTheDocument();
    expect(screen.getByText('团队协作改善')).toBeInTheDocument();
    expect(screen.getByText('工作负载优化')).toBeInTheDocument();
  });

  it('should show performance distribution chart', async () => {
    render(<SAMDashboard />);
    
    await waitFor(() => {
      expect(screen.getByText('绩效分布')).toBeInTheDocument();
      expect(screen.getByText('25%')).toBeInTheDocument(); // High performers
      expect(screen.getByText('65%')).toBeInTheDocument(); // Solid performers
      expect(screen.getByText('10%')).toBeInTheDocument(); // Low performers
    });
  });

  it('should display opportunities section', async () => {
    render(<SAMDashboard />);
    
    await waitFor(() => {
      expect(screen.getByText('机会识别')).toBeInTheDocument();
      expect(screen.getByText('人才优化')).toBeInTheDocument();
      expect(screen.getByText('流程改进')).toBeInTheDocument();
    });

    // Check opportunity details
    expect(screen.getByText('内部人才流动')).toBeInTheDocument();
    expect(screen.getByText('跨部门人才配置优化')).toBeInTheDocument();
    expect(screen.getByText('招聘流程')).toBeInTheDocument();
    expect(screen.getByText('平均招聘周期45天')).toBeInTheDocument();
  });

  it('should handle real-time data updates', async () => {
    render(<SAMDashboard />);
    
    await waitFor(() => {
      expect(screen.getByText('组织健康度')).toBeInTheDocument();
    });

    // Check last update timestamp
    expect(screen.getByText(/最后更新:/)).toBeInTheDocument();
  });

  it('should display charts correctly', async () => {
    render(<SAMDashboard />);
    
    await waitFor(() => {
      expect(screen.getByTestId('line-chart')).toBeInTheDocument();
      expect(screen.getByTestId('bar-chart')).toBeInTheDocument();
      expect(screen.getByTestId('doughnut-chart')).toBeInTheDocument();
    });
  });

  it('should handle filter changes for department analysis', async () => {
    render(<SAMDashboard />);
    
    await waitFor(() => {
      expect(screen.getByText('组织健康度')).toBeInTheDocument();
    });

    // Find department filter dropdown
    const departmentFilter = screen.getByDisplayValue('全部部门');
    fireEvent.mouseDown(departmentFilter);
    
    await waitFor(() => {
      const researchOption = screen.getByText('研发部');
      fireEvent.click(researchOption);
    });

    // Should update analysis for selected department
    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('department=研发部'),
        expect.any(Object)
      );
    });
  });

  it('should display compliance risks with appropriate severity indicators', async () => {
    render(<SAMDashboard />);
    
    await waitFor(() => {
      expect(screen.getByText('劳动法合规')).toBeInTheDocument();
      expect(screen.getByText('部分员工超时工作')).toBeInTheDocument();
      expect(screen.getByText('中等')).toBeInTheDocument(); // MEDIUM severity
    });
  });

  it('should show talent flight risk indicators', async () => {
    render(<SAMDashboard />);
    
    await waitFor(() => {
      expect(screen.getByText('李四')).toBeInTheDocument();
      expect(screen.getByText('72%')).toBeInTheDocument(); // Flight risk percentage
      expect(screen.getByText('3个月内')).toBeInTheDocument(); // Time frame
    });

    // Check risk indicators
    expect(screen.getByText('薪资不满')).toBeInTheDocument();
    expect(screen.getByText('晋升机会少')).toBeInTheDocument();
    expect(screen.getByText('工作压力大')).toBeInTheDocument();
  });
});