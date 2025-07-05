package runtime

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var lastEventTimestamp = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "plcmirror_last_op_timestamp",
	Help: "Timestamp of the last operation received from upstream.",
})
