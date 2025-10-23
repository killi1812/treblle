package controller_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"treblle/controller"
	"treblle/dto"
	"treblle/model"
	"treblle/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// --- Mock RequestCrudService ---
// We only need to mock the methods used by RequestCtn, which is List in this case.
type MockRequestCrudService struct {
	mock.Mock
}

func (m *MockRequestCrudService) List(params service.ListRequestsParams) ([]model.Request, int64, error) {
	args := m.Called(params)
	// Handle potential nil return for the slice
	var requests []model.Request
	if args.Get(0) != nil {
		requests = args.Get(0).([]model.Request)
	}
	return requests, args.Get(1).(int64), args.Error(2)
}

func (m *MockRequestCrudService) GetStatistics(start, end *time.Time) (*service.AllRequestStatistics, error) {
	args := m.Called(start, end)
	// Handle potential nil return for the slice
	var requests service.AllRequestStatistics
	if args.Get(0) != nil {
		requests = args.Get(0).(service.AllRequestStatistics)
	}
	return &requests, args.Error(1)
}

// --- RequestController Test Suite ---
type RequestControllerTestSuite struct {
	suite.Suite
	router                 *gin.Engine
	mockRequestCrudService *MockRequestCrudService
	logger                 *zap.SugaredLogger
	logObserver            *observer.ObservedLogs
}

// SetupSuite runs once before all tests in the suite
func (suite *RequestControllerTestSuite) SetupSuite() {
	core, obs := observer.New(zap.InfoLevel)
	suite.logger = zap.New(core).Sugar()
	suite.logObserver = obs
	zap.ReplaceGlobals(zap.New(core)) // Set the global logger
	gin.SetMode(gin.TestMode)

	// Although RequestCtn doesn't directly use config, setting it up might be
	// necessary if underlying app initialization depends on it.
	// config.AppConfig = &config.AppConfiguration{ /* ... */ }

	suite.mockRequestCrudService = new(MockRequestCrudService)

	suite.router = gin.Default()
	apiGroup := suite.router.Group("/api") // Match the group used in your actual setup

	requestCtrl := controller.RequestCtn{
		Logger:  suite.logger,
		CrudSrv: suite.mockRequestCrudService,
	}
	requestCtrl.RegisterEndpoints(apiGroup)
}

// TearDownSuite runs once after all tests
func (suite *RequestControllerTestSuite) TearDownSuite() {
	if suite.logger != nil {
		_ = suite.logger.Sync()
	}
}

// SetupTest runs before each test
func (suite *RequestControllerTestSuite) SetupTest() {
	// Reset mocks before each test
	suite.mockRequestCrudService.ExpectedCalls = nil
	suite.mockRequestCrudService.Calls = nil
}

// TestRequestController runs the test suite
func TestRequestController(t *testing.T) {
	suite.Run(t, new(RequestControllerTestSuite))
}

// --- Test Cases ---

func (suite *RequestControllerTestSuite) TestListRequests_Success_Defaults() {
	// Arrange
	mockRequests := []model.Request{
		{ID: 1, Method: "GET", Path: "/api/test1", Response: 200, CreatedAt: time.Now().Add(-1 * time.Hour), ResponseTime: time.Now().Add(-1 * time.Hour).Add(50 * time.Millisecond), Latency: 50 * time.Millisecond},
		{ID: 2, Method: "POST", Path: "/api/test2", Response: 201, CreatedAt: time.Now(), ResponseTime: time.Now().Add(100 * time.Millisecond), Latency: 100 * time.Millisecond},
	}
	mockTotal := int64(50) // Example total count

	// Expect the service's List method to be called with default params
	expectedParams := service.ListRequestsParams{
		Limit:  20, // Default limit from controller
		Offset: 0,  // Default offset
		// SortBy and Order are empty, service should apply default sorting
	}
	suite.mockRequestCrudService.On("List", expectedParams).Return(mockRequests, mockTotal, nil).Once()

	// Act
	req, _ := http.NewRequest(http.MethodGet, "/api/requests", nil) // No query params, use defaults
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var responseDto dto.ResDataDto
	err := json.Unmarshal(w.Body.Bytes(), &responseDto)
	assert.NoError(suite.T(), err)

	assert.Len(suite.T(), responseDto.Data, 2)
	assert.Equal(suite.T(), uint(1), responseDto.Data[0].ID) // Check if conversion is correct
	assert.Equal(suite.T(), "GET", responseDto.Data[0].Method)
	assert.Equal(suite.T(), int64(50), responseDto.Data[0].Latency)

	assert.Equal(suite.T(), mockTotal, responseDto.Pagination.Total)
	assert.Equal(suite.T(), 20, responseDto.Pagination.Limit) // Default limit
	assert.Equal(suite.T(), 0, responseDto.Pagination.Offset) // Default offset

	suite.mockRequestCrudService.AssertExpectations(suite.T())
}

