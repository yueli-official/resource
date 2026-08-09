// Package bootstrap reconciles Resource-owned records after schema migration.
package bootstrap

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/yueli-official/foundation/go/identifier"
)

type executor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

const initialClassificationCatalog = `
INSERT INTO resource_classification_catalogs (id, catalog_key)
VALUES ($1::uuid, 'resource')
ON CONFLICT (catalog_key) DO NOTHING`

const initialClassificationPolicy = `
INSERT INTO resource_classification_policy_profiles (
	catalog_id, policy_key, schema_version, policy_revision,
	category_policy, facet_policies, tag_policy, discovery_policy
)
SELECT id, 'resource.item.default', 1, 1,
	'{"minAssignments":0,"maxAssignments":8,"requirePrimary":false,"leafOnly":false,"maxDepth":0}'::jsonb,
	'[]'::jsonb,
	'{"unknown":"reject","minAssignments":0,"maxAssignments":20}'::jsonb,
	'{"defaultSort":"name_asc"}'::jsonb
FROM resource_classification_catalogs
WHERE catalog_key = 'resource'
ON CONFLICT (catalog_id, policy_key) DO NOTHING`

func ReconcileClassification(ctx context.Context, target executor) error {
	catalogID, err := identifier.New()
	if err != nil {
		return fmt.Errorf("generate Resource classification catalog identifier: %w", err)
	}
	if _, err := target.ExecContext(ctx, initialClassificationCatalog, catalogID.String()); err != nil {
		return fmt.Errorf("install Resource classification catalog: %w", err)
	}
	if _, err := target.ExecContext(ctx, initialClassificationPolicy); err != nil {
		return fmt.Errorf("install Resource classification policy: %w", err)
	}
	return nil
}
