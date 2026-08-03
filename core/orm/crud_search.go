package orm

import (
	"context"
	"database/sql"
	"fmt"

	"sumeru/core/applog"
)

// SearchInterceptor allows addons to intercept and modify search domains.
type SearchInterceptor func(ctx context.Context, model string, domain [][]interface{}) ([][]interface{}, error)

var (
	SearchInterceptors []SearchInterceptor
)

// RegisterSearchInterceptor adds an interceptor to the global ORM search pipeline.
func RegisterSearchInterceptor(fn SearchInterceptor) {
	SearchInterceptors = append(SearchInterceptors, fn)
}

// Search finds records matching the criteria
func Search(ctx context.Context, modelName string, domain [][]interface{}) (results []map[string]interface{}, err error) {
	defer func() {
		n := 0
		if results != nil {
			n = len(results)
		}
		applog.ORMOp(ctx, "search", modelName, err, "rows", n)
	}()
	if _, ok := Registry[modelName]; !ok {
		return nil, fmt.Errorf("model %s not found", modelName)
	}
	uid := SecurityUID(ctx)
	if err := CheckModelAccess(ctx, uid, modelName, "read"); err != nil {
		return nil, err
	}

	// Call registered interceptors (e.g. sumeru_ai)
	for _, interceptor := range SearchInterceptors {
		var err error
		domain, err = interceptor(ctx, modelName, domain)
		if err != nil {
			return nil, err
		}
	}

	domain, err = MergeRuleDomainsIntoSearch(ctx, uid, modelName, "read", domain)
	if err != nil {
		return nil, err
	}

	var whereClause string
	var args []interface{}
	whereClause, args, err = buildSearchWhereClause(domain)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE %s", GetTableName(modelName), whereClause)
	var rows *sql.Rows
	rows, err = DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	results = []map[string]interface{}{}

	for rows.Next() {
		vals := make([]interface{}, len(cols))
		valPtrs := make([]interface{}, len(cols))
		for i := range vals {
			valPtrs[i] = &vals[i]
		}

		if err = rows.Scan(valPtrs...); err != nil {
			return nil, err
		}

		m := make(map[string]interface{})
		for i, col := range cols {
			m[col] = vals[i]
		}
		results = append(results, m)
	}

	return results, rows.Err()
}

// SearchLimit returns up to limit rows for modelName matching domain, ordered by id.
// limit must be positive; otherwise it defaults to 500.
func SearchLimit(ctx context.Context, modelName string, domain [][]interface{}, limit int) (results []map[string]interface{}, err error) {
	defer func() {
		n := 0
		if results != nil {
			n = len(results)
		}
		applog.ORMOp(ctx, "search_limit", modelName, err, "rows", n, "limit", limit)
	}()
	if _, ok := Registry[modelName]; !ok {
		return nil, fmt.Errorf("model %s not found", modelName)
	}
	if limit <= 0 {
		limit = 500
	}
	uid := SecurityUID(ctx)
	if err := CheckModelAccess(ctx, uid, modelName, "read"); err != nil {
		return nil, err
	}

	// Call registered interceptors
	for _, interceptor := range SearchInterceptors {
		var err error
		domain, err = interceptor(ctx, modelName, domain)
		if err != nil {
			return nil, err
		}
	}

	domain, err = MergeRuleDomainsIntoSearch(ctx, uid, modelName, "read", domain)
	if err != nil {
		return nil, err
	}

	var whereClause string
	var args []interface{}
	whereClause, args, err = buildSearchWhereClause(domain)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE %s ORDER BY id ASC LIMIT %d",
		GetTableName(modelName), whereClause, limit)
	var rows *sql.Rows
	rows, err = DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	results = nil

	for rows.Next() {
		vals := make([]interface{}, len(cols))
		valPtrs := make([]interface{}, len(cols))
		for i := range vals {
			valPtrs[i] = &vals[i]
		}
		if err = rows.Scan(valPtrs...); err != nil {
			return nil, err
		}
		m := make(map[string]interface{})
		for i, col := range cols {
			m[col] = vals[i]
		}
		results = append(results, m)
	}
	return results, rows.Err()
}
