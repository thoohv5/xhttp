package rpc

import (
	"context"

	uh "github.com/thoohv5/xhttp/http"
	"github.com/thoohv5/xhttp/util/mapstructure"
)

// IRpc Rpc标准
type IRpc interface {
	// Get Get
	Get(ctx context.Context, url string, param interface{}, data interface{}, op ...uh.Option) (err error)
	// Post Post
	Post(ctx context.Context, url string, param interface{}, data interface{}, op ...uh.Option) (err error)
	// Put Put
	Put(ctx context.Context, url string, param interface{}, data interface{}, op ...uh.Option) (err error)
	// Delete Delete
	Delete(ctx context.Context, url string, param interface{}, data interface{}, op ...uh.Option) (err error)
}

type rpc struct {
	opts []uh.Option
	// transform
	transform func(ctx context.Context, param interface{}) (map[string]interface{}, error)
}

// New 创建
func New(opts ...Option) IRpc {
	r := &rpc{
		transform: func(ctx context.Context, param interface{}) (map[string]interface{}, error) {
			toMap := make(map[string]interface{})
			decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
				Squash: true,
				Result: &toMap,
			})
			if err != nil {
				return toMap, err
			}
			if err = decoder.Decode(param); nil != err {
				return toMap, err
			}
			return toMap, nil
		},
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *rpc) pre(ctx context.Context, param interface{}, opts ...Option) (toMap map[string]interface{}, err error) {
	for _, opt := range opts {
		opt(r)
	}

	if param == nil {
		return make(map[string]interface{}), nil
	}

	// 检查
	if ret, ok := param.(map[string]interface{}); ok {
		toMap = ret
		return
	}

	if r.transform != nil {
		ret, err := r.transform(ctx, param)
		if err != nil {
			return nil, err
		}
		toMap = ret
	}

	return
}

func (r *rpc) Get(ctx context.Context, url string, param interface{}, data interface{}, op ...uh.Option) (err error) {
	toMap, err := r.pre(ctx, param, WithTransForm(func(ctx context.Context, param interface{}) (map[string]interface{}, error) {
		// struct to map
		toMap := make(map[string]interface{})
		decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			Squash:  true,
			TagName: "form",
			Result:  &toMap,
		})
		if err != nil {
			return nil, err
		}
		if err = decoder.Decode(param); nil != err {
			return nil, err
		}
		return toMap, nil
	}))
	if err != nil {
		return err
	}

	// 请求
	if err = uh.Get(ctx, url, data, append(append(r.opts, uh.WithParam(toMap)), op...)...); err != nil {
		return
	}

	return
}

func (r *rpc) Post(ctx context.Context, url string, param interface{}, data interface{}, op ...uh.Option) (err error) {
	toMap, err := r.pre(ctx, param)
	if err != nil {
		return
	}

	// 请求
	if err = uh.Post(ctx, url, toMap, data, append(append(r.opts, uh.WithParam(toMap)), op...)...); err != nil {
		return
	}

	return
}

func (r *rpc) Put(ctx context.Context, url string, param interface{}, data interface{}, op ...uh.Option) (err error) {
	toMap, err := r.pre(ctx, param)
	if err != nil {
		return
	}

	// 请求
	if err = uh.Put(ctx, url, toMap, data, append(append(r.opts, uh.WithParam(toMap)), op...)...); err != nil {
		return
	}
	return
}

func (r *rpc) Delete(ctx context.Context, url string, param interface{}, data interface{}, op ...uh.Option) (err error) {
	toMap, err := r.pre(ctx, param)
	if err != nil {
		return
	}

	// 请求
	if err = uh.Delete(ctx, url, toMap, data, append(append(r.opts, uh.WithParam(toMap)), op...)...); err != nil {
		return
	}

	return
}
