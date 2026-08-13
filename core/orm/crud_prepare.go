package orm

import (
	"context"
	"fmt"
)

func prepareCreateWrite(ctx context.Context, model Model, values map[string]interface{}, opts PrepareOptions) (prepared map[string]interface{}, uid int, err error) {
	uid = SecurityUID(ctx)
	if err := CheckModelAccess(ctx, uid, model.ModelName(), "create"); err != nil {
		return nil, 0, err
	}
	prepared, err = PrepareValues(model, values, WriteOpCreate, opts)
	if err != nil {
		return nil, 0, err
	}
	if err := CheckFieldWriteAccess(ctx, uid, model.ModelName(), prepared); err != nil {
		return nil, 0, err
	}
	if err := CheckRecordRules(ctx, uid, model.ModelName(), "create", prepared); err != nil {
		return nil, 0, err
	}
	if len(prepared) == 0 {
		return nil, 0, fmt.Errorf("create requires at least one column")
	}
	return prepared, uid, nil
}

func requireRegisteredModel(modelName string) (Model, error) {
	inst, ok := Registry[modelName]
	if !ok || inst == nil {
		return nil, fmt.Errorf("model %s not found", modelName)
	}
	return inst, nil
}
