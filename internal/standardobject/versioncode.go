package standardobject

import (
	"fmt"
	"strings"
	"time"
)

const (
	versionCodeDateLayout = "20060102"
	versionCodeTimeLayout = "150405.000Z0700"
	maxEntropyLength      = 8
)

// MakeVersionCode builds a deterministic identifier for a temporal version.
// The format follows {CODE}-{YYYYMMDD}-{HHMMSS.mmmZ0700}-{entropy}, where the
// final entropy segment is optional. This keeps version codes stable while
// still allowing multiple corrections for the same effective date.
func MakeVersionCode(code string, effective time.Time, updatedAt time.Time, entropy string) string {
	baseCode := strings.ToUpper(strings.TrimSpace(code))
	if baseCode == "" {
		baseCode = "UNKNOWN"
	}
	effectiveUTC := coalesceTime(effective, time.Now()).UTC()
	updatedUTC := coalesceTime(updatedAt, time.Now()).UTC()

	base := fmt.Sprintf("%s-%s", baseCode, effectiveUTC.Format(versionCodeDateLayout))
	timestamp := updatedUTC.Format(versionCodeTimeLayout)
	token := normalizeEntropy(entropy)

	if token == "" {
		return fmt.Sprintf("%s-%s", base, timestamp)
	}
	return fmt.Sprintf("%s-%s-%s", base, timestamp, token)
}

func coalesceTime(value time.Time, fallback time.Time) time.Time {
	if value.IsZero() {
		return fallback
	}
	return value
}

func normalizeEntropy(value string) string {
	clean := strings.ToUpper(strings.TrimSpace(strings.ReplaceAll(value, "-", "")))
	if clean == "" {
		return ""
	}
	if len(clean) > maxEntropyLength {
		return clean[:maxEntropyLength]
	}
	return clean
}
