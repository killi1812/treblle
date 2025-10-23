package main

import (
	"treblle/app"
	"treblle/controller"
	"treblle/service"

	"go.uber.org/zap"
)

func init() {
	app.Setup()
}

func main() {
	// Provide logger
	app.Provide(zap.S)

	app.Provide(service.NewRequestService)
	app.Provide(service.NewRequestCrudService)

	app.RegisterController(controller.NewInfoCnt)
	app.RegisterController(controller.NewRequestCtn)

	app.Start()
}
