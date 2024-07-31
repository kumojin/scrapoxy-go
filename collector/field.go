package collector

type Fields struct {
	// CPU
	CpuCount       int64 `json:"cpu.count"`
	GoroutineCount int64 `json:"cpu.goroutines"`
	CgoCalls       int64 `json:"cpu.cgo_calls"`

	// General
	Alloc      int64 `json:"mem.alloc"`
	TotalAlloc int64 `json:"mem.total"`
	Sys        int64 `json:"mem.sys"`
	Lookups    int64 `json:"mem.lookups"`
	Mallocs    int64 `json:"mem.malloc"`
	Frees      int64 `json:"mem.frees"`

	// Heap
	HeapAlloc    int64 `json:"mem.heap.alloc"`
	HeapSys      int64 `json:"mem.heap.sys"`
	HeapIdle     int64 `json:"mem.heap.idle"`
	HeapInuse    int64 `json:"mem.heap.inuse"`
	HeapReleased int64 `json:"mem.heap.released"`
	HeapObjects  int64 `json:"mem.heap.objects"`

	// Stack
	StackInuse  int64 `json:"mem.stack.inuse"`
	StackSys    int64 `json:"mem.stack.sys"`
	MSpanInuse  int64 `json:"mem.stack.mspan_inuse"`
	MSpanSys    int64 `json:"mem.stack.mspan_sys"`
	MCacheInuse int64 `json:"mem.stack.mcache_inuse"`
	MCacheSys   int64 `json:"mem.stack.mcache_sys"`

	OtherSys int64 `json:"mem.othersys"`

	// GC
	GCSys          int64 `json:"mem.gc.sys"`
	NextGC         int64 `json:"mem.gc.next"`
	LastGC         int64 `json:"mem.gc.last"`
	PauseTotalNs   int64 `json:"mem.gc.pause_total"`
	PauseNs        int64 `json:"mem.gc.pause"`
	NumGC          int64 `json:"mem.gc.count"`
	GCCPUFractionM int64 `json:"mem.gc.cpu_fraction_m"`

	Goarch  string `json:"-"`
	Goos    string `json:"-"`
	Version string `json:"-"`
}

func (f Fields) ToMap() map[string]int64 {
	return map[string]int64{
		"CpuCount":       f.CpuCount,
		"GoroutineCount": f.GoroutineCount,
		"CgoCalls":       f.CgoCalls,

		"Alloc":      f.Alloc,
		"TotalAlloc": f.TotalAlloc,
		"Sys":        f.Sys,
		"Lookups":    f.Lookups,
		"Mallocs":    f.Mallocs,
		"Frees":      f.Frees,

		"HeapAlloc":    f.HeapAlloc,
		"HeapSys":      f.HeapSys,
		"HeapIdle":     f.HeapIdle,
		"HeapInuse":    f.HeapInuse,
		"HeapReleased": f.HeapReleased,
		"HeapObjects":  f.HeapObjects,

		"StackInuse":  f.StackInuse,
		"StackSys":    f.StackSys,
		"MSpanInuse":  f.MSpanInuse,
		"MSpanSys":    f.MSpanSys,
		"MCacheInuse": f.MCacheInuse,
		"MCacheSys":   f.MCacheSys,

		"OtherSys": f.OtherSys,

		// GC
		"GCSys":          f.GCSys,
		"NextGC":         f.NextGC,
		"LastGC":         f.LastGC,
		"PauseTotalNs":   f.PauseTotalNs,
		"PauseNs":        f.PauseNs,
		"NumGC":          f.NumGC,
		"GCCPUFractionM": f.GCCPUFractionM,
	}
}
