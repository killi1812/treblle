package auth_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"treblle/app"
	"treblle/model"
	"treblle/util/auth"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// --- Test Suite Definition ---
type MiddlewareTestSuite struct {
	suite.Suite
	router *gin.Engine
}

// SetupSuite runs once before all tests in the suite
func (suite *MiddlewareTestSuite) SetupSuite() {
	app.AccessKey = "test-middleware-access-key"
	app.RefreshKey = "test-middleware-refresh-key"

	// Setup Gin router for testing
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()

	suite.router.GET("/protected/general", auth.Protect(), func(c *gin.Context) {
		c.String(http.StatusOK, "general_access_granted")
	})

	suite.router.GET("/protected/admin", auth.Protect("admin"), func(c *gin.Context) {
		c.String(http.StatusOK, "admin_access_granted")
	})
}

// Helper to make HTTP requests
func (suite *MiddlewareTestSuite) performRequest(method, path, token string, body ...string) *httptest.ResponseRecorder {
	var reqBody *strings.Reader
	if len(body) > 0 {
		reqBody = strings.NewReader(body[0])
	} else {
		reqBody = strings.NewReader("")
	}

	req, _ := http.NewRequest(method, path, reqBody)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if method == http.MethodPost || method == http.MethodPut {
		req.Header.Set("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	return w
}

// Helper to generate a token
func (suite *MiddlewareTestSuite) generateToken(userID, username string, userRole model.UserRole, expiresAt time.Time) string {
	claims := &auth.Claims{
		Username: username,
		Role:     userRole,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        userID,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-1 * time.Minute)), // Allow for slight clock skew
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(app.AccessKey))
	if err != nil {
		suite.T().Fatalf("Failed to sign token for test: %v", err)
	}
	return tokenString
}

// --- Test Cases for Protect Middleware ---

func (suite *MiddlewareTestSuite) TestProtect_ValidToken_GeneralAccess() {
	testUserID := "123"
	token := suite.generateToken(testUserID, "test@example.com", "", time.Now().Add(5*time.Minute))

	w := suite.performRequest(http.MethodGet, "/protected/general", token)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.Equal(suite.T(), "general_access_granted", w.Body.String())
}

func (suite *MiddlewareTestSuite) TestProtect_NoToken() {
	w := suite.performRequest(http.MethodGet, "/protected/general", "")

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	assert.Contains(suite.T(), w.Body.String(), "Missing token")
}

func (suite *MiddlewareTestSuite) TestProtect_InvalidTokenFormat_NoBearer() {
	w := httptest.NewRecorder() // Use httptest directly for more control over header
	req, _ := http.NewRequest(http.MethodGet, "/protected/general", nil)
	req.Header.Set("Authorization", "InvalidTokenWithoutBearerPrefix")
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	assert.Contains(suite.T(), w.Body.String(), "Invalid token format")
}

func (suite *MiddlewareTestSuite) TestProtect_InvalidTokenFormat_TooShort() {
	w := suite.performRequest(http.MethodGet, "/protected/general", "short")

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	assert.Contains(suite.T(), w.Body.String(), "Invalid token format")
}

func (suite *MiddlewareTestSuite) TestProtect_MalformedToken() {
	// This token is not a valid JWT structure
	malformedToken := "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ" // Missing signature part
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected/general", nil)
	req.Header.Set("Authorization", malformedToken) // Set directly to bypass "Bearer " prefixing in helper for this specific case
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	// The error message might vary based on JWT library, "Invalid token" or "token contains an invalid number of segments"
	// For this test, we check for "Invalid token" as per the middleware's generic error for parsing issues.
	// A more specific check might involve parsing the JSON response if your middleware returns one.
	// The current middleware returns "Invalid token" for jwt.ParseWithClaims errors.
	assert.Contains(suite.T(), w.Body.String(), "Invalid token")
}

func (suite *MiddlewareTestSuite) TestProtect_ExpiredToken() {
	testUserID := "1234"
	expiredToken := suite.generateToken(testUserID, "expired@example.com", "", time.Now().Add(-5*time.Minute))

	w := suite.performRequest(http.MethodGet, "/protected/general", expiredToken)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	// The middleware uses `token.Valid` which would be false for an expired token.
	// The specific error from `jwt.ParseWithClaims` would be `jwt.ErrTokenExpired`.
	// The middleware then returns "Invalid token".
	assert.Contains(suite.T(), w.Body.String(), "Invalid token")
}

func (suite *MiddlewareTestSuite) TestProtect_WrongSigningKey() {
	claims := &auth.Claims{
		Username: "wrongkey@example.com",
		//Role:     model.RoleOsoba,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        "12345",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Sign with a different key
	wrongKeyToken, _ := token.SignedString([]byte("a-completely-different-secret-key"))

	w := suite.performRequest(http.MethodGet, "/protected/general", wrongKeyToken)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	// Error from `jwt.ParseWithClaims` would be `jwt.ErrSignatureInvalid`.
	// Middleware returns "Invalid token".
	assert.Contains(suite.T(), w.Body.String(), "Invalid token")
}

func (suite *MiddlewareTestSuite) TestProtect_ValidToken_CorrectRole() {
	testUserUUID := "123456"
	adminToken := suite.generateToken(testUserUUID, "admin@example.com", "admin", time.Now().Add(5*time.Minute))

	w := suite.performRequest(http.MethodGet, "/protected/admin", adminToken)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.Equal(suite.T(), "admin_access_granted", w.Body.String())
}

func (suite *MiddlewareTestSuite) TestProtect_ValidToken_InsufficientRole() {
	testUserUUID := "1234567"
	userToken := suite.generateToken(testUserUUID, "user@example.com", "", time.Now().Add(5*time.Minute))

	w := suite.performRequest(http.MethodGet, "/protected/admin", userToken)

	assert.Equal(suite.T(), http.StatusForbidden, w.Code)
	// No body expected for 403 from this middleware implementation
}

// --- Run Test Suite ---
func TestMiddlewareSuite(t *testing.T) {
	suite.Run(t, new(MiddlewareTestSuite))
}
