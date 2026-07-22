package controller

import (
	"github.com/gogf/gf/v2/net/ghttp"
)

// Healthz is the public liveness endpoint (no auth) using the raw health DTO.
func Healthz(r *ghttp.Request) {
	r.Response.WriteJson(map[string]any{"status": "up"})
}
