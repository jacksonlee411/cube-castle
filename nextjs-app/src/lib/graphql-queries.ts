// src/lib/graphql-queries.ts
import { gql } from '@apollo/client';

// Employee Queries
export const GET_EMPLOYEES = gql`
  query GetEmployees(
    $filters: EmployeeFilters
    $first: Int
    $after: String
  ) {
    employees(filters: $filters, first: $first, after: $after) {
      edges {
        node {
          id
          employeeId
          legalName
          preferredName
          email
          status
          hireDate
          terminationDate
          currentPosition {
            positionTitle
            department
            jobLevel
            location
            employmentType
            effectiveDate
          }
        }
        cursor
      }
      pageInfo {
        hasNextPage
        hasPreviousPage
        startCursor
        endCursor
      }
      totalCount
    }
  }
`;

export const GET_EMPLOYEE = gql`
  query GetEmployee($id: ID!) {
    employee(id: $id) {
      id
      employeeId
      legalName
      preferredName
      email
      status
      hireDate
      terminationDate
      currentPosition {
        id
        positionTitle
        department
        jobLevel
        location
        employmentType
        reportsToEmployeeId
        effectiveDate
        minSalary
        maxSalary
        currency
      }
      manager {
        id
        legalName
        email
      }
      directReports {
        id
        legalName
        email
        currentPosition {
          positionTitle
          department
        }
      }
    }
  }
`;

// Position History Queries
export const GET_POSITION_HISTORY = gql`
  query GetPositionHistory(
    $employeeId: ID!
    $fromDate: Time
    $toDate: Time
    $limit: Int
  ) {
    employee(id: $employeeId) {
      id
      legalName
      employeeId
      positionHistory(fromDate: $fromDate, toDate: $toDate, limit: $limit) {
        edges {
          node {
            id
            positionTitle
            department
            jobLevel
            location
            employmentType
            reportsToEmployeeId
            effectiveDate
            endDate
            changeReason
            isRetroactive
            minSalary
            maxSalary
            currency
          }
          cursor
        }
        pageInfo {
          hasNextPage
          hasPreviousPage
          startCursor
          endCursor
        }
        totalCount
      }
    }
  }
`;

export const GET_POSITION_TIMELINE = gql`
  query GetPositionTimeline($employeeId: ID!, $maxEntries: Int) {
    employee(id: $employeeId) {
      id
      legalName
      employeeId
      positionTimeline(maxEntries: $maxEntries) {
        id
        positionTitle
        department
        jobLevel
        location
        employmentType
        reportsToEmployeeId
        effectiveDate
        endDate
        changeReason
        isRetroactive
        minSalary
        maxSalary
        currency
      }
    }
  }
`;

// Position Change Mutations
export const CREATE_POSITION_CHANGE = gql`
  mutation CreatePositionChange($input: CreatePositionChangeInput!) {
    createPositionChange(input: $input) {
      positionHistory {
        id
        positionTitle
        department
        effectiveDate
        isRetroactive
      }
      workflowId
      errors {
        message
        field
        code
      }
    }
  }
`;

export const VALIDATE_POSITION_CHANGE = gql`
  mutation ValidatePositionChange($employeeId: ID!, $effectiveDate: Time!) {
    validatePositionChange(employeeId: $employeeId, effectiveDate: $effectiveDate) {
      isValid
      errors {
        code
        message
        field
      }
      warnings {
        code
        message
        severity
      }
    }
  }
`;

// Workflow Queries
export const GET_WORKFLOW_STATUS = gql`
  query GetWorkflowStatus($workflowId: String!) {
    workflowStatus(workflowId: $workflowId) {
      workflowId
      status
      currentStep
      progress
      startedAt
      updatedAt
      completedAt
      error
    }
  }
`;

// Approval Mutations
export const APPROVE_POSITION_CHANGE = gql`
  mutation ApprovePositionChange($workflowId: String!, $comments: String) {
    approvePositionChange(workflowId: $workflowId, comments: $comments) {
      success
      workflowId
      errors {
        message
        field
        code
      }
    }
  }
`;

