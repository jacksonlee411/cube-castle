package outbox

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLoadConfigDefaults(t *testing.T) {
	t.Setenv("OUTBOX_DISPATCH_INTERVAL", "")
	t.Setenv("OUTBOX_DISPATCH_BATCH_SIZE", "")
	t.Setenv("OUTBOX_DISPATCH_MAX_RETRY", "")
	t.Setenv("OUTBOX_DISPATCH_BACKOFF_BASE", "")
	t.Setenv("OUTBOX_DISPATCH_METRIC_PREFIX", "")

	cfg, err := LoadConfig()
	require.NoError(t, err)
	require.Equal(t, 5*time.Second, cfg.PollInterval)
	require.Equal(t, 50, cfg.BatchSize)
	require.Equal(t, 10, cfg.MaxRetry)
	require.Equal(t, 5*time.Second, cfg.BackoffBase)
	require.Equal(t, "outbox_dispatch", cfg.MetricNamespace)
}

func TestLoadConfigOverrides(t *testing.T) {
	t.Setenv("OUTBOX_DISPATCH_INTERVAL", "10s")
	t.Setenv("OUTBOX_DISPATCH_BATCH_SIZE", "25")
	t.Setenv("OUTBOX_DISPATCH_MAX_RETRY", "3")
	t.Setenv("OUTBOX_DISPATCH_BACKOFF_BASE", "1s")
	t.Setenv("OUTBOX_DISPATCH_METRIC_PREFIX", "custom_prefix")

	cfg, err := LoadConfig()
	require.NoError(t, err)
	require.Equal(t, 10*time.Second, cfg.PollInterval)
	require.Equal(t, 25, cfg.BatchSize)
	require.Equal(t, 3, cfg.MaxRetry)
	require.Equal(t, time.Second, cfg.BackoffBase)
	require.Equal(t, "custom_prefix", cfg.MetricNamespace)
}

func TestLoadConfigInvalid(t *testing.T) {
	t.Setenv("OUTBOX_DISPATCH_INTERVAL", "abc")
	_, err := LoadConfig()
	require.Error(t, err)

	// reset environment for subsequent tests
	t.Setenv("OUTBOX_DISPATCH_INTERVAL", "")
	t.Setenv("OUTBOX_DISPATCH_BATCH_SIZE", "not-int")
	_, err = LoadConfig()
	require.Error(t, err)

	os.Clearenv()
}
