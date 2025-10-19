# PeopleSoft Core HR 功能菜单参考

**文档编号**: 79
**创建日期**: 2025-10-13
**当前版本**: 3.1
**最后更新**: 2025-10-19
**文档类型**: 参考资料

## 概述

本文档记录 PeopleSoft Core HR（核心人力资源）的完整功能菜单结构，涵盖 22 个功能模块，包括：

- **基础运营模块**（1-10）：组织、人员、职位、人事、工作信息、薪酬、福利、时间考勤、自助服务、报表分析
- **人才管理模块**（11-14）：招聘、绩效、培训、人才管理
- **薪资与合同合规模块**（15-16, 22）：薪资计算、劳动合同管理、合规管理
- **专项管理模块**（17-20）：缺勤、员工关系、劳动力规划、健康安全
- **横向支撑功能**（21）：工作流、审批、通知、集成、系统管理

同时提供模块依赖关系图、典型业务流程、用户角色视图等实用内容，作为企业级 Core HR 系统功能设计与需求分析的完整参考基准。

---

## 1. 组织管理 (Organization Management)

### 主要功能
- 组织架构设置
- 部门/业务单元管理
- 职位管理
- 汇报关系

### 说明
负责整个企业的组织结构定义，包括层级关系、部门划分、职位设置等基础组织框架。

---

## 2. 人员管理 (Workforce Administration)

### 主要功能
- **员工主数据** (参考中国大陆主流 HCM 软件字段体系)

#### 模块划分优化建议（参考主流 HCM）
- **2.1 Worker Core Profile 管理（人员基础档案）**：聚焦身份标识、组织归属、岗位信息等核心字段，支撑跨模块引用。
- **2.2 Personal & Family Data Maintenance（个人及家庭信息维护）**：涵盖联系方式、紧急联系人、家庭成员与受益人等高频变更数据。
- **2.3 Employment Lifecycle Events（雇佣生命周期事件）**：管理入职、转岗、晋升、离职等人事事件，保障审批与审计闭环。
- **2.4 Workforce Status Governance（员工状态治理）**：维护用工状态、兼岗、停薪留职等状态型数据，并联动考勤与薪资。
- **独立模块：Employment Contract Management（劳动合同管理）**：承接合同模板、签署、续签、派遣与外包信息，移出人员管理模块以满足合规与审计需求（见第 22 节）。

#### 2.1 基础身份信息
  - **员工标识**：员工编号 (employeeNumber)、工号 (staffId)、全球员工ID (globalEmployeeId)
  - **姓名信息**：法定姓名 (legalName)、曾用名 (formerName)、姓名拼音 (namePinyin)、英文名 (englishName)、首选称呼 (preferredName)
  - **基本属性**：性别 (gender)、出生日期 (dateOfBirth)、血型 (bloodType)
  - **员工照片**：照片 (photo)、照片URL (photoUrl)
  - **证件信息**：
    - 证件类型 (idType): 居民身份证、护照、港澳居民来往内地通行证、台湾居民来往大陆通行证、外国人永久居留身份证
    - 证件号码 (idNumber)
    - 证件签发机关 (issuingAuthority)
    - 证件有效期起始日期 (idValidFrom)
    - 证件有效期终止日期 (idValidTo)
    - 是否长期有效 (isPermanent)
  - **国籍民族**：国籍 (nationality)、籍贯 (nativePlace)、民族 (ethnicity)、政治面貌 (politicalStatus)
  - **婚育状况**：婚姻状况 (maritalStatus)、生育状况 (parentalStatus)、子女数量 (numberOfChildren)

> **注**: 年龄 (age)、司龄 (companyYears)、工龄 (totalYears) 等为计算字段，不在主数据中存储

#### 2.2 户籍与居住信息
  - **户籍信息**：
    - 户籍类型 (residenceType): 城镇居民户口、农村居民户口、集体户口、军籍
    - 户籍所在省 (residenceProvince)
    - 户籍所在市 (residenceCity)
    - 户籍所在区/县 (residenceDistrict)
    - 户籍详细地址 (residenceAddress)
    - 户口所在派出所 (policeStation)
  - **居住证信息**（非本地户籍适用）：
    - 居住证编号 (residencePermitNumber)
    - 居住证签发日期 (residencePermitIssueDate)
    - 居住证有效期 (residencePermitExpiryDate)
  - **现居住地址**：
    - 现居住省 (currentProvince)
    - 现居住市 (currentCity)
    - 现居住区/县 (currentDistrict)
    - 现居住详细地址 (currentAddress)
    - 邮政编码 (postalCode)
    - 是否长期居住 (isPermanentResidence)

