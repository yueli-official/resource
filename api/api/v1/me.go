package v1

import "github.com/gogf/gf/v2/frame/g"

type GetResourceMeReq struct {
	g.Meta `path:"/api/v1/me" method:"get" tags:"resource" summary:"Get my Resource access"`
}

type GetResourceMeRes struct {
	IsAdministrator bool     `json:"isAdministrator"`
	Roles           []string `json:"roles"`
	Capabilities    []string `json:"capabilities"`
}
