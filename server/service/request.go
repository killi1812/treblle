package service

import (
	"net/http"
	"treblle/app"
	"treblle/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type RequestService struct {
	db     *gorm.DB
	logger *zap.SugaredLogger
}

func NewRequestService() app.RequestLogger {
	var service *RequestService

	app.Invoke(func(db *gorm.DB, logger *zap.SugaredLogger) {
		service = &RequestService{
			db:     db,
			logger: logger,
		}
	})

	return service
}

func (r *RequestService) LogRequest(req *http.Request) error {
	var request model.Request
	if err := request.FromRequest(req); err != nil {
		r.logger.Errorf("Failed logging request, error = %v", err)
		return err
	}

	rez := r.db.Create(&request)
	if rez.Error != nil {
		r.logger.Errorf("Failed logging request, error = %v", rez.Error)
		return rez.Error
	}

	return nil
}
