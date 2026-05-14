package mail

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"sumeru/core/orm"
)

// Subtype values for mail.message rows.
const (
	SubtypeComment      = "comment"
	SubtypeNotification = "notification"
	SubtypeModule       = "module"
)

// Row is one persisted message (subset of columns used by UI and hooks).
type Row struct {
	Body        string
	Subtype     string
	Author      string
	CreateDate  time.Time
	Model       string
	CoreID     int64
}

// firstCompanyID returns the primary company row id, or 0 if none.
func firstCompanyID(ctx context.Context) int64 {
	if orm.DB == nil {
		return 0
	}
	tn := orm.GetTableName("core.company")
	var id sql.NullInt64
	if err := orm.DB.QueryRowContext(ctx, `SELECT id FROM `+tn+` ORDER BY id ASC LIMIT 1`).Scan(&id); err != nil || !id.Valid {
		return 0
	}
	return id.Int64
}

// CompanyChatterEnabled reads mail_chatter_enabled from the first core.company row (default true).
func CompanyChatterEnabled(ctx context.Context) bool {
	if orm.DB == nil {
		return true
	}
	tn := orm.GetTableName("core.company")
	var b sql.NullBool
	err := orm.DB.QueryRowContext(ctx, `SELECT mail_chatter_enabled FROM `+tn+` ORDER BY id ASC LIMIT 1`).Scan(&b)
	if err != nil || !b.Valid {
		return true
	}
	return b.Bool
}

// CompanyActivityPanelEnabled reads mail_activity_panel_enabled from the first core.company row (default true).
func CompanyActivityPanelEnabled(ctx context.Context) bool {
	if orm.DB == nil {
		return true
	}
	tn := orm.GetTableName("core.company")
	var b sql.NullBool
	err := orm.DB.QueryRowContext(ctx, `SELECT mail_activity_panel_enabled FROM `+tn+` ORDER BY id ASC LIMIT 1`).Scan(&b)
	if err != nil || !b.Valid {
		return true
	}
	return b.Bool
}

// PostMessage inserts a mail.message row. author may be empty (stored as "System").
func PostMessage(ctx context.Context, model string, coreID int64, body, subtype, author string) error {
	if orm.DB == nil {
		return fmt.Errorf("database not initialized")
	}
	model = strings.TrimSpace(model)
	body = strings.TrimSpace(body)
	subtype = strings.TrimSpace(subtype)
	if model == "" || body == "" || subtype == "" {
		return fmt.Errorf("model, body, and subtype are required")
	}
	if _, ok := orm.Registry[model]; !ok {
		return fmt.Errorf("unknown model %q", model)
	}
	author = strings.TrimSpace(author)
	if author == "" {
		author = "System"
	}
	vals := map[string]interface{}{
		"model":        model,
		"core_id":       int(coreID),
		"body":         body,
		"subtype":      subtype,
		"author":       author,
		"create_date":  time.Now().UTC(),
	}
	if cid := firstCompanyID(ctx); cid > 0 {
		vals["company_id"] = int(cid)
	}
	_, err := orm.Create(ctx, orm.MailMessage{}, vals)
	return err
}

// ListCommentsForRecord returns user chatter lines (subtype comment) for a record, oldest first.
func ListCommentsForRecord(ctx context.Context, model string, coreID int64, limit int) ([]Row, error) {
	if orm.DB == nil {
		return nil, nil
	}
	if limit <= 0 || limit > 500 {
		limit = 120
	}
	tn := orm.GetTableName("mail.message")
	q := `SELECT body, subtype, author, create_date, model, core_id FROM ` + tn +
		` WHERE model = $1 AND core_id = $2 AND subtype = $3 ORDER BY create_date ASC, id ASC LIMIT $4`
	rows, err := orm.DB.QueryContext(ctx, q, model, coreID, SubtypeComment, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMessageRows(rows)
}

// ListForRecord returns messages for a model/res_id pair, newest first.
func ListForRecord(ctx context.Context, model string, coreID int64, limit int) ([]Row, error) {
	if orm.DB == nil {
		return nil, nil
	}
	if limit <= 0 || limit > 500 {
		limit = 80
	}
	tn := orm.GetTableName("mail.message")
	q := `SELECT body, subtype, author, create_date, model, core_id FROM ` + tn +
		` WHERE model = $1 AND core_id = $2 ORDER BY create_date DESC, id DESC LIMIT $3`
	rows, err := orm.DB.QueryContext(ctx, q, model, coreID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMessageRows(rows)
}

// QueryActivityLog returns audit-oriented lines (module events, notifications, record saves).
// User chatter comments are excluded. Optional ctxModel/ctxID adds notifications on that record only.
func QueryActivityLog(ctx context.Context, limit int, ctxModel string, ctxID int64) ([]Row, error) {
	if orm.DB == nil {
		return nil, nil
	}
	if limit <= 0 || limit > 200 {
		limit = 40
	}
	tn := orm.GetTableName("mail.message")
	ctxModel = strings.TrimSpace(ctxModel)
	var rows *sql.Rows
	var err error
	if ctxModel != "" && ctxID > 0 {
		if _, ok := orm.Registry[ctxModel]; !ok {
			ctxModel, ctxID = "", 0
		}
	}
	if ctxModel != "" && ctxID > 0 {
		q := `SELECT body, subtype, author, create_date, model, core_id FROM ` + tn +
			` WHERE model = 'sys.module' OR subtype = 'module' OR (subtype = 'notification' AND model = $1 AND core_id = $2)` +
			` ORDER BY create_date DESC, id DESC LIMIT $3`
		rows, err = orm.DB.QueryContext(ctx, q, ctxModel, ctxID, limit)
	} else {
		q := `SELECT body, subtype, author, create_date, model, core_id FROM ` + tn +
			` WHERE (model = 'sys.module' OR subtype IN ('notification','module'))` +
			` ORDER BY create_date DESC, id DESC LIMIT $1`
		rows, err = orm.DB.QueryContext(ctx, q, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMessageRows(rows)
}

func scanMessageRows(rows *sql.Rows) ([]Row, error) {
	var out []Row
	for rows.Next() {
		var r Row
		var ts time.Time
		if err := rows.Scan(&r.Body, &r.Subtype, &r.Author, &ts, &r.Model, &r.CoreID); err != nil {
			return out, err
		}
		r.CreateDate = ts.UTC()
		out = append(out, r)
	}
	return out, rows.Err()
}

// LogModuleEvent posts a row on sys.module for the given technical module name.
func LogModuleEvent(ctx context.Context, moduleName, verb, detail string) {
	moduleName = strings.TrimSpace(moduleName)
	if moduleName == "" || orm.DB == nil {
		return
	}
	row, err := orm.SearchOne(ctx, "sys.module", map[string]interface{}{"name": moduleName})
	if err != nil {
		return
	}
	id, ok := orm.CoerceInt64(row["id"])
	if !ok || id <= 0 {
		return
	}
	verb = strings.TrimSpace(verb)
	if verb == "" {
		verb = "event"
	}
	detail = strings.TrimSpace(detail)
	body := fmt.Sprintf("%s: %s", verb, moduleName)
	if detail != "" {
		body = body + " — " + detail
	}
	_ = PostMessage(ctx, "sys.module", id, body, SubtypeModule, "System")
}
