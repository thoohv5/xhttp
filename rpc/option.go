package rpc

import (
	"context"

	uh "github.com/thoohv5/xhttp/http"
)

type Option func(*rpc)

func WithOption(opts ...uh.Option) Option {
	return func(r *rpc) {
		r.opts = opts
	}
}

func WithTransForm(tf func(ctx context.Context, param interface{}) (map[string]interface{}, error)) Option {
	return func(r *rpc) {
		r.transform = tf
	}
}
