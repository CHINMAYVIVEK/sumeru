package module

import (
	"sync"
)

var (
	// LoadedAddons contains addons that are both discovered and fully loaded (metadata/rows).
	LoadedAddons = map[string]*Addon{}
	// DiscoveredAddons contains all addons found on the filesystem during discovery.
	DiscoveredAddons = map[string]*Addon{}
	// installMu ensures serial execution of module install/uninstall/update operations.
	installMu sync.Mutex
)
