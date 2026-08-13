package orm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"sumeru/core/cache"
)

const xmlIDCacheTTL = 30 * time.Second

type xmlIDCacheHit struct {
	ID    int
	Model string
}

// ResolveXmlId returns the database ID for a given XML ID (module.name).
// The name segment may contain dots (e.g. base.action_core.company → module base, name action_core.company).
func ResolveXmlId(ctx context.Context, xmlID string) (int, string, error) {
	cacheKey := fmt.Sprintf("xmlid:%d:%t:%s", SecurityUID(ctx), SecurityBypass(ctx), xmlID)
	if v, ok := cache.Get(cacheKey); ok {
		if hit, ok := v.(xmlIDCacheHit); ok {
			return hit.ID, hit.Model, nil
		}
	}

	parts := strings.Split(xmlID, ".")
	module := ""
	name := xmlID
	if len(parts) >= 2 {
		module = parts[0]
		name = strings.Join(parts[1:], ".")
	}

	criteria := map[string]interface{}{"name": name}
	if module != "" {
		criteria["module"] = module
	}

	data, err := SearchOne(ctx, "sys.model.data", criteria)
	if err != nil {
		return 0, "", err
	}
	rid, ok := CoerceInt64(data["core_id"])
	if !ok {
		return 0, "", fmt.Errorf("invalid core_id in sys.model.data")
	}
	model := AsString(data["model"])
	cache.Set(cacheKey, xmlIDCacheHit{ID: int(rid), Model: model}, xmlIDCacheTTL)
	return int(rid), model, nil
}

// InvalidateXmlIDCache drops cached XML id resolutions (call with rule cache after security sync).
func InvalidateXmlIDCache() {
	cache.DeletePrefix("xmlid:")
}
