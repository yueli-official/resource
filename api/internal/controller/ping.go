package controller

import (
	"context"

	foundationauth "github.com/yueli-official/foundation/go/auth"
	v1 "github.com/yueli-official/resource/api/api/v1"
	"github.com/yueli-official/resource/api/internal/reserr"
)

// Ping is the authenticated probe — proof the Foundation auth chain is wired.
type Ping struct{}

func (Ping) Ping(ctx context.Context, _ *v1.PingReq) (*v1.PingRes, error) {
	p, ok := foundationauth.FromContext(ctx)
	if !ok {
		return nil, reserr.Forbidden()
	}
	return &v1.PingRes{Subject: p.Subject}, nil
}