export const REJECT_POSITION_CHANGE = gql`
  mutation RejectPositionChange($workflowId: String!, $reason: String!) {
    rejectPositionChange(workflowId: $workflowId, reason: $reason) {
      success
      workflowId
      errors {
        message
        field
        code
      }
    }
  }
`;

// Organization Queries
export const GET_ORGANIZATION_CHART = gql`
  query GetOrganizationChart(
    $rootDepartment: String
    $asOfDate: Time
    $maxLevels: Int
  ) {
    organizationChart(
      rootDepartment: $rootDepartment
      asOfDate: $asOfDate
      maxLevels: $maxLevels
    ) {
      department
      employees {
        id
        legalName
        currentPosition {
          positionTitle
          jobLevel
        }
      }
      subDepartments {
        department
        employees {
          id
          legalName
          currentPosition {
            positionTitle
            jobLevel
          }
        }
        managerCount
        totalEmployees
      }
      managerCount
      totalEmployees
    }
  }
`;

// Analytics Queries
export const GET_POSITION_CHANGE_ANALYTICS = gql`
  query GetPositionChangeAnalytics(
    $fromDate: Time!
    $toDate: Time!
    $groupBy: AnalyticsGroupBy
  ) {
    positionChangeAnalytics(
      fromDate: $fromDate
      toDate: $toDate
      groupBy: $groupBy
    ) {
      period
      totalChanges
      newHires
      promotions
      lateralMoves
      departmentChanges
      retroactiveChanges
      averageSalaryChange
    }
  }
`;

// Subscriptions
export const EMPLOYEE_POSITION_CHANGED = gql`
  subscription EmployeePositionChanged($employeeId: ID) {
    employeePositionChanged(employeeId: $employeeId) {
      id
      employeeId
      positionTitle
      department
      effectiveDate
      isRetroactive
    }
  }
`;

export const WORKFLOW_STATUS_CHANGED = gql`
  subscription WorkflowStatusChanged($workflowId: String!) {
    workflowStatusChanged(workflowId: $workflowId) {
      workflowId
      status
      currentStep
      progress
      startedAt
      updatedAt
      completedAt
      error
    }
  }
`;

export const POSITION_CHANGE_APPROVAL_REQUIRED = gql`
  subscription PositionChangeApprovalRequired($approverId: ID!) {
    positionChangeApprovalRequired(approverId: $approverId) {
      workflowId
      employeeId
      currentPosition {
        positionTitle
        department
      }
      newPosition {
        positionTitle
        department
        minSalary
        maxSalary
        currency
      }
      requestedBy
      requestedAt
      dueDate
      priority
    }
  }
`;

// Organization Chart Queries
export const GET_ORGANIZATION_METRICS = gql`
  query GetOrganizationMetrics($department: String, $asOfDate: Time) {
    organizationMetrics(department: $department, asOfDate: $asOfDate) {
      totalEmployees
      totalDepartments
      averageTeamSize
      maxReportingDepth
      spanOfControl
      departmentMetrics {
        department
        employeeCount
        managerCount
        averageSpan
        maxDepth
      }
    }
  }
`;

export const GET_REPORTING_HIERARCHY = gql`
  query GetReportingHierarchy($managerId: ID!, $maxDepth: Int) {
    getReportingHierarchy(managerId: $managerId, maxDepth: $maxDepth) {
      manager {
        id
        legalName
        currentPosition {
          positionTitle
          jobLevel
        }
      }
      directReports {
        id
        legalName
        currentPosition {
          positionTitle
          jobLevel
        }
      }
      allReports {
        id
        legalName
        currentPosition {
          positionTitle
          jobLevel
        }
      }
      depth
    }
  }
`;

export const GET_EMPLOYEE_INFLUENCE = gql`
  query GetEmployeeInfluence($employeeId: ID!, $analysisDepth: Int) {
    getEmployeeInfluence(employeeId: $employeeId, analysisDepth: $analysisDepth) {
      employee {
        id
        legalName
        currentPosition {
          positionTitle
          jobLevel
        }
      }
      influenceScore
      connectionsCount
      departments
      keyConnections {
        employee {
          id
          legalName
          currentPosition {
            positionTitle
            jobLevel
          }
        }
        connectionType
        strength
      }
    }
  }
`;

