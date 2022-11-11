package model

type Response struct {
	Code  int    `json:"code"`
	Error string `json:"error,omitempty"`
}

type ResponseSuccess struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}
