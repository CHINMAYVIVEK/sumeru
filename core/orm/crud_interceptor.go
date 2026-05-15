package orm

import "context"

// SearchInterceptor allows addons to intercept and modify search domains.
type SearchInterceptor func(ctx context.Context, model string, domain [][]interface{}) ([][]interface{}, error)

var (
	SearchInterceptors []SearchInterceptor
)

// RegisterSearchInterceptor adds an interceptor to the global ORM search pipeline.
func RegisterSearchInterceptor(fn SearchInterceptor) {
	SearchInterceptors = append(SearchInterceptors, fn)
}