export const GET_COLLABORATION_NETWORK = gql`
  query GetCollaborationNetwork($department: String!, $includeExternal: Boolean) {
    getCollaborationNetwork(department: $department, includeExternal: $includeExternal) {
      department
      employees {
        id
        legalName
        currentPosition {
          positionTitle
          jobLevel
        }
      }
      connections {
        fromEmployee {
          id
          legalName
        }
        toEmployee {
          id
          legalName
        }
        connectionType
        weight
      }
      networkDensity
      averagePathLength
    }
  }
`;

export const GET_TEAM_STRUCTURE = gql`
  query GetTeamStructure($teamId: ID, $managerId: ID, $department: String) {
    getTeamStructure(teamId: $teamId, managerId: $managerId, department: $department) {
      teamId
      teamName
      manager {
        id
        legalName
        currentPosition {
          positionTitle
          jobLevel
        }
      }
      members {
        employee {
          id
          legalName
          currentPosition {
            positionTitle
            jobLevel
          }
        }
        role
        joinDate
        isCore
      }
      subTeams {
        teamId
        teamName
        manager {
          id
          legalName
        }
        members {
          employee {
            id
            legalName
          }
          role
          isCore
        }
        teamMetrics {
          memberCount
          averageTenure
          diversityIndex
          collaborationScore
        }
      }
      teamMetrics {
        memberCount
        averageTenure
        diversityIndex
        collaborationScore
      }
    }
  }
`;

export const GET_SUCCESSION_CANDIDATES = gql`
  query GetSuccessionCandidates($positionId: ID!, $criteria: SuccessionCriteria) {
    getSuccessionCandidates(positionId: $positionId, criteria: $criteria) {
      employee {
        id
        legalName
        currentPosition {
          positionTitle
          jobLevel
        }
      }
      readinessScore
      skillMatch
      experienceMatch
      riskFactors
      developmentNeeds
    }
  }
`;

export const GET_ORGANIZATIONAL_HEALTH = gql`
  query GetOrganizationalHealth($department: String, $timeRange: TimeRange) {
    getOrganizationalHealth(department: $department, timeRange: $timeRange) {
      overallScore
      turnoverRate
      promotionRate
      averageTenure
      spanOfControlHealth
      communicationHealth
      departmentHealthScores {
        department
        healthScore
        turnoverRate
        averageTenure
        managerEffectiveness
        teamCohesion
      }
    }
  }
`;

// Graph Sync Mutations
export const SYNC_EMPLOYEE_TO_GRAPH = gql`
  mutation SyncEmployeeToGraph($employeeId: ID!) {
    syncEmployeeToGraph(employeeId: $employeeId)
  }
`;

export const SYNC_POSITION_TO_GRAPH = gql`
  mutation SyncPositionToGraph($positionId: ID!, $employeeId: ID!) {
    syncPositionToGraph(positionId: $positionId, employeeId: $employeeId)
  }
`;

export const CREATE_REPORTING_RELATIONSHIP = gql`
  mutation CreateReportingRelationship($managerId: ID!, $reporteeId: ID!) {
    createReportingRelationship(managerId: $managerId, reporteeId: $reporteeId)
  }
`;

export const FULL_GRAPH_SYNC = gql`
  mutation FullGraphSync {
    fullGraphSync {
      success
      syncedEmployees
      syncedPositions
      syncedRelationships
      errors
    }
  }
`;

export const SYNC_DEPARTMENT = gql`
  mutation SyncDepartment($department: String!) {
    syncDepartment(department: $department) {
      success
      syncedEmployees
      syncedPositions
      syncedRelationships
      errors
    }
  }
`;

// Organization Subscriptions
export const ORGANIZATION_STRUCTURE_CHANGED = gql`
  subscription OrganizationStructureChanged($department: String) {
    organizationStructureChanged(department: $department) {
      type
      department
      affectedEmployees
      timestamp
    }
  }
`;

