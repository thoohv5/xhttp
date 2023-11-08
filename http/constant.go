package http

import "time"

type Method string

const (
	DefaultTimeOut        = 3 * time.Second
	MethodPost     Method = "POST"
	MethodGet      Method = "GET"
	MethodPut      Method = "PUT"
	MethodDelete   Method = "DELETE"
)
