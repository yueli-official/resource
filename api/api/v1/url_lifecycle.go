package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/yueli-official/foundation/go/urllifecycle"
)

type ResolveURLReq struct {
	g.Meta `path:"/api/v1/url-lifecycle/resolve" method:"get" tags:"url-lifecycle" summary:"Resolve a public Resource URL"`
	Path   string `json:"path" v:"required"`
	Query  string `json:"query"`
}

type ResolveURLRes struct {
	Resolution urllifecycle.Resolution `json:"resolution"`
}
