package orm

import (
	"context"
	"fmt"
	"strings"
)

// ApplyRelatedFields fills virtual related fields on rec (in place).
func ApplyRelatedFields(ctx context.Context, model string, rec map[string]interface{}) error {
	if rec == nil {
		return nil
	}
	inst, ok := Registry[model]
	if !ok || inst == nil {
		return nil
	}
	for _, fd := range inst.Fields() {
		if fd.Related == "" || fd.RelatedStore {
			continue
		}
		val, err := resolveRelatedValue(ctx, model, rec, fd.Related)
		if err != nil {
			return fmt.Errorf("related %s.%s: %w", model, fd.Name, err)
		}
		rec[fd.Name] = val
	}
	return nil
}

func resolveRelatedValue(ctx context.Context, model string, rec map[string]interface{}, path string) (interface{}, error) {
	parts := strings.Split(path, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid related path %q", path)
	}
	relField := strings.TrimSpace(parts[0])
	targetField := strings.TrimSpace(parts[1])
	if relField == "" || targetField == "" {
		return nil, fmt.Errorf("invalid related path %q", path)
	}
	relID, ok := CoerceInt64(rec[relField])
	if !ok || relID <= 0 {
		return nil, nil
	}
	fd := FieldDef(model, relField)
	if fd == nil || fd.Relation == "" {
		return nil, fmt.Errorf("relation field %q not found on %s", relField, model)
	}
	target, err := SearchOne(ctx, fd.Relation, map[string]interface{}{"id": int(relID)})
	if err != nil {
		return nil, err
	}
	if target == nil {
		return nil, nil
	}
	return target[targetField], nil
}
