package app

import (
	"github.com/gin-gonic/gin"
)

type Controller interface {
	RegisterEndpoints(router *gin.RouterGroup)
}

var controllers []Controller

// RegisterController registers a controller to a router
func RegisterController(newCtn func() Controller) {
	controllers = append(controllers, newCtn())
}
