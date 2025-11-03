package eventbus

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenericJSONEvent(t *testing.T) {
	payload := json.RawMessage(`{"foo":"bar"}`)
	event := NewGenericJSONEvent("organization.created", "agg-1", "organization", payload)

	require.Equal(t, "organization.created", event.EventType())
	require.Equal(t, "agg-1", event.AggregateID())
	require.Equal(t, "organization", event.AggregateType())

	out := event.Payload()
	require.JSONEq(t, string(payload), string(out))

	out[0] = '{'
	require.JSONEq(t, string(payload), string(event.Payload()), "payload should be immutable")
}
