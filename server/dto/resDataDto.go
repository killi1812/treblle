package dto

import (
	"treblle/model"
)

type ResDataDto struct {
	Data       []RequestsDto `json:"data"`
	Pagination Pagination    `json:"pagination"`
}

type Pagination struct {
	Total  int64 `json:"total"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
}

type RequestsDto struct {
	ID           uint   `json:"id"`
	Method       string `json:"method"`
	Response     int    `json:"response"`
	Path         string `json:"path"`
	ResponseTime string `json:"responseTime"`
	CreatedAt    string `json:"createdAt"`
	Latency      int64  `json:"latency"` //Latency in Milliseconds
}

func (dto *RequestsDto) FromModel(m model.Request) error {
	dto.ID = m.ID
	dto.Method = m.Method
	dto.Response = m.Response
	dto.Path = m.Path
	dto.ResponseTime = m.ResponseTime.String()
	dto.CreatedAt = m.CreatedAt.String()
	dto.Latency = m.Latency.Milliseconds()

	return nil
}
