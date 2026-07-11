package controller

import (
	"github.com/gogf/gf/v2/net/ghttp"

	"platform/gokit/response"
)

// Healthz is the public liveness endpoint (no auth), enveloped by ghttpx.
func Healthz(r *ghttp.Request) {
	r.Response.WriteJson(response.OK(map[string]any{"status": "up"}))
}
