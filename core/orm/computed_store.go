package orm

import (
	"context"
	"fmt"
)

// MergeStoredComputes runs stored compute handlers and merges results into values for SQL write.
func MergeStoredComputes(ctx context.Context, modelName string, rec map[string]interface{}) error {
	if rec == nil {
		return nil
	}
	inst, ok := Registry[modelName]
	if !ok || inst == nil {
		return nil
	}
	stored := storedComputeFields(inst)
	if len(stored) == 0 {
		return nil
	}
	if err := ApplyComputes(ctx, modelName, rec); err != nil {
		return err
	}
	for _, name := range stored {
		if v, ok := rec[name]; ok {
			rec[name] = v
		}
	}
	return nil
}

func storedComputeFields(model Model) []string {
	var names []string
	for _, f := range model.Fields() {
		if f.Compute != "" && f.ComputeStore {
			names = append(names, f.Name)
		}
	}
	return names
}

// RejectVirtualWrites returns an error if values touch virtual or readonly-compute fields.
func RejectVirtualWrites(model Model, values map[string]interface{}) error {
	if model == nil || len(values) == 0 {
		return nil
	}
	byName := map[string]FieldDefinition{}
	for _, f := range model.Fields() {
		byName[f.Name] = f
	}
	for k := range values {
		fd, ok := byName[k]
		if !ok {
			continue
		}
		if IsVirtualField(fd) {
			return fmt.Errorf("field %q on %s is read-only", k, model.ModelName())
		}
		if fd.Compute != "" && fd.ComputeStore {
			return fmt.Errorf("field %q on %s is computed", k, model.ModelName())
		}
		if fd.Related != "" {
			return fmt.Errorf("field %q on %s is related", k, model.ModelName())
		}
	}
	return nil
}
