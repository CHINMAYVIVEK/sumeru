package module

import (
	"sumeru/core/engine/parser"
)

// recordValuesFromXML builds an upsert map from a data-file record, skipping named fields.
func recordValuesFromXML(xmlRecord parser.Record, skipFields ...string) map[string]interface{} {
	fieldMap := parser.RecordFieldMap(xmlRecord)
	skip := make(map[string]bool, len(skipFields))
	for _, name := range skipFields {
		skip[name] = true
	}
	values := make(map[string]interface{}, len(fieldMap))
	for key, val := range fieldMap {
		if skip[key] {
			continue
		}
		values[key] = val
	}
	return values
}
