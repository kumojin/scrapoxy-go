package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"sync"
)

type MetricInfo struct {
	Desc              *prometheus.Desc
	Type              prometheus.ValueType
	HasDynamicsLabels bool
}
type metrics map[string]MetricInfo

func NewMetric(namespace, subsystem, metricName string, docString string, t prometheus.ValueType, constLabels prometheus.Labels, varLabels []string) MetricInfo {
	return MetricInfo{
		Desc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, metricName),
			docString,
			varLabels,
			constLabels,
		),
		Type:              t,
		HasDynamicsLabels: len(varLabels) > 0,
	}
}

type MetricValue struct {
	Value     int64
	Labels    []string
	Collector func(chan<- prometheus.Metric)
}
type ExtraCollector interface {
	CollectStats() map[string]map[string]MetricValue
	GetMetrics() map[string]MetricInfo
}
type PrometheusMetricsCollector struct {
	collector     Collector
	serverMetrics map[string]MetricInfo

	extraCollector ExtraCollector
	mutex          sync.RWMutex
}

func NewPrometheusMetrics(c Collector, extraCollector ExtraCollector) *PrometheusMetricsCollector {
	runtimeInfo := c.collectRuntimeInfo()
	labels := prometheus.Labels{
		"go_arch":    runtimeInfo.Goarch,
		"go_os":      runtimeInfo.Goos,
		"go_version": runtimeInfo.Version,
	}
	prometheusMetrics := &PrometheusMetricsCollector{
		collector: c,
		serverMetrics: metrics{
			"CpuCount":       NewMetric(c.Namespace, c.Subsystem, "cpu_count", "", prometheus.GaugeValue, labels, []string{}),
			"GoroutineCount": NewMetric(c.Namespace, c.Subsystem, "goroutine_count", "", prometheus.GaugeValue, labels, []string{}),
			"CgoCalls":       NewMetric(c.Namespace, c.Subsystem, "cgo_calls", "", prometheus.GaugeValue, labels, []string{}),

			"Alloc":      NewMetric(c.Namespace, c.Subsystem, "mem_alloc", "", prometheus.GaugeValue, labels, []string{}),
			"TotalAlloc": NewMetric(c.Namespace, c.Subsystem, "mem_total_alloc", "", prometheus.GaugeValue, labels, []string{}),
			"Sys":        NewMetric(c.Namespace, c.Subsystem, "mem_sys", "", prometheus.GaugeValue, labels, []string{}),
			"OtherSys":   NewMetric(c.Namespace, c.Subsystem, "mem_othersys", "", prometheus.GaugeValue, labels, []string{}),
			"Lookups":    NewMetric(c.Namespace, c.Subsystem, "mem_lookups", "", prometheus.GaugeValue, labels, []string{}),
			"Mallocs":    NewMetric(c.Namespace, c.Subsystem, "mem_malloc", "", prometheus.GaugeValue, labels, []string{}),
			"Frees":      NewMetric(c.Namespace, c.Subsystem, "mem_frees", "", prometheus.GaugeValue, labels, []string{}),

			"HeapAlloc":    NewMetric(c.Namespace, c.Subsystem, "mem_heap_alloc", "", prometheus.GaugeValue, labels, []string{}),
			"HeapSys":      NewMetric(c.Namespace, c.Subsystem, "mem_heap_sys", "", prometheus.GaugeValue, labels, []string{}),
			"HeapIdle":     NewMetric(c.Namespace, c.Subsystem, "mem_heap_idle", "", prometheus.GaugeValue, labels, []string{}),
			"HeapInuse":    NewMetric(c.Namespace, c.Subsystem, "mem_heap_inuse", "", prometheus.GaugeValue, labels, []string{}),
			"HeapReleased": NewMetric(c.Namespace, c.Subsystem, "mem_heap_released", "", prometheus.GaugeValue, labels, []string{}),
			"HeapObjects":  NewMetric(c.Namespace, c.Subsystem, "mem_heap_objects", "", prometheus.GaugeValue, labels, []string{}),

			"StackInuse":  NewMetric(c.Namespace, c.Subsystem, "mem_stack_inuse", "", prometheus.GaugeValue, labels, []string{}),
			"StackSys":    NewMetric(c.Namespace, c.Subsystem, "mem_stack_sys", "", prometheus.GaugeValue, labels, []string{}),
			"MSpanInuse":  NewMetric(c.Namespace, c.Subsystem, "mem_stack_mspan_inuse", "", prometheus.GaugeValue, labels, []string{}),
			"MSpanSys":    NewMetric(c.Namespace, c.Subsystem, "mem_stack_mspan_sys", "", prometheus.GaugeValue, labels, []string{}),
			"MCacheInuse": NewMetric(c.Namespace, c.Subsystem, "mem_stack_mcache_inuse", "", prometheus.GaugeValue, labels, []string{}),
			"MCacheSys":   NewMetric(c.Namespace, c.Subsystem, "mem_stack_mcache_sys", "", prometheus.GaugeValue, labels, []string{}),

			"GCSys":          NewMetric(c.Namespace, c.Subsystem, "mem_gc_sys", "", prometheus.GaugeValue, labels, []string{}),
			"NextGC":         NewMetric(c.Namespace, c.Subsystem, "mem_gc_next", "", prometheus.GaugeValue, labels, []string{}),
			"LastGC":         NewMetric(c.Namespace, c.Subsystem, "mem_gc_last", "", prometheus.GaugeValue, labels, []string{}),
			"PauseTotalNs":   NewMetric(c.Namespace, c.Subsystem, "mem_gc_pause_total", "", prometheus.GaugeValue, labels, []string{}),
			"PauseNs":        NewMetric(c.Namespace, c.Subsystem, "mem_gc_pause", "", prometheus.GaugeValue, labels, []string{}),
			"NumGC":          NewMetric(c.Namespace, c.Subsystem, "mem_gc_count", "", prometheus.GaugeValue, labels, []string{}),
			"GCCPUFractionM": NewMetric(c.Namespace, c.Subsystem, "mem_gc_cpu_fraction", "", prometheus.GaugeValue, labels, []string{}),
		},
		extraCollector: extraCollector,
	}
	prometheus.MustRegister()
	return prometheusMetrics
}

func (c *PrometheusMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range c.serverMetrics {
		ch <- m.Desc
	}

	for _, m := range c.extraCollector.GetMetrics() {
		ch <- m.Desc
	}
}

func (c *PrometheusMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock() // To protect metrics from concurrent collects.
	defer c.mutex.Unlock()

	stats := c.collector.CollectStats()
	for k, v := range stats.ToMap() {
		metric := c.serverMetrics[k]
		ch <- prometheus.MustNewConstMetric(metric.Desc, metric.Type, float64(v))
	}

	if c.extraCollector != nil {
		extraStats := c.extraCollector.CollectStats()
		extraMetrics := c.extraCollector.GetMetrics()
		for k, vals := range extraStats {
			metric := extraMetrics[k]
			if metric.HasDynamicsLabels {
				for _, v := range vals {
					if v.Collector != nil {
						v.Collector(ch)
					} else {
						ch <- prometheus.MustNewConstMetric(metric.Desc, metric.Type, float64(v.Value), v.Labels...)
					}
				}
			} else {
				v := vals["default"]
				if v.Collector != nil {
					v.Collector(ch)
				} else {
					ch <- prometheus.MustNewConstMetric(metric.Desc, metric.Type, float64(v.Value))
				}
			}

		}
	}
}
