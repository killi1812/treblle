package model

import (
	"time"
)

type Request struct {
	ID           uint `gorm:"primarykey"`
	Method       string
	Response     string
	Path         string
	ResponseTime time.Time
	CreatedAt    time.Time
}
