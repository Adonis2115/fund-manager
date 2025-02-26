// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: copyfrom.go

package repository

import (
	"context"
)

// iteratorForBulkCreateStocks implements pgx.CopyFromSource.
type iteratorForBulkCreateStocks struct {
	rows                 []BulkCreateStocksParams
	skippedFirstNextCall bool
}

func (r *iteratorForBulkCreateStocks) Next() bool {
	if len(r.rows) == 0 {
		return false
	}
	if !r.skippedFirstNextCall {
		r.skippedFirstNextCall = true
		return true
	}
	r.rows = r.rows[1:]
	return len(r.rows) > 0
}

func (r iteratorForBulkCreateStocks) Values() ([]interface{}, error) {
	return []interface{}{
		r.rows[0].ID,
		r.rows[0].Name,
		r.rows[0].Symbol,
		r.rows[0].Customsymbol,
		r.rows[0].Scripttype,
		r.rows[0].Industry,
		r.rows[0].Isin,
		r.rows[0].Fno,
	}, nil
}

func (r iteratorForBulkCreateStocks) Err() error {
	return nil
}

func (q *Queries) BulkCreateStocks(ctx context.Context, arg []BulkCreateStocksParams) (int64, error) {
	return q.db.CopyFrom(ctx, []string{"stocks"}, []string{"id", "name", "symbol", "customsymbol", "scripttype", "industry", "isin", "fno"}, &iteratorForBulkCreateStocks{rows: arg})
}
