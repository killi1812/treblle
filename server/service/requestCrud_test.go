package service_test

import (
	"testing"
	"time"
	"treblle/app"
	"treblle/model"
	"treblle/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// --- RequestCrudService Test Suite ---
type RequestCrudServiceTestSuite struct {
	suite.Suite
	db          *gorm.DB
	logger      *zap.SugaredLogger
	logObserver *observer.ObservedLogs
	crudService service.IRequestCrudService
	// Store seeded requests for easy access in tests
	seededRequests []model.Request
}

// SetupSuite runs once before all tests in the suite.
func (suite *RequestCrudServiceTestSuite) SetupSuite() {
	// --- Logger Setup ---
	core, obs := observer.New(zap.InfoLevel)
	suite.logger = zap.New(core).Sugar()
	suite.logObserver = obs

	// --- Database Setup ---
	// Using in-memory SQLite for isolated testing
	db, err := gorm.Open(sqlite.Open("file:req_crud_test.db?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Keep GORM logs quiet during tests
	})
	suite.Require().NoError(err, "Failed to connect to SQLite for RequestCrudService tests")
	suite.db = db

	// --- Schema Migration ---
	// Migrate only the necessary model for this service
	err = suite.db.AutoMigrate(&model.Request{})
	suite.Require().NoError(err, "Failed to migrate database schema for Request")

	// --- Service Initialization ---
	// Initialize the *real* service, but provide the test DB and logger via DI
	app.Test() // Setup test DI container
	app.Provide(func() *gorm.DB { return suite.db })
	app.Provide(func() *zap.SugaredLogger { return suite.logger })
	suite.crudService = service.NewRequestCrudService() // This will now get the test DB and logger
	suite.Require().NotNil(suite.crudService)

	// --- Seed Data ---
	suite.seedTestData()
}

// seedTestData populates the database with initial data for tests.
func (suite *RequestCrudServiceTestSuite) seedTestData() {
	now := time.Now()
	requests := []model.Request{
		{Method: "GET", Path: "/api/users", Response: 200, CreatedAt: now.Add(-5 * time.Minute), ResponseTime: now.Add(-5 * time.Minute).Add(50 * time.Millisecond), Latency: 50 * time.Millisecond},
		{Method: "POST", Path: "/api/users", Response: 201, CreatedAt: now.Add(-4 * time.Minute), ResponseTime: now.Add(-4 * time.Minute).Add(150 * time.Millisecond), Latency: 150 * time.Millisecond},
		{Method: "GET", Path: "/api/products/123", Response: 200, CreatedAt: now.Add(-3 * time.Minute), ResponseTime: now.Add(-3 * time.Minute).Add(75 * time.Millisecond), Latency: 75 * time.Millisecond},
		{Method: "PUT", Path: "/api/products/123", Response: 200, CreatedAt: now.Add(-2 * time.Minute), ResponseTime: now.Add(-2 * time.Minute).Add(250 * time.Millisecond), Latency: 250 * time.Millisecond},
		{Method: "GET", Path: "/api/orders", Response: 404, CreatedAt: now.Add(-1 * time.Minute), ResponseTime: now.Add(-1 * time.Minute).Add(30 * time.Millisecond), Latency: 30 * time.Millisecond},
		{Method: "DELETE", Path: "/api/users/456", Response: 204, CreatedAt: now, ResponseTime: now.Add(120 * time.Millisecond), Latency: 120 * time.Millisecond}, // ID 6
	}

	result := suite.db.Create(&requests)
	suite.Require().NoError(result.Error, "Failed to seed test data")
	suite.seededRequests = requests
}

// TearDownSuite runs once after all tests in the suite have finished.
func (suite *RequestCrudServiceTestSuite) TearDownSuite() {
	if suite.logger != nil {
		_ = suite.logger.Sync()
	}
	if suite.db != nil {
		sqlDB, _ := suite.db.DB()
		err := sqlDB.Close()
		suite.Require().NoError(err)
	}
}

// TestRequestCrudServiceTestSuite is the entry point for running the test suite.
func TestRequestCrudServiceTestSuite(t *testing.T) {
	suite.Run(t, new(RequestCrudServiceTestSuite))
}

// --- Test Cases ---

func (suite *RequestCrudServiceTestSuite) TestList_Defaults() {
	params := service.ListRequestsParams{Limit: 10, Offset: 0} // Default limit/offset
	requests, total, err := suite.crudService.List(params)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(len(suite.seededRequests)), total)
	assert.Len(suite.T(), requests, len(suite.seededRequests)) // Since limit > seeded count
	// Default sort is created_at desc
	assert.Equal(suite.T(), uint(6), requests[0].ID) // Most recent first
	assert.Equal(suite.T(), uint(1), requests[len(requests)-1].ID)
}

