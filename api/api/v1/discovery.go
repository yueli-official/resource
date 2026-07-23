package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"platform/gokit/discoveryapi"
)

type GetDiscoveryArtifactReq struct {
	g.Meta `path:"/api/v1/discovery/artifact" method:"get" tags:"discovery" summary:"Get one artifact from the current atomic discovery publication"`
	Name   string `json:"name" v:"required"`
}

type GetDiscoveryArtifactRes struct {
	Result discoveryapi.ArtifactResult `json:"result"`
}