export const REPORTING_RELATIONSHIP_CHANGED = gql`
  subscription ReportingRelationshipChanged($employeeId: ID) {
    reportingRelationshipChanged(employeeId: $employeeId) {
      type
      employeeId
      oldManagerId
      newManagerId
      affectedReports
      timestamp
    }
  }
`;

export const TEAM_COMPOSITION_CHANGED = gql`
  subscription TeamCompositionChanged($teamId: ID, $department: String) {
    teamCompositionChanged(teamId: $teamId, department: $department) {
      type
      teamId
      employeeId
      changeDetails
      timestamp
    }
  }
`;

// SAM (Situational Awareness Model) Queries
export const GET_SITUATIONAL_CONTEXT = gql`
  query GetSituationalContext($filters: AnalysisFilters) {
    getSituationalContext(filters: $filters) {
      timestamp
      alertLevel
      organizationHealth {
        overallScore
        turnoverRate
        engagementLevel
        productivityIndex
        spanOfControlHealth
        departmentHealth {
          department
          healthScore
          turnoverRate
          averageTenure
          managerEffectiveness
          teamCohesion
          workloadBalance
          lastAssessment
        }
        trendAnalysis {
          trend
          trendStrength
          keyDrivers
          predictedHealth
          confidence
        }
      }
      talentMetrics {
        talentPipelineHealth
        successionReadiness
        skillGapAnalysis {
          skillArea
          currentLevel
          requiredLevel
          gapSize
          priority
          affectedRoles
          closureStrategy
        }
        performanceDistribution {
          highPerformers
          solidPerformers
          lowPerformers
          performanceGaps
        }
        learningDevelopmentROI
        internalMobilityRate
      }
      riskAssessment {
        overallRiskScore
        keyPersonRisks {
          employeeId
          employeeName
          position
          department
          riskScore
          riskFactors
          businessImpact
          mitigationSteps
          lastAssessment
        }
        complianceRisks {
          riskType
          severity
          description
          affectedAreas
          complianceGaps
          remediationPlan
          deadline
        }
        operationalRisks {
          riskCategory
          description
          probability
          impact
          riskScore
          affectedTeams
          contingencyPlan
        }
        talentFlightRisks {
          employeeId
          employeeName
          flightRisk
          riskIndicators
          retentionActions
          timeFrame
        }
        riskMitigation {
          riskType
          mitigationAction
          effectiveness
          timeline
          responsibleParty
        }
      }
      opportunities {
        talentOptimization {
          opportunityType
          description
          affectedRoles
          expectedBenefit
          implementationSteps
        }
        processImprovements {
          processArea
          currentState
          proposedState
          efficiencyGain
          implementationComplexity
        }
        structuralChanges {
          changeType
          description
          rationale
          affectedTeams
          implementationPhases
        }
        investmentPriorities {
          investmentArea
          priority
          estimatedCost
          expectedROI
          justification
        }
        capabilityGaps {
          capabilityArea
          currentLevel
          requiredLevel
          gapSize
          closureStrategy
        }
      }
      recommendations {
        id
        type
        priority
        category
        title
        description
        businessImpact
        implementation {
          timeline
          phases {
            phaseNumber
            phaseName
            duration
            activities
            dependencies
            deliverables
          }
          resources {
            resourceType
            quantity
            skillRequirements
            timeCommitment
          }
          keyMilestones {
            name
            description
            targetDate
            successMetrics
          }
          successCriteria
        }
        roiEstimate {
          costSavings
          revenueIncrease
          efficiencyGains
          riskReduction
          timeToBreakeven
          confidenceLevel
        }
        riskFactors
        successMetrics {
          kpis {
            name
            description
            currentValue
            targetValue
            measurement
          }
          measurementPlan
          reportingCadence
        }
        dependencies
        confidence
      }
    }
  }
`;