func (suite *RequestControllerTestSuite) TestListRequests_Success_WithParams() {
	// Arrange
	mockRequests := []model.Request{
		{ID: 3, Method: "PUT", Path: "/api/update/item", Response: 200, CreatedAt: time.Now(), ResponseTime: time.Now().Add(75 * time.Millisecond), Latency: 75 * time.Millisecond},
	}
	mockTotal := int64(1)
	search := "update"
	method := "PUT"
	responseCode := 200

	// Expect the service's List method to be called with specific params
	expectedParams := service.ListRequestsParams{
		Search:   &search,
		Method:   &method,
		Response: &responseCode,
		Limit:    10,
		Offset:   5,
		SortBy:   "response_time",
		Order:    "asc",
	}
	suite.mockRequestCrudService.On("List", expectedParams).Return(mockRequests, mockTotal, nil).Once()

	// Act
	reqUrl := "/api/requests?search=update&method=PUT&response=200&limit=10&offset=5&sort_by=response_time&order=asc"
	req, _ := http.NewRequest(http.MethodGet, reqUrl, nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var responseDto dto.ResDataDto
	err := json.Unmarshal(w.Body.Bytes(), &responseDto)
	assert.NoError(suite.T(), err)

	assert.Len(suite.T(), responseDto.Data, 1)
	assert.Equal(suite.T(), uint(3), responseDto.Data[0].ID)
	assert.Equal(suite.T(), "PUT", responseDto.Data[0].Method)

	assert.Equal(suite.T(), mockTotal, responseDto.Pagination.Total)
	assert.Equal(suite.T(), 10, responseDto.Pagination.Limit)
	assert.Equal(suite.T(), 5, responseDto.Pagination.Offset)

	suite.mockRequestCrudService.AssertExpectations(suite.T())
}

func (suite *RequestControllerTestSuite) TestListRequests_BindingError() {
	// Act
	// Invalid limit value
	req, _ := http.NewRequest(http.MethodGet, "/api/requests?limit=abc", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var errorDto dto.ErrorDto
	err := json.Unmarshal(w.Body.Bytes(), &errorDto)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), errorDto.Error, "Invalid query parameters")

	// Ensure the service method was NOT called
	suite.mockRequestCrudService.AssertNotCalled(suite.T(), "List", mock.Anything)
}

func (suite *RequestControllerTestSuite) TestListRequests_ServiceError() {
	// Arrange
	expectedError := errors.New("database connection failed")
	// Expect List to be called but return an error
	suite.mockRequestCrudService.On("List", mock.AnythingOfType("service.ListRequestsParams")).Return(nil, int64(0), expectedError).Once()

	// Act
	req, _ := http.NewRequest(http.MethodGet, "/api/requests", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)

	var errorDto dto.ErrorDto
	err := json.Unmarshal(w.Body.Bytes(), &errorDto)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Could not retrieve requests", errorDto.Error)

	suite.mockRequestCrudService.AssertExpectations(suite.T())
}
func (suite *RequestControllerTestSuite) TestListRequests_Success_BadOffset() {
	// Arrange
	mockRequests := []model.Request{
		{ID: 1, Method: "GET", Path: "/api/test1?offset=-1", Response: 200, CreatedAt: time.Now().Add(-1 * time.Hour), ResponseTime: time.Now().Add(-1 * time.Hour).Add(50 * time.Millisecond), Latency: 50 * time.Millisecond},
	}
	mockTotal := int64(50) // Example total count

	// Expect the service's List method to be called with default params
	expectedParams := service.ListRequestsParams{
		Limit:  20,
		Offset: 0,
	}
	suite.mockRequestCrudService.On("List", expectedParams).Return(mockRequests, mockTotal, nil).Once()

	// Act
	req, _ := http.NewRequest(http.MethodGet, "/api/requests", nil) // No query params, use defaults
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var responseDto dto.ResDataDto
	err := json.Unmarshal(w.Body.Bytes(), &responseDto)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), uint(1), responseDto.Data[0].ID) // Check if conversion is correct
	assert.Equal(suite.T(), "GET", responseDto.Data[0].Method)
	assert.Equal(suite.T(), int64(50), responseDto.Data[0].Latency)

	assert.Equal(suite.T(), mockTotal, responseDto.Pagination.Total)
	assert.Equal(suite.T(), 20, responseDto.Pagination.Limit) // Default limit
	assert.Equal(suite.T(), 0, responseDto.Pagination.Offset) // Default offset

	suite.mockRequestCrudService.AssertExpectations(suite.T())
}
