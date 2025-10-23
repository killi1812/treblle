package service

import (
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

type RequestCrudService struct {
	db     *gorm.DB
	logger *zap.SugaredLogger
}

type IRequestCrudService interface {
	List(params ListRequestsParams) ([]model.Request, int64, error)
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
