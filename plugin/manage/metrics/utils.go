package metrics

import "github.com/prometheus/client_golang/prometheus"

func RegisterMetrics(c prometheus.Collector) error {
	return prometheus.Register(c)
}

func UnregisterMetrics(c prometheus.Collector) bool {
	return prometheus.Unregister(c)
}
