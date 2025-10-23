package service

import (
	"treblle/app"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ApiService struct {
	db     *gorm.DB
	logger *zap.SugaredLogger
}

func NewApiService() *ApiService {
	var service *ApiService

	app.Invoke(func(db *gorm.DB, logger *zap.SugaredLogger) {
		service = &ApiService{
			db:     db,
			logger: logger,
		}
	})

	return service
}
