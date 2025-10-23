package service

import (
	"errors"
	"sort"
	"strings"
	"time"
	"treblle/app"
	"treblle/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ListRequestsParams struct {
	Search   *string // Search term for 'path' field
	Method   *string // Filter by method (e.g., "GET")
	Response *int    // Filter by response code (e.g., 404)

	// Pagination
	Limit  int
	Offset int

	// Sorting
	SortBy string // "created_at" or "response_time" or "latency"
	Order  string // "asc" or "desc"
}

// PathStatistics holds the detailed statistics grouped by path
type PathStatistics struct {
	Path             string  `json:"path"`
	RequestCount     int64   `json:"request_count"`
	AverageLatencyMs float64 `json:"average_latency_ms"`
	ClientErrorCount int64   `json:"client_error_count"`
	ServerErrorCount int64   `json:"server_error_count"`
}

// Result struct specifically for the GORM Scan operation
type pathStatsQueryResult struct {
	Path             string
	RequestCount     int64
	AvgLatencyNanos  float64
	ClientErrorCount int64
	ServerErrorCount int64
}

// AllRequestStatistics holds the aggregated statistics for the requested period
type AllRequestStatistics struct {
	StatsPerPath []PathStatistics `json:"stats_per_path"`
}

type RequestCrudService struct {
	db     *gorm.DB
	logger *zap.SugaredLogger
}

type IRequestCrudService interface {
	List(params ListRequestsParams) ([]model.Request, int64, error)
	GetStatistics(startTime, endTime *time.Time) (*AllRequestStatistics, error)
}

// NewRequestCRUDService is your constructor from the snippet.
func NewRequestCrudService() IRequestCrudService {
	var service *RequestCrudService

	app.Invoke(func(db *gorm.DB, logger *zap.SugaredLogger) {
		service = &RequestCrudService{
			db:     db,
			logger: logger,
		}
	})

	return service
}

// List returns a paginated list of requests based on filter and search parameters.
// It also returns the total count of records that match the query (before pagination).
func (s *RequestCrudService) List(params ListRequestsParams) ([]model.Request, int64, error) {
	var requests []model.Request
	var total int64

	query := s.db.Model(&model.Request{})

	// --- Apply Search ---
	if params.Search != nil && *params.Search != "" {
		query = query.Where("path LIKE ?", "%"+*params.Search+"%")
	}

	if params.Method != nil && *params.Method != "" {
		query = query.Where("method = ?", *params.Method)
	}
	if params.Response != nil {
		// Use a pointer to allow filtering for '0', though '0' is not a real HTTP status.
		query = query.Where("response = ?", *params.Response)
	}

	// --- Get Total Count (before pagination) ---
	// This is crucial for the UI to know how many pages there are.
	err := query.Count(&total).Error
	if err != nil {
		s.logger.Errorf("Failed to count requests: %v", err)
		return nil, 0, err
	}

	// --- Apply Sorting ---
	if params.SortBy != "" {
		order := "asc" // default order
		if params.Order == "desc" {
			order = "desc"
		}

		var column string
		switch params.SortBy {
		case "created_at":
			column = "created_at"
		case "response_time":
			column = "response_time"
		case "latency":
			column = "latency"

		}
		query = query.Order(column + " " + order)
	} else {
		// Default sort if nothing is provided
		query = query.Order("created_at desc")
	}

	// --- Apply Pagination ---
	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}
	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}

	// --- Execute Query ---
	err = query.Find(&requests).Error
	if err != nil {
		s.logger.Errorf("Failed to get requests from DB: %v", err)
		return nil, 0, err
	}

	return requests, total, nil
}
func (s *RequestCrudService) GetStatistics(startTime, endTime *time.Time) (*AllRequestStatistics, error) {
	var results []pathStatsQueryResult // Use the intermediate struct for scanning
	var allStats AllRequestStatistics

	query := s.db.Model(&model.Request{})

	// Apply time range filter if provided
	if startTime != nil {
		query = query.Where("created_at >= ?", *startTime)
	}
	if endTime != nil {
		query = query.Where("created_at <= ?", *endTime)
	}

	// Select path and aggregated statistics
	// Using SUM with CASE WHEN (or equivalent) for conditional counting
	query = query.Select(`
		path,
		count(*) as request_count,
		avg(latency) as avg_latency_nanos,
		sum(case when response >= 400 and response < 500 then 1 else 0 end) as client_error_count,
		sum(case when response >= 500 then 1 else 0 end) as server_error_count
	`).Group("path").Order("request_count desc") // Order by most frequent paths first

	err := query.Scan(&results).Error
	if err != nil {
		// Handle potential "Scan error... converting NULL to float64" specifically if needed,
		// though with counts, this is less likely than with just AVG.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// No records found in the time range, return empty stats
			return &allStats, nil // Return empty, not nil error
		}
		s.logger.Errorf("Failed to calculate statistics per path: %v", err)
		return nil, err
	}

	// Convert intermediate results to the final structure
	allStats.StatsPerPath = make([]PathStatistics, len(results))
	for i, res := range results {
		allStats.StatsPerPath[i] = PathStatistics{
			Path:             res.Path,
			RequestCount:     res.RequestCount,
			AverageLatencyMs: res.AvgLatencyNanos / float64(time.Millisecond), // Convert ns to ms
			ClientErrorCount: res.ClientErrorCount,
			ServerErrorCount: res.ServerErrorCount,
		}
	}

	cleanedStatsMap := make(map[string]PathStatistics)
	for _, pathStat := range allStats.StatsPerPath {
		basePath, _, _ := strings.Cut(pathStat.Path, "?")
		existing, ok := cleanedStatsMap[basePath]
		if !ok {
			existing = PathStatistics{Path: basePath}
		}

		// Aggregate counts
		existing.RequestCount += pathStat.RequestCount
		existing.ClientErrorCount += pathStat.ClientErrorCount
		existing.ServerErrorCount += pathStat.ServerErrorCount

		// Calculate weighted average for latency
		// Avoid division by zero if RequestCount is somehow 0
		if existing.RequestCount > 0 {
			totalLatencyMs := (existing.AverageLatencyMs * float64(existing.RequestCount-pathStat.RequestCount)) + (pathStat.AverageLatencyMs * float64(pathStat.RequestCount))
			existing.AverageLatencyMs = totalLatencyMs / float64(existing.RequestCount)
		} else {
			existing.AverageLatencyMs = 0 // Or handle as appropriate
		}

		cleanedStatsMap[basePath] = existing
	}

	// Convert map back to slice
	cleanedSlice := make([]PathStatistics, 0, len(cleanedStatsMap))
	for _, stat := range cleanedStatsMap {
		cleanedSlice = append(cleanedSlice, stat)
	}

	// Optionally re-sort the cleaned slice (e.g., by RequestCount descending)
	sort.Slice(cleanedSlice, func(i, j int) bool {
		return cleanedSlice[i].RequestCount > cleanedSlice[j].RequestCount
	})

	allStats.StatsPerPath = cleanedSlice // Replace original slice with cleaned one

	return &allStats, nil
}
