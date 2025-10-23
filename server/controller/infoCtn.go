package controller

import (
	"net/http"
	"treblle/app"
	"treblle/dto"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type InfoCtn struct {
	logger *zap.SugaredLogger
}

// NewImageCnt creates a new controller for images.
func NewInfoCnt() app.Controller {
	var controller *InfoCtn
	app.Invoke(func(logger *zap.SugaredLogger) {
		controller = &InfoCtn{
			logger: logger,
		}
	})
	return controller
}

// RegisterEndpoints registers the image manipulation endpoints.
func (cnt *InfoCtn) RegisterEndpoints(router *gin.RouterGroup) {
	router.GET("/info", cnt.getServerInfo)
}

// getServerInfo godoc
//
//	@Summary		Get server info
//	@Description	return information about the server build, version, etc ...
//	@Tags			info
//	@Produce		json
//	@Success		200	{object}	dto.ServerInfoDto	"Information about server"
//	@Router			/info [get]
func (ctn *InfoCtn) getServerInfo(c *gin.Context) {
	serverInfo := dto.ServerInfoDto{
		Build:          app.Build,
		Version:        app.Version,
		CommitHash:     app.CommitHash,
		BuildTimestamp: app.BuildTimestamp,
	}
	c.AbortWithStatusJSON(http.StatusOK, serverInfo)
}