#### 2.3 联系方式
  - **个人联系方式**：
    - 手机号码 (mobilePhone)
    - 个人固定电话 (homePhone)
    - 个人电子邮箱 (personalEmail)
  - **企业联系方式**：
    - 企业邮箱 (workEmail)
    - 办公电话 (officePhone)
    - 办公分机号 (extension)
    - 企业微信号 (workWechatId)
    - 企业即时通讯账号 (workInstantMessagingId)
  - **紧急联系人**（支持多人）：
    - 紧急联系人列表 (emergencyContacts): Array
      - 姓名 (name)
      - 关系 (relationship)
      - 电话 (phone)
      - 地址 (address)
      - 优先级 (priority): 第一联系人、第二联系人、第三联系人

#### 2.4 组织与雇佣信息
  - **组织归属**：
    - 法人主体 (legalEntity)、公司名称 (companyName)
    - 成本中心 (costCenter)
    - 所属部门 (department)、部门编码 (departmentCode)
    - 业务单元 (businessUnit)
  - **职位信息**：
    - 职位 (position)、职位编码 (positionCode)
    - 岗位 (job)、岗位编码 (jobCode)
    - 职务 (jobTitle)
    - 职级 (grade)、职等 (rank)、职档 (step)
    - 序列 (jobFamily)、子序列 (jobSubFamily)
  - **汇报关系**：
    - 直接上级 (directManager)、直接上级员工编号 (directManagerId)
    - 职能上级 (functionalManager)、职能上级员工编号 (functionalManagerId)

> **注**: 职能上级用于矩阵式管理中的虚线汇报关系
  - **工作地点**：
    - 工作地点 (workLocation)
    - 办公地点 (officeLocation)
    - 工作城市 (workCity)
    - 是否远程办公 (isRemoteWorker)
  - **雇佣属性**：
    - 雇佣类型 (employmentType): 正式员工、劳务派遣、实习生、退休返聘、外包人员
    - 用工性质 (workArrangement): 全职、兼职
    - 员工类别 (employeeCategory): 管理人员、技术人员、销售人员、生产人员、支持人员
    - 员工状态 (employmentStatus): 在职、试用期、离职、停薪留职、内退
  - **重要日期**：
    - 入职日期 (hireDate)
    - 首次入职日期 (originalHireDate)
    - 试用期起始日期 (probationStartDate)
    - 试用期结束日期 (probationEndDate)
    - 转正日期 (regularizationDate)
    - 预计离职日期 (expectedTerminationDate)
    - 实际离职日期 (actualTerminationDate)
  - **离职信息**：
    - 离职原因 (terminationReason)
    - 离职类型 (terminationType): 主动离职、被动离职、协商解除、合同到期、退休
    - 是否可再雇佣 (rehireEligibility)
  - **入职来源**：
    - 招聘渠道 (recruitmentChannel): 校园招聘、社会招聘、内部推荐、猎头推荐
    - 推荐人 (referrer)、推荐人员工编号 (referrerId)

#### 2.5 模块交叉：劳动合同信息
> 为保持模块高内聚，劳动合同相关主数据已迁移至 **"22. 劳动合同管理 (Employment Contract Management)"**，包含合同模板、签署、续签、派遣外包及试用期等字段。人员管理模块仅保留与组织归属密切相关的用工信息。

#### 2.6 薪酬基线与发薪设置
  - **薪酬等级**：
    - 薪酬等级 (compensationGrade)
    - 薪级 (salaryLevel)
    - 薪档 (salaryStep)
  - **发薪设置**：
    - 薪资币种 (salaryCurrency)
    - 发薪周期 (payFrequency): 月薪、半月薪、周薪、日薪、小时工资
    - 发薪日 (payDay)
    - 发薪方式 (paymentMethod): 银行转账、现金、支票
  - **银行账户**：
    - 开户银行 (bankName)
    - 银行账号 (bankAccountNumber)
    - 开户支行 (branchName)
    - 银行卡类型 (cardType): 储蓄卡、工资卡

> **注**: 具体薪资组成（基本工资、岗位工资、绩效工资、津贴补贴等）请参见 **"6. 薪酬管理 (Compensation)"** 模块

