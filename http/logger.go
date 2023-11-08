package http

// ILogger 日志
type ILogger interface {
	Println(msg string, a ...interface{})
}
