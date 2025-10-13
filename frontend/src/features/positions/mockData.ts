import type { PositionLifecycleEvent } from './types'

export interface PositionMock {
  code: string
  title: string
  jobFamilyGroup: string
  jobFamily: string
  jobRole: string
  jobLevel: string
  organization: {
    code: string
    name: string
  }
  supervisor: {
    name: string
    code: string
  }
  headcountCapacity: number
  headcountInUse: number
  status: 'PLANNED' | 'ACTIVE' | 'FILLED' | 'VACANT' | 'INACTIVE'
  effectiveDate: string
  location: string
  shiftPattern?: string
  notes?: string
  lifecycle: PositionLifecycleEvent[]
}

export const mockPositions: PositionMock[] = [
  {
    code: 'P1000101',
    title: '物业保洁员',
    jobFamilyGroup: 'OPER',
    jobFamily: 'OPER-OPS',
    jobRole: 'OPER-OPS-CLEAN',
    jobLevel: 'S1',
    organization: {
      code: '2000010',
      name: '上海虹桥商务区物业项目',
    },
    supervisor: {
      name: '李雪',
      code: 'P2000008',
    },
    headcountCapacity: 8,
    headcountInUse: 6,
    status: 'FILLED',
    effectiveDate: '2024-01-01',
    location: '上海·虹桥商务区',
    shiftPattern: '早晚双班制',
    notes: '该岗位采用“一岗多人”模式，节假日阶段性增编。',
    lifecycle: [
      {
        id: 'evt-1001',
        type: 'FILL',
        label: '批量入职',
        operator: '赵伟',
        occurredAt: '2024-03-01',
        summary: '新增 4 名保洁员，编制占用达到 6/8。',
      },
      {
        id: 'evt-1002',
        type: 'VACATE',
        label: '兼职离岗',
        operator: '赵伟',
        occurredAt: '2024-05-18',
        summary: '两名兼职人员离岗，编制释放 1.0 FTE。',
      },
      {
        id: 'evt-1003',
        type: 'FILL',
        label: '增编',
        operator: '赵伟',
        occurredAt: '2024-08-10',
        summary: '根据业主扩容请求增加 1.0 FTE。',
      },
    ],
  },
  {
    code: 'P1000102',
    title: '保洁主管',
    jobFamilyGroup: 'OPER',
    jobFamily: 'OPER-OPS',
    jobRole: 'OPER-OPS-SUPV',
    jobLevel: 'M1',
    organization: {
      code: '2000010',
      name: '上海虹桥商务区物业项目',
    },
    supervisor: {
      name: '王晨',
      code: 'P2000001',
    },
    headcountCapacity: 1,
    headcountInUse: 1,
    status: 'FILLED',
    effectiveDate: '2023-10-01',
    location: '上海·虹桥商务区',
    notes: '负责 8 名保洁员的日常排班和质量检查。',
    lifecycle: [
      {
        id: 'evt-2001',
        type: 'FILL',
        label: '岗位填充',
        operator: '刘洋',
        occurredAt: '2023-10-01',
        summary: '由南京项目调入，完成交叉培训。',
      },
      {
        id: 'evt-2002',
        type: 'TRANSFER',
        label: '组织变更',
        operator: '刘洋',
        occurredAt: '2024-02-15',
        summary: '并入虹桥物业项目，汇报对象调整为项目经理。',
      },
    ],
  },
  {
    code: 'P3000501',
    title: '高级后端工程师',
    jobFamilyGroup: 'PROF',
    jobFamily: 'PROF-IT',
    jobRole: 'PROF-IT-BKND',
    jobLevel: 'P5',
    organization: {
      code: '1000001',
      name: '数字技术中心',
    },
    supervisor: {
      name: '陈静',
      code: 'P4000101',
    },
    headcountCapacity: 2,
    headcountInUse: 1,
    status: 'ACTIVE',
    effectiveDate: '2024-06-01',
    location: '上海·张江',
    notes: '承担组织命令服务的 CQRS 架构演进工作。',
    lifecycle: [
      {
        id: 'evt-3001',
        type: 'CREATE',
        label: '职位创建',
        operator: '李雷',
        occurredAt: '2024-04-20',
        summary: '根据人力预算审批创建新职位。',
      },
      {
        id: 'evt-3002',
        type: 'FILL',
        label: '首位入职',
        operator: '李雷',
        occurredAt: '2024-06-10',
        summary: '罗明加入团队，编制占用 0.5 FTE。',
      },
    ],
  },
  {
    code: 'P5000201',
    title: '总部行政专员',
    jobFamilyGroup: 'CORP',
    jobFamily: 'CORP-ADMIN',
    jobRole: 'CORP-ADMIN-OPS',
    jobLevel: 'S2',
    organization: {
      code: '3000001',
      name: '总部共享服务中心',
    },
    supervisor: {
      name: '何梅',
      code: 'P5000001',
    },
    headcountCapacity: 1,
    headcountInUse: 0,
    status: 'PLANNED',
    effectiveDate: '2025-01-01',
    location: '北京·金融街',
    notes: '新财年规划职位，等待预算批复。',
    lifecycle: [
      {
        id: 'evt-4001',
        type: 'CREATE',
        label: '规划职位',
        operator: '周伟',
        occurredAt: '2024-11-20',
        summary: '列入 2025 年行政支持编制计划。',
      },
    ],
  },
]
