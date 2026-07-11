package controller

import (
	"context"

	"platform/gokit/authjwt"
	v1 "platform/products/resource/api/api/v1"
	"platform/products/resource/api/internal/reserr"
)

// Ping is the authenticated probe — proof the authjwt chain is wired.
type Ping struct{}

func (Ping) Ping(ctx context.Context, _ *v1.PingReq) (*v1.PingRes, error) {
	p, ok := authjwt.From(ctx)
	if !ok {
		return nil, reserr.Forbidden()
	}
	return &v1.PingRes{Subject: p.Subject}, nil
}
