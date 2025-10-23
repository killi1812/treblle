package service

import (
	"net/http"
	"time"
	"treblle/app"
	"treblle/model"
	"treblle/util/read"

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

func (r *RequestService) LogRequest(req *http.Request) (*model.Request, error) {
	var request model.Request
	if err := request.FromRequest(req); err != nil {
		r.logger.Errorf("Failed logging request, error = %v", err)
		return nil, err
	}

	rez := r.db.Create(&request)
	if rez.Error != nil {
		r.logger.Errorf("Failed logging request, error = %v", rez.Error)
		return nil, rez.Error
	}

	return &request, nil
}

func (r *RequestService) LogResponse(id uint, resp *http.Response) (*model.Request, error) {
	var request model.Request

	if rez := r.db.First(&request, id); rez.Error != nil {
		r.logger.Errorf("Failed reading request, error = %v", rez.Error)
		return nil, rez.Error
	}

	request.ResponseTime = time.Now()
	data, err := read.ReadBody(resp)
	if err != nil {
		zap.S().Errorf("Error reading response, error = %v", err)
		return nil, nil
	}
	request.Response = string(data)
	zap.S().Debugf("req latency is: %v ", time.Since(request.CreatedAt))

	if rez := r.db.Save(&request); rez.Error != nil {
		r.logger.Errorf("Failed logging request, error = %v", rez.Error)
		return nil, rez.Error
	}

	return &request, nil
}
