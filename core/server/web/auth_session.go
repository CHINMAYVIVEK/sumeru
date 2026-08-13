package web

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"sumeru/core/orm"
	"sumeru/core/server/config"
)

const sessionCookieName = "sumeru_session"
const sessionDuration = 7 * 24 * time.Hour

func randomSessionID() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// CreateSession stores a new session and sets an HttpOnly cookie.
func CreateSession(w http.ResponseWriter, userID int) error {
	if orm.DB == nil {
		return fmt.Errorf("no database")
	}
	sid, err := randomSessionID()
	if err != nil {
		return err
	}
	exp := time.Now().UTC().Add(sessionDuration)
	tbl := orm.MustQuotedTableName("sys.session")
	if _, err := orm.DB.Exec(`INSERT INTO `+tbl+` (sid, user_id, expires_at) VALUES ($1, $2, $3)`, sid, userID, exp); err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sid,
		Path:     "/",
		MaxAge:   int(sessionDuration.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   !config.AppConfig.DevMode,
	})
	return nil
}

// ClearSessionCookie removes the session cookie (client-side).
func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   !config.AppConfig.DevMode,
	})
}

// SessionUserID returns core.user id from cookie session, or 0.
func SessionUserID(r *http.Request) int {
	if orm.DB == nil {
		return 0
	}
	c, err := r.Cookie(sessionCookieName)
	if err != nil || c.Value == "" {
		return 0
	}
	tbl := orm.MustQuotedTableName("sys.session")
	var uid int
	err = orm.DB.QueryRow(
		`SELECT user_id FROM `+tbl+` WHERE sid = $1 AND expires_at > NOW()`,
		c.Value,
	).Scan(&uid)
	if err != nil {
		return 0
	}
	return uid
}

// DestroySession removes the session row and clears the cookie.
func DestroySession(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie(sessionCookieName)
	if err == nil && c.Value != "" {
		tbl := orm.MustQuotedTableName("sys.session")
		_, _ = orm.DB.Exec(`DELETE FROM `+tbl+` WHERE sid = $1`, c.Value)
	}
	ClearSessionCookie(w)
}
