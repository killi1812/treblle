package service

import (
	"treblle/app"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type RequestService struct {
	db     *gorm.DB
	logger *zap.SugaredLogger
}

func NewRequestService() *RequestService {
	var service *RequestService

	app.Invoke(func(db *gorm.DB, logger *zap.SugaredLogger) {
		service = &RequestService{
			db:     db,
			logger: logger,
		}
	})

	return service
}
