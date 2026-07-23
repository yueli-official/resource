package dao

import (
	"context"
	"database/sql"
	"errors"

	"github.com/gogf/gf/v2/database/gdb"
)

var ErrSiteSettingsRevisionConflict = errors.New("resource site settings revision conflict")

type TransactionHook func(context.Context, *sql.Tx) error

func ComposeTransactionHooks(hooks ...TransactionHook) TransactionHook {
	return func(ctx context.Context, tx *sql.Tx) error {
		for _, hook := range hooks {
			if hook != nil {
				if err := hook(ctx, tx); err != nil {
					return err
				}
			}
		}
		return nil
	}
}

func runTransactionHook(ctx context.Context, tx gdb.TX, hook TransactionHook) error {
	if hook == nil {
		return nil
	}
	return hook(ctx, tx.GetSqlTX())
}
