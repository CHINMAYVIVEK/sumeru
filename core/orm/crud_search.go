package orm

import (
	"context"
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

func execSearchQuery(ctx context.Context, modelName string, domain [][]interface{}, orderLimit string) ([]map[string]interface{}, error) {
	if _, ok := Registry[modelName]; !ok {
		return nil, fmt.Errorf("model %s not found", modelName)
	}
	uid := SecurityUID(ctx)
	if err := CheckModelAccess(ctx, uid, modelName, "read"); err != nil {
		return nil, err
	}

	for _, interceptor := range SearchInterceptors {
		var err error
		domain, err = interceptor(ctx, modelName, domain)
		if err != nil {
			return nil, err
		}
	}

	domain, err := MergeRuleDomainsIntoSearch(ctx, uid, modelName, "read", domain)
	if err != nil {
		return nil, err
	}

	whereClause, args, err := buildSearchWhereClause(domain)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE %s", GetTableName(modelName), whereClause)
	if orderLimit != "" {
		query += " " + orderLimit
	}
	rows, err := DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	var results []map[string]interface{}
	for rows.Next() {
		m, err := scanRowToMap(cols, rows)
		if err != nil {
			return nil, err
		}
		results = append(results, m)
	}
	return results, rows.Err()
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
	results, err = execSearchQuery(ctx, modelName, domain, "")
	if err != nil {
		return nil, err
	}
	if results == nil {
		results = []map[string]interface{}{}
	}
	return results, nil
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
	if limit <= 0 {
		limit = 500
	}
	orderLimit := fmt.Sprintf("ORDER BY id ASC LIMIT %d", limit)
	return execSearchQuery(ctx, modelName, domain, orderLimit)
}
