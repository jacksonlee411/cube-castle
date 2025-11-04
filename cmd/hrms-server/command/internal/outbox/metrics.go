package outbox

import "github.com/prometheus/client_golang/prometheus"

type metrics struct {
	publishSuccess prometheus.Counter
	publishFailure prometheus.Counter
	retryScheduled prometheus.Counter
	activeGauge    prometheus.Gauge
}

func newMetrics(prefix string, reg prometheus.Registerer) *metrics {
	m := &metrics{
		publishSuccess: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prefix + "_success_total",
			Help: "Number of outbox events successfully published",
		}),
		publishFailure: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prefix + "_failure_total",
			Help: "Number of outbox events failed to publish",
		}),
		retryScheduled: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prefix + "_retry_total",
			Help: "Number of outbox events scheduled for retry",
		}),
		activeGauge: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: prefix + "_active",
			Help: "Indicator whether dispatcher is actively polling",
		}),
	}

	if reg != nil {
		reg.MustRegister(m.publishSuccess, m.publishFailure, m.retryScheduled, m.activeGauge)
	}

	return m
}

func (m *metrics) reset() {
	if m == nil {
		return
	}
	m.activeGauge.Set(0)
}
