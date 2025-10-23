package dto

type ListQuery struct {
	Search   string `form:"search"`
	Method   string `form:"method" binding:"oneof= GET POST PUT DELETE PATCH HEAD OPTION TRACE CONNECT"`
	Response int    `form:"response,one"` // Gin binds '0' if not present
	Limit    int    `form:"limit"`
	Offset   int    `form:"offset"`
	SortBy   string `form:"sort_by"`
	Order    string `form:"order"`
}
