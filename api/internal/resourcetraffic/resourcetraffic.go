// Package resourcetraffic declares the Resource consumer's traffic vocabulary
// and maintains the catalog's read projection from Foundation Traffic truth.
package resourcetraffic

import (
	"context"
	"fmt"
	"strings"

	"github.com/yueli-official/foundation/go/traffic"

	"github.com/yueli-official/resource/api/internal/dao"
)

const (
	ResourceItem traffic.ResourceKind = "resource"
)

func Definition(timeZone string) traffic.Definition {
	return traffic.Definition{
		Version:  traffic.DefinitionVersion,
		TimeZone: strings.TrimSpace(timeZone),
		ResourceKinds: []traffic.ResourceKindDefinition{
			{Key: ResourceItem},
		},
	}
}

func ReconcileProjections(ctx context.Context, module traffic.Module, store *dao.PG) error {
	ids, err := store.ListResourceIDs(ctx)
	if err != nil {
		return fmt.Errorf("list resources for traffic projection: %w", err)
	}
	if len(ids) == 0 {
		return nil
	}
	resources := make([]traffic.Resource, 0, len(ids))
	for _, id := range ids {
		resources = append(resources, traffic.Resource{Kind: ResourceItem, ID: id})
	}
	totals, err := module.Totals(ctx, resources)
	if err != nil {
		return fmt.Errorf("read resource traffic totals: %w", err)
	}
	for _, item := range totals {
		if err := store.ReplaceViewProjection(ctx, item.Resource.ID, item.Totals.Views); err != nil {
			return fmt.Errorf("reconcile resource view projection for %s: %w", item.Resource.ID, err)
		}
	}
	return nil
}
