package controller

import (
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/yueli-official/resource/api/internal/reserr"
)

// CauseMappingMiddleware is the Resource application boundary: domain and
// provider adapters return typed causes, and this layer maps them once to the
// public catalog before Foundation performs HTTP projection.
func CauseMappingMiddleware(request *ghttp.Request) {
	request.Middleware.Next()
	if err := request.GetError(); err != nil {
		request.SetError(reserr.MapCause(err))
	}
}
