package controller

import (
	"net/http"
	"treblle/app"
	"treblle/dto"
	"treblle/service"
	"treblle/util/auth"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthCtn struct {
	auth   service.IAuthService
	logger *zap.SugaredLogger
}

func NewAuthCtn() app.Controller {
	var controller *AuthCtn

	// Use the mock service for testing
	app.Invoke(func(loginService service.IAuthService, logger *zap.SugaredLogger) {
		// create controller
		controller = &AuthCtn{
			auth:   loginService,
			logger: logger,
		}
	})

	return controller
}

func (c *AuthCtn) RegisterEndpoints(api *gin.RouterGroup) {
	// create a group with the name of the router
	group := api.Group("/auth")

	// register Endpoints
	group.POST("/login", c.login)
	group.POST("/refresh", auth.Protect(), c.refreshToken)
	group.POST("/logout", auth.Protect(), c.logout)
}

// Login godoc
//
//	@Summary		User login
//	@Description	Authenticates a user and returns access and refresh tokens
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			loginDto	body		dto.LoginDto	true	"Login credentials"
//	@Success		200			{object}	dto.TokenDto
//	@Router			/auth/login [post]
func (l *AuthCtn) login(c *gin.Context) {
	var loginDto dto.LoginDto

	if err := c.BindJSON(&loginDto); err != nil {
		l.logger.Errorf("Invalid login request err = %+v", err)
		return
	}

	accessToken, err := l.auth.Login(loginDto.Email, loginDto.Password)
	if err != nil {
		l.logger.Errorf("Login failed err = %+v", err)
		c.JSON(http.StatusUnauthorized, err.Error())
		return
	}

	c.JSON(http.StatusOK, dto.TokenDto{
		AccessToken: accessToken,
	})
}

// Refresh godoc
//
//	@Summary		Refresh Access Token
//	@Description	Generates a new access token using a valid refresh token
//	@Tags			auth
//	@Produce		json
//	@Success		200	{object}	dto.TokenDto
//	@Router			/auth/refresh [post]
func (l *AuthCtn) refreshToken(c *gin.Context) {
	tokenStr := c.Request.Header.Get("Authorization")
	token, err := l.auth.RefreshTokens(tokenStr)
	if err != nil {
		l.logger.Errorf("Refresh failed err = %w", err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, dto.TokenDto{
		AccessToken: token,
	})
}

// Refresh godoc
//
//	@Summary		Refresh Access Token
//	@Description	Generates a new access token using a valid refresh token
//	@Tags			auth
//	@Produce		json
//	@Success		200	{object}	dto.TokenDto
//	@Router			/auth/logout [post]
func (l *AuthCtn) logout(c *gin.Context) {
	_, claims, err := auth.ParseToken(c.Request.Header.Get("Authorization"))
	if err != nil {
		l.logger.Errorf("Logout failed err = %w", err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	err = l.auth.Logout(claims.ID)
	if err != nil {
		l.logger.Errorf("Logout failed err = %w", err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.AbortWithStatus(http.StatusOK)
}