#### 2.7 社保公积金基础信息
  - **社保信息**：
    - 社保参保地省份 (socialInsuranceProvince)
    - 社保参保地城市 (socialInsuranceCity)
    - 社保账号 (socialInsuranceAccountNumber)
    - 社保卡号 (socialInsuranceCardNumber)
    - 参保日期 (socialInsuranceStartDate)
    - 参保险种 (insuranceTypes): 养老保险、医疗保险、失业保险、工伤保险、生育保险
  - **公积金信息**：
    - 公积金账号 (providentFundAccountNumber)
    - 公积金参缴地 (providentFundCity)

> **注**: 社保公积金的缴纳基数、个人比例、企业比例等计算参数请参见 **"15. 薪资计算 (Payroll)"** 模块

#### 2.8 个税基础信息
  - **纳税人信息**：
    - 纳税人识别号 (taxpayerIdentificationNumber)
    - 个税居民类型 (taxResidencyType): 居民纳税人、非居民纳税人
    - 首次入境日期 (firstEntryDate)（非居民适用）
    - 是否享受税收协定 (isTaxTreatyApplicable)

> **注**: 专项附加扣除、累计减除费用、减免税额等计算参数请参见 **"15. 薪资计算 (Payroll)"** 模块

#### 2.9 工时制度与排班规则
  - **工时制度**：
    - 工时类型 (workingHoursType): 标准工时制、综合工时制、不定时工作制
    - 周工作天数 (workDaysPerWeek)
    - 日标准工作小时数 (standardDailyHours)
    - 周标准工作小时数 (standardWeeklyHours)
  - **排班信息**：
    - 排班组 (shiftGroup)
    - 默认班次 (defaultShift)
    - 工作日历 (workCalendar)
  - **加班规则**：
    - 加班核算方式 (overtimeCalculationMethod): 自动、手工、混合
    - 是否允许加班 (isOvertimeAllowed)
    - 平日加班倍率 (weekdayOvertimeRate)
    - 周末加班倍率 (weekendOvertimeRate)
    - 法定节假日加班倍率 (holidayOvertimeRate)

> **注**: 假期额度、休假记录等请参见 **"8. 时间与考勤"** 和 **"17. 缺勤管理"** 模块

#### 2.10 教育背景
  - **最高学历信息**：
    - 最高学历 (highestEducation): 博士研究生、硕士研究生、大学本科、大学专科、中专/技校、高中、初中及以下
    - 最高学位 (highestDegree): 博士、硕士、学士、无学位
    - 专业 (major)
    - 毕业院校 (graduatedFrom)
    - 毕业时间 (graduationDate)
    - 学历类型 (educationType): 全日制、非全日制、在职
  - **第二学历**（如有）：
    - 第二学历 (secondEducation)
    - 第二学位 (secondDegree)
    - 第二专业 (secondMajor)
    - 毕业院校 (secondGraduatedFrom)
    - 毕业时间 (secondGraduationDate)
  - **在读状态**：
    - 是否在职学习 (isStudyingPartTime)
    - 在读学历层次 (studyingEducationLevel)
    - 预计毕业时间 (expectedGraduationDate)

#### 2.11 资质与能力
  - **职称信息**：
    - 职称等级 (professionalTitle): 正高级、副高级、中级、初级、员级
    - 职称名称 (titleName)
    - 职称获得时间 (titleAcquiredDate)
    - 职称评审单位 (titleIssuingAuthority)
  - **职业资格证书**：
    - 证书名称 (certificationName)
    - 证书编号 (certificationNumber)
    - 颁发机构 (issuingOrganization)
    - 获证日期 (certificationDate)
    - 有效期至 (certificationExpiryDate)
    - 证书等级 (certificationLevel)
  - **语言能力**：
    - 语种 (language)
    - 熟练程度 (proficiencyLevel): 母语、精通、熟练、一般
    - 语言考试成绩 (languageTestScore)：如英语四六级、雅思、托福等
  - **技能等级**：
    - 技能类别 (skillCategory)
    - 技能名称 (skillName)
    - 技能等级 (skillLevel): 专家、熟练、一般、初级
    - 技能认证 (skillCertification)