export const GET_ORGANIZATION_INSIGHTS = gql`
  query GetOrganizationInsights($department: String, $timeRange: TimeRange) {
    getOrganizationInsights(department: $department, timeRange: $timeRange) {
      insightType
      summary
      keyFindings {
        category
        finding
        impact
        confidence
        evidence
        recommendation
      }
      trendAnalysis {
        trend
        trendStrength
        keyDrivers
        predictedOutcome
        confidence
      }
      actionItems {
        priority
        category
        action
        timeline
        responsibleParty
        expectedImpact
      }
      generatedAt
    }
  }
`;

export const GET_TALENT_ANALYTICS = gql`
  query GetTalentAnalytics(
    $analysisType: String
    $department: String
    $filters: AnalysisFilters
  ) {
    getTalentAnalytics(
      analysisType: $analysisType
      department: $department
      filters: $filters
    ) {
      talentHealth {
        overallScore
        engagementLevel
        retentionRate
        developmentIndex
        successionReadiness
      }
      skillAnalysis {
        skillGaps {
          skillArea
          currentLevel
          requiredLevel
          gapSize
          priority
          affectedRoles
        }
        developmentPriorities
        emergingSkillNeeds
        skillInventory {
          skillArea
          totalProficiency
          employeeCount
          averageLevel
          expertCount
        }
      }
      performanceInsights {
        highPerformersRatio
        performanceGaps
        improvementAreas
        performanceTrends {
          metric
          trend
          changePercentage
          period
        }
      }
      careerPathAnalysis {
        internalMobilityRate
        promotionReadiness
        careerPathClarity
        developmentGaps
        successorMapping {
          position
          readyCandidates
          developingCandidates
          gapAnalysis
        }
      }
      generatedAt
    }
  }
`;

export const GET_RISK_INSIGHTS = gql`
  query GetRiskInsights(
    $riskCategory: String
    $department: String
    $filters: AnalysisFilters
  ) {
    getRiskInsights(
      riskCategory: $riskCategory
      department: $department
      filters: $filters
    ) {
      overallRiskLevel
      riskScore
      keyRisks {
        riskType
        severity
        probability
        impact
        description
        affectedAreas
        mitigationPlan
        timeline
        monitoringKPIs
      }
      trendAnalysis {
        trend
        riskEvolution
        emergingRisks
        riskMitigation
      }
      recommendations {
        priority
        action
        expectedReduction
        implementation
        monitoringPlan
      }
      generatedAt
    }
  }
`;

export const GET_PERFORMANCE_PREDICTIONS = gql`
  query GetPerformancePredictions(
    $predictionType: String
    $parameters: PredictionParameters
  ) {
    getPerformancePredictions(
      predictionType: $predictionType
      parameters: $parameters
    ) {
      predictionHorizon
      confidence
      predictedMetrics {
        metricName
        currentValue
        predictedValue
        changePercentage
        trend
        influencingFactors
      }
      scenarioAnalysis {
        scenarioName
        probability
        expectedOutcome
        keyAssumptions
        impactMetrics {
          metric
          baselineValue
          scenarioValue
          impact
        }
      }
      generatedAt
    }
  }
`;

export const GET_STRATEGIC_RECOMMENDATIONS = gql`
  query GetStrategicRecommendations(
    $category: RecommendationCategory
    $priority: Priority
    $department: String
  ) {
    getStrategicRecommendations(
      category: $category
      priority: $priority
      department: $department
    ) {
      id
      type
      priority
      category
      title
      description
      businessImpact
      implementation {
        timeline
        phases {
          phaseNumber
          phaseName
          duration
          activities
          dependencies
          deliverables
        }
        resources {
          resourceType
          quantity
          skillRequirements
          timeCommitment
        }
        keyMilestones {
          name
          description
          targetDate
          successMetrics
        }
        successCriteria
      }
      roiEstimate {
        costSavings
        revenueIncrease
        efficiencyGains
        riskReduction
        timeToBreakeven
        confidenceLevel
      }
      riskFactors
      successMetrics {
        kpis {
          name
          description
          currentValue
          targetValue
          measurement
        }
        measurementPlan
        reportingCadence
      }
      dependencies
      confidence
    }
  }
`;

