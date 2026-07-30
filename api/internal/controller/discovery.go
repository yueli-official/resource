package controller

import (
	"context"

	v1 "github.com/yueli-official/resource/api/api/v1"
	"github.com/yueli-official/resource/api/internal/discoveryapi"
	"github.com/yueli-official/resource/api/internal/resourcediscovery"
)

type PublicDiscovery struct {
	manager *resourcediscovery.Manager
}

func NewPublicDiscovery(manager *resourcediscovery.Manager) *PublicDiscovery {
	return &PublicDiscovery{manager: manager}
}

func (controller *PublicDiscovery) GetDiscoveryArtifact(
	ctx context.Context,
	req *v1.GetDiscoveryArtifactReq,
) (*v1.GetDiscoveryArtifactRes, error) {
	snapshot, err := controller.manager.Snapshot(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetDiscoveryArtifactRes{
		Result: discoveryapi.FindArtifact(snapshot, req.Name),
	}, nil
}
