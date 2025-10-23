package service_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
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

// --- ReqLogger Test Suite ---
type ReqLoggerTestSuite struct {
	suite.Suite
	db          *gorm.DB
	logger      *zap.SugaredLogger
	reqLogger   app.RequestLogger
	logObserver *observer.ObservedLogs
}

// SetupSuite runs once before the entire suite (minimal setup).
func (suite *ReqLoggerTestSuite) SetupSuite() {
	core, obs := observer.New(zap.InfoLevel)
	suite.logger = zap.New(core).Sugar()
	suite.logObserver = obs
}

// TearDownSuite runs once after all tests in the suite have finished.
func (suite *ReqLoggerTestSuite) TearDownSuite() {
	if suite.logger != nil {
		_ = suite.logger.Sync()
	}
}

// SetupTest runs before each test method - Creates isolated DB and service instance.
func (suite *ReqLoggerTestSuite) SetupTest() {
	dsn := fmt.Sprintf("file:req_log_test_%s?mode=memory&cache=private", suite.T().Name()) // Unique DSN per test
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Keep GORM logs quiet
	})
	suite.Require().NoError(err, "Failed to connect to SQLite for test %s", suite.T().Name())
	suite.db = db // Assign the test-specific DB

	// --- Schema Migration (Per Test DB) ---
	err = suite.db.AutoMigrate(&model.Request{})
	suite.Require().NoError(err, "Failed to migrate database schema for test %s", suite.T().Name())

	// --- Service Initialization (Per Test) ---
	suite.reqLogger = &service.ReqLogger{
		Db:     db,
		Logger: suite.logger,
	}
	suite.Require().NotNil(suite.reqLogger)
}

// TearDownTest runs after each test method - Cleans up the test-specific DB.
func (suite *ReqLoggerTestSuite) TearDownTest() {
	if suite.db != nil {
		sqlDB, _ := suite.db.DB()
		err := sqlDB.Close()
		suite.Require().NoError(err, "Failed to close SQLite connection for test %s", suite.T().Name())
	}
}

// TestReqLoggerTestSuite is the entry point for running the test suite.
func TestReqLoggerTestSuite(t *testing.T) {
	suite.Run(t, new(ReqLoggerTestSuite))
}

// --- Test Cases ---

func (suite *ReqLoggerTestSuite) TestLogRequest_Success() {
	// Arrange
	mockReq := httptest.NewRequest(http.MethodGet, "/api/test/path?query=1", nil)

	// Act
	loggedReq, err := suite.reqLogger.LogRequest(mockReq)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), loggedReq)
	assert.Positive(suite.T(), loggedReq.ID) // DB should assign an ID
	assert.Equal(suite.T(), http.MethodGet, loggedReq.Method)
	assert.Equal(suite.T(), "/api/test/path", loggedReq.Path) // Should only store the path, not query
	assert.WithinDuration(suite.T(), time.Now(), loggedReq.CreatedAt, 1*time.Second)
	assert.Equal(suite.T(), 0, loggedReq.Response)               // Response not set yet
	assert.True(suite.T(), loggedReq.ResponseTime.IsZero())      // ResponseTime not set yet
	assert.Equal(suite.T(), time.Duration(0), loggedReq.Latency) // Latency not set yet

	// Verify in DB
	var dbReq model.Request
	result := suite.db.First(&dbReq, loggedReq.ID)
	assert.NoError(suite.T(), result.Error)
	assert.Equal(suite.T(), loggedReq.Method, dbReq.Method)
	assert.Equal(suite.T(), loggedReq.Path, dbReq.Path)
}

