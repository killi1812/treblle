package model

// AllRequestStatistics holds the aggregated statistics for the requested period
type AllRequestStatistics struct {
	StatsPerPath []PathStatistics `json:"stats_per_path"`
}

// PathStatistics holds the detailed statistics grouped by path
type PathStatistics struct {
	Path             string  `json:"path"`
	RequestCount     int64   `json:"request_count"`
	AverageLatencyMs float64 `json:"average_latency_ms"`
	ClientErrorCount int64   `json:"client_error_count"`
	ServerErrorCount int64   `json:"server_error_count"`
}
