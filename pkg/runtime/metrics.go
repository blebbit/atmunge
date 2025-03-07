package runtime

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var lastEventTimestamp = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "plcmirror_last_op_timestamp",
	Help: "Timestamp of the last operation received from upstream.",
})

var requestCount = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "plcmirror_inbound_requests_total",
	Help: "Counter of received requests.",
}, []string{"status"})

var requestLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "plcmirror_response_latency_millisecond",
	Help:    "Latency of responses.",
	Buckets: prometheus.ExponentialBucketsRange(0.1, 30000, 20),
}, []string{"status"})
