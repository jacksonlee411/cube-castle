package validator

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	ruleOutcomeLabelSuccess = "success"
	ruleOutcomeLabelWarning = "warning"
	ruleOutcomeLabelFailed  = "failed"
	ruleOutcomeLabelError   = "error"

	chainOutcomeLabelSuccess   = "success"
	chainOutcomeLabelFailed    = "failed"
	chainOutcomeLabelCancelled = "cancelled"
)

var (
	metricsOnce sync.Once

	validatorRuleDurationSeconds *prometheus.HistogramVec
	validatorRuleOutcomeTotal    *prometheus.CounterVec
	validatorChainDuration       *prometheus.HistogramVec
	validatorChainOutcomeTotal   *prometheus.CounterVec
)

func ensureValidatorMetricsRegistered() {
	metricsOnce.Do(func() {
		validatorRuleDurationSeconds = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "validator_rule_duration_seconds",
				Help:    "Histogram of validation rule execution durations.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"rule_id"},
		)

		validatorRuleOutcomeTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "validator_rule_outcome_total",
				Help: "Total number of validation rule executions grouped by outcome.",
			},
			[]string{"rule_id", "outcome"},
		)

		validatorChainDuration = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "validator_chain_duration_seconds",
				Help:    "Histogram of validation chain execution durations.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"operation"},
		)

		validatorChainOutcomeTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "validator_chain_outcome_total",
				Help: "Total number of validation chain executions grouped by outcome.",
			},
			[]string{"operation", "outcome"},
		)

		prometheus.MustRegister(
			validatorRuleDurationSeconds,
			validatorRuleOutcomeTotal,
			validatorChainDuration,
			validatorChainOutcomeTotal,
		)
	})
}

func observeRuleMetrics(ruleID, outcome string, duration time.Duration) {
	if ruleID == "" {
		ruleID = "UNKNOWN"
	}
	if outcome == "" {
		outcome = ruleOutcomeLabelSuccess
	}

	ensureValidatorMetricsRegistered()
	validatorRuleDurationSeconds.WithLabelValues(ruleID).Observe(duration.Seconds())
	validatorRuleOutcomeTotal.WithLabelValues(ruleID, outcome).Inc()
}

func observeChainMetrics(operation, outcome string, duration time.Duration) {
	if operation == "" {
		operation = "unknown"
	}
	if outcome == "" {
		outcome = chainOutcomeLabelSuccess
	}

	ensureValidatorMetricsRegistered()
	validatorChainDuration.WithLabelValues(operation).Observe(duration.Seconds())
	validatorChainOutcomeTotal.WithLabelValues(operation, outcome).Inc()
}
