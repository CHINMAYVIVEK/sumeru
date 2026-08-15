package web

import (
	"context"
	"strings"

	"sumeru/core/orm"
)

func actionListDomain(ctx context.Context, actionData map[string]interface{}) [][]interface{} {
	raw := strings.TrimSpace(orm.AsString(actionData["domain"]))
	if raw == "" || raw == "[]" {
		return nil
	}
	dom, err := orm.ParseDomainJSON(raw)
	if err != nil || len(dom) == 0 {
		return nil
	}
	dom = orm.ResolveDomainXMLRefs(ctx, dom)
	return orm.SubstituteDomainUID(dom, orm.SecurityUID(ctx))
}
