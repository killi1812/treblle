package model

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type Request struct {
	ID           uint `gorm:"primarykey"`
	Method       string
	Response     string
	Path         string
	ResponseTime time.Time
	CreatedAt    time.Time
}

func (r *Request) FromRequest(req *http.Request) error {
	r.Method = req.Method
	r.Path = req.URL.Path
	r.CreatedAt = time.Now()
	zap.S().Debugf("Populating request data, rez %+v", *r)
	return nil
}
