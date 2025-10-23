package controller

import (
	"net/http"
	"treblle/app"
	"treblle/dto"
	"treblle/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type RequestCtn struct {
	logger  *zap.SugaredLogger
	crudSrv *service.RequestCrudService
}

// NewRequestCtn crates new controller with its sependencies
func NewRequestCtn() app.Controller {
	var controller *RequestCtn
	app.Invoke(func(logger *zap.SugaredLogger, service *service.RequestCrudService) {
		controller = &RequestCtn{
			logger:  logger,
			crudSrv: service,
		}
	})
	return controller
}

// RegisterEndpoints registers the image manipulation endpoints.
func (cnt *RequestCtn) RegisterEndpoints(router *gin.RouterGroup) {
	router.GET("/requests", cnt.ListRequests)
}

// ListRequests godoc
//
//	@Summary		List API requests
//	@Description	Get a paginated list of recorded API requests, with filtering and sorting.
//	@Tags			Requests
//	@Accept			json
//	@Produce		json
//	@Param			search		query		string	false	"Search term for request path"
//	@Param			method		query		string	false	"Filter by HTTP method (e.Example, GET, POST)"	enums(GET, POST, PUT, DELETE, PATCH)
//	@Param			response	query		int		false	"Filter by response status code (e.Example, 200, 404)"
//	@Param			limit		query		int		false	"Pagination limit"	default(20)
//	@Param			offset		query		int		false	"Pagination offset"
//	@Param			sort_by		query		string	false	"Sort by field (created_at or response_time)"	enums(created_at, response_time)
//	@Param			order		query		string	false	"Sort order (asc or desc)"						enums(asc, desc)
//	@Success		200			{object}	dto.ResDataDto
//	@Failure		400			{object}	dto.ErrorDto
//	@Failure		500			{object}	dto.ErrorDto
//	@Router			/requests [get]
func (cnt *RequestCtn) ListRequests(c *gin.Context) {
	var q dto.ListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		cnt.logger.Errorf("Failed to bind query params: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters: " + err.Error()})
		return
	}

	// --- 2. Set defaults for pagination ---
	if q.Limit <= 0 {
		q.Limit = 20 // Default limit
	}
	if q.Offset < 0 {
		q.Offset = 0
	}

	// --- 3. Map query params to the service params struct ---
	// This layer of mapping is good practice.
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

	// --- 4. Call the service ---
	requests, total, err := cnt.crudSrv.List(params)
	if err != nil {
		cnt.logger.Errorf("Service failed to list requests: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve requests"})
		return
	}

	// --- 5. Return a structured JSON response ---
	c.JSON(http.StatusOK, dto.ResDataDto{
		Data: requests,
		Pagination: dto.Pagination{
			Total:  total,
			Limit:  q.Limit,
			Offset: q.Offset,
		},
	})
}