#### 2.12 健康与安全
  - **健康信息**：
    - 入职体检日期 (preEmploymentCheckupDate)
    - 入职体检结论 (preEmploymentCheckupResult)
    - 定期体检日期 (regularCheckupDate)
    - 定期体检结论 (regularCheckupResult)
    - 健康状况 (healthStatus): 健康、良好、一般、较差
    - 健康限制说明 (healthRestrictions)
    - 残疾情况 (disabilityStatus)
    - 残疾等级 (disabilityLevel)
  - **职业健康**：
    - 是否接触职业病危害 (isExposedToOccupationalHazards)
    - 职业病危害因素 (occupationalHazardFactors)
    - 职业病危害告知签署日期 (hazardDisclosureSignDate)
    - 岗前职业健康检查日期 (preJobHealthCheckDate)
    - 在岗期间职业健康检查周期 (onJobHealthCheckFrequency)
  - **特种作业资格**：
    - 特种作业类型 (specialOperationType): 电工作业、焊接作业、高处作业、制冷作业等
    - 特种作业证编号 (specialOperationCertNumber)
    - 发证日期 (certIssueDate)
    - 复审日期 (certReviewDate)
    - 有效期至 (certExpiryDate)
  - **安全培训**：
    - 三级安全教育完成日期 (safetyEducationCompletionDate)
    - 最近安全培训日期 (latestSafetyTrainingDate)
    - 安全培训记录 (safetyTrainingRecords)

> **注**: 工伤事件记录（工伤发生日期、工伤类型、工伤等级、工伤认定文号）请参见 **"20. 健康安全 (Health & Safety)"** 模块的安全事件管理

#### 2.13 家庭关系
  - **家庭成员信息**（统一结构，支持多人）：
    - 家庭成员列表 (familyMembers): Array
      - 姓名 (name)
      - 关系 (relationship): 配偶、子女、父亲、母亲、其他
      - 性别 (gender)
      - 出生日期 (dateOfBirth)
      - 证件号码 (idNumber)
      - 工作单位/就读学校 (employerOrSchool)
      - 联系电话 (phoneNumber)
      - 是否在读 (isStudying)（仅子女适用）
      - 是否抚养 (isDependent)

#### 2.14 系统账号（基础）
  - **登录凭证**：
    - 企业域账号 (domainAccount)
    - 企业邮箱账号 (workEmail)

> **注**: IT资产管理、权限管理、个性化偏好设置等请参见 **"21. 横向支撑功能 (Cross-Functional Services)"** 模块的系统管理部分

- **个人信息维护**
  - 员工自助信息更新
  - 信息变更审批流程
  - 信息修改历史记录

- **雇佣记录**
  - 历史任职记录
  - 跨法人调动记录
  - 岗位变动轨迹

- **员工状态管理**
  - 在职状态监控
  - 试用期管理
  - 离职流程跟踪

### 说明
管理员工的基本信息和雇佣关系，是人力资源管理的核心数据基础。

---

## 3. 职位管理 (Position Management)

### 主要功能
- 职位创建与维护
- 职位预算控制
- 职位层级结构

### 说明
以职位为中心的管理模式，支持职位预算、编制控制等企业人力资源规划功能。

---

## 4. 人事管理 (Personnel Management)

### 主要功能
- 入职 (Hire)
- 转岗 (Transfer)
- 晋升 (Promotion)
- 离职 (Termination)
- 调薪 (Compensation Change)

### 说明
处理员工在职期间的各类人事变动，是 HR 日常操作的核心模块。

---

## 5. 工作信息 (Job Data)

### 主要功能
- 职务 (Job Code)
- 职级 (Grade)
- 薪酬计划
- 工作地点

### 说明
定义和管理职务、职级等工作相关的分类信息，为人事决策提供标准化依据。

---

## 6. 薪酬管理 (Compensation)

### 主要功能
- **薪资结构设计**
  - **薪资组成**：
    - 基本工资 (baseSalary)
    - 岗位工资 (positionSalary)
    - 绩效工资 (performanceSalary)
    - 年度目标奖金 (annualBonus)
  - **津贴补贴**：
    - 交通补贴 (transportationAllowance)
    - 通讯补贴 (communicationAllowance)
    - 餐补 (mealAllowance)
    - 住房补贴 (housingAllowance)
    - 其他补贴 (otherAllowances)
- **调薪管理**
  - 调薪历史记录
  - 调薪审批流程
  - 调薪原因分类
- **薪酬预算控制**
  - 部门薪酬预算
  - 薪酬预算执行监控
  - 薪酬成本分析

> **注**: 员工薪酬等级、薪级、薪档等基线信息存储在"2. 人员管理"模块

### 说明
管理员工的薪酬信息，包括薪资结构设计、调薪历史追踪、薪酬预算控制等。薪酬管理聚焦于薪资结构定义和调薪管理，具体薪资计算、社保公积金计算请参见"15. 薪资计算"模块。

---

