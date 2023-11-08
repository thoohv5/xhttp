package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	neturl "net/url"
	"reflect"

	"github.com/thoohv5/xhttp/util/transform"
)

// IHttp 标准
type IHttp interface {
	// Get get
	Get(ctx context.Context, url string, result interface{}, opts ...Option) error
	// Post post
	Post(ctx context.Context, url string, param map[string]interface{}, result interface{}, opts ...Option) error
	// Put put
	Put(ctx context.Context, url string, param map[string]interface{}, result interface{}, opts ...Option) error
	// Delete delete
	Delete(ctx context.Context, url string, param map[string]interface{}, result interface{}, opts ...Option) error
}

type entity struct {
	*parameter
}

// New 创建
func New() IHttp {
	h := &entity{
		parameter: &parameter{
			method:  MethodGet,
			timeout: DefaultTimeOut,
			header: map[string]string{
				"Connection":   "close",
				"Content-Type": "application/json",
			},
			param: map[string]interface{}{},
			tLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
			deleteUriFlag: true,
		},
	}
	return h
}

func (r *entity) withOpt(opts ...Option) error {
	for _, o := range opts {
		o.apply(r.parameter)
	}
	return nil
}

func (r *entity) request(ctx context.Context, url string, result interface{}, opts ...Option) (err error) {
	opts = append([]Option{WithUrl(url)}, opts...)
	// 可选参数
	if err = r.withOpt(opts...); nil != err {
		return fmt.Errorf("request withOpt err, opts: %v, %w", opts, err)
	}

	// 预处理
	for _, beforeRequest := range r.beforeRequest {
		if err = beforeRequest(r.parameter); nil != err {
			return fmt.Errorf("request callback err, r: %v, %w", r, err)
		}
	}

	// 组装request
	req, err := http.NewRequestWithContext(ctx, string(r.method), r.url, r.body)
	if nil != err {
		return fmt.Errorf("request NewRequestWithContext err, url: %s, body: %s, %w", r.url, r.body, err)
	}

	// 组装header
	for key, value := range r.header {
		req.Header.Set(key, value)
	}

	// 发送请求
	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig: r.tLSClientConfig,
	}, Timeout: r.timeout}
	resp, err := client.Do(req)
	if nil != err {
		return fmt.Errorf("request do err, param: %v, %w", req, err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); nil != closeErr {
			errStr := ""
			if err != nil {
				errStr = fmt.Sprintf("(%s)", err.Error())
			}
			err = fmt.Errorf("resp body close err, %v %w", errStr, closeErr)
		}
	}()

	var bodyByte []byte
	// 完整Response
	if r.response != nil {
		*r.response = *resp
		// 读取请求
		if bodyByte, err = ioutil.ReadAll(resp.Body); nil != err {
			return fmt.Errorf("request read err, bodyByte: %v, %w", bodyByte, err)
		}
		r.response.Body = ioutil.NopCloser(bytes.NewBuffer(bodyByte))
	}

	// 不需要解析返回值
	if result == nil {
		if _, err = io.Copy(ioutil.Discard, resp.Body); err != nil {
			return fmt.Errorf("resp body clear err, %w", err)
		}
		return
	}

	// 读取请求
	if len(bodyByte) == 0 {
		if bodyByte, err = ioutil.ReadAll(resp.Body); nil != err {
			return fmt.Errorf("request read err, bodyByte: %v, %w", bodyByte, err)
		}
	}

	// 没有内容
	if len(bodyByte) == 0 {
		return
	}

	// 按照JSON解析返回值
	if json.Valid(bodyByte) {
		if err = json.Unmarshal(bodyByte, &result); nil != err {
			return fmt.Errorf("request json un err, result: %v, %w", result, err)
		}
		return
	}

	// 按照字符串解析返回值
	rv := reflect.ValueOf(result)
	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("result must be a pointer")
	}

	rvv := rv.Elem()
	if rvv.Kind() != reflect.String {
		return fmt.Errorf("result must be a string")
	}
	if !rvv.CanSet() {
		return fmt.Errorf("result can not set")
	}
	rvv.SetString(string(bodyByte))

	return
}

