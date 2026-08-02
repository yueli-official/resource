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
	if !ok || p == nil {
		return nil, reserr.Forbidden()
	}
	kind, _ := p.Claim("subject_kind")
	actor := ""
	switch kind {
	case "user", "guest":
		actor = p.Subject
	case "client":
		actor = p.ClientID
	}
	if actor == "" {
		return nil, reserr.Forbidden()
	}
	return &v1.PingRes{Subject: actor}, nil
}