## 7. 福利管理 (Benefits Administration)

### 主要功能
- 福利计划
- 福利资格
- 福利登记
- 生命事件

### 说明
管理企业提供的各类福利计划，支持员工福利选择和生命事件触发的福利变更。

---

## 8. 时间与考勤 (Time and Labor)

### 主要功能
- 工时报告
- 休假管理
- 考勤记录
- 排班管理

### 说明
记录和管理员工的工作时间、考勤状态、休假申请等时间相关信息。

---

## 9. 自助服务 (Self Service)

### 主要功能
- 员工自助服务 (ESS - Employee Self Service)
- 经理自助服务 (MSS - Manager Self Service)
- 信息查询与更新

### 说明
提供员工和经理的自助服务门户，支持信息查询、申请提交、审批等自助操作。

---

## 10. 报表与分析 (Reporting & Analytics)

### 主要功能
- 标准报表
- 查询工具
- 仪表板
- 数据分析

### 说明
提供各类人力资源报表、查询工具和数据分析功能，支持管理决策。

---

## 11. 招聘管理 (Recruiting / Talent Acquisition)

### 主要功能
- **职位需求管理**
  - 招聘需求申请与审批
  - 职位发布与管理
  - 招聘预算控制
- **候选人管理**
  - 候选人信息录入与追踪
  - 简历解析与筛选
  - 面试安排与反馈
  - 背景调查
- **录用流程**
  - Offer 管理
  - 入职准备

### 说明
贯穿招聘需求、候选人管理、面试、录用全流程，支持招聘 KPI 分析与协作。

---

## 12. 绩效管理 (Performance Management)

### 主要功能
- 目标设定
- 绩效评估
- 绩效校准
- 绩效反馈

### 说明
构建目标管理、绩效评估、绩效反馈的闭环体系，支持绩效数据沉淀与分析。

---

## 13. 培训与发展 (Learning & Development)

### 主要功能
- 培训计划
- 课程管理
- 学员管理
- 培训效果评估

### 说明
支持培训资源管理、项目执行、学习记录追踪与培训效果衡量。

---

## 14. 人才管理 (Talent Management)

### 主要功能
- 继任计划
- 潜力评估
- 人才盘点
- 职业发展计划

### 说明
帮助企业识别和保留核心人才，支持继任与发展规划。

---

## 15. 薪资计算 (Payroll)

### 主要功能
- **薪资计算引擎**
  - 薪资周期计算
  - 薪资公式配置
  - 薪资计算规则
- **社保公积金计算**
  - **养老保险**：
    - 养老保险缴纳基数 (pensionBase)
    - 养老保险个人比例 (pensionEmployeeRate)
    - 养老保险企业比例 (pensionEmployerRate)
  - **医疗保险**：
    - 医疗保险缴纳基数 (medicalBase)
    - 医疗保险个人比例 (medicalEmployeeRate)
    - 医疗保险企业比例 (medicalEmployerRate)
  - **失业保险**：
    - 失业保险缴纳基数 (unemploymentBase)
    - 失业保险个人比例 (unemploymentEmployeeRate)
    - 失业保险企业比例 (unemploymentEmployerRate)
  - **工伤保险**：
    - 工伤保险缴纳基数 (workInjuryBase)
    - 工伤保险企业比例 (workInjuryEmployerRate)
  - **生育保险**：
    - 生育保险缴纳基数 (maternityBase)
    - 生育保险企业比例 (maternityEmployerRate)
  - **公积金**：
    - 公积金缴纳基数 (providentFundBase)
    - 公积金个人比例 (providentFundEmployeeRate)
    - 公积金企业比例 (providentFundEmployerRate)
    - 补充公积金个人比例 (supplementaryFundEmployeeRate)
    - 补充公积金企业比例 (supplementaryFundEmployerRate)
- **个税计算**
  - **专项附加扣除**：
    - 子女教育 (childEducationDeduction)
    - 继续教育 (continuingEducationDeduction)
    - 大病医疗 (medicalDeduction)
    - 住房贷款利息 (housingLoanInterestDeduction)
    - 住房租金 (housingRentDeduction)
    - 赡养老人 (elderCareDeduction)
    - 婴幼儿照护 (infantCareDeduction)
  - **税务计算参数**：
    - 累计减除费用 (cumulativeDeduction)
    - 减免税额 (taxExemptionAmount)
  - 个税申报
  - 个税汇算清缴
- **薪资核对**
  - 薪资核对清单
  - 薪资差异分析
