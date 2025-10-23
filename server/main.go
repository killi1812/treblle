package main

import (
	"treblle/app"
	"treblle/controller"
	"treblle/service"
	"treblle/util/seed"

	"go.uber.org/zap"
)

//	@securitydefinitions.bearerauth	BearerAuth

func init() {
	app.Setup()
}

func main() {
	// Provide logger
	app.Provide(zap.S)
	//app.Provide(minio.New)

	app.Provide(service.NewDiscordService)
	app.Provide(service.NewUserCrudService)
	app.Provide(service.NewAuthService)

	app.RegisterController(controller.NewGameCnt)
	app.RegisterController(controller.NewInfoCnt)
	app.RegisterController(controller.NewUserCtn)
	app.RegisterController(controller.NewAuthCtn)

	seed.Insert()

	app.Start()
}
