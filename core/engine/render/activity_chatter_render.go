package render

import (
	"context"
	"fmt"
	"html/template"
	"strings"
	"time"

	"sumeru/addons/mail"
	"sumeru/core/applog"
	"sumeru/core/engine/parser"
)

func writeActivityChatterPanel(ctx context.Context, sb *strings.Builder, c *parser.Chatter, vr *ViewRecordData, viewModel string) {
	_ = c
	if !mail.CompanyChatterEnabled(ctx) || vr == nil || vr.RecordID <= 0 {
		return
	}
	model := strings.TrimSpace(viewModel)
	if model == "" {
		model = strings.TrimSpace(vr.ResModel)
	}
	if model == "" {
		return
	}

	msgs, err := mail.ListCommentsForRecord(ctx, model, int64(vr.RecordID), 120)
	if err != nil {
		applog.WarnMsg(ctx, "render", "chatter", "activity chatter list failed", err,
			map[string]interface{}{"model": model, "record_id": vr.RecordID})
		msgs = nil
	}

	nextURL := "/web"
	if q := strings.TrimSpace(vr.FormBaseQuery); q != "" {
		nextURL = "/web?" + q
	}

	sb.WriteString(`<div class="sum-msg-shell">`)
	sb.WriteString(`<div class="sum-msg-thread">`)
	for _, m := range msgs {
		meta := strings.TrimSpace(m.Author)
		if meta == "" {
			meta = "User"
		}
		timeStr := ""
		if !m.CreateDate.IsZero() {
			timeStr = m.CreateDate.Local().Format("Jan 02, 15:04")
		}
		sb.WriteString(`<article class="sum-msg-card">`)
		sb.WriteString(`<header class="sum-msg-card-head">`)
		sb.WriteString(`<span class="sum-msg-author">` + template.HTMLEscapeString(meta) + `</span>`)
		if timeStr != "" {
			sb.WriteString(`<time class="sum-msg-time" datetime="` + template.HTMLEscapeString(m.CreateDate.Format(time.RFC3339)) + `">` + template.HTMLEscapeString(timeStr) + `</time>`)
		}
		sb.WriteString(`</header>`)
		sb.WriteString(`<div class="sum-msg-body">` + template.HTMLEscapeString(m.Body) + `</div>`)
		sb.WriteString(`</article>`)
	}
	if len(msgs) == 0 {
		sb.WriteString(`<div class="sum-msg-thread-empty" role="status">`)
		sb.WriteString(`<p class="sum-msg-thread-empty-title">No messages yet</p>`)
		sb.WriteString(`<p class="sum-msg-thread-empty-hint">Be the first to leave a comment on this record.</p>`)
		sb.WriteString(`</div>`)
	}
	sb.WriteString(`</div>`)
	sb.WriteString(`<footer class="sum-msg-composer">`)
	sb.WriteString(`<form method="post" action="/web/chatter/post" class="sum-msg-form">`)
	sb.WriteString(`<input type="hidden" name="model" value="` + template.HTMLEscapeString(model) + `" />`)
	sb.WriteString(`<input type="hidden" name="res_id" value="` + template.HTMLEscapeString(fmt.Sprintf("%d", vr.RecordID)) + `" />`)
	sb.WriteString(`<input type="hidden" name="next" value="` + template.HTMLEscapeString(nextURL) + `" />`)
	sb.WriteString(`<label class="sr-only" for="sum-chatter-body">Message</label>`)
	sb.WriteString(`<textarea id="sum-chatter-body" name="body" rows="3" class="sum-msg-input" placeholder="Write a message…"></textarea>`)
	sb.WriteString(`<div class="sum-msg-form-actions">`)
	sb.WriteString(`<p class="sum-msg-form-hint">Ctrl+Enter or ⌘+Enter to send</p>`)
	sb.WriteString(`<button type="submit" class="sum-msg-send">Send</button></div>`)
	sb.WriteString(`</form></footer></div>`)
}
