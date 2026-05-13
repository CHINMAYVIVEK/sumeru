package orm

import "sync"

var (
	securityMu       sync.RWMutex
	securityUID      int
	securityBypassACL bool
)

// SetSecurityBypass disables ACL checks (module data sync / CLI).
func SetSecurityBypass(v bool) {
	securityMu.Lock()
	securityBypassACL = v
	securityMu.Unlock()
}

func SecurityBypass() bool {
	securityMu.RLock()
	defer securityMu.RUnlock()
	return securityBypassACL
}

// SetSecurityUID sets the current request's res.users id (0 = anonymous).
func SetSecurityUID(uid int) {
	securityMu.Lock()
	securityUID = uid
	securityMu.Unlock()
}

// SecurityUID returns the current request user id.
func SecurityUID() int {
	securityMu.RLock()
	defer securityMu.RUnlock()
	return securityUID
}

// ClearSecurityUID resets the request user (call via defer after each HTTP request).
func ClearSecurityUID() {
	SetSecurityUID(0)
}
