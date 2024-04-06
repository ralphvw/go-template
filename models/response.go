package models

type Response struct {
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
	TotalPages int         `json:"totalPages"`
}

type SingleResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
