package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"proxy/collector"
	"strconv"
)

type Collector struct {
	repo    Repository
	metrics map[string]collector.MetricInfo
}

func NewCollector(repo Repository, namespace, subsystem string) Collector {
	return Collector{
		repo: repo,
		metrics: map[string]collector.MetricInfo{
			"proxy_status": collector.NewMetric(namespace, subsystem, "proxy_status", "", prometheus.GaugeValue, nil, []string{"status"}),
			"proxy_count":  collector.NewMetric(namespace, subsystem, "proxy_count", "", prometheus.GaugeValue, nil, []string{"removing"}),
		},
	}
}

func (c Collector) GetMetrics() map[string]collector.MetricInfo {
	return c.metrics
}

func (c Collector) CollectStats() map[string]map[string]collector.MetricValue {
	stats := make(map[string]map[string]collector.MetricValue)
	proxyCount := c.repo.GetProxyCountByStatus()
	for k, v := range proxyCount.Status {
		if stats["proxy_status"] == nil {
			stats["proxy_status"] = make(map[string]collector.MetricValue)
		}
		stats["proxy_status"][k] = collector.MetricValue{Value: v, Labels: []string{k}}
	}
	for k, v := range proxyCount.Removing {
		if stats["proxy_count"] == nil {
			stats["proxy_count"] = make(map[string]collector.MetricValue)
		}
		stats["proxy_count"][strconv.FormatBool(k)] = collector.MetricValue{Value: v, Labels: []string{strconv.FormatBool(k)}}
	}
	return stats
}
