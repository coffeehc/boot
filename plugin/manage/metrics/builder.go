package metrics

import (
	"fmt"
	"time"

	"git.xiagaogao.com/coffee/boot/configuration"
	"github.com/prometheus/client_golang/prometheus"
)

type CollectorOpt struct {
	Namespace   string
	Subsystem   string
	Name        string
	ConstLabels prometheus.Labels
}

func BuildConstLabels(serviceInfo *configuration.ServiceInfo, serviceAddr string) (labels prometheus.Labels) {
	if serviceInfo != nil {
		labels["service"] = serviceInfo.ServiceName
	}
	if serviceAddr != "" {
		labels["service_addr"] = serviceAddr
	}
	return
}

func NewCounterVec(opt CollectorOpt, labelNames []string) *prometheus.CounterVec {
	collector := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace:   opt.Namespace,
		Subsystem:   opt.Subsystem,
		Name:        opt.Name,
		ConstLabels: opt.ConstLabels,
		Help:        fmt.Sprintf("%s-%s-%s:%s", opt.Namespace, opt.Subsystem, opt.Name, "CounterVer"),
	}, labelNames)
	RegisterMetrics(collector)
	return collector
}

func NewGaugeVec(opt CollectorOpt, labelNames []string) *prometheus.GaugeVec {
	collector := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   opt.Namespace,
		Subsystem:   opt.Subsystem,
		Name:        opt.Name,
		ConstLabels: opt.ConstLabels,
		Help:        fmt.Sprintf("%s-%s-%s:%s", opt.Namespace, opt.Subsystem, opt.Name, "GaugeVec"),
	}, labelNames)
	RegisterMetrics(collector)
	return collector
}

func NewHistogramVec(opt CollectorOpt, buckets []float64, labelNames []string) *prometheus.HistogramVec {
	collector := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace:   opt.Namespace,
		Subsystem:   opt.Subsystem,
		Name:        opt.Name,
		ConstLabels: opt.ConstLabels,
		Help:        fmt.Sprintf("%s-%s-%s:%s", opt.Namespace, opt.Subsystem, opt.Name, "HistogramVec"),
		Buckets:     buckets,
	}, labelNames)
	RegisterMetrics(collector)
	return collector
}

func NewSummaryVec(opt CollectorOpt, maxAge time.Duration, ageBuckets uint32, bufCap uint32, labelNames []string) *prometheus.SummaryVec {
	collector := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace:   opt.Namespace,
		Subsystem:   opt.Subsystem,
		Name:        opt.Name,
		ConstLabels: opt.ConstLabels,
		Help:        fmt.Sprintf("%s-%s-%s:%s", opt.Namespace, opt.Subsystem, opt.Name, "SummaryVec"),
		MaxAge:      maxAge,
		AgeBuckets:  ageBuckets,
		BufCap:      bufCap,
	}, labelNames)
	RegisterMetrics(collector)
	return collector
}
