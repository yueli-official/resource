package resourceauthz

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/yueli-official/foundation/go/authorization"
)

func SyncResourceScopes(
	ctx context.Context,
	db *sql.DB,
	runtime authorization.ResourceScopeRegistry,
) error {
	rows, err := db.QueryContext(ctx, `SELECT id::text FROM resources ORDER BY id`)
	if err != nil {
		return fmt.Errorf("list resources for authorization scope sync: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return fmt.Errorf("scan resource authorization scope: %w", err)
		}
		if err := ensureScope(ctx, runtime, authorization.RegisterScopeCommand{
			ID: ResourceScopeID(id), Type: ScopeResource, ParentID: RootScopeID,
		}); err != nil {
			return err
		}
	}
	return rows.Err()
}

func ensureScope(
	ctx context.Context,
	runtime authorization.ResourceScopeRegistry,
	command authorization.RegisterScopeCommand,
) error {
	_, err := runtime.RegisterScope(ctx, command)
	if err == nil || authorization.Is(err, authorization.ErrorConflict) {
		return nil
	}
	return fmt.Errorf("register resource authorization scope %q: %w", command.ID, err)
}
