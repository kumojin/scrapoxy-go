package collector

import "runtime"

// Collector implements the grabbing of metrics
type Collector struct {
	// Namespace is a component of the fully-qualified name of the Metric
	Namespace string

	// Subsystem is a component of the fully-qualified name of the Metric
	Subsystem string

	// EnableCPU determines whether CPU statistics will be output. Defaults to true.
	EnableCPU bool

	// EnableMem determines whether memory statistics will be output. Defaults to true.
	EnableMem bool
}

func (c *Collector) collectRuntimeInfo() Fields {
	fields := Fields{}

	fields.Goos = runtime.GOOS
	fields.Goarch = runtime.GOARCH
	fields.Version = runtime.Version()

	return fields
}

func (c *Collector) CollectStats() Fields {
	fields := Fields{}

	if c.EnableCPU {
		fields.GoroutineCount = int64(runtime.NumGoroutine())
		fields.CgoCalls = int64(runtime.NumCgoCall())
		fields.CpuCount = int64(runtime.NumCPU())
	}

	if c.EnableMem {
		m := &runtime.MemStats{}
		runtime.ReadMemStats(m)

		// General
		fields.Alloc = int64(m.Alloc)
		fields.TotalAlloc = int64(m.TotalAlloc)
		fields.Sys = int64(m.Sys)
		fields.Lookups = int64(m.Lookups)
		fields.Mallocs = int64(m.Mallocs)
		fields.Frees = int64(m.Frees)

		// Heap
		fields.HeapAlloc = int64(m.HeapAlloc)
		fields.HeapSys = int64(m.HeapSys)
		fields.HeapIdle = int64(m.HeapIdle)
		fields.HeapInuse = int64(m.HeapInuse)
		fields.HeapReleased = int64(m.HeapReleased)
		fields.HeapObjects = int64(m.HeapObjects)

		// Stack
		fields.StackInuse = int64(m.StackInuse)
		fields.StackSys = int64(m.StackSys)
		fields.MSpanInuse = int64(m.MSpanInuse)
		fields.MSpanSys = int64(m.MSpanSys)
		fields.MCacheInuse = int64(m.MCacheInuse)
		fields.MCacheSys = int64(m.MCacheSys)

		fields.OtherSys = int64(m.OtherSys)

		// Garbage Collector
		fields.GCSys = int64(m.GCSys)
		fields.NextGC = int64(m.NextGC)
		fields.LastGC = int64(m.LastGC)
		fields.PauseTotalNs = int64(m.PauseTotalNs)
		fields.PauseNs = int64(m.PauseNs[(m.NumGC+255)%256])
		fields.NumGC = int64(m.NumGC)
		fields.GCCPUFractionM = int64(float64(m.GCCPUFraction) * 1000)
	}

	fields.Goos = runtime.GOOS
	fields.Goarch = runtime.GOARCH
	fields.Version = runtime.Version()

	return fields
}
