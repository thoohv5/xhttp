package http

import (
	"crypto/tls"
	"io"
	nethttp "net/http"
	"time"
)

// parameter 参数
type parameter struct {
	// url
	url string
	// 请求方式
	method Method
	// 超时时间
	timeout time.Duration
	// header
	header map[string]string
	// param
	param map[string]interface{}
	// beforeRequest
	beforeRequest []func(r *parameter) error
	// reader
	body io.Reader
	// tls
	tLSClientConfig *tls.Config
	// log
	log ILogger
	// deleteUriFlag
	deleteUriFlag bool

	// 返回值
	response *nethttp.Response
}

func (p *parameter) SetBody(body io.Reader) {
	p.body = body
}

type Option interface {
	apply(*parameter)
}

type optionFunc func(*parameter)

func (f optionFunc) apply(o *parameter) {
	f(o)
}

// WithUrl URL
func WithUrl(url string) Option {
	return optionFunc(func(r *parameter) {
		r.url = url
	})
}

// WithMethod 方法
func WithMethod(method Method) Option {
	return optionFunc(func(r *parameter) {
		r.method = method
	})
}

// WithTimeout 超时
func WithTimeout(timeout time.Duration) Option {
	return optionFunc(func(r *parameter) {
		r.timeout = timeout * time.Second
	})
}

// WithParam 参数
func WithParam(params map[string]interface{}) Option {
	return optionFunc(func(r *parameter) {
		for key, val := range params {
			r.param[key] = val
		}
	})
}

// WithHeader Header
func WithHeader(headers map[string]string) Option {
	return optionFunc(func(r *parameter) {
		for key, val := range headers {
			r.header[key] = val
		}
	})
}

// WithBeforeRequest 请求前置
func WithBeforeRequest(preDeal func(r *parameter) error) Option {
	return optionFunc(func(r *parameter) {
		r.beforeRequest = append(r.beforeRequest, preDeal)
	})
}

// WithTLSClientConfig 证书
func WithTLSClientConfig(tLSClientConfig *tls.Config) Option {
	return optionFunc(func(r *parameter) {
		r.tLSClientConfig = tLSClientConfig
	})
}

// WithResponse 返回值
func WithResponse(response *nethttp.Response) Option {
	return optionFunc(func(r *parameter) {
		r.response = response
	})
}

// WithLogger 日志
func WithLogger(log ILogger) Option {
	return optionFunc(func(r *parameter) {
		r.log = log
	})
}

// WithDeleteURIFlag DELETE URL
func WithDeleteURIFlag(flag bool) Option {
	return optionFunc(func(r *parameter) {
		r.deleteUriFlag = flag
	})
}