func (r *entity) Get(ctx context.Context, url string, result interface{}, opts ...Option) error {
	// withMethod, WithBeforeRequest
	opts = append(opts, WithMethod(MethodGet), WithBeforeRequest(func(r *parameter) error {
		// 组装url
		params := neturl.Values{}
		netUrl, err := neturl.Parse(url)
		if err != nil {
			return fmt.Errorf("get json ma err, param: %s, %w", url, err)
		}
		for key, value := range r.param {
			params.Add(key, transform.Strval(value))
		}
		netUrl.RawQuery = params.Encode()
		r.url = netUrl.String()
		if r.log != nil {
			r.log.Println("Get url", r.header, r.url)
		}
		return nil
	}))

	return r.request(ctx, url, result, opts...)
}

func (r *entity) Post(ctx context.Context, url string, param map[string]interface{}, result interface{}, opts ...Option) error {
	// withMethod, withParam, WithBeforeRequest
	opts = append(opts, WithMethod(MethodPost), WithParam(param), WithBeforeRequest(func(r *parameter) error {
		if nil == r.param {
			return nil
		}
		// 组装param
		data, err := json.Marshal(r.param)
		if nil != err {
			return fmt.Errorf("post json ma err, param: %s, %w", param, err)
		}
		r.SetBody(bytes.NewBuffer(data))
		if r.log != nil {
			r.log.Println("Post url", r.header, r.url, string(data))
		}
		return nil
	}))

	return r.request(ctx, url, result, opts...)
}

func (r *entity) Put(ctx context.Context, url string, param map[string]interface{}, result interface{}, opts ...Option) error {
	// withMethod, withParam, WithBeforeRequest
	opts = append(opts, WithMethod(MethodPut), WithParam(param), WithBeforeRequest(func(r *parameter) error {
		if nil == r.param {
			return nil
		}
		// 组装param
		data, err := json.Marshal(r.param)
		if nil != err {
			return fmt.Errorf("put json ma err, param: %s, %w", param, err)
		}
		r.SetBody(bytes.NewBuffer(data))
		if r.log != nil {
			r.log.Println("Put url", r.header, r.url, string(data))
		}
		return nil
	}))

	return r.request(ctx, url, result, opts...)
}

func (r *entity) Delete(ctx context.Context, url string, param map[string]interface{}, result interface{}, opts ...Option) error {
	// withMethod, WithBeforeRequest
	opts = append(opts, WithMethod(MethodDelete), WithParam(param), WithBeforeRequest(func(r *parameter) error {
		if nil == r.param {
			return nil
		}
		// 组装param
		data, err := json.Marshal(r.param)
		if nil != err {
			return fmt.Errorf("post json ma err, param: %s, %w", param, err)
		}
		r.SetBody(bytes.NewBuffer(data))
		if r.log != nil {
			r.log.Println("Delete url", r.header, r.url, string(data))
		}
		// 组装url
		if r.deleteUriFlag {
			params := neturl.Values{}
			netUrl, err := neturl.Parse(url)
			if err != nil {
				return fmt.Errorf("get json ma err, param: %s, %w", url, err)
			}
			for key, value := range r.param {
				params.Add(key, transform.Strval(value))
			}
			netUrl.RawQuery = params.Encode()
			r.url = netUrl.String()
			if r.log != nil {
				r.log.Println("Delete url", r.url)
			}
		}
		return nil
	}))

	return r.request(ctx, url, result, opts...)
}

var defaultHttp = New()

func Get(ctx context.Context, url string, result interface{}, opts ...Option) error {
	return defaultHttp.Get(ctx, url, result, opts...)
}

func Post(ctx context.Context, url string, param map[string]interface{}, result interface{}, opts ...Option) error {
	return defaultHttp.Post(ctx, url, param, result, opts...)
}

func Put(ctx context.Context, url string, param map[string]interface{}, result interface{}, opts ...Option) error {
	return defaultHttp.Put(ctx, url, param, result, opts...)
}

func Delete(ctx context.Context, url string, param map[string]interface{}, result interface{}, opts ...Option) error {
	return defaultHttp.Delete(ctx, url, param, result, opts...)
}
