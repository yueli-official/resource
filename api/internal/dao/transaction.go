package dao

import (
	"context"
	"database/sql"
	"errors"

	"github.com/gogf/gf/v2/database/gdb"
)

var ErrSiteSettingsRevisionConflict = errors.New("resource site settings revision conflict")

type TransactionHook func(context.Context, *sql.Tx) error

func runTransactionHook(ctx context.Context, tx gdb.TX, hook TransactionHook) error {
	if hook == nil {
		return nil
	}
	return hook(ctx, tx.GetSqlTX())
}
