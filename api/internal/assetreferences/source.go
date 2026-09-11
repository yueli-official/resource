// Package assetreferences owns this consumer's authoritative business usage.
package assetreferences

import (
	"context"
	"database/sql"
	"github.com/yueli-official/asset/referencesync"
)

func Source(origins ...string) func(context.Context, *sql.Tx) ([]referencesync.Snapshot, error) {
	return referencesync.QuerySource([]referencesync.Query{
		{RefType: "resource-cover", Kind: "asset", SQL: `SELECT id::text,title,'/manage/'||id::text,cover_asset_id FROM resources`},
		{RefType: "resource-file", Kind: "asset", SQL: `SELECT r.id::text,r.title,'/manage/'||r.id::text,a.asset_id FROM resource_assets a JOIN resources r ON r.id=a.resource_id`},
		{RefType: "resource-content", Kind: "markdown", SQL: `SELECT id::text,title,'/manage/'||id::text,description FROM resources`},
		{RefType: "resource-seo", Kind: "markdown", SQL: `SELECT r.id::text,r.title,'/manage/'||r.id::text,s.og_image FROM resource_seo s JOIN resources r ON r.id=s.resource_id`},
	}, origins...)
}