- **发薪管理**
  - 发薪批次管理
  - 银行代发文件生成
  - 发薪确认与回执
- **薪资报表**
  - 工资条生成与发放
  - 薪资汇总报表
  - 社保公积金汇总表
  - 个税申报表

> **注**: 员工社保公积金账号、个税纳税人识别号等基础信息存储在"2. 人员管理"模块

### 说明
支持薪资周期计算、社保公积金计算、个税计算、薪资审核、发薪与税务处理。薪资计算模块聚焦于薪资、社保、公积金、个税的计算逻辑和发薪流程。

---

## 16. 合规管理 (Regulatory & Compliance)

### 主要功能
- 劳动法规管理
- 合规审计
- 数据安全与隐私
- 风险管理

### 说明
保障企业遵守劳动法规、行业规范和内部政策，支持审计跟踪和风险控制。

---

## 17. 缺勤管理 (Absence Management)

### 主要功能
- **缺勤政策配置**
  - 假期类型定义
  - 假期计算规则
  - 假期结转规则
- **假期额度管理**
  - **年假管理**：
    - 年假额度 (annualLeaveEntitlement)
    - 当年已休年假 (annualLeaveTaken)
    - 剩余年假 (annualLeaveRemaining)
  - **调休管理**：
    - 调休额度 (compensatoryLeaveBalance)
  - **其他假期累计**：
    - 病假累计 (sickLeaveTaken)
    - 事假累计 (personalLeaveTaken)
- **缺勤申请与审批**
  - 请假申请
  - 请假审批流程
  - 销假管理
- **缺勤分析与报表**
  - 假期余额查询
  - 部门缺勤统计
  - 缺勤趋势分析

> **注**: 员工工时制度、排班信息、加班规则等基础信息存储在"2. 人员管理"模块

### 说明
统一管理假期与缺勤流程，支持假期余额计算与缺勤分析。缺勤管理聚焦于假期额度计算、缺勤申请审批和缺勤数据分析。

---

## 18. 员工关系 (Employee Relations)

### 主要功能
- 员工投诉
- 劳资纠纷
- 纪律处分
- 员工关怀

### 说明
处理员工关系事件，支持调查记录、行动计划和跟踪反馈。

---

## 19. 劳动力规划 (Workforce Planning)

### 主要功能
- 编制规划
- 人力需求预测
- 人员结构分析
- 情景模拟

### 说明
帮助企业进行中长期人力规划，支持编制管理与人力成本预算。

---

## 20. 健康安全 (Health & Safety)

### 主要功能
- **安全事件管理**
  - **工伤记录**：
    - 工伤发生日期 (workInjuryDate)
    - 工伤类型 (workInjuryType)
    - 工伤等级 (workInjuryLevel)
    - 工伤认定文号 (workInjuryRecognitionNumber)
  - 安全事故调查
  - 事故整改跟踪
- **健康检查管理**
  - 体检计划制定
  - 体检结果管理
  - 职业病筛查
- **安全培训管理**
  - 安全培训计划
  - 培训记录管理
  - 安全考试与认证
- **合规报告**
  - 安全事故报告
  - 职业健康报告
  - 安全检查报告

> **注**: 员工健康基础信息（入职体检、定期体检、特种作业资格等）存储在"2. 人员管理"模块

### 说明
关注员工健康与工作安全，支持工伤事件记录、安全事故管理、预防措施与合规报告。健康安全模块聚焦于安全事件处理和健康安全合规管理。

---

## 21. 横向支撑功能 (Cross-Functional Services)

### 主要功能
- 工作流与审批引擎
- 通知与提醒
- 集成接口管理
- 系统参数与安全

### 说明
为核心业务模块提供统一的流程、通知、权限和集成能力支撑。

---

## 22. 劳动合同管理 (Employment Contract Management)

### 主要功能
- 合同模板维护与版本控制（区分固定期限、无固定期限、派遣、劳务、实习等模板）
- 劳动合同签署、续签、补签与电子签约流程
- 合同试用期、到期、续签提醒与批量计划
- 派遣/外包/劳务派遣协议管理及双主体合规校验
- 合同变更、解除、终止审批流与归档管理
- 合同合规审计（对接政策变更、风险预警、法规清单）

