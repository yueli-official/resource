package controller

import (
	"context"

	"github.com/yueli-official/foundation/go/urllifecycle"

	v1 "github.com/yueli-official/resource/api/api/v1"
)

type PublicURLLifecycle struct {
	resolver urllifecycle.Resolver
}

func NewPublicURLLifecycle(resolver urllifecycle.Resolver) *PublicURLLifecycle {
	return &PublicURLLifecycle{resolver: resolver}
}

func (controller *PublicURLLifecycle) ResolveURL(
	ctx context.Context,
	req *v1.ResolveURLReq,
) (*v1.ResolveURLRes, error) {
	resolution, err := controller.resolver.Resolve(ctx, urllifecycle.Lookup{
		EscapedPath: req.Path,
		RawQuery:    req.Query,
	})
	if err != nil {
		return nil, err
	}
	return &v1.ResolveURLRes{Resolution: resolution}, nil
}
