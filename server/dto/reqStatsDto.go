package dto

import "treblle/service"

type RequestStatistics struct {
	RequestCount     int64            `json:"request_count"`
	AverageLatencyMs float64          `json:"average_latency_ms"`
	ClientErrorCount int64            `json:"client_error_count"`
	ServerErrorCount int64            `json:"server_error_count"`
	RequestsPerPath  []PathStatistics `json:"requests_per_path"`
}

// PathStatistics holds the detailed statistics grouped by path
type PathStatistics struct {
	Path             string  `json:"path"`
	RequestCount     int64   `json:"request_count"`
	AverageLatencyMs float64 `json:"average_latency_ms"`
	ClientErrorCount int64   `json:"client_error_count"`
	ServerErrorCount int64   `json:"server_error_count"`
}

// FromModel populates the RequestStatistics DTO from the service's AllRequestStatistics struct.
// It receives a pointer to modify the DTO instance directly.
func (dto *RequestStatistics) FromModel(stats *service.AllRequestStatistics) {
	if stats == nil || stats.StatsPerPath == nil {
		// Ensure it's an empty slice, not nil
		dto.RequestsPerPath = []PathStatistics{}
		return
	}
	var sum float64

	dto.RequestsPerPath = make([]PathStatistics, len(stats.StatsPerPath))
	for i, serviceStat := range stats.StatsPerPath {

		dto.ClientErrorCount += serviceStat.ClientErrorCount
		dto.ServerErrorCount += serviceStat.ServerErrorCount
		dto.RequestCount += serviceStat.RequestCount
		sum += serviceStat.AverageLatencyMs * float64(serviceStat.RequestCount)

		dto.RequestsPerPath[i] = PathStatistics{
			Path:             serviceStat.Path,
			RequestCount:     serviceStat.RequestCount,
			AverageLatencyMs: serviceStat.AverageLatencyMs, // Already in ms from the service
			ClientErrorCount: serviceStat.ClientErrorCount,
			ServerErrorCount: serviceStat.ServerErrorCount,
		}
	}
	dto.AverageLatencyMs = sum / float64(dto.RequestCount)
}
