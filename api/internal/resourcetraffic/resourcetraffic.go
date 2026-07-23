// Package resourcetraffic declares the Resource consumer's traffic vocabulary
// and its one-time bridge from the legacy resources.view_count column.
package resourcetraffic

import (
	"context"
	"fmt"
	"strings"

	"github.com/yueli-official/foundation/go/traffic"

	"platform/products/resource/api/internal/dao"
)

const (
	ResourceItem   traffic.ResourceKind = "resource"
	baselineSource                      = "resource.resources.view_count"
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

type LegacySnapshot struct {
	InitialBaselines []traffic.BaselineImport
	Resources        []traffic.Resource
}

func SnapshotLegacy(ctx context.Context, store *dao.PG) (LegacySnapshot, error) {
	rows, err := store.ListViewProjections(ctx)
	if err != nil {
		return LegacySnapshot{}, fmt.Errorf("list resource traffic projections: %w", err)
	}
	snapshot := LegacySnapshot{
		InitialBaselines: make([]traffic.BaselineImport, 0, len(rows)),
		Resources:        make([]traffic.Resource, 0, len(rows)),
	}
	for _, row := range rows {
		resource := traffic.Resource{Kind: ResourceItem, ID: row.ResourceID}
		snapshot.Resources = append(snapshot.Resources, resource)
		snapshot.InitialBaselines = append(snapshot.InitialBaselines, traffic.BaselineImport{
			Source: baselineSource, Resource: resource, Views: row.Views,
		})
	}
	return snapshot, nil
}

func Reconcile(ctx context.Context, module traffic.Module, store *dao.PG, resources []traffic.Resource) error {
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
