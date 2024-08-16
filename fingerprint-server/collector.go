package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"proxy/collector"
)

type Collector struct {
	metrics map[string]collector.MetricInfo
}

func NewCollector(namespace, subsystem string) Collector {
	return Collector{
		metrics: map[string]collector.MetricInfo{
			"requests_count": collector.NewMetric(namespace, subsystem, "requests_count", "", prometheus.CounterValue, nil, []string{}),
		},
	}
}

func (c Collector) GetMetrics() map[string]collector.MetricInfo {
	return c.metrics
}

func (c Collector) CollectStats() map[string]map[string]collector.MetricValue {
	stats := make(map[string]map[string]collector.MetricValue)

	stats["requests_count"] = map[string]collector.MetricValue{"default": collector.MetricValue{Collector: requestCounter.Collect}}

	return stats
}
