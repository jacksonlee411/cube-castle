package validator

import (
	"context"
	"errors"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	pkglogger "cube-castle/pkg/logger"
)

// RuleSeverity 标准化严重级别枚举，保持与审计/错误码对齐。
type RuleSeverity string

const (
	SeverityCritical RuleSeverity = "CRITICAL"
	SeverityHigh     RuleSeverity = "HIGH"
	SeverityMedium   RuleSeverity = "MEDIUM"
	SeverityLow      RuleSeverity = "LOW"

	defaultRulePriority = 100
)

// RuleHandler 执行单条规则，返回增量校验结果。
type RuleHandler func(ctx context.Context, subject interface{}) (*RuleOutcome, error)

// RuleOutcome 表示规则执行后产生的错误、警告及附加上下文。
type RuleOutcome struct {
	Errors   []ValidationError
	Warnings []ValidationWarning
	Context  map[string]interface{}
}

// Rule 定义链式执行中的单条业务校验规则。
type Rule struct {
	ID            string
	Description   string
	Severity      RuleSeverity
	Priority      int
	ShortCircuit  bool
	Handler       RuleHandler
	TelemetryOnly bool // 如为 true，失败不会影响 Valid（后续规则仍可执行）
}

// ValidationChain 负责按照优先级顺序执行规则并聚合结果。
type ValidationChain struct {
	mu          sync.RWMutex
	rules       []*Rule
	sorted      bool
	baseContext map[string]interface{}
	logger      pkglogger.Logger
	operation   string
}

// ChainOption 配置链式执行器。
type ChainOption func(*ValidationChain)

// WithBaseContext 预设链路级上下文，会注入到最终 ValidationResult.Context。
func WithBaseContext(ctx map[string]interface{}) ChainOption {
	return func(chain *ValidationChain) {
		if len(ctx) == 0 {
			return
		}
		copied := make(map[string]interface{}, len(ctx))
		for k, v := range ctx {
			copied[k] = v
		}
		chain.baseContext = copied
	}
}

// WithOperationLabel 为验证链记录 operation 标签，便于指标聚合。
func WithOperationLabel(operation string) ChainOption {
	return func(chain *ValidationChain) {
		operation = strings.TrimSpace(operation)
		if operation == "" {
			return
		}
		chain.operation = operation
	}
}

// NewValidationChain 创建链式执行器，若 logger 为空则使用静默记录器。
func NewValidationChain(logger pkglogger.Logger, opts ...ChainOption) *ValidationChain {
	if logger == nil {
		logger = pkglogger.NewLogger(
			pkglogger.WithWriter(io.Discard),
			pkglogger.WithLevel(pkglogger.LevelError),
		)
	}

	chain := &ValidationChain{
		rules:       make([]*Rule, 0),
		baseContext: map[string]interface{}{},
		logger: logger.WithFields(pkglogger.Fields{
			"component": "validator",
			"layer":     "chain",
		}),
	}

	for _, opt := range opts {
		if opt != nil {
			opt(chain)
		}
	}
	return chain
}

// Register 向链式执行器注册规则。
func (c *ValidationChain) Register(rule *Rule) error {
	if rule == nil {
		return errors.New("validator chain: nil rule provided")
	}
	ruleID := strings.TrimSpace(strings.ToUpper(rule.ID))
	if ruleID == "" {
		return errors.New("validator chain: rule ID is required")
	}
	if rule.Handler == nil {
		return errors.New("validator chain: handler is required")
	}

	if rule.Priority == 0 {
		rule.Priority = defaultRulePriority
	}
	rule.ID = ruleID
	rule.Severity = normalizeSeverity(rule.Severity)

	c.mu.Lock()
	defer c.mu.Unlock()
	c.rules = append(c.rules, rule)
	c.sorted = false
	return nil
}

