package constraints

import (
	"fmt"
	"time"
)

// RangeWindow describes a bi-temporal window.
type RangeWindow struct {
	From time.Time
	To   *time.Time
}

// ConstraintType encodes SAP 风格的时间约束。
type ConstraintType string

const (
	ConstraintTypeTC1 ConstraintType = "TC1"
	ConstraintTypeTC2 ConstraintType = "TC2"
	ConstraintTypeTC3 ConstraintType = "TC3"
)

// ValidationResult 描述约束检查后的操作建议。
type ValidationResult struct {
	RequireContiguousValidity bool
}

// Validate 根据时间约束校验 incoming window，必要时要求收敛旧版本。
func Validate(constraint ConstraintType, existing []RangeWindow, incoming RangeWindow) (ValidationResult, error) {
	result := ValidationResult{}
	if len(existing) == 0 {
		result.RequireContiguousValidity = constraint == ConstraintTypeTC1
		return result, nil
	}
	last := existing[len(existing)-1]
	switch constraint {
	case ConstraintTypeTC1:
		if last.To != nil {
			if incoming.From.Before(*last.To) {
				return result, fmt.Errorf("tc1 overlap detected: incoming=%s lastEnd=%s", incoming.From, last.To)
			}
			if incoming.From.After(*last.To) {
				return result, fmt.Errorf("tc1 gap detected: incoming=%s lastEnd=%s", incoming.From, last.To)
			}
		}
		result.RequireContiguousValidity = true
	case ConstraintTypeTC2:
		if last.To != nil && incoming.From.Before(*last.To) {
			return result, fmt.Errorf("tc2 overlap detected: incoming=%s lastEnd=%s", incoming.From, last.To)
		}
	default:
		// TC3 允许重叠/空窗，无需额外处理
	}
	return result, nil
}

// ValidateAppendOnly 向后兼容保留 TC1 校验。
func ValidateAppendOnly(existing []RangeWindow, incoming RangeWindow) error {
	_, err := Validate(ConstraintTypeTC1, existing, incoming)
	return err
}