func (suite *RequestCrudServiceTestSuite) TestList_WithSearch() {
	search := "products"
	params := service.ListRequestsParams{Search: &search, Limit: 10, Offset: 0}
	requests, total, err := suite.crudService.List(params)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(2), total) // Found GET and PUT /api/products/123
	assert.Len(suite.T(), requests, 2)
	assert.Contains(suite.T(), requests[0].Path, "products")
	assert.Contains(suite.T(), requests[1].Path, "products")
}

func (suite *RequestCrudServiceTestSuite) TestList_WithMethodFilter() {
	method := "GET"
	params := service.ListRequestsParams{Method: &method, Limit: 10, Offset: 0}
	requests, total, err := suite.crudService.List(params)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(3), total) // 3 GET requests were seeded
	assert.Len(suite.T(), requests, 3)
	for _, req := range requests {
		assert.Equal(suite.T(), "GET", req.Method)
	}
}

func (suite *RequestCrudServiceTestSuite) TestList_WithResponseFilter() {
	response := 200
	params := service.ListRequestsParams{Response: &response, Limit: 10, Offset: 0}
	requests, total, err := suite.crudService.List(params)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(3), total) // 3 requests with status 200
	assert.Len(suite.T(), requests, 3)
	for _, req := range requests {
		assert.Equal(suite.T(), 200, req.Response)
	}
}

func (suite *RequestCrudServiceTestSuite) TestList_WithPagination() {
	params := service.ListRequestsParams{Limit: 2, Offset: 1} // Get 2nd and 3rd most recent
	requests, total, err := suite.crudService.List(params)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(len(suite.seededRequests)), total)
	assert.Len(suite.T(), requests, 2)
	// Default sort is created_at desc, so offset 1 skips ID 6
	assert.Equal(suite.T(), uint(5), requests[0].ID) // Second most recent
	assert.Equal(suite.T(), uint(4), requests[1].ID) // Third most recent
}

func (suite *RequestCrudServiceTestSuite) TestList_SortByResponseTimeAsc() {
	params := service.ListRequestsParams{SortBy: "latency", Order: "asc", Limit: 10, Offset: 0}
	requests, total, err := suite.crudService.List(params)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(len(suite.seededRequests)), total)
	assert.Len(suite.T(), requests, len(suite.seededRequests))

	// Check if sorted by Latency (ascending)
	assert.True(suite.T(), requests[0].Latency <= requests[1].Latency)
	assert.True(suite.T(), requests[1].Latency <= requests[2].Latency)
	assert.True(suite.T(), requests[2].Latency <= requests[3].Latency)
	// ... and so on
	assert.Equal(suite.T(), int64(30), requests[0].Latency.Milliseconds())                // Should be the fastest (ID 5)
	assert.Equal(suite.T(), int64(250), requests[len(requests)-1].Latency.Milliseconds()) // Should be the slowest (ID 4)
}

func (suite *RequestCrudServiceTestSuite) TestList_SortByCreatedAtDesc() {
	params := service.ListRequestsParams{SortBy: "created_at", Order: "desc", Limit: 10, Offset: 0}
	requests, total, err := suite.crudService.List(params)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(len(suite.seededRequests)), total)
	assert.Len(suite.T(), requests, len(suite.seededRequests))

	// Check if sorted by CreatedAt (descending) - most recent first
	assert.True(suite.T(), requests[0].CreatedAt.After(requests[1].CreatedAt) || requests[0].CreatedAt.Equal(requests[1].CreatedAt))
	assert.True(suite.T(), requests[1].CreatedAt.After(requests[2].CreatedAt) || requests[1].CreatedAt.Equal(requests[2].CreatedAt))
	assert.Equal(suite.T(), uint(6), requests[0].ID) // Most recent
}

func (suite *RequestCrudServiceTestSuite) TestList_CombinedFiltersAndSort() {
	method := "GET"
	response := 200
	params := service.ListRequestsParams{
		Method:   &method,
		Response: &response,
		SortBy:   "response_time",
		Order:    "desc", // Slowest first among GET 200s
		Limit:    10,
		Offset:   0,
	}
	requests, total, err := suite.crudService.List(params)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(2), total) // Only GET requests with status 200 (IDs 1 and 3)
	assert.Len(suite.T(), requests, 2)
	assert.Equal(suite.T(), uint(3), requests[0].ID) // ID 3 has latency 75ms
	assert.Equal(suite.T(), uint(1), requests[1].ID) // ID 1 has latency 50ms
	assert.True(suite.T(), requests[0].Latency >= requests[1].Latency)
}

func (suite *RequestCrudServiceTestSuite) TestList_NoResults() {
	search := "nonexistentpath"
	params := service.ListRequestsParams{Search: &search, Limit: 10, Offset: 0}
	requests, total, err := suite.crudService.List(params)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(0), total)
	assert.Len(suite.T(), requests, 0)
}