### 数据字段参考
- **合同基本信息**：合同编号 (contractNumber)、签订主体 (contractingEntity)、签订日期 (contractSignDate)、首次签订日期 (firstContractDate)。
- **合同期限**：起始日期 (contractStartDate)、终止日期 (contractEndDate)、合同类型 (contractType)、合同期限 (contractDuration)、续签次数 (renewalCount)、是否无固定期限 (isOpenEnded)。
- **试用期管理**：试用期时长 (probationPeriodMonths)、试用期工资比例 (probationSalaryRate)、试用期评估节点 (probationReviewDate)。
- **派遣与外包信息**：派遣/外包单位 (dispatchAgency)、协议编号 (dispatchAgreementNumber)、派遣起止日期 (dispatchStartDate/dispatchEndDate)、派遣岗位 (dispatchPosition)、派遣费用结算方式 (dispatchSettlementMode)。
- **合规与提醒设置**：合同到期提醒日 (contractExpiryAlertDate)、是否存在竞业限制 (hasNonCompete)、竞业限制期限/补偿 (nonCompeteTerm/Compensation)、特殊审批编号 (specialApprovalNumber)。
- **归档与附件**：电子合同文件ID (contractDocumentId)、纸质合同存档位置 (physicalArchiveLocation)、归档状态 (archiveStatus)、电子签名证书编号 (esignCertificateNumber)。

### 说明
围绕劳动合同全生命周期管理，提供模板、签署、续签、派遣外包、合规审计的集中治理，与人员管理模块通过 `person_id`、`employment_id` 进行关联。模块需要与薪资、合规、员工关系等场景联动，确保符合《劳动合同法》《劳动派遣暂行规定》等监管要求，并支持提醒与审计追踪。

---

## 模块依赖关系概览

```
组织管理 (1) → 职位管理 (3) → 人员管理 (2) → 人事管理 (4)
                   ↘ 工作信息 (5) → 薪酬管理 (6)
                                   ↘ 薪资计算 (15)
人员管理 (2) → 绩效管理 (12) → 人才管理 (14)
                                    ↘ 培训管理 (13)
人事/时间考勤数据 (4/8/17) → 薪资计算 (15)
差旅/缺勤/加班 (8/17) → 薪资计算 (15)
```

---

## 典型业务流程示例

### 1. 招聘与入职流程

```
招聘需求 (11) → 招聘审批 (21) → 职位发布 (3/11)
   ↓
候选人筛选 (11) → 面试安排 (11) → Offer 审批 (11/21)
   ↓
入职准备 (4) → 入职手续 (4/2) → 新员工培训 (13)
```

### 2. 绩效与薪酬联动流程

```
目标设定 (12) → 中期检查 (12) → 绩效评估 (12)
   ↓
绩效校准 (12) → 绩效结果 (12)
   ↓
薪酬调整建议 (6) → 调薪审批 (4/21)
   ↓
薪资计算 (15) → 薪资发放 (15)
```

### 3. 人才盘点流程

```
确定盘点范围 (14) → 绩效数据收集 (12) → 潜力评估 (14)
    ↓
九宫格制作 (14) → 校准会议 (14) → 人才分类 (14)
    ↓
高潜人才：发展计划 (14) + 继任安排 (14)
低绩效人才：改进计划 (12) / 淘汰 (4)
核心骨干：保留激励 (6/7)
```

---

## 按用户角色的功能视图

### HR 专员视角

**日常操作高频模块**：
- 人员管理 (2)：员工信息维护
- 人事管理 (4)：入转调离办理
- 招聘管理 (11)：候选人跟进、面试安排
- 培训管理 (13)：培训班次组织、学员管理
- 时间考勤 (8/17)：考勤异常处理、假期审批

**周期性操作**：
- 薪资计算 (15)：月度薪资处理
- 绩效管理 (12)：绩效周期启动、数据收集
- 报表分析 (10)：月度/季度人力报表

### 直线经理视角

**管理职责模块**：
- 自助服务 (9)：团队信息查看、审批处理
- 绩效管理 (12)：目标设定、绩效评估、反馈辅导
- 培训管理 (13)：团队培训需求提报、培训效果评估
- 人才管理 (14)：高潜识别、IDP 制定
- 招聘管理 (11)：招聘需求提报、候选人面试

**审批事项**：
- 假期审批 (17)
- 加班审批 (8)
- 培训申请审批 (13)
- 转岗/晋升审批 (4)

### 员工视角 (ESS)

**自助查询**：
- 个人信息 (2)：查看和更新个人资料
- 薪资信息 (6/15)：工资单查询、薪资历史
- 假期余额 (17)：查看各类假期余额

**自助申请**：
- 请假申请 (17)
- 加班申请 (8)
- 培训申请 (13)
- 证明开具 (2)

