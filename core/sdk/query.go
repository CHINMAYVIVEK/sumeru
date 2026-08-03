package sdk

import (
	"context"
	"sumeru/core/orm"
)

// SearchOneInput finds one row by exact equality on all map keys.
type SearchOneInput struct {
	ModelName string
	Criteria  map[string]interface{}
}

// SearchOne delegates to the ORM.
func SearchOne(ctx context.Context, in SearchOneInput) (map[string]interface{}, error) {
	return orm.SearchOne(ctx, in.ModelName, in.Criteria)
}

// SearchInput runs a simple AND domain (see ORM Search for domain shape).
type SearchInput struct {
	ModelName string
	Domain    [][]interface{}
}

// Search delegates to the ORM.
func Search(ctx context.Context, in SearchInput) ([]map[string]interface{}, error) {
	return orm.Search(ctx, in.ModelName, in.Domain)
}

// CreateInput inserts one row for the given model.
type CreateInput struct {
	Model  Model
	Values map[string]interface{}
}

// Create delegates to the ORM.
func Create(ctx context.Context, in CreateInput) (int, error) {
	return orm.Create(ctx, in.Model, in.Values)
}

// UpsertInput upserts on a unique conflict column.
type UpsertInput struct {
	Model       Model
	Values      map[string]interface{}
	ConflictCol string
}

// Upsert delegates to the ORM.
func Upsert(ctx context.Context, in UpsertInput) (int, error) {
	return orm.Upsert(ctx, in.Model, in.Values, in.ConflictCol)
}

// ResolveXmlIdInput resolves module.xml_id to database id and model name.
type ResolveXmlIdInput struct {
	XMLID string
}

// ResolveXmlId delegates to the ORM.
func ResolveXmlId(ctx context.Context, in ResolveXmlIdInput) (int, string, error) {
	return orm.ResolveXmlId(ctx, in.XMLID)
}