// Execute 依次执行规则并聚合结果，支持短路控制。
func (c *ValidationChain) Execute(ctx context.Context, subject interface{}) *ValidationResult {
	overallStart := time.Now()
	result := NewValidationResult()
	for k, v := range c.baseContext {
		result.Context[k] = v
	}

	rules := c.snapshotRules()
	executedRuleIDs := make([]string, 0, len(rules))

	for _, rule := range rules {
		if ctx.Err() != nil {
			c.logger.WithFields(pkglogger.Fields{
				"ruleId": rule.ID,
				"error":  ctx.Err(),
			}).Warn("validation chain aborted due to context cancellation")
			result.Context["cancelled"] = true
			break
		}

		start := time.Now()
		outcome, err := rule.Handler(ctx, subject)
		duration := time.Since(start)

		c.logger.WithFields(pkglogger.Fields{
			"ruleId":       rule.ID,
			"durationMs":   duration.Milliseconds(),
			"shortCircuit": rule.ShortCircuit,
		}).Debug("validation rule executed")

		executedRuleIDs = append(executedRuleIDs, rule.ID)

		ruleOutcomeLabel := ruleOutcomeLabelSuccess

		if err != nil {
			c.logger.WithFields(pkglogger.Fields{
				"ruleId": rule.ID,
				"error":  err,
			}).Error("validation rule execution failed")

			result.Errors = append(result.Errors, ValidationError{
				Code:     "VALIDATION_RULE_EXECUTION_ERROR",
				Message:  err.Error(),
				Severity: string(SeverityCritical),
				Context: map[string]interface{}{
					"ruleId":       rule.ID,
					"internal":     true,
					"shortCircuit": rule.ShortCircuit,
				},
			})

			ruleOutcomeLabel = ruleOutcomeLabelError
			observeRuleMetrics(rule.ID, ruleOutcomeLabel, duration)

			if rule.ShortCircuit && !rule.TelemetryOnly {
				break
			}
			continue
		}

		if outcome == nil {
			observeRuleMetrics(rule.ID, ruleOutcomeLabel, duration)
			continue
		}

		mergeRuleOutcome(result, rule, outcome)

		if len(outcome.Errors) > 0 {
			ruleOutcomeLabel = ruleOutcomeLabelFailed
		} else if len(outcome.Warnings) > 0 {
			ruleOutcomeLabel = ruleOutcomeLabelWarning
		}

		observeRuleMetrics(rule.ID, ruleOutcomeLabel, duration)

		if rule.ShortCircuit && len(outcome.Errors) > 0 && !rule.TelemetryOnly {
			break
		}
	}

	if len(executedRuleIDs) > 0 {
		result.Context["executedRules"] = executedRuleIDs
	} else if _, ok := result.Context["executedRules"]; !ok {
		result.Context["executedRules"] = []string{}
	}

	result.Valid = len(result.Errors) == 0

	operation, _ := result.Context["operation"].(string)
	if operation == "" {
		operation = c.operation
	}

	outcomeLabel := chainOutcomeLabelSuccess
	if cancelled, ok := result.Context["cancelled"].(bool); ok && cancelled {
		outcomeLabel = chainOutcomeLabelCancelled
	} else if !result.Valid {
		outcomeLabel = chainOutcomeLabelFailed
	}
	observeChainMetrics(operation, outcomeLabel, time.Since(overallStart))

	return result
}

// SeverityToHTTPStatus 将严重级别映射到 HTTP 状态码，用于统一错误转换。
func SeverityToHTTPStatus(severity string) int {
	switch strings.ToUpper(strings.TrimSpace(severity)) {
	case string(SeverityCritical):
		return http.StatusBadRequest
	case string(SeverityHigh):
		return http.StatusBadRequest
	case string(SeverityMedium):
		return http.StatusUnprocessableEntity
	case string(SeverityLow):
		return http.StatusOK
	default:
		return http.StatusBadRequest
	}
}

func (c *ValidationChain) snapshotRules() []*Rule {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.sorted {
		sort.SliceStable(c.rules, func(i, j int) bool {
			if c.rules[i].Priority == c.rules[j].Priority {
				return c.rules[i].ID < c.rules[j].ID
			}
			return c.rules[i].Priority < c.rules[j].Priority
		})
		c.sorted = true
	}

	copied := make([]*Rule, len(c.rules))
	copy(copied, c.rules)
	return copied
}

func mergeRuleOutcome(result *ValidationResult, rule *Rule, outcome *RuleOutcome) {
	if len(outcome.Context) > 0 {
		result.Context["rule:"+rule.ID] = outcome.Context
	}

	for _, errItem := range outcome.Errors {
		if errItem.Context == nil {
			errItem.Context = map[string]interface{}{}
		}
		if _, exists := errItem.Context["ruleId"]; !exists {
			errItem.Context["ruleId"] = rule.ID
		}
		if errItem.Severity == "" {
			errItem.Severity = string(rule.Severity)
		}
		result.Errors = append(result.Errors, errItem)
	}

	for _, warnItem := range outcome.Warnings {
		result.Warnings = append(result.Warnings, warnItem)
	}
}

func normalizeSeverity(severity RuleSeverity) RuleSeverity {
	switch strings.ToUpper(strings.TrimSpace(string(severity))) {
	case string(SeverityCritical):
		return SeverityCritical
	case string(SeverityHigh):
		return SeverityHigh
	case string(SeverityMedium):
		return SeverityMedium
	case string(SeverityLow):
		return SeverityLow
	default:
		return SeverityHigh
	}
}