export const GET_SUCCESSION_PLANNING = gql`
  query GetSuccessionPlanning($department: String, $criticalRolesOnly: Boolean) {
    getSuccessionPlanning(department: $department, criticalRolesOnly: $criticalRolesOnly) {
      position
      currentHolder {
        id
        legalName
        currentPosition {
          positionTitle
          department
        }
      }
      readyCandidates {
        employee {
          id
          legalName
          currentPosition {
            positionTitle
            department
          }
        }
        readinessScore
        skillMatch
        experienceMatch
        developmentNeeds
        timeToReadiness
      }
      developingCandidates {
        employee {
          id
          legalName
          currentPosition {
            positionTitle
            department
          }
        }
        readinessScore
        skillMatch
        experienceMatch
        developmentNeeds
        timeToReadiness
      }
      riskLevel
      developmentPlan {
        actionType
        description
        timeline
        cost
        expectedOutcome
      }
    }
  }
`;

export const GET_SKILL_GAP_ANALYSIS = gql`
  query GetSkillGapAnalysis($department: String, $skillCategory: String) {
    getSkillGapAnalysis(department: $department, skillCategory: $skillCategory) {
      skillArea
      currentLevel
      requiredLevel
      gapSize
      priority
      affectedRoles
      closureStrategy
    }
  }
`;

export const GET_BENCHMARK_ANALYSIS = gql`
  query GetBenchmarkAnalysis($benchmarkType: String, $comparisonGroup: String) {
    getBenchmarkAnalysis(benchmarkType: $benchmarkType, comparisonGroup: $comparisonGroup) {
      benchmarkType
      currentPerformance
      industryAverage
      topQuartile
      percentileRank
      improvementPotential
      recommendations
    }
  }
`;

// SAM Subscriptions
export const SITUATIONAL_CONTEXT_UPDATED = gql`
  subscription SituationalContextUpdated {
    situationalContextUpdated {
      timestamp
      alertLevel
      organizationHealth {
        overallScore
        turnoverRate
        engagementLevel
      }
      riskAssessment {
        overallRiskScore
      }
    }
  }
`;

export const RISK_ALERT_TRIGGERED = gql`
  subscription RiskAlertTriggered($department: String, $riskLevel: RiskLevel) {
    riskAlertTriggered(department: $department, riskLevel: $riskLevel) {
      alertId
      riskType
      severity
      description
      affectedAreas
      recommendedActions
      timestamp
    }
  }
`;

export const PERFORMANCE_THRESHOLD_BREACHED = gql`
  subscription PerformanceThresholdBreached($metric: String, $threshold: Float) {
    performanceThresholdBreached(metric: $metric, threshold: $threshold) {
      alertId
      metric
      currentValue
      threshold
      trend
      department
      timestamp
    }
  }
`;

export const TALENT_PIPELINE_ALERT = gql`
  subscription TalentPipelineAlert($department: String, $alertType: String) {
    talentPipelineAlert(department: $department, alertType: $alertType) {
      alertId
      alertType
      description
      affectedEmployees
      recommendedActions
      priority
      timestamp
    }
  }
`;

// Bulk operations
export const BULK_CREATE_POSITION_CHANGES = gql`
  mutation BulkCreatePositionChanges($changes: [CreatePositionChangeInput!]!) {
    bulkCreatePositionChanges(changes: $changes) {
      successCount
      errorCount
      results {
        positionHistory {
          id
          positionTitle
          department
        }
        workflowId
        errors {
          message
          field
          code
        }
      }
    }
  }
`;

// Search and filtering
export const FIND_REPORTING_PATH = gql`
  query FindReportingPath($fromEmployeeId: ID!, $toEmployeeId: ID!) {
    findReportingPath(fromEmployeeId: $fromEmployeeId, toEmployeeId: $toEmployeeId) {
      distance
      pathType
      path {
        employee {
          id
          legalName
          currentPosition {
            positionTitle
            department
          }
        }
        relationship
      }
    }
  }
`;

export const FIND_COMMON_MANAGER = gql`
  query FindCommonManager($employeeIds: [ID!]!) {
    findCommonManager(employeeIds: $employeeIds) {
      id
      legalName
      currentPosition {
        positionTitle
        department
      }
    }
  }
`;