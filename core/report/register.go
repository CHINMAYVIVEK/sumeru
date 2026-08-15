package report

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"sumeru/core/orm"
)

func init() {
	orm.RegisterObjectAction(BulkModelName, "action_confirm_import", actionConfirmImport)
	orm.RegisterObjectAction(BulkModelName, "action_cancel_import", actionCancelImport)
}

func actionConfirmImport(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if model != BulkModelName || id <= 0 {
		return "", fmt.Errorf("invalid batch")
	}
	batch, err := orm.SearchOne(ctx, BulkModelName, map[string]interface{}{"id": id})
	if err != nil {
		return "", err
	}
	mapping, err := ParseMappingJSON(vals["column_mapping"])
	if err != nil {
		mapping, _ = ParseMappingJSON(orm.AsString(batch["column_mapping"]))
	}
	skipInvalid := vals["skip_invalid"] == "1" || vals["skip_invalid"] == "true"
	result, err := ExecuteBulkImport(ctx, ExecuteBulkImportInput{
		BatchID:     id,
		Mapping:     mapping,
		SkipInvalid: skipInvalid,
	})
	if err != nil {
		return "", err
	}
	next := orm.AsString(batch["next_url"])
	if next == "" {
		next = "/web/home"
	}
	sep := "?"
	if strings.Contains(next, "?") {
		sep = "&"
	}
	return next + sep + "msg=" + ImportFlashMessage(result), nil
}

func actionCancelImport(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	if err := CancelBatch(ctx, id); err != nil {
		return "", err
	}
	batch, _ := orm.SearchOne(ctx, BulkModelName, map[string]interface{}{"id": id})
	next := orm.AsString(batch["next_url"])
	if next == "" {
		next = "/web/home"
	}
	return next, nil
}

// BulkImportFormActionID resolves the window action id for bulk import mapping form.
func BulkImportFormActionID(ctx context.Context) int {
	id, _, err := orm.ResolveXmlId(ctx, "base.action_bulk_import")
	if err != nil {
		return 0
	}
	return id
}

// MappingFormURL builds workspace URL for a batch mapping form.
func MappingFormURL(batchID, actionID int) string {
	if actionID <= 0 {
		return fmt.Sprintf("/web?model=%s&id=%d&view_type=form", BulkModelName, batchID)
	}
	return fmt.Sprintf("/web?action=%d&id=%d&view_type=form", actionID, batchID)
}

// ParseFieldsParam splits comma-separated field list from query/form.
func ParseFieldsParam(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var out []string
	for _, p := range strings.Split(raw, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// ParseActionIDParam parses action id from request.
func ParseActionIDParam(raw string) int {
	id, _ := strconv.Atoi(strings.TrimSpace(raw))
	return id
}
