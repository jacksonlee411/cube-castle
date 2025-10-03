package utils

import "strings"

const RootParentCode = "0000000"

var legacyRootCodes = []string{"0", RootParentCode}

// NormalizeParentCodePointer normalizes parent code input by trimming spaces and
// treating legacy root codes ("0"、"0000000") 或空值为 nil，以便服务内部统一使用 nil 表示根节点。
// 返回的指针会指向一个新的字符串副本，避免直接引用调用方的可变变量。
func NormalizeParentCodePointer(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" || IsRootParentCode(trimmed) {
		return nil
	}

	normalized := trimmed
	return &normalized
}

// IsRootParentCode 判断给定的父组织编码是否代表根节点占位符。
func IsRootParentCode(value string) bool {
	trimmed := strings.TrimSpace(value)
	for _, candidate := range legacyRootCodes {
		if trimmed == candidate {
			return true
		}
	}
	return false
}
