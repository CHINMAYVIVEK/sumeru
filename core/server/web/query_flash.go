package web

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"sumeru/core/engine/render"
)

var importFlashPattern = regexp.MustCompile(`^imported_(\d+)_updated_(\d+)_skipped_(\d+)$`)

// FlashFromQueryMessage converts ?msg= query values into workspace flash banners.
func FlashFromQueryMessage(msg string) (render.FlashMessage, bool) {
	return flashFromQueryMessage(msg)
}

func flashFromQueryMessage(msg string) (render.FlashMessage, bool) {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return render.FlashMessage{}, false
	}
	if m := importFlashPattern.FindStringSubmatch(msg); len(m) == 4 {
		created, _ := strconv.Atoi(m[1])
		updated, _ := strconv.Atoi(m[2])
		skipped, _ := strconv.Atoi(m[3])
		return render.FlashMessage{
			Kind:  "success",
			Title: "Import complete",
			Body:  fmt.Sprintf("Created %d, updated %d, skipped %d row(s).", created, updated, skipped),
		}, true
	}
	if strings.HasPrefix(msg, "imported_") {
		n := strings.TrimPrefix(msg, "imported_")
		return render.FlashMessage{
			Kind:  "success",
			Title: "Import complete",
			Body:  fmt.Sprintf("Imported %s row(s).", n),
		}, true
	}
	switch msg {
	case resetPasswordMsg:
		return render.FlashMessage{Kind: "info", Title: "Password reset", Body: "If the account exists, reset instructions were sent."}, true
	case "api_key_created":
		return render.FlashMessage{Kind: "success", Title: "API key created", Body: "Copy the key from the banner above if shown."}, true
	case moduleMsgSaved:
		return render.FlashMessage{Kind: "success", Title: "Saved", Body: "Changes were saved."}, true
	default:
		if strings.HasPrefix(msg, "installed_") || strings.HasPrefix(msg, "uninstalled_") || strings.HasPrefix(msg, "upgraded_") {
			return render.FlashMessage{Kind: "success", Title: "Apps", Body: strings.ReplaceAll(msg, "_", " ")}, true
		}
		return render.FlashMessage{Kind: "info", Title: "", Body: msg}, true
	}
}

func appendQueryFlashesToViewRecord(r *http.Request, viewRecord *render.ViewRecordData) {
	if viewRecord == nil || r == nil {
		return
	}
	msg := strings.TrimSpace(r.URL.Query().Get(flashMessageParam))
	if msg == "" {
		return
	}
	if flash, ok := flashFromQueryMessage(msg); ok {
		viewRecord.FlashMessages = append(viewRecord.FlashMessages, flash)
	}
}
