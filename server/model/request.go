package model

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type Request struct {
	ID           uint   `gorm:"primarykey"`
	Method       string `gorm:"type:varchar(5);not null"`
	Response     int    `gorm:"type:int;null"`
	Path         string `gorm:"type:varchar(150);not null"`
	ResponseTime time.Time
	CreatedAt    time.Time     `gorm:"not null"`
	Latency      time.Duration `gorm:"null"`
}

func (r *Request) FromRequest(req *http.Request) error {
	r.Method = req.Method
	// BUG: fiter the path so it desn't contain proxy/
	r.Path = req.URL.Path
	r.CreatedAt = time.Now()

	zap.S().Debugf("Populating request data, rez %+v", *r)
	return nil
}
