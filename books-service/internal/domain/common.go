package domain

import "math"

type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type ErrorResponse struct {
	Code      string      `json:"code"`
	Message   string      `json:"message"`
	Details   ErrorDetail `json:"details"`
	RequestID string      `json:"request_id"`
}

type ErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"messages"`
}

func NewPagination(page, limit, total int) Pagination {
	return Pagination{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: int(math.Ceil(float64(total) / float64(limit))),
	}
}

type ListParams struct {
	Search string
	Sort   string
	Order  string
	Page   int
	Limit  int
}