func (suite *ReqLoggerTestSuite) TestLogResponse_Success() {
	// Arrange: First log a request to get a valid ID
	mockReq := httptest.NewRequest(http.MethodPost, "/api/data", nil)
	initialReq, err := suite.reqLogger.LogRequest(mockReq)
	suite.Require().NoError(err)
	suite.Require().NotNil(initialReq)
	initialID := initialReq.ID

	// Let some time pass to simulate processing
	time.Sleep(50 * time.Millisecond)

	// Create a mock response
	mockResp := &http.Response{
		StatusCode: http.StatusCreated, // 201
		Request:    mockReq,            // Associate response with original request
	}

	// Act
	updatedReq, err := suite.reqLogger.LogResponse(initialID, mockResp)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), updatedReq)
	assert.Equal(suite.T(), initialID, updatedReq.ID)
	assert.Equal(suite.T(), http.StatusCreated, updatedReq.Response)
	assert.WithinDuration(suite.T(), time.Now(), updatedReq.ResponseTime, 1*time.Second)
	assert.False(suite.T(), updatedReq.ResponseTime.IsZero())
	assert.True(suite.T(), updatedReq.Latency >= 50*time.Millisecond) // Check latency calculation

	// Verify in DB
	var dbReq model.Request
	result := suite.db.First(&dbReq, initialID)
	assert.NoError(suite.T(), result.Error)
	assert.Equal(suite.T(), http.StatusCreated, dbReq.Response)
	assert.False(suite.T(), dbReq.ResponseTime.IsZero())
	assert.Equal(suite.T(), updatedReq.Latency, dbReq.Latency)
}

func (suite *ReqLoggerTestSuite) TestLogResponse_RequestNotFound() {
	// Arrange
	nonExistentID := uint(999)
	mockResp := &http.Response{
		StatusCode: http.StatusOK,
		Request:    httptest.NewRequest(http.MethodGet, "/", nil), // Needs a non-nil request
	}

	// Act
	updatedReq, err := suite.reqLogger.LogResponse(nonExistentID, mockResp)

	// Assert
	assert.Error(suite.T(), err)
	assert.True(suite.T(), errors.Is(err, gorm.ErrRecordNotFound)) // Check for specific GORM error
	assert.Nil(suite.T(), updatedReq)
}

// --- Helper for setting DB on test instance ---
// Add this method to your actual service.ReqLogger struct if it doesn't exist
// This is needed because DI happens *before* SetupTest creates the unique DB.
// NOTE: Add this method to the *actual* service/reqLog.go file
/*
func (r *ReqLogger) SetDB(db *gorm.DB) {
	r.db = db
}
*/

// --- Mock Context Value (If needed by LogResponse) ---
// If LogResponse relies on context values set by the proxy (like request ID),
// you'll need to mock that context in the response's request.
func (suite *ReqLoggerTestSuite) TestLogResponse_WithMockContext() {
	// Arrange: First log a request to get a valid ID
	mockReq := httptest.NewRequest(http.MethodPost, "/api/context", nil)
	initialReq, err := suite.reqLogger.LogRequest(mockReq)
	suite.Require().NoError(err)
	initialID := initialReq.ID

	// Create context with the expected key/value
	// Use the *actual* key your proxy uses (replace "YourRequestIDKeyType" and "YourRequestIDKey")
	// type YourRequestIDKeyType string
	// const YourRequestIDKey YourRequestIDKeyType = "RequestIdKey" // Or whatever key is used in proxy.go
	// ctxWithID := context.WithValue(mockReq.Context(), YourRequestIDKey, initialID)
	ctxWithID := context.WithValue(mockReq.Context(), "TestKey", initialID) // Example key
	reqWithContext := mockReq.WithContext(ctxWithID)

	mockResp := &http.Response{
		StatusCode: http.StatusOK,
		Request:    reqWithContext, // Use the request with the mocked context
	}

	// Act
	updatedReq, err := suite.reqLogger.LogResponse(initialID, mockResp)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), updatedReq)
	assert.Equal(suite.T(), initialID, updatedReq.ID)
	assert.Equal(suite.T(), http.StatusOK, updatedReq.Response)

	// Verify in DB
	var dbReq model.Request
	result := suite.db.First(&dbReq, initialID)
	assert.NoError(suite.T(), result.Error)
	assert.Equal(suite.T(), http.StatusOK, dbReq.Response)
}
