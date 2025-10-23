package service

import (
	"treblle/model"
	"treblle/util/auth"
	"treblle/util/cerror"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// --- Auth Service Test Suite ---
type authTestSuite struct {
	suite.Suite
	db            *gorm.DB
	logger        *zap.SugaredLogger
	logObserver   *observer.ObservedLogs
	authService   IAuthService
	seededUser    *model.User
	seededRawPass string
}

// SetupSuite runs once before all tests in the suite.
func (suite *authTestSuite) SetupSuite() {
	// --- Logger Setup ---
	core, obs := observer.New(zap.InfoLevel)
	suite.logger = zap.New(core).Sugar()
	suite.logObserver = obs
	zap.ReplaceGlobals(zap.New(core))
	// --- Database Setup ---
	db, err := gorm.Open(sqlite.Open("file:auth_test.db?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	suite.Require().NoError(err, "Failed to connect to SQLite for AuthService tests")
	suite.db = db

	// --- Schema Migration ---
	err = suite.db.AutoMigrate(model.GetAllModels()...)
	suite.Require().NoError(err, "Failed to migrate database schema")

	// --- Service Initialization ---
	suite.authService = &AuthService{
		db:     db,
		logger: suite.logger,
	}
	suite.Require().NotNil(suite.authService)

	// Seed a test user
	suite.seededRawPass = "password123"
	hashedPassword, err := auth.HashPassword(suite.seededRawPass)
	suite.Require().NoError(err)

	user := model.User{
		Uuid:         uuid.New(),
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
		Role:         model.ROLE_USER,
	}
	result := suite.db.Create(&user)
	suite.Require().NoError(result.Error)
	suite.seededUser = &user
}

func (suite *authTestSuite) SetupTest() {

}

// TearDownSuite runs once after all tests in the suite have finished.
func (suite *authTestSuite) TearDownSuite() {
	if suite.logger != nil {
		_ = suite.logger.Sync()
	}
	if suite.db != nil {
		sqlDB, _ := suite.db.DB()
		err := sqlDB.Close()
		suite.Require().NoError(err)
	}
}

// TestAuthTestSuite is the entry point for running the test suite.
func TestAuthTestSuite(t *testing.T) {
	suite.Run(t, new(authTestSuite))
}

// --- Test Cases ---

func (suite *authTestSuite) TestLogin_Success() {
	// Act
	accessToken, err := suite.authService.Login(suite.seededUser.Email, suite.seededRawPass)

	// Assert
	suite.NoError(err)
	suite.NotEmpty(accessToken)

	// Verify a session was created in the database
	var session model.Session
	result := suite.db.Where("user_uuid = ?", suite.seededUser.Uuid).First(&session)
	suite.NoError(result.Error)
	suite.NotEmpty(session.RefreshToken)
}

func (suite *authTestSuite) TestLogin_UserNotFound() {
	// Act
	accessToken, err := suite.authService.Login("nonexistent@example.com", "password")

	// Assert
	suite.ErrorIs(err, cerror.ErrInvalidCredentials)
	suite.Empty(accessToken)
}

func (suite *authTestSuite) TestLogin_InvalidPassword() {
	// Act
	accessToken, err := suite.authService.Login(suite.seededUser.Email, "wrongpassword")

	// Assert
	suite.ErrorIs(err, cerror.ErrInvalidCredentials)
	suite.Empty(accessToken)
}

func (suite *authTestSuite) TestLogin_ExistingSessionIsReplaced() {
	// Arrange: Log the user in once to create a session
	_, err := suite.authService.Login(suite.seededUser.Email, suite.seededRawPass)
	suite.Require().NoError(err)

	var firstSession model.Session
	suite.db.Where("user_uuid = ?", suite.seededUser.Uuid).First(&firstSession)
	suite.Require().NotEmpty(firstSession.RefreshToken)

	// Act: Log the user in a second time
	_, err = suite.authService.Login(suite.seededUser.Email, suite.seededRawPass)
	suite.Require().NoError(err)

	// Assert: Check that the session has been updated
	var secondSession model.Session
	suite.db.Where("user_uuid = ?", suite.seededUser.Uuid).First(&secondSession)
	suite.Require().NoError(err)
	suite.NotEqual(firstSession.RefreshToken, secondSession.RefreshToken, "A new refresh token should be generated")
}

func (suite *authTestSuite) TestLogout_Success() {
	// Arrange: Log in to create a session
	_, err := suite.authService.Login(suite.seededUser.Email, suite.seededRawPass)
	suite.Require().NoError(err)

	// Act
	err = suite.authService.Logout(suite.seededUser.Uuid.String())

	// Assert
	suite.NoError(err)

	// Verify the session was deleted
	var session model.Session
	result := suite.db.Where("user_uuid = ?", suite.seededUser.Uuid).First(&session)
	suite.ErrorIs(result.Error, gorm.ErrRecordNotFound)
}

func (suite *authTestSuite) TestRefreshTokens_Success() {
	// Arrange: Log in to get a valid token and create a session
	originalAccessToken, err := suite.authService.Login(suite.seededUser.Email, suite.seededRawPass)
	suite.Require().NoError(err)
	suite.Require().NotEmpty(originalAccessToken)

	// Act
	newAccessToken, err := suite.authService.RefreshTokens("Bearer " + originalAccessToken)

	// Assert
	suite.NoError(err)
	suite.NotEmpty(newAccessToken)
	suite.NotEqual(originalAccessToken, newAccessToken, "A new access token should be generated")
}

func (suite *authTestSuite) TestRefreshTokens_InvalidToken() {
	// Act
	newAccessToken, err := suite.authService.RefreshTokens("Bearer invalidtoken")

	// Assert
	suite.Error(err)
	suite.Empty(newAccessToken)
}

func (suite *authTestSuite) TestRefreshTokens_NoSession() {
	// Arrange: Generate a token but don't create a session
	token, _, err := auth.GenerateTokens(suite.seededUser)
	suite.Require().NoError(err)

	// Act
	newAccessToken, err := suite.authService.RefreshTokens("Bearer " + token)

	// Assert
	suite.ErrorIs(err, gorm.ErrRecordNotFound)
	suite.Empty(newAccessToken)
}
