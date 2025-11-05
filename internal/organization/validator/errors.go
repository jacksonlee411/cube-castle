package validator

import (
	"fmt"
)

// ValidationFailedError 表示业务规则校验未通过，并携带完整的 ValidationResult。
type ValidationFailedError struct {
	operation string
	result    *ValidationResult
}

// NewValidationFailedError 构造校验失败错误，operation 为调用方上下文（如命令名称）。
func NewValidationFailedError(operation string, result *ValidationResult) error {
	if result == nil {
		return fmt.Errorf("validator: validation failed without result for %s", operation)
	}
	return &ValidationFailedError{
		operation: operation,
		result:    result,
	}
}

func (e *ValidationFailedError) Error() string {
	if e.result == nil {
		return fmt.Sprintf("validation failed: %s (no details)", e.operation)
	}
	if len(e.result.Errors) > 0 {
		return fmt.Sprintf("validation failed: %s (%s)", e.operation, e.result.Errors[0].Code)
	}
	return fmt.Sprintf("validation failed: %s", e.operation)
}

// Result 返回完整的验证结果。
func (e *ValidationFailedError) Result() *ValidationResult {
	return e.result
}

// Operation 返回触发验证的操作名称。
func (e *ValidationFailedError) Operation() string {
	return e.operation
}