**自助服务**：
- 绩效管理 (12)：目标查看、自评、反馈查看
- 培训学习 (13)：在线学习、培训报名
- 职业发展 (14)：职业路径查看、IDP 制定

### 系统管理员视角

**系统配置**：
- 组织管理 (1)：组织架构维护
- 工作信息 (5)：职务职级配置
- 系统管理 (21)：用户权限、系统参数

**数据治理**：
- 合规管理 (16)：审计日志、数据安全
- 集成接口 (21)：接口配置、数据同步
- 报表分析 (10)：自定义报表开发

**运维监控**：
- 工作流引擎 (21)：流程配置、流程监控
- 通知管理 (21)：通知模板、发送监控

### HRBP (人力资源业务伙伴) 视角

**战略支持**：
- 劳动力规划 (19)：人力需求预测、编制规划
- 人才管理 (14)：人才盘点、继任计划
- 报表分析 (10)：人力分析、决策支持

**业务协同**：
- 招聘管理 (11)：业务部门招聘支持
- 绩效管理 (12)：绩效体系设计、校准会议
- 组织发展：组织变革支持 (1)

**员工关系**：
- 员工关系 (18)：投诉处理、劳资纠纷、员工沟通
- 合规管理 (16)：劳动法规遵从、风险管理

---

## 应用说明

### 在 Cube Castle 项目中的参考价值

1. **功能完整性对标**: 本菜单可作为企业级 Core HR 系统的功能完整性检查清单
2. **模块边界定义**: 帮助明确各功能模块的职责边界和接口设计
3. **需求分析参考**: 在需求收集阶段，可参考标准菜单结构进行功能覆盖度评估
4. **术语标准化**: 统一使用行业标准术语，提升系统的专业性和可理解性

### 注意事项

- PeopleSoft 的具体菜单项和层级可能因版本和企业配置而异
- 本文档记录的是通用功能框架，实际实现需根据业务需求裁剪
- 建议结合项目的 API 契约（`docs/api/openapi.yaml`）和实现清单（`docs/reference/02-IMPLEMENTATION-INVENTORY.md`）进行功能映射

---

## 变更记录

| 日期 | 版本 | 变更说明 | 作者 |
|------|------|---------|------|
| 2025-10-19 | 3.1 | **模块划分优化**：新增第 22 模块“劳动合同管理 (Employment Contract Management)”并整合合同字段；人员管理模块补充主流 HCM 模块划分建议，2.5 节改为模块交叉指引；概述更新为 22 个模块并重组“薪资与合同合规模块”描述 | Claude |
| 2025-10-19 | 3.0 | **重大架构调整**：严格限定模块边界，确保人员管理聚焦于核心员工主数据<br>- **2. 人员管理**：移除计算字段（age, companyYears, totalYears）、补充缺失字段（photo, terminationReason等）、简化为14个子分类<br>- **6. 薪酬管理**：接收薪资组成、津贴补贴内容<br>- **15. 薪资计算**：接收社保公积金计算参数、个税计算参数<br>- **17. 缺勤管理**：接收假期额度管理内容<br>- **20. 健康安全**：接收工伤记录管理内容<br>- 删除2.15个性化偏好（属于系统管理），大幅简化2.14系统账号 | Claude |
| 2025-10-19 | 2.1 | 详细补充员工主数据字段分类与具体字段：新增15个子分类（基础身份信息、户籍与居住信息、联系方式、组织与雇佣信息、劳动合同信息、薪酬信息基线、社保公积金信息、个税信息、考勤与工时规则、教育背景、资质与能力、健康与安全、家庭关系、系统访问与资产、个性化偏好），覆盖200+字段，参考中国大陆主流HCM软件实践 | Claude |
| 2025-10-13 | 2.0 | 全面补充：新增11个核心模块（招聘、绩效、培训、人才、薪资、合规、缺勤、员工关系、劳动力规划、健康安全、横向支撑），添加模块依赖关系、典型业务流程、用户角色视图 | Claude |
| 2025-10-13 | 1.0 | 初始版本，记录 PeopleSoft Core HR 基础 10 个功能菜单 | Claude |

---

## 相关文档

- `docs/reference/02-IMPLEMENTATION-INVENTORY.md` - 系统实现清单
- `docs/api/openapi.yaml` - REST API 契约
- `docs/api/schema.graphql` - GraphQL 查询契约
- `80-position-management-with-temporal-tracking.md` - 职位管理实现计划
