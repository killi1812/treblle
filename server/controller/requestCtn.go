package controller

import (
	"net/http"
	"time"
	"treblle/app"
	"treblle/dto"
	"treblle/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type RequestCtn struct {
	Logger  *zap.SugaredLogger
	CrudSrv service.IRequestCrudService
}

// NewRequestCtn crates new controller with its sependencies
func NewRequestCtn() app.Controller {
	var controller *RequestCtn
	app.Invoke(func(logger *zap.SugaredLogger, service service.IRequestCrudService) {
		controller = &RequestCtn{
			Logger:  logger,
			CrudSrv: service,
		}
	})
	return controller
}

// RegisterEndpoints registers the image manipulation endpoints.
func (cnt *RequestCtn) RegisterEndpoints(router *gin.RouterGroup) {
	router.GET("/requests", cnt.ListRequests)
	router.GET("/requests/statistics", cnt.GetRequestStatistics) // Register the new endpoint
}

// ListRequests godoc
//
//	@Summary		List API requests
//	@Description	Get a paginated list of recorded API requests, with filtering and sorting.
//	@Tags			Requests
//	@Accept			json
//	@Produce		json
//	@Param			search		query		string	false	"Search term for request path"
//	@Param			method		query		string	false	"Filter by HTTP method (e.Example, GET, POST)"	enums(GET, POST, PUT, DELETE, PATCH, HEAD, OPTION, TRACE,CONNECT)
//	@Param			response	query		int		false	"Filter by response status code (e.Example, 200, 404)"
//	@Param			limit		query		int		false	"Pagination limit"	default(20)
//	@Param			offset		query		int		false	"Pagination offset"
//	@Param			sort_by		query		string	false	"Sort by field (created_at or response_time)"	enums(created_at, response_time, latency)
//	@Param			order		query		string	false	"Sort order (asc or desc)"						enums(asc, desc)
//	@Success		200			{object}	dto.ResDataDto
//	@Failure		400			{object}	dto.ErrorDto
//	@Failure		500			{object}	dto.ErrorDto
//	@Router			/requests [get]
func (cnt *RequestCtn) ListRequests(c *gin.Context) {
	var q dto.ListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		cnt.Logger.Errorf("Failed to bind query params: %v", err)
		c.JSON(http.StatusBadRequest, dto.ErrorDto{Error: "Invalid query parameters: " + err.Error()})
		return
	}

	// --- 2. Set defaults for pagination ---
	if q.Limit <= 0 {
		q.Limit = 20 // Default limit
	}
	if q.Offset < 0 {
		q.Offset = 0
	}

	params := service.ListRequestsParams{
		Limit:  q.Limit,
		Offset: q.Offset,
		SortBy: q.SortBy,
		Order:  q.Order,
	}

	// Use pointers for optional filters
	if q.Search != "" {
		params.Search = &q.Search
	}
	if q.Method != "" {
		params.Method = &q.Method
	}
	// 'response' is an int. If it's '0', it's likely not set by the user.
	// Adjust this logic if '0' is a valid response code you want to filter by.
	if q.Response != 0 {
		params.Response = &q.Response
	}

	requests, total, err := cnt.CrudSrv.List(params)
	if err != nil {
		cnt.Logger.Errorf("Service failed to list requests: %v", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorDto{Error: "Could not retrieve requests"})
		return
	}

	var reqDto = make([]dto.RequestsDto, len(requests))
	for i := range reqDto {
		reqDto[i].FromModel(requests[i])
	}

	// --- 5. Return a structured JSON response ---
	c.JSON(http.StatusOK, dto.ResDataDto{
		Data: reqDto,
		Pagination: dto.Pagination{
			Total:  total,
			Limit:  q.Limit,
			Offset: q.Offset,
		},
	})
}

// GetRequestStatistics godoc
//
//	@Summary		Get request statistics
//	@Description	Calculates statistics like average latency and error counts per path, optionally filtered by a time range.
//	@Tags			Requests
//	@Accept			json
//	@Produce		json
//	@Param			start_time	query		string					false	"Start time for filtering (RFC3339 format, e.g., 2023-10-26T00:00:00Z)"	format(date-time)
//	@Param			end_time	query		string					false	"End time for filtering (RFC3339 format, e.g., 2023-10-26T23:59:59Z)"	format(date-time)
//	@Success		200			{object}	dto.RequestStatistics	"Aggregated statistics per path"
//	@Failure		400			{object}	dto.ErrorDto
//	@Failure		500			{object}	dto.ErrorDto
//	@Router			/requests/statistics [get]
func (cnt *RequestCtn) GetRequestStatistics(c *gin.Context) {
	var startTimePtr *time.Time
	var endTimePtr *time.Time
	var err error

	// Parse optional start_time
	startTimeStr := c.Query("start_time")
	if startTimeStr != "" {
		parsedTime, err := time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			cnt.Logger.Warnf("Invalid start_time format: %v", err)
			c.JSON(http.StatusBadRequest, dto.ErrorDto{Error: "Invalid start_time format. Use RFC3339 (e.g., 2023-10-26T00:00:00Z)"})
			return
		}
		startTimePtr = &parsedTime
	}

	// Parse optional end_time
	endTimeStr := c.Query("end_time")
	if endTimeStr != "" {
		parsedTime, err := time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			cnt.Logger.Warnf("Invalid end_time format: %v", err)
			c.JSON(http.StatusBadRequest, dto.ErrorDto{Error: "Invalid end_time format. Use RFC3339 (e.g., 2023-10-26T23:59:59Z)"})
			return
		}
		endTimePtr = &parsedTime
	}

	// Validate time range if both are provided
	if startTimePtr != nil && endTimePtr != nil && startTimePtr.After(*endTimePtr) {
		c.JSON(http.StatusBadRequest, dto.ErrorDto{Error: "start_time cannot be after end_time"})
		return
	}

	// Call the service
	stats, err := cnt.CrudSrv.GetStatistics(startTimePtr, endTimePtr)
	if err != nil {
		cnt.Logger.Errorf("Service failed to get request statistics: %v", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorDto{Error: "Could not retrieve request statistics"})
		return
	}
	var ret dto.RequestStatistics
	ret.FromModel(stats)

	// Return the statistics
	c.JSON(http.StatusOK, ret)
}
