package service

import (
	"net/http"
	"time"
	"treblle/app"
	"treblle/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ReqLogger struct {
	db     *gorm.DB
	logger *zap.SugaredLogger
}

func NewRequestService() app.RequestLogger {
	var service *ReqLogger

	app.Invoke(func(db *gorm.DB, logger *zap.SugaredLogger) {
		service = &ReqLogger{
			db:     db,
			logger: logger,
		}
	})

	return service
}

func (r *ReqLogger) LogRequest(req *http.Request) (*model.Request, error) {
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

func (r *ReqLogger) LogResponse(id uint, resp *http.Response) (*model.Request, error) {
	var request model.Request

	if rez := r.db.First(&request, id); rez.Error != nil {
		r.logger.Errorf("Failed reading request, error = %v", rez.Error)
		return nil, rez.Error
	}

	request.ResponseTime = time.Now()
	request.Response = resp.StatusCode
	request.Latency = request.ResponseTime.Sub(request.CreatedAt)
	zap.S().Debugf("req latency is: %v ", request.Latency.Milliseconds())

	if rez := r.db.Save(&request); rez.Error != nil {
		r.logger.Errorf("Failed logging request, error = %v", rez.Error)
		return nil, rez.Error
	}

	return &request, nil
}
