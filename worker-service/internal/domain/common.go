package domain

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
